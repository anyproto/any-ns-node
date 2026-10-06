package cache

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
)

// test hook, nil in production: runs in a lookup between the canonical miss and the alias check
var hookBeforeAliasCheck func()

// test hook, nil in production: runs in the write transaction of an observation, after the
// record of the name was read
var hookAfterRead func(obs *NameDataItem)

// a socket read or write of the cache's Mongo client never takes longer than this (unless the
// connection string says otherwise). it bounds what withTx leaves behind when it stops waiting
const mongoSocketTimeout = 30 * time.Second

// storeTimeout bounds one write of the cache: a transaction with its retries and its commit, or
// a single update (a lease, a backoff). a var only so that tests can shrink it
var storeTimeout = 10 * time.Second

// supportsTransactions: a replica set member or a mongos
func supportsTransactions(ctx context.Context, client *mongo.Client) (bool, error) {
	var res struct {
		SetName string `bson:"setName"`
		Msg     string `bson:"msg"`
	}
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&res)
	if err != nil {
		// before MongoDB 4.4.2
		err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "isMaster", Value: 1}}).Decode(&res)
	}
	if err != nil {
		return false, err
	}
	return res.SetName != "" || res.Msg == "isdbgrid", nil
}

// majority, but never waits for it longer than timeout (without wtimeout an unsatisfied
// majority blocks a write forever)
func majorityWrite(timeout time.Duration) *writeconcern.WriteConcern {
	return &writeconcern.WriteConcern{W: "majority", WTimeout: timeout}
}

// boundedCtx: a single Mongo write outside of a transaction (a lease, a backoff) never waits
// longer than storeTimeout
func boundedCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, storeTimeout)
}

// withTx runs fn in a transaction (on the primary, majority read and write concern) and
// returns within storeTimeout, whatever Mongo does.
// fn can run more than once: a transaction that conflicts with another one is retried.
// without a replica set (an explicitly allowed unsafe mode) fn runs as is.
//
// the driver commits and aborts with a context that has no deadline (and WithTransaction
// retries a commit for up to two minutes), so the transaction runs in its own goroutine and
// the caller stops waiting for it after storeTimeout. a commit that is late still lands, and
// that is fine: every write is ordered by its chain block, it can not overwrite a newer one
func withTx[T any](ctx context.Context, cs *cacheService, fn func(ctx context.Context) (T, error)) (T, error) {
	v, _, err := withTxExit(ctx, cs, fn)
	return v, err
}

// withTxExit is withTx that also returns a channel closed when the transaction (and its
// session) actually ended: withTx can stop waiting for it earlier
func withTxExit[T any](ctx context.Context, cs *cacheService, fn func(ctx context.Context) (T, error)) (T, <-chan struct{}, error) {
	exited := make(chan struct{})
	// read once: the transaction can outlive this call
	timeout := storeTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !cs.txSupported {
		defer close(exited)
		v, err := fn(ctx)
		return v, exited, err
	}

	type result struct {
		v   T
		err error
	}
	// buffered: the goroutine never blocks on a caller that is gone
	done := make(chan result, 1)
	cs.inflight.Add(1)
	cs.txActive.Add(1)
	go func() {
		defer cs.inflight.Done()
		defer close(exited)
		defer cs.txActive.Add(-1)
		v, err := runTx(ctx, cs.itemColl.Database().Client(), timeout, fn)
		done <- result{v, err}
	}()

	select {
	case r := <-done:
		return r.v, exited, r.err
	case <-ctx.Done():
		var zero T
		return zero, exited, fmt.Errorf("cache transaction: %w", ctx.Err())
	}
}

// runTx owns its session and its result: nothing it touches is shared with the caller of
// withTx, that can be gone when it returns
func runTx[T any](ctx context.Context, client *mongo.Client, timeout time.Duration, fn func(ctx context.Context) (T, error)) (T, error) {
	var out T

	// the client can be configured to read from secondaries: a transaction must not
	sess, err := client.StartSession(options.Session().SetDefaultReadPreference(readpref.Primary()))
	if err != nil {
		return out, err
	}
	defer func() {
		// ends (aborts) a transaction that is still open: bounded, ctx can be over already
		endCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		sess.EndSession(endCtx)
	}()

	opts := options.Transaction().
		SetReadPreference(readpref.Primary()).
		SetReadConcern(readconcern.Majority()).
		SetWriteConcern(majorityWrite(timeout)).
		SetMaxCommitTime(&timeout)
	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		// WithTransaction itself does not stop on ctx: it retries anything labeled transient
		// (an unreachable server is) for two minutes. a plain ctx error ends the retries
		if err := sc.Err(); err != nil {
			return nil, err
		}
		v, err := fn(sc)
		if err != nil {
			return nil, err
		}
		out = v
		return nil, nil
	}, opts)
	if err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

