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
// needs a refresh to it through an in-memory queue (dropped when full: the periodic scan finds
// the record later), and every record that needs one is also due in Mongo (repair_at, indexed),
// where the periodic scan of every node finds it. a lease in Mongo keeps the nodes from
// refreshing the same name at once, a backoff keeps a failing name from taking every round.

const (
	// a lookup hands an expired record (or a tombstone) to the background refresh, but not more
	// often than this: the record can have been renewed (or registered again) without the cache
	// knowing
	expiredRefreshInterval = 10 * time.Minute
	// an expired record that the chain still has registered (the grace period) is re-read by the
	// periodic scan once per this: a renewal elsewhere does not touch the cache
	expiredRepairInterval = 24 * time.Hour
	// after a failed refresh nobody (on any node) re-reads the name for this long
	refreshFailureBackoff = time.Minute

	// the default interval of the periodic scan, and how many names it re-reads per round
	defaultRepairInterval = time.Minute
	repairBatch           = 20

	// the in-memory queue of the names that lookups hand to the background refresh
	refreshQueueSize = 1024
)

// refreshLease is how long a claimed refresh blocks the other nodes (it must outlive the
// refresh: the confirmation, the enrichment, the finalized read and the write are bounded by
// 10 seconds each; it only matters if the claiming node dies in the middle).
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
	name, err := cs.canonical(fullName)
	if err != nil {
		return err
	}
	times := cs.rereadTimes()
	_, err = withTx(ctx, cs, func(ctx context.Context) (struct{}, error) {
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
		update := bson.M{"$set": set}
		if len(unset) > 0 {
			update["$unset"] = unset
		}
		_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, update)
		return struct{}{}, err
	})
	return err
}

// dueBy: the periodic scan takes the record at t at the latest (an earlier due time stays), never
// before its lease or backoff
func dueBy(t int64) bson.D {
	return bson.D{{Key: "$set", Value: bson.M{
		"repair_at": bson.M{"$max": bson.A{
			bson.M{"$min": bson.A{bson.M{"$ifNull": bson.A{"$repair_at", t}}, t}},
			bson.M{"$ifNull": bson.A{"$refresh_next_at", 0}},
		}},
	}}}
}

// retryLater makes the periodic scan take the cached registration of the name after the backoff:
// a removal that is not final yet (the backfill). the data stays as it is
func (cs *cacheService) retryLater(ctx context.Context, fullName string) error {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()
	_, err := cs.itemColl.UpdateOne(ctx, bson.M{"name": fullName, "removed": bson.M{"$ne": true}},
		mongo.Pipeline{dueBy(cs.now().Add(refreshFailureBackoff).UnixMilli())})
	return err
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
// (unix ms), sorted, coalesced, at most maxRereads
func mergeRereads(a, b []int64, doneAt int64) []int64 {
	all := append(append(make([]int64, 0, len(a)+len(b)), a...), b...)
	slices.Sort(all)
	var out []int64
	for _, t := range all {
		if t <= doneAt || (len(out) > 0 && t-out[len(out)-1] < rereadCoalesce.Milliseconds()) {
			continue
		}
		if len(out) == maxRereads {
			break
		}
		out = append(out, t)
	}
	return out
}

// refreshRequest: a name for the background worker
type refreshRequest struct {
	name string
	// after a completed operation: no lease, and the re-reads (see RefreshAfterOperation)
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

// needsRefresh: a record that a lookup hands to the background refresh
func needsRefresh(item *NameDataItem, now time.Time) bool {
	// somebody is refreshing it right now, or the last refresh failed
	if now.UnixMilli() < item.RefreshNextAt {
		return false
	}
	readAgo := now.Sub(time.UnixMilli(item.ObservedAt))
	switch {
	case item.RefreshNeeded:
		return true
	case item.Removed:
		// the name could have been registered since
		return readAgo >= expiredRefreshInterval
	case now.Unix() < item.NameExpires:
		return false
	case isLapsed(item.NameExpires, now):
		// only the contracts can say that a name is free
		return true
	}
	// expired, in the grace period: was it renewed since we read it?
	return readAgo >= expiredRefreshInterval
}

// repairAt: when the periodic scan takes the record (unix ms; 0: never, the field is unset):
//   - an incomplete record (or one marked after a failed refresh): now (after its backoff)
//   - a scheduled re-read (see rereadDelays)
//   - a registration: when it expires (a renewal?); expired: when it lapses, and once per
//     expiredRepairInterval (a renewal elsewhere)
//
// never before the record's lease or backoff (RefreshNextAt)
func repairAt(d *NameDataItem) int64 {
	var due int64
	if d.RefreshNeeded {
		due = 1
	} else {
		if len(d.Rereads) > 0 {
			due = slices.Min(d.Rereads)
		}
		if !d.Removed && d.NameExpires > 0 {
			if e := expiryCheckAt(d); due == 0 || e < due {
				due = e
			}
		}
	}
	if due == 0 {
		return 0
	}
	return max(due, d.RefreshNextAt)
}

func expiryCheckAt(d *NameDataItem) int64 {
	expires := d.NameExpires * 1000
	if expires > d.ObservedAt {
		return expires
	}
	next := d.ObservedAt + expiredRepairInterval.Milliseconds()
	if lapses := (d.NameExpires + gracePeriodSec + 1) * 1000; lapses > d.ObservedAt && lapses < next {
		return lapses
	}
	return next
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

// backOff keeps every node from re-reading the name for refreshFailureBackoff.
// it does not touch the data or the observation
func (cs *cacheService) backOff(ctx context.Context, fullName string) {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()

	_, err := cs.itemColl.UpdateOne(ctx, bson.M{"name": fullName}, holdUntil(cs.now().Add(refreshFailureBackoff).UnixMilli()))
	if err != nil {
		log.Warn("failed to store the refresh backoff", zap.String("FullName", fullName), zap.Error(err))
	}
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

// failed: the refresh did not settle the record (it stays as it was): try again after a backoff
func failed(err error) bool {
	if err == nil || errors.Is(err, ErrNameDataIncomplete) {
		return false
	}
	return errors.Is(err, errNotFinal) || !errors.Is(err, ErrNameNotRegistered)
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
	o := refreshOpts{background: true, rereads: cs.rereadTimes()}
	_, err := cs.refresh(ctx, fullName, o)
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
	claimed, err := cs.claimRefresh(ctx, fullName, cs.now())
	if err != nil {
		log.Warn("cache repair: failed to take the refresh lease", zap.String("FullName", fullName), zap.Error(err))
		return false
	}
	if !claimed {
		return false
	}
	if _, err = cs.refresh(ctx, fullName, refreshOpts{background: true}); failed(err) {
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

// repairOnce re-reads up to limit names that are due (see repairAt), the longest waiting first:
// a record that keeps failing moves back with its backoff, it can not starve the others.
// returns how many it re-read
func (cs *cacheService) repairOnce(ctx context.Context, limit int64) (int, error) {
	findCtx, cancel := boundedCtx(ctx)
	defer cancel()
	cur, err := cs.itemColl.Find(findCtx, repairDue(cs.now()),
		options.Find().SetLimit(limit).SetSort(repairOrder).SetProjection(bson.M{"name": 1}))
	if err != nil {
		return 0, err
	}
	var docs []NameDataItem
	if err = cur.All(findCtx, &docs); err != nil {
		return 0, err
	}

	done := 0
	for _, d := range docs {
		if ctx.Err() != nil {
			break
		}
		if cs.refreshLeased(ctx, d.FullName) {
			done++
		}
	}
	return done, nil
}
