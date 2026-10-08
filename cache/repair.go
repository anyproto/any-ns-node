package cache

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// the background refresh of the cache. requests never wait for it: a lookup hands a record that
// needs a refresh (an incomplete one, or one whose refresh after an operation failed) to it
// through an in-memory queue (dropped when full: the periodic scan finds the record later), and
// every record that needs one is also due in Mongo (repair_at, indexed), where the periodic scan
// of every node finds it. a lease in Mongo keeps the nodes from refreshing the same name at once,
// a backoff keeps a failing name from taking every round.
//
// chain reads are driven by operations, never by lookups or time: the cache is the record of who
// holds a name (every registration goes through an operation of ours), an expired or lapsed name
// stays reserved for its owner and needs no read. the background reads only the re-reads after an
// operation (see rereadDelays) and the records that need a refresh

const (
	// after a failed refresh nobody (on any node) re-reads the name for this long; it doubles
	// with every failure in a row, up to maxFailureBackoff (see failureBackoff)
	refreshFailureBackoff = time.Minute
	maxFailureBackoff     = 24 * time.Hour
	// a record under a non-canonical spelling is not handed to the background again for this long
	// (see settleAlias)
	aliasSettleInterval = 10 * time.Minute

	// the default interval of the periodic scan, and how many names it re-reads per round
	defaultRepairInterval = time.Minute
	repairBatch           = 20

	// the in-memory queue of the names that lookups hand to the background refresh
	refreshQueueSize = 1024
)

// failureBackoff: the backoff after the failures-th failed refresh in a row (1: the first):
// refreshFailureBackoff * 2^(failures-1), at most maxFailureBackoff
func failureBackoff(failures int) time.Duration {
	d := refreshFailureBackoff
	for i := 1; i < failures && d < maxFailureBackoff; i++ {
		d *= 2
	}
	return min(d, maxFailureBackoff)
}

// refreshLease is how long a claimed refresh blocks the other nodes (it must outlive the
// refresh: the confirmation, the enrichment and the write are bounded by 10 seconds each; it
// only matters if the claiming node dies in the middle).
// a var only so that tests can shrink it
var refreshLease = time.Minute

// rereadDelays: after a completed operation the name is read again at these delays (by the
// periodic scan, see Rereads), whatever the first refresh found:
//   - 5 minutes: the contracts provider (load balanced backends) can lag behind the one that
//     returned the operation's receipt; a few minutes later it has the operation's block
//   - 30 minutes: longer than the Sepolia finality (~13 minutes), so the operation's block is
//     final by then: a reorg that dropped or changed the operation shows up in this read
var rereadDelays = []time.Duration{5 * time.Minute, 30 * time.Minute}

const (
	// re-reads closer than this to each other are one (repeated polls of an operation)
	rereadCoalesce = time.Minute
	// at most this many pending re-reads per record
	maxRereads = 8
)

// scheduleTimeout bounds the background write of the re-reads (see scheduleRereads).
// a var only so that tests can shrink it
var scheduleTimeout = 2 * time.Second

// maxScheduling: at most this many re-read schedules run at once (each bounded by
// scheduleTimeout); more are skipped (the in-memory request still runs the refresh)
const maxScheduling = 64

// scheduleRereadsAsync stores the re-reads of a name changed by an operation completed now on its
// record, in the background: the caller (a GetOperation poll) never waits for the write. tracked,
// Close waits for it
func (cs *cacheService) scheduleRereadsAsync(fullName string) {
	cs.asyncMu.Lock()
	defer cs.asyncMu.Unlock()
	if cs.closed || cs.itemColl == nil {
		return
	}
	select {
	case cs.scheduling <- struct{}{}:
	default:
		log.Warn("too many re-read schedules in flight, skipping one", zap.String("FullName", fullName))
		return
	}
	cs.async.Add(1)
	go func() {
		defer cs.async.Done()
		defer func() { <-cs.scheduling }()
		ctx, cancel := context.WithTimeout(context.Background(), scheduleTimeout)
		defer cancel()
		if err := cs.scheduleRereads(ctx, fullName); err != nil {
			log.Warn("failed to schedule the re-reads of a name", zap.String("FullName", fullName), zap.Error(err))
		}
	}()
}

