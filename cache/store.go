package cache

import (
	"context"
	"fmt"
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
	// read once: the transaction can outlive this call
	timeout := storeTimeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if !cs.txSupported {
		return fn(ctx)
	}

	type result struct {
		v   T
		err error
	}
	// buffered: the goroutine never blocks on a caller that is gone
	done := make(chan result, 1)
	go func() {
		v, err := runTx(ctx, cs.itemColl.Database().Client(), timeout, fn)
		done <- result{v, err}
	}()

	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		var zero T
		return zero, fmt.Errorf("cache transaction: %w", ctx.Err())
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

// applyObservation stores what the contracts said about the name at a block, in one transaction:
// it reads the record of the name and replaces it only if the observation is newer (see newer).
// returns the record that is in the cache afterwards (nil: none)
func (cs *cacheService) applyObservation(ctx context.Context, obs *NameDataItem) (stored *NameDataItem, err error) {
	// two first inserts of a name can race. with the unique index on the name the loser gets a
	// duplicate key error: run it again, it will see the winner's record then
	for i := 0; i < 3; i++ {
		stored, err = withTx(ctx, cs, func(ctx context.Context) (*NameDataItem, error) {
			return cs.applyObservationTx(ctx, obs, false)
		})
		if !mongo.IsDuplicateKeyError(err) {
			return stored, err
		}
	}
	return stored, err
}

// applyObservationTx is the body of the transaction: it can run more than once, it starts from
// scratch every time and does not change obs.
// dry: decide exactly the same, but write nothing (the dry run of the backfill)
func (cs *cacheService) applyObservationTx(ctx context.Context, obs *NameDataItem, dry bool) (*NameDataItem, error) {
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
		if item.ObservedBlock == stored.ObservedBlock && item.ObservedBlockHash == stored.ObservedBlockHash && item.ObservedAt > readAt(stored) {
			stored.ForkReadAt = item.ObservedAt
			if !dry {
				if _, err := cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, bson.M{"$set": bson.M{"fork_read_at": stored.ForkReadAt}}); err != nil {
					return nil, err
				}
			}
		}
		return stored, nil
	}

	if item.RefreshNeeded && stored != nil {
		mergeConfirmedOwner(&item, stored)
	}
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

// nameIndex: is there an index on exactly {name: 1}, and is it unique? ("" if there is none)
func (cs *cacheService) nameIndex(ctx context.Context) (name string, unique bool, err error) {
	specs, err := cs.itemColl.Indexes().ListSpecifications(ctx)
	if err != nil {
		return "", false, err
	}
	for _, spec := range specs {
		elems, err := spec.KeysDocument.Elements()
		if err != nil {
			return "", false, err
		}
		if len(elems) != 1 || elems[0].Key() != "name" {
			continue
		}
		if dir, ok := elems[0].Value().AsInt64OK(); !ok || dir != 1 {
			continue
		}
		return spec.Name, spec.Unique != nil && *spec.Unique, nil
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