// newer: does a replace b? the one order of the observations of a name, for every write:
//  1. the higher block
//  2. at one height, two forks (block hashes): the one that was read last (ObservedAt, or
//     ForkReadAt: see applyObservationTx) is the one the chain moved to. ties: the larger hash
//  3. in one fork (the same chain state): a complete record over an incomplete one, then the
//     later read
//
// equal: false, the stored record stays
func newer(a, b *NameDataItem) bool {
	if a.ObservedBlock != b.ObservedBlock {
		return a.ObservedBlock > b.ObservedBlock
	}
	if a.ObservedBlockHash != b.ObservedBlockHash {
		if ra, rb := readAt(a), readAt(b); ra != rb {
			return ra > rb
		}
		return a.ObservedBlockHash > b.ObservedBlockHash
	}
	if a.RefreshNeeded != b.RefreshNeeded {
		return !a.RefreshNeeded
	}
	return a.ObservedAt > b.ObservedAt
}

// readAt: the latest read of the record's fork that the record knows of
func readAt(d *NameDataItem) int64 {
	return max(d.ObservedAt, d.ForkReadAt)
}

// mergeConfirmedOwner fills what an incomplete observation could not read from the stored
// record, only where the observation itself proves that it can not have changed: the NameWrapper
// returned the same wallet, only the wallet's owner (EOA) could not be read.
// nothing else is carried over: the owner, AnyID and space ID are read together (one failed
// read leaves all of them unknown, a successful one can be empty on purpose), and an unexpired
// name can still be transferred
func mergeConfirmedOwner(obs *NameDataItem, stored *NameDataItem) {
	if stored.Removed || obs.OwnerEthAddress != "" || obs.OwnerScwEthAddress == "" {
		return
	}
	if strings.EqualFold(obs.OwnerScwEthAddress, stored.OwnerScwEthAddress) {
		obs.OwnerEthAddress = stored.OwnerEthAddress
	}
}

func setOrUnset(set, unset bson.M, key string, value interface{}, empty bool) {
	if empty {
		unset[key] = ""
	} else {
		set[key] = value
	}
}

// markRefreshNeeded marks the record of the name (if there is one) for a refresh by the periodic
// scan after the backoff, and adds the re-reads to it. the data stays as it is: the name is still
// served as it was (a registration as taken)
func (cs *cacheService) markRefreshNeeded(ctx context.Context, fullName string, rereads []int64) error {
	_, err := withTx(ctx, cs, func(ctx context.Context) (struct{}, error) {
		stored, err := cs.getNameData(ctx, fullName)
		if err != nil || stored == nil {
			return struct{}{}, err
		}
		stored.RefreshNeeded = true
		stored.RefreshNextAt = max(stored.RefreshNextAt, cs.now().Add(refreshFailureBackoff).UnixMilli())
		stored.Rereads = mergeRereads(stored.Rereads, rereads, 0)
		stored.RepairAt = repairAt(stored)
		_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, bson.M{"$set": bson.M{
			"refresh_needed":  true,
			"refresh_next_at": stored.RefreshNextAt,
			"rereads":         stored.Rereads,
			"repair_at":       stored.RepairAt,
		}})
		return struct{}{}, err
	})
	return err
}

// applyObservation stores what the contracts said about the name at a block, in one transaction:
// it reads the record of the name and replaces it only if the observation is newer (see newer).
// returns the record that is in the cache afterwards (nil: none)
func (cs *cacheService) applyObservation(ctx context.Context, obs *NameDataItem, o refreshOpts) (stored *NameDataItem, err error) {
	// two first inserts of a name can race. with the unique index on the name the loser gets a
	// duplicate key error: run it again, it will see the winner's record then
	for i := 0; i < 3; i++ {
		stored, err = withTx(ctx, cs, func(ctx context.Context) (*NameDataItem, error) {
			return cs.applyObservationTx(ctx, obs, o)
		})
		if !mongo.IsDuplicateKeyError(err) {
			return stored, err
		}
	}
	return stored, err
}