// scheduleRereads adds the re-reads (see rereadDelays) to the record of the name, if there is
// one, in one transaction: merged with the pending ones (coalesced, at most maxRereads, see
// mergeRereads), so repeated polls of an operation can not grow it. the periodic scan runs them
// (never before the record's lease or backoff)
func (cs *cacheService) scheduleRereads(ctx context.Context, fullName string) error {
	exited, err := cs.storeRereads(ctx, fullName)
	// the caller holds a scheduling slot: it is given back only when the transaction really
	// ended (a commit can outlive the wait), so the cap holds for the transactions themselves
	<-exited
	return err
}

// storeRereads is the write of scheduleRereads; exited is closed when its transaction ended
func (cs *cacheService) storeRereads(ctx context.Context, fullName string) (<-chan struct{}, error) {
	name, err := cs.canonical(fullName)
	if err != nil {
		done := make(chan struct{})
		close(done)
		return done, err
	}
	times := cs.rereadTimes()
	_, exited, err := withTxExit(ctx, cs, func(ctx context.Context) (struct{}, error) {
		stored, err := cs.getNameData(ctx, name)
		if err != nil || stored == nil {
			return struct{}{}, err
		}
		rereads := mergeRereads(stored.Rereads, times, 0)
		if slices.Equal(rereads, stored.Rereads) {
			return struct{}{}, nil
		}
		stored.Rereads = rereads
		stored.RepairAt = repairAt(stored)
		set, unset := bson.M{"rereads": rereads}, bson.M{}
		setOrUnset(set, unset, "repair_at", stored.RepairAt, stored.RepairAt == 0)
		_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, updateDoc(set, unset))
		return struct{}{}, err
	})
	return exited, err
}

// rereadTimes: the re-reads of a name changed by an operation completed now
func (cs *cacheService) rereadTimes() []int64 {
	now := cs.now()
	out := make([]int64, 0, len(rereadDelays))
	for _, d := range rereadDelays {
		out = append(out, now.Add(d).UnixMilli())
	}
	return out
}

// mergeRereads: the pending re-reads of a and b, without the ones done by a read at doneAt
// (unix ms), sorted, coalesced, at most maxRereads (the earliest, and the latest)
func mergeRereads(a, b []int64, doneAt int64) []int64 {
	all := append(append(make([]int64, 0, len(a)+len(b)), a...), b...)
	slices.Sort(all)
	var out []int64
	for _, t := range all {
		if t <= doneAt || (len(out) > 0 && t-out[len(out)-1] < rereadCoalesce.Milliseconds()) {
			continue
		}
		out = append(out, t)
	}
	// capped: the earliest ones, and always the latest (the finality follow-up of the newest
	// change: dropping it could leave nothing after a reorg)
	if len(out) > maxRereads {
		out = append(out[:maxRereads-1], out[len(out)-1])
	}
	return out
}

// refreshRequest: a name for the background worker
type refreshRequest struct {
	name string
	// after a completed operation: the re-reads (see RefreshAfterOperation)
	afterOp bool
}

// refreshQueue is the in-memory queue of the background worker: non-blocking, deduplicated
type refreshQueue struct {
	ch     chan refreshRequest
	mu     sync.Mutex
	queued map[refreshRequest]bool
}

func newRefreshQueue(size int) *refreshQueue {
	return &refreshQueue{ch: make(chan refreshRequest, size), queued: map[refreshRequest]bool{}}
}

// push never blocks: false if the queue is full (the request is dropped)
func (q *refreshQueue) push(r refreshRequest) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.queued[r] {
		return true
	}
	select {
	case q.ch <- r:
		q.queued[r] = true
		return true
	default:
		return false
	}
}

func (q *refreshQueue) done(r refreshRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.queued, r)
}

// requestRefresh hands the name to the background worker. it never blocks and never writes:
// lookups call it
func (cs *cacheService) requestRefresh(r refreshRequest) {
	if cs.queue == nil {
		return
	}
	if !cs.queue.push(r) {
		log.Warn("the refresh queue is full, dropping the request (the periodic scan will find the record)",
			zap.String("FullName", r.name))
	}
}

// needsRefresh: a record that a lookup hands to the background refresh: an incomplete one, or
// one whose refresh after an operation failed (not while it is leased or backing off). an expired,
// lapsed or removed record is never handed to it: it needs no chain read
func needsRefresh(item *NameDataItem, now time.Time) bool {
	return item.RefreshNeeded && now.UnixMilli() >= item.RefreshNextAt
}

// repairAt: when the periodic scan takes the record (unix ms; 0: never, the field is unset):
//   - an incomplete record (or one marked after a failed refresh): now (after its backoff)
//   - a scheduled re-read (see rereadDelays)
//
// never before the record's lease or backoff (RefreshNextAt). an expiry is no reason to read the
// chain: a renewal is an operation of ours, it schedules its own re-reads
func repairAt(d *NameDataItem) int64 {
	var due int64
	switch {
	case d.RefreshNeeded:
		due = 1
	case len(d.Rereads) > 0:
		due = slices.Min(d.Rereads)
	default:
		return 0
	}
	return max(due, d.RefreshNextAt)
}

// claimRefresh atomically takes the refresh lease of the name (shared by all nodes through Mongo)
func (cs *cacheService) claimRefresh(ctx context.Context, fullName string, now time.Time) (bool, error) {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()

	filter := bson.M{
		"name": fullName,
		"$or": bson.A{
			bson.M{"refresh_next_at": bson.M{"$lte": now.UnixMilli()}},
			bson.M{"refresh_next_at": bson.M{"$exists": false}},
		},
	}
	res, err := cs.itemColl.UpdateOne(ctx, filter, holdUntil(now.Add(refreshLease).UnixMilli()))
	if err != nil {
		return false, err
	}
	return res.ModifiedCount > 0, nil
}

// backOff keeps every node from re-reading the name for its failure backoff (see
// failureBackoff). it does not touch the data or the observation
func (cs *cacheService) backOff(ctx context.Context, fullName string) {
	if err := cs.retryAfterBackoff(ctx, bson.M{"name": fullName}); err != nil {
		log.Warn("failed to store the refresh backoff", zap.String("FullName", fullName), zap.Error(err))
	}
}

// retryAfterBackoff counts a failed refresh of the matching record (refresh_failures): nobody
// refreshes it before now + its backoff (a longer lease or backoff stays), and the periodic scan
// takes it right then: repair_at is set to it, whatever it was, so a failed refresh is always
// retried. one update: the nodes share the count
func (cs *cacheService) retryAfterBackoff(ctx context.Context, filter bson.M) error {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()
	now := cs.now().UnixMilli()
	// refreshFailureBackoff * 2^(failures-1), at most maxFailureBackoff (the exponent is capped
	// first: 2^11 minutes is more than a day already)
	backoff := bson.M{"$min": bson.A{
		maxFailureBackoff.Milliseconds(),
		bson.M{"$multiply": bson.A{
			refreshFailureBackoff.Milliseconds(),
			bson.M{"$pow": bson.A{2, bson.M{"$min": bson.A{11, bson.M{"$subtract": bson.A{"$refresh_failures", 1}}}}}},
		}},
	}}
	_, err := cs.itemColl.UpdateOne(ctx, filter, mongo.Pipeline{
		{{Key: "$set", Value: bson.M{"refresh_failures": bson.M{"$add": bson.A{bson.M{"$ifNull": bson.A{"$refresh_failures", 0}}, 1}}}}},
		{{Key: "$set", Value: bson.M{"refresh_next_at": bson.M{"$max": bson.A{
			bson.M{"$ifNull": bson.A{"$refresh_next_at", int64(0)}},
			bson.M{"$toLong": bson.M{"$add": bson.A{now, backoff}}},
		}}}}},
		{{Key: "$set", Value: bson.M{"repair_at": "$refresh_next_at"}}},
	})
	return err
}

// holdUntil: nobody refreshes the record before until (a lease, a backoff). its scan time (if it
// has one, see repairAt) moves with it
func holdUntil(until int64) mongo.Pipeline {
	return mongo.Pipeline{{{Key: "$set", Value: bson.M{
		"refresh_next_at": until,
		"repair_at": bson.M{"$cond": bson.A{
			bson.M{"$eq": bson.A{bson.M{"$type": "$repair_at"}, "missing"}},
			"$$REMOVE",
			bson.M{"$max": bson.A{"$repair_at", until}},
		}},
	}}}}
}

// failed: the refresh did not settle the record (it stays as it was): try again after a backoff.
// "not registered" is settled: nothing is cached, or the cached record stays (errNotOnChain: the
// scheduled re-reads pick up a lagging provider, never a retry loop)
func failed(err error) bool {
	return err != nil && !errors.Is(err, ErrNameDataIncomplete) && !errors.Is(err, ErrNameNotRegistered)
}

// the repair index: the periodic scan reads it only (no collection scan, no in-memory sort).
// non-unique and idempotent, so it is created at startup (an index build on the cache, a small
// collection, does not block reads and writes for long); a failure only makes the scan slower
const repairIndexName = "repair_at"