// applyObservationTx is the body of the transaction: it can run more than once, it starts from
// scratch every time and does not change obs.
// o.dry: decide exactly the same, but write nothing (the dry run of the backfill).
// o.rereads are scheduled on whatever record stays; the ones due before the read are done
func (cs *cacheService) applyObservationTx(ctx context.Context, obs *NameDataItem, o refreshOpts) (*NameDataItem, error) {
	dry := o.dry
	item := *obs
	item.ID = primitive.NilObjectID
	item.OwnerScwEthAddress = strings.ToLower(item.OwnerScwEthAddress)
	item.OwnerEthAddress = strings.ToLower(item.OwnerEthAddress)

	stored, err := cs.getNameData(ctx, item.FullName)
	if err != nil {
		return nil, err
	}
	if hookAfterRead != nil {
		hookAfterRead(obs)
	}

	if stored != nil && !newer(&item, stored) {
		// the stored record stays. a read of its own fork that lost to it (e.g. an incomplete
		// one) still counts for which fork was read last
		log.Info("the cache has a newer record of the name, keeping it", zap.String("FullName", item.FullName),
			zap.Int64("cached block", stored.ObservedBlock), zap.Int64("read block", item.ObservedBlock))
		set, unset := bson.M{}, bson.M{}
		sameState := item.ObservedBlock == stored.ObservedBlock && item.ObservedBlockHash == stored.ObservedBlockHash
		if sameState && item.ObservedAt > readAt(stored) {
			stored.ForkReadAt = item.ObservedAt
			set["fork_read_at"] = stored.ForkReadAt
		}
		// the due re-reads are done only by a read of the stored state itself (the same block):
		// a stale read (an older block, a fork that lost, a removal the cache has a newer
		// registration against) decides nothing, they stay due (the lease, or the backoff of a
		// failed refresh, delays the next try)
		doneAt := int64(0)
		if sameState {
			doneAt = item.ObservedAt
		}
		if rereads := mergeRereads(stored.Rereads, o.rereads, doneAt); !slices.Equal(rereads, stored.Rereads) {
			stored.Rereads = rereads
			setOrUnset(set, unset, "rereads", rereads, len(rereads) == 0)
			stored.RepairAt = repairAt(stored)
			setOrUnset(set, unset, "repair_at", stored.RepairAt, stored.RepairAt == 0)
		}
		if !dry && len(set)+len(unset) > 0 {
			update := bson.M{}
			if len(set) > 0 {
				update["$set"] = set
			}
			if len(unset) > 0 {
				update["$unset"] = unset
			}
			if _, err := cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, update); err != nil {
				return nil, err
			}
		}
		return stored, nil
	}

	if item.RefreshNeeded && stored != nil {
		mergeConfirmedOwner(&item, stored)
	}
	// a replacement in the same fork keeps the latest read of that fork (the stored record's,
	// if it was read later: e.g. an incomplete read replaced by an earlier complete one)
	if stored != nil && item.ObservedBlock == stored.ObservedBlock && item.ObservedBlockHash == stored.ObservedBlockHash {
		if read := readAt(stored); read > item.ObservedAt {
			item.ForkReadAt = max(item.ForkReadAt, read)
		}
	}
	var before []int64
	if stored != nil {
		before = stored.Rereads
	}
	item.Rereads = mergeRereads(before, o.rereads, item.ObservedAt)
	item.RepairAt = repairAt(&item)
	switch {
	case dry:
	case stored == nil:
		res, err := cs.itemColl.InsertOne(ctx, item)
		if err != nil {
			return nil, err
		}
		item.ID = res.InsertedID.(primitive.ObjectID)
	default:
		item.ID = stored.ID
		if _, err := cs.itemColl.ReplaceOne(ctx, bson.M{"_id": item.ID}, item); err != nil {
			return nil, err
		}
	}
	return &item, nil
}

// nameIndex: is there an index on exactly {name: 1} (with the simple collation: the alias index
// name_ci has the same key), and is it unique? ("" if there is none)
func (cs *cacheService) nameIndex(ctx context.Context) (name string, unique bool, err error) {
	cur, err := cs.itemColl.Indexes().List(ctx)
	if err != nil {
		return "", false, err
	}
	var specs []struct {
		Name      string   `bson:"name"`
		Key       bson.Raw `bson:"key"`
		Unique    bool     `bson:"unique"`
		Collation bson.Raw `bson:"collation"`
	}
	if err = cur.All(ctx, &specs); err != nil {
		return "", false, err
	}
	for _, spec := range specs {
		if spec.Collation != nil {
			if locale, ok := spec.Collation.Lookup("locale").StringValueOK(); !ok || locale != "simple" {
				continue
			}
		}
		elems, err := spec.Key.Elements()
		if err != nil {
			return "", false, err
		}
		if len(elems) != 1 || elems[0].Key() != "name" {
			continue
		}
		if dir, ok := elems[0].Value().AsInt64OK(); !ok || dir != 1 {
			continue
		}
		return spec.Name, spec.Unique, nil
	}
	return "", false, nil
}

// ErrNameIndexNotUnique is returned when there is an index on exactly {name: 1}, but it is not
// unique. it is never dropped automatically (an operator decides)
var ErrNameIndexNotUnique = fmt.Errorf("the name cache has a non-unique index on {name: 1}: drop it by hand, " +
	"resolve the duplicates, then run -dedupe-cache -refresh-apply")

// ensureNameIndex: the cache keeps one record per name, the unique index on {name: 1} enforces
// it (prod has it, as name_1). any unique index on exactly {name: 1} is accepted whatever its
// name; a missing one is created (it fails if the cache has duplicates)
func (cs *cacheService) ensureNameIndex(ctx context.Context, create bool) (status string, err error) {
	idxName, unique, err := cs.nameIndex(ctx)
	switch {
	case err != nil:
		return "", err
	case idxName != "" && !unique:
		return "", fmt.Errorf("%w (index %q)", ErrNameIndexNotUnique, idxName)
	case idxName != "":
		return "exists", nil
	case !create:
		return "missing", nil
	}
	_, err = cs.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return "", fmt.Errorf("create the unique index on {name: 1} (duplicates of a name must be resolved by hand first): %w", err)
	}
	return "created", nil
}