func (cs *cacheService) ensureRepairIndex(ctx context.Context) {
	_, err := cs.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "repair_at", Value: 1}, {Key: "_id", Value: 1}},
		Options: options.Index().SetName(repairIndexName).SetSparse(true),
	})
	if err != nil {
		log.Warn("failed to create the repair index of the name cache", zap.Error(err))
	}
}

func (cs *cacheService) startWorker() {
	ctx, cancel := context.WithCancel(context.Background())
	cs.stopWorker = cancel
	cs.workerDone = make(chan struct{})
	go cs.worker(ctx)
}

// worker runs the background refresh: the names that lookups hand to it, and the periodic scan
// (every repairInterval, if it is on)
func (cs *cacheService) worker(ctx context.Context) {
	defer close(cs.workerDone)

	var tick <-chan time.Time
	if cs.repairInterval > 0 {
		ticker := time.NewTicker(cs.repairInterval)
		defer ticker.Stop()
		tick = ticker.C
	}
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-cs.queue.ch:
			cs.queue.done(r)
			cs.handle(ctx, r)
		case <-tick:
			if _, err := cs.repairOnce(ctx, repairBatch); err != nil && ctx.Err() == nil {
				log.Warn("cache repair failed", zap.Error(err))
			}

		}
	}
}

// handle runs a request of the queue
func (cs *cacheService) handle(ctx context.Context, r refreshRequest) {
	if r.afterOp {
		cs.refreshAfterOperation(ctx, r.name)
		return
	}
	cs.refreshLeased(ctx, r.name)
}

// refreshAfterOperation: see RefreshAfterOperation. it does not wait for the lease: the re-reads
// must be scheduled whatever another refresh does
func (cs *cacheService) refreshAfterOperation(ctx context.Context, fullName string) {
	if name, err := cs.canonical(fullName); err == nil {
		fullName = name
	}
	o := refreshOpts{rereads: cs.rereadTimes()}
	// the shared lease and the backoff apply here too (repeated polls during an incident must
	// not multiply the reads): a held record is left to the holder and the scan, its re-reads are
	// stored anyway (here, and by scheduleRereadsAsync). a name that is not cached has nothing to
	// lease
	claimed, err := cs.claimRefresh(ctx, fullName, cs.now())
	if err != nil {
		log.Warn("failed to take the refresh lease after an operation", zap.String("FullName", fullName), zap.Error(err))
		return
	}
	if !claimed {
		stored, err := cs.getNameData(ctx, fullName)
		if err != nil {
			log.Warn("failed to read a name after its operation", zap.String("FullName", fullName), zap.Error(err))
			return
		}
		if stored != nil {
			// held by another refresh (or just written, or backing off): the operation's re-reads
			// must be on the record anyway, also if the background schedule of the request was
			// skipped or timed out. coalesced with it (nothing new: no write). bounded (withTx),
			// and the worker does not wait for a late commit
			if _, err := cs.storeRereads(ctx, fullName); err != nil {
				log.Warn("failed to store the re-reads of a name after its operation", zap.String("FullName", fullName), zap.Error(err))
			}
			return
		}
	}
	_, err = cs.refresh(ctx, fullName, o)
	if !failed(err) {
		return
	}
	log.Warn("failed to refresh a name after its operation, marking it", zap.String("FullName", fullName), zap.Error(err))
	if err = cs.markRefreshNeeded(ctx, fullName, o.rereads); err != nil {
		log.Warn("failed to mark a name for a refresh", zap.String("FullName", fullName), zap.Error(err))
	}
}

// refreshLeased refreshes the name under its lease (skips it if another refresh holds it).
// a failure backs off. returns false if the name was skipped
func (cs *cacheService) refreshLeased(ctx context.Context, fullName string) bool {
	// a record under a non-canonical spelling: its canonical name is what is refreshed, the
	// record itself leaves the scan (until -dedupe-cache migrates it). the canonical record may
	// not exist (nothing to lease): the settled alias throttles this
	if canonical, err := cs.canonical(fullName); err == nil && canonical != fullName {
		cs.settleAlias(ctx, fullName)
		if c, err := cs.getNameData(ctx, canonical); err == nil && c == nil {
			if _, err = cs.refresh(ctx, canonical, refreshOpts{}); failed(err) {
				log.Warn("cache repair: failed to re-read a name", zap.String("FullName", canonical), zap.Error(err))
			}
			return true
		}
		fullName = canonical
	}
	claimed, err := cs.claimRefresh(ctx, fullName, cs.now())
	if err != nil {
		log.Warn("cache repair: failed to take the refresh lease", zap.String("FullName", fullName), zap.Error(err))
		return false
	}
	if !claimed {
		return false
	}
	if _, err = cs.refresh(ctx, fullName, refreshOpts{}); failed(err) {
		log.Warn("cache repair: failed to re-read a name", zap.String("FullName", fullName), zap.Error(err))
		cs.backOff(ctx, fullName)
	}
	return true
}

// the periodic scan: what is due, in which order. served by the repair index alone
func repairDue(now time.Time) bson.M {
	return bson.M{"repair_at": bson.M{"$lte": now.UnixMilli()}}
}

var repairOrder = bson.D{{Key: "repair_at", Value: 1}, {Key: "_id", Value: 1}}

// maxReschedules bounds the records one round of the scan reschedules without a read (see
// repairOnce): after an upgrade from 0.7.1 they come in thousands, the first rounds work them off
const maxReschedules = 500

// repairOnce re-reads up to limit names that are due (see repairAt), the longest waiting first:
// a record that keeps failing moves back with its backoff, it can not starve the others.
// a record whose repair_at was set under other rules (0.7.1 scheduled expiries, lapses and daily
// re-reads) is not read: its repair_at is recomputed from its fields and stored, without a
// contract call. such records do not count toward limit (they could fill every round and starve
// the records that are really due): the scan pages on (by repair_at and _id), at most
// maxReschedules of them per round. returns how many it re-read
func (cs *cacheService) repairOnce(ctx context.Context, limit int64) (int, error) {
	now := cs.now()
	filter := repairDue(now)
	done, reads, rescheduled := 0, int64(0), 0
	for reads < limit && rescheduled < maxReschedules && ctx.Err() == nil {
		docs, err := cs.repairPage(ctx, filter, limit)
		if err != nil {
			return done, err
		}
		for i := range docs {
			d := &docs[i]
			if ctx.Err() != nil || reads >= limit || rescheduled >= maxReschedules {
				return done, nil
			}
			if due := repairAt(d); due == 0 || due > now.UnixMilli() {
				cs.reschedule(ctx, d, due)
				rescheduled++
				continue
			}
			reads++
			if cs.refreshLeased(ctx, d.FullName) {
				done++
			}
		}
		if int64(len(docs)) < limit {
			break
		}
		// the next page: after the last record of this one, whatever happened to the records
		last := docs[len(docs)-1]
		filter = bson.M{"repair_at": bson.M{"$lte": now.UnixMilli()}, "$or": bson.A{
			bson.M{"repair_at": bson.M{"$gt": last.RepairAt}},
			bson.M{"repair_at": last.RepairAt, "_id": bson.M{"$gt": last.ID}},
		}}
	}
	return done, nil
}

// repairPage: one page of the scan, served by the repair index
func (cs *cacheService) repairPage(ctx context.Context, filter bson.M, limit int64) ([]NameDataItem, error) {
	findCtx, cancel := boundedCtx(ctx)
	defer cancel()
	cur, err := cs.itemColl.Find(findCtx, filter,
		options.Find().SetLimit(limit).SetSort(repairOrder).SetProjection(bson.M{
			"name": 1, "repair_at": 1, "refresh_needed": 1, "refresh_next_at": 1, "rereads": 1,
		}))
	if err != nil {
		return nil, err
	}
	var docs []NameDataItem
	err = cur.All(findCtx, &docs)
	return docs, err
}

// reschedule stores the repair_at of a record that the scan took by an outdated one (0: unset),
// unless the record changed in between (then its writer set it)
func (cs *cacheService) reschedule(ctx context.Context, d *NameDataItem, due int64) {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()
	set, unset := bson.M{}, bson.M{}
	setOrUnset(set, unset, "repair_at", due, due == 0)
	_, err := cs.itemColl.UpdateOne(ctx, bson.M{"_id": d.ID, "repair_at": d.RepairAt}, updateDoc(set, unset))
	if err != nil {
		log.Warn("cache repair: failed to reschedule a record", zap.String("FullName", d.FullName), zap.Error(err))
		return
	}
	log.Debug("cache repair: not due, rescheduled without a read", zap.String("FullName", d.FullName), zap.Int64("repair_at", due))
}
