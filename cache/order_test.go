package cache

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anyproto/any-sync/app"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
)

func requireReplicaSet(t *testing.T, fx *fixture) {
	t.Helper()
	if !fx.txSupported {
		t.Skipf("SKIPPED %s: needs a replica set (transactions), the test Mongo is standalone: see testMongoURI", t.Name())
	}
}

// testService: a cache service on the collection, with its own contracts (not started: no
// background worker)
func testService(coll *mongo.Collection, m contracts.ContractsService, txSupported bool) *cacheService {
	return &cacheService{itemColl: coll, contracts: m, txSupported: txSupported, now: time.Now}
}

// a second cache service on the same collection, with its own contracts and chain head: the
// other ns node, or another request on this one
func otherCacheService(t *testing.T, fx *fixture, block int64, at time.Time) (*cacheService, *mock_contracts.MockContractsService) {
	m := mock_contracts.NewMockContractsService(gomock.NewController(t))
	m.EXPECT().LatestBlock(gomock.Any()).Return(testBlock(block, at), nil).AnyTimes()
	return testService(fx.itemColl, m, fx.txSupported), m
}

// a cache service on the same collection, through a Mongo client of its own
func serviceWithClient(t *testing.T, fx *fixture, opts *options.ClientOptions) (*cacheService, *mock_contracts.MockContractsService) {
	client, err := mongo.Connect(ctx, opts.ApplyURI(testMongoURI()))
	require.NoError(t, err)
	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = client.Disconnect(dctx)
	})

	m := mock_contracts.NewMockContractsService(gomock.NewController(t))
	m.EXPECT().LatestBlock(gomock.Any()).DoAndReturn(func(context.Context) (*contracts.Block, error) {
		return fx.latestBlock(), nil
	}).AnyTimes()
	coll := client.Database(testDbName).Collection("cache", options.Collection().SetWriteConcern(majorityWrite(storeTimeout)))
	return testService(coll, m, fx.txSupported), m
}

// how long a stalled command blocks: long enough to tell a bounded wait (storeTimeout) from
// an unbounded one
const stallBlock = 2500 * time.Millisecond

// failCommand sets the failCommand fail point of the test replica set (it needs
// --setParameter enableTestCommands=1, see testMongoURI)
func failCommand(t *testing.T, data bson.M, mode bson.M) {
	t.Helper()
	// a client of its own: the cleanups run after the fixture closed its client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(ctx) })
	admin := client.Database("admin")
	err = admin.RunCommand(ctx, bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: mode},
		{Key: "data", Value: data},
	}).Err()
	if err != nil && strings.Contains(err.Error(), "no such command") {
		t.Skipf("SKIPPED %s: needs the failCommand fail point, the test replica set does not allow it: "+
			"start it with --setParameter enableTestCommands=1 (see testMongoURI)", t.Name())
	}
	require.NoError(t, err)
	// a blocked command still runs when its block ends, and the transaction it belongs to
	// stays on the server (with its locks) until transactionLifetimeLimitSeconds. the next
	// test drops the database: let the server expire it quickly
	setLifetime := func(sec int) {
		_ = admin.RunCommand(ctx, bson.D{{Key: "setParameter", Value: 1}, {Key: "transactionLifetimeLimitSeconds", Value: sec}}).Err()
	}
	setLifetime(1)
	t.Cleanup(func() {
		_ = admin.RunCommand(ctx, bson.D{{Key: "configureFailPoint", Value: "failCommand"}, {Key: "mode", Value: "off"}}).Err()
		deadline := time.Now().Add(watchdog)
		for time.Now().Before(deadline) {
			n, err := admin.Aggregate(ctx, mongo.Pipeline{
				{{Key: "$currentOp", Value: bson.M{"allUsers": true, "idleSessions": true}}},
				{{Key: "$match", Value: bson.M{"transaction": bson.M{"$exists": true}}}},
			})
			var open []bson.M
			if err == nil {
				err = n.All(ctx, &open)
			}
			if err != nil || len(open) == 0 {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		setLifetime(60)
	})
}

// refreshName refreshes testFullName as the background (and the backfill) does
func (cs *cacheService) refreshName() error {
	_, err := cs.refresh(ctx, testFullName, refreshOpts{})
	return err
}

func obsAt(block int64, hash string, observedAt int64) *NameDataItem {
	return &NameDataItem{FullName: testFullName, OwnerEthAddress: testEoa, OwnerScwEthAddress: testScw, NameExpires: notExpired,
		ObservedBlock: block, ObservedBlockHash: hash, ObservedAt: observedAt}
}

// the order of the observations of a name
func TestNewer(t *testing.T) {
	a := obsAt(10, "0xa", 100)

	// the higher block wins, whatever the read times or completeness
	require.True(t, newer(obsAt(11, "0xa", 1), a))
	require.False(t, newer(obsAt(9, "0xa", 1000), a))
	incomplete := obsAt(11, "0xb", 1)
	incomplete.RefreshNeeded = true
	require.True(t, newer(incomplete, a))
	lapse := &NameDataItem{FullName: testFullName, Lapsed: true, ObservedBlock: 9, ObservedBlockHash: "0x9", ObservedAt: 1000}
	require.False(t, newer(lapse, a), "an older lapse never replaces a newer registration")
	require.False(t, newer(obsAt(8, "0x8", 2000), &NameDataItem{Lapsed: true, ObservedBlock: 9, ObservedBlockHash: "0x9"}),
		"an older registration never replaces a newer lapse")

	// one height, two forks: the one read last
	require.True(t, newer(obsAt(10, "0xb", 101), a))
	require.False(t, newer(obsAt(10, "0xb", 99), a))
	// ... counting the reads that lost to the stored record (ForkReadAt)
	aRead := *a
	aRead.ForkReadAt = 200
	require.False(t, newer(obsAt(10, "0xb", 150), &aRead))
	// ... ties: the larger hash
	require.True(t, newer(obsAt(10, "0xb", 100), a))
	require.False(t, newer(obsAt(10, "0x9", 100), a))

	// one chain state: a complete one over an incomplete one, then the later read
	inc := obsAt(10, "0xa", 500)
	inc.Incomplete = true
	require.False(t, newer(inc, a))
	require.True(t, newer(a, inc))
	require.True(t, newer(obsAt(10, "0xa", 101), a))
	require.False(t, newer(obsAt(10, "0xa", 100), a), "equal: the stored record stays")

	// a legacy record (no block) loses to any read
	require.True(t, newer(obsAt(1, "0x1", 1), &NameDataItem{FullName: testFullName}))
}

// the cache is ordered by the block the contracts were read at: an older chain state never
// replaces a newer one, whatever the local clocks or the order of the writes
func TestCacheService_ChainOrder(t *testing.T) {
	t.Run("skewed clock: an older-block read that finishes last loses to a newer-block registration", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 50, 0)

		// node A: its clock is an hour ahead, it reads block 100 (the old registration), slowly
		fx.now = func() time.Time { return time.Now().Add(time.Hour) }
		fx.setHead(100, time.Now())
		p := newPause(t)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(context.Context, [32]byte, common.Hash) (common.Address, error) {
			p.wait()
			return common.HexToAddress(nameWrapper), nil
		})
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(inGrace), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(testEoa), nil)

		done := make(chan error, 1)
		go func() {
			done <- fx.refreshName()
		}()
		p.Entered(t)

		// node B: a new registration is cached at block 101
		b, bContracts := otherCacheService(t, fx, 101, time.Now())
		expectRegistered(bContracts, otherScw, otherEoa, otherAnyID, notExpired)
		require.NoError(t, b.refreshName())

		p.Release()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(watchdog):
			t.Fatal("the refresh did not finish")
		}

		item := cachedItem(t, fx)
		require.Equal(t, int64(101), item.ObservedBlock)
		require.Equal(t, otherEoa, item.OwnerEthAddress)
		require.Equal(t, notExpired, item.NameExpires)
	})

	t.Run("lagging provider: a later read of an older block does not overwrite a renewal", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.setHead(200, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.refreshName())

		lagging, laggingContracts := otherCacheService(t, fx, 150, time.Now())
		expectRegistered(laggingContracts, testScw, testEoa, testAnyID, inGrace)
		require.NoError(t, lagging.refreshName())

		require.Equal(t, notExpired, cachedItem(t, fx).NameExpires)
	})

	t.Run("lagging provider: a lapse read at an older block does not overwrite a newer registration", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.setHead(200, time.Now())
		expectRegistered(fx.contracts, otherScw, otherEoa, otherAnyID, notExpired)
		require.NoError(t, fx.refreshName())

		// a lagging provider: block 150, before the registration (a re-registration after a lapse)
		lagging, laggingContracts := otherCacheService(t, fx, 150, time.Now())
		expectLapsed(laggingContracts)
		require.NoError(t, lagging.refreshName())

		item := cachedItem(t, fx)
		require.False(t, item.Lapsed)
		require.Equal(t, int64(200), item.ObservedBlock)
		require.Equal(t, notExpired, item.NameExpires)
		require.Equal(t, otherEoa, isNameAvailable(t, fx.cacheService).OwnerEthAddress)
	})

	t.Run("a lapse is not overwritten by an older observation, the owner stays", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)

		// the lapse read at block 300
		fx.setHead(300, time.Now())
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())

		// a backfill that read block 250 (still in the grace period then) finishes after it
		backfill, backfillContracts := otherCacheService(t, fx, 250, time.Now().Add(time.Hour))
		expectRegistered(backfillContracts, testScw, testEoa, testAnyID, inGrace)
		require.NoError(t, backfill.refreshName())

		item := cachedItem(t, fx)
		require.True(t, item.Lapsed)
		require.Equal(t, lapsed, item.NameExpires)
		require.Equal(t, int64(300), item.ObservedBlock)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)

		// the owner still resolves: the name is reserved for it
		require.True(t, byAnyID(t, fx, testAnyID).Found)
		res, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: testScw})
		require.NoError(t, err)
		require.True(t, res.Found)
	})

	t.Run("a registration after a lapse (or a 0.7.1 tombstone) replaces it", func(t *testing.T) {
		for _, tombstone := range []bool{false, true} {
			fx := newFixture(t)
			if tombstone {
				insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(300), "observed_block_hash": blockHash(300).Hex()})
			} else {
				seedItem(t, fx, lapsed, 10, 0)
				fx.setHead(300, time.Now())
				expectLapsed(fx.contracts)
				require.NoError(t, fx.refreshName())
			}

			fx.setHead(301, time.Now())
			expectRegistered(fx.contracts, otherScw, otherEoa, otherAnyID, notExpired)
			require.NoError(t, fx.refreshName())

			item := cachedItem(t, fx)
			require.False(t, item.Lapsed)
			require.False(t, item.Removed)
			require.Equal(t, otherEoa, isNameAvailable(t, fx.cacheService).OwnerEthAddress)
			require.EqualValues(t, 1, countRecords(t, fx))
			fx.finish(t)
		}
	})

	t.Run("the same block: an incomplete read does not replace a complete record", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.setHead(400, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.refreshName())

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))
		// the cache has the complete record of that block: that is what the caller gets
		require.NoError(t, fx.refreshName())

		item := cachedItem(t, fx)
		require.False(t, item.RefreshNeeded)
		require.Equal(t, testEoa, item.OwnerEthAddress)
		// the lost read still counts for its fork
		require.Greater(t, item.ForkReadAt, item.ObservedAt)
	})

	t.Run("an incomplete read keeps the wallet's owner, but only for the same wallet", func(t *testing.T) {
		for _, c := range []struct {
			scw   string
			owner string
		}{{testScw, testEoa}, {otherScw, ""}} {
			fx := newFixture(t)
			seedItem(t, fx, inGrace, 10, 0)

			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
			fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(c.scw, testAnyID, "", nil)
			fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
			require.ErrorIs(t, fx.refreshName(), ErrNameDataIncomplete)

			item := cachedItem(t, fx)
			require.Equal(t, c.owner, item.OwnerEthAddress)
			require.Equal(t, c.scw, item.OwnerScwEthAddress)
			require.True(t, item.RefreshNeeded)
			fx.finish(t)
		}
	})

	t.Run("a replacement in the same fork keeps the fork's latest read: a stale other fork can not win", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// fork A at block 600, incomplete, read at 300
		incomplete := obsAt(600, "0xa", 300)
		incomplete.Incomplete, incomplete.RefreshNeeded, incomplete.OwnerEthAddress = true, true, ""
		_, err := fx.applyObservation(ctx, incomplete, refreshOpts{})
		require.NoError(t, err)
		// a delayed complete read of fork A from 100 replaces it (complete wins in one fork)
		complete := obsAt(600, "0xa", 100)
		_, err = fx.applyObservation(ctx, complete, refreshOpts{})
		require.NoError(t, err)
		item := cachedItem(t, fx)
		require.False(t, item.Incomplete)
		require.Equal(t, int64(300), item.ForkReadAt)

		// a delayed read of fork B from 200: fork A was read later, it stays
		b := obsAt(600, "0xb", 200)
		b.OwnerEthAddress = otherEoa
		_, err = fx.applyObservation(ctx, b, refreshOpts{})
		require.NoError(t, err)
		item = cachedItem(t, fx)
		require.Equal(t, "0xa", item.ObservedBlockHash)
		require.Equal(t, testEoa, item.OwnerEthAddress)
	})

	t.Run("a legacy record (raw BSON, no observation fields) is replaced by any read", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		_, err := fx.itemColl.InsertOne(ctx, bson.M{
			"name": testFullName, "owner_eth_address": testEoa, "owner_scw_eth_address": testScw,
			"owner_any_address": testAnyID, "space_id": "", "name_expires": inGrace,
		})
		require.NoError(t, err)

		fx.setHead(1, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.refreshName())

		require.Equal(t, notExpired, cachedItem(t, fx).NameExpires)
		require.EqualValues(t, 1, countRecords(t, fx))
	})
}

// a 0.7.1 tombstone never resolves in a reverse lookup, whatever fields its document has
func TestCacheService_ReverseLookupSkipsTombstones(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)
	_, err := fx.itemColl.InsertOne(ctx, bson.M{"name": testFullName, "removed": true, "observed_block": 10,
		"owner_scw_eth_address": testScw, "owner_any_address": testAnyID})
	require.NoError(t, err)

	res, err := fx.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{AnyAddress: testAnyID})
	require.NoError(t, err)
	require.False(t, res.Found)
	res, err = fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: testScw})
	require.NoError(t, err)
	require.False(t, res.Found)
}

// a lapse (expired and past the grace period) keeps the record: a name stays reserved to the
// identity that registered it. the lapse is stored in place, the owner fields are never wiped
func TestCacheService_Lapse(t *testing.T) {
	t.Run("the owner, AnyID, space and canon stay; lapsed, the expiry and the block are stored; no enrichment read", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)
		// incomplete and marked, with a backoff count: a lapse has nothing more to read
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{
			"canon": testFullName, "incomplete": true, "refresh_needed": true, "refresh_failures": 3, "repair_at": int64(1)}})
		require.NoError(t, err)

		fx.setHead(300, time.Now())
		// (only these two reads: an enrichment read fails the test)
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())

		item := cachedItem(t, fx)
		require.True(t, item.Lapsed)
		require.Equal(t, lapsed, item.NameExpires)
		require.Equal(t, int64(300), item.ObservedBlock)
		require.Equal(t, blockHash(300).Hex(), item.ObservedBlockHash)
		require.Equal(t, strings.ToLower(nameWrapper), item.RegistryOwner)
		require.Equal(t, testEoa, item.OwnerEthAddress)
		require.Equal(t, testScw, item.OwnerScwEthAddress)
		require.Equal(t, testAnyID, item.OwnerAnyAddress)
		require.Equal(t, "space", item.SpaceId)
		require.Equal(t, testFullName, item.Canon)
		require.False(t, item.Incomplete)
		require.False(t, item.RefreshNeeded)
		require.Zero(t, item.RefreshFailures)
		require.Zero(t, item.RepairAt, "nothing to read: not scanned")
		require.False(t, item.Removed)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)
		require.Zero(t, fx.runQueued())

		// idempotent: read again at a later block, the same record
		before := rawItem(t, fx, testFullName)
		fx.setHead(301, time.Now())
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())
		after := rawItem(t, fx, testFullName)
		require.Equal(t, int64(301), after["observed_block"])
		for _, k := range []string{"owner_eth_address", "owner_scw_eth_address", "owner_any_address", "space_id", "name_expires", "lapsed", "canon", "_id"} {
			require.Equal(t, before[k], after[k], k)
		}
		require.EqualValues(t, 1, countRecords(t, fx))
	})

	t.Run("an older lapse observation does not overwrite a newer record", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		fx.setHead(400, time.Now())
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())

		item := cachedItem(t, fx)
		require.False(t, item.Lapsed)
		require.Equal(t, notExpired, item.NameExpires)
		require.Equal(t, int64(500), item.ObservedBlock)
	})

	t.Run("a 0.7.1 tombstone: the lapse is stored, removed stays (it has no owner to keep)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(10), "observed_block_hash": blockHash(10).Hex()})

		fx.setHead(300, time.Now())
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())
		item := cachedItem(t, fx)
		require.True(t, item.Lapsed)
		require.True(t, item.Removed)
		require.Equal(t, lapsed, item.NameExpires)
		out := isNameAvailable(t, fx.cacheService)
		require.False(t, out.Available)
		require.Equal(t, lapsed, out.NameExpires)
	})

	t.Run("lapsed and not cached: nothing is written, not registered", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		expectLapsed(fx.contracts)
		require.ErrorIs(t, fx.refreshName(), ErrNameNotRegistered)
		require.Zero(t, countRecords(t, fx))
	})

	t.Run("a dry run decides the same, writes nothing", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)
		before := rawItem(t, fx, testFullName)

		expectLapsed(fx.contracts)
		stored, err := fx.refresh(ctx, testFullName, refreshOpts{dry: true})
		require.NoError(t, err)
		require.True(t, stored.Lapsed)
		require.Equal(t, testEoa, stored.OwnerEthAddress)
		require.Equal(t, before, rawItem(t, fx, testFullName))
	})
}

// the chain says that a name was never registered: a cached record is never removed (a lagging
// provider right after a registration); nothing cached stays nothing
func TestCacheService_NotOnChain(t *testing.T) {
	noOwner := func(m *mock_contracts.MockContractsService) {
		m.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil)
		m.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(0), nil)
	}

	t.Run("cached: kept as it is, a settled refresh (no backoff loop), served as taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, 0)
		before := rawItem(t, fx, testFullName)

		fx.setHead(300, time.Now())
		noOwner(fx.contracts)
		err := fx.refreshName()
		require.ErrorIs(t, err, ErrNameNotRegistered)
		require.ErrorIs(t, err, errNotOnChain)
		require.False(t, failed(err), "settled: no backoff")
		require.Equal(t, before, rawItem(t, fx, testFullName))
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)

		// a record that is not marked: no backoff, nothing marked, not due
		noOwner(fx.contracts)
		require.True(t, fx.refreshLeased(ctx, testFullName))
		item := cachedItem(t, fx)
		require.Zero(t, item.RefreshFailures)
		require.False(t, item.RefreshNeeded)
		require.Zero(t, item.RepairAt)
		require.Equal(t, testEoa, item.OwnerEthAddress)
	})

	t.Run("cached and marked (e.g. restored for a read): kept marked, retried after a growing backoff, not every lease", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, 0)
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"refresh_needed": true, "repair_at": int64(1)}})
		require.NoError(t, err)
		old := refreshLease
		refreshLease = time.Millisecond
		defer func() { refreshLease = old }()

		at := time.Now()
		for i, want := range []time.Duration{time.Minute, 2 * time.Minute} {
			fx.now = func() time.Time { return at }
			noOwner(fx.contracts)
			require.Equal(t, 1, repairRound(t, fx))
			item := cachedItem(t, fx)
			require.True(t, item.RefreshNeeded)
			require.Equal(t, i+1, item.RefreshFailures)
			require.Equal(t, at.Add(want).UnixMilli(), item.RepairAt)
			require.Equal(t, testAnyID, item.OwnerAnyAddress)
			fx.now = func() time.Time { return at.Add(want - time.Second) }
			require.Zero(t, repairRound(t, fx))
			at = at.Add(want + time.Second)
		}
	})

	t.Run("not cached: ErrNameNotRegistered (not errNotOnChain), nothing written", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		noOwner(fx.contracts)
		err := fx.refreshName()
		require.ErrorIs(t, err, ErrNameNotRegistered)
		require.NotErrorIs(t, err, errNotOnChain)
		require.Zero(t, countRecords(t, fx))
	})
}

// an older observation reads the record in its transaction, then a newer one commits: the
// older one must not overwrite it (it conflicts, runs again and sees the newer record)
func TestCacheService_TransactionInterleave(t *testing.T) {
	for _, c := range []struct {
		name          string
		older, newerO *NameDataItem
	}{
		{"registration over registration", obsAt(100, blockHash(100).Hex(), time.Now().UnixMilli()), obsAt(200, blockHash(200).Hex(), time.Now().UnixMilli())},
		{"registration over a lapse", obsAt(100, blockHash(100).Hex(), time.Now().UnixMilli()),
			&NameDataItem{FullName: testFullName, Lapsed: true, NameExpires: lapsed, ObservedBlock: 200, ObservedBlockHash: blockHash(200).Hex(), ObservedAt: time.Now().UnixMilli()}},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			requireReplicaSet(t, fx)
			seedItem(t, fx, notExpired, 50, 0)

			p := newPause(t)
			var once sync.Once
			hookAfterRead = func(o *NameDataItem) {
				if o.ObservedBlock == 100 {
					once.Do(p.wait)
				}
			}
			defer func() { hookAfterRead = nil }()

			type result struct {
				stored *NameDataItem
				err    error
			}
			done := make(chan result, 1)
			go func() {
				w, err := fx.applyObservation(ctx, c.older, refreshOpts{})
				done <- result{w, err}
			}()
			p.Entered(t)

			// the newer observation commits while the older one is in its transaction
			w, err := fx.applyObservation(ctx, c.newerO, refreshOpts{})
			require.NoError(t, err)
			require.Equal(t, int64(200), w.ObservedBlock)

			p.Release()
			select {
			case r := <-done:
				require.NoError(t, r.err)
				require.Equal(t, int64(200), r.stored.ObservedBlock)
			case <-time.After(watchdog):
				t.Fatal("the older observation did not finish")
			}

			item := cachedItem(t, fx)
			require.Equal(t, int64(200), item.ObservedBlock)
			require.Equal(t, c.newerO.Lapsed, item.Lapsed)
			require.Equal(t, testEoa, item.OwnerEthAddress, "the owner stays")
			require.EqualValues(t, 1, countRecords(t, fx))
		})
	}
}

// concurrent writers of one name: whatever the order, the newest block wins (only real
// transactions make read-compare-write atomic)
func TestCacheService_ConcurrentWriters(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)
	requireReplicaSet(t, fx)

	for round := 0; round < 5; round++ {
		_, err := fx.itemColl.DeleteMany(ctx, bson.M{})
		require.NoError(t, err)

		const writers = 16
		var wg sync.WaitGroup
		errs := make(chan error, writers)
		start := make(chan struct{})
		for i := 1; i <= writers; i++ {
			wg.Add(1)
			go func(block int64) {
				defer wg.Done()
				<-start
				o := obsAt(block, blockHash(block).Hex(), time.Now().UnixMilli())
				// every third one a lapse (stored in place, or nothing while there is no record)
				if block%3 == 0 {
					o = &NameDataItem{FullName: testFullName, Lapsed: true, NameExpires: lapsed, ObservedBlock: block, ObservedBlockHash: blockHash(block).Hex()}
				}
				_, err := fx.applyObservation(ctx, o, refreshOpts{})
				errs <- err
			}(int64(i))
		}
		close(start)
		within(t, "concurrent writes", wg.Wait)
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}

		item := cachedItem(t, fx)
		require.Equal(t, int64(writers), item.ObservedBlock, "round %d", round)
		require.EqualValues(t, 1, countRecords(t, fx))
	}
}

// two first inserts of a name: one record (the unique index), the loser runs again
func TestCacheService_ConcurrentFirstWrites(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)

	for round := 0; round < 10; round++ {
		_, err := fx.itemColl.DeleteMany(ctx, bson.M{})
		require.NoError(t, err)

		const writers = 8
		start := make(chan struct{})
		errs := make(chan error, writers)
		for i := 1; i <= writers; i++ {
			go func(block int64) {
				<-start
				_, err := fx.applyObservation(ctx, obsAt(block, blockHash(block).Hex(), time.Now().UnixMilli()), refreshOpts{})
				errs <- err
			}(int64(i))
		}
		close(start)
		for i := 0; i < writers; i++ {
			select {
			case err := <-errs:
				require.NoError(t, err, "round %d", round)
			case <-time.After(watchdog):
				t.Fatal("a write did not return")
			}
		}
		require.EqualValues(t, 1, countRecords(t, fx))
	}
}

// the driver commits and aborts without a deadline: a stalled commit or abort must not hold
// the caller (a GetOperation poll, a readFromCache: false lookup) longer than storeTimeout
func TestCacheService_StalledTransactionIsBounded(t *testing.T) {
	for _, c := range []struct {
		name     string
		commands []string
	}{
		{"commit", []string{"commitTransaction"}},
		// the write stalls (the callback fails on the deadline), then the abort stalls
		{"abort", []string{"update", "abortTransaction"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			requireReplicaSet(t, fx)
			seedItem(t, fx, notExpired, 10, 0)

			old := storeTimeout
			storeTimeout = 300 * time.Millisecond
			defer func() { storeTimeout = old }()

			// only this client's commands stall
			const appName = "go7567-stall"
			cs, m := serviceWithClient(t, fx, options.Client().SetAppName(appName))
			// failCommand skips the test when fail points are unavailable; set it
			// before the mock expectations so a skip leaves none unmet.
			failCommand(t, bson.M{
				"failCommands":    c.commands,
				"appName":         appName,
				"blockConnection": true,
				"blockTimeMS":     stallBlock.Milliseconds(),
			}, bson.M{"times": len(c.commands)})
			expectRegistered(m, testScw, testEoa, testAnyID, notExpired+1)

			start := time.Now()
			var err error
			within(t, "UpdateInCache", func() {
				err = cs.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
			})
			elapsed := time.Since(start)

			require.Error(t, err)
			require.NotErrorIs(t, err, ErrNameNotRegistered)
			require.NotErrorIs(t, err, ErrNameDataIncomplete)
			require.Less(t, elapsed, stallBlock-time.Second, "the transaction was waited for")
		})
	}
}

// the client reads from secondaries (here: only, so that a single-node set can tell): transactions
// must still read from the primary
func TestCacheService_TransactionsReadFromThePrimary(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)
	requireReplicaSet(t, fx)

	cs, m := serviceWithClient(t, fx, options.Client().SetReadPreference(readpref.Secondary()))
	expectRegistered(m, testScw, testEoa, testAnyID, notExpired)

	require.NoError(t, cs.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
	require.Equal(t, testEoa, cachedItem(t, fx).OwnerEthAddress)
}

func startCache(t *testing.T, conf *config.Config) error {
	m := mock_contracts.NewMockContractsService(gomock.NewController(t))
	m.EXPECT().Name().Return(contracts.CName).AnyTimes()
	m.EXPECT().Init(gomock.Any()).AnyTimes()

	a := new(app.App)
	a.Register(rpctest.NewTestServer()).Register(conf).Register(m).Register(New())
	err := a.Start(ctx)
	_ = a.Close(ctx)
	return err
}

func TestCacheService_StandaloneNeedsAnExplicitFlag(t *testing.T) {
	fx := newFixture(t)
	txSupported := fx.txSupported
	fx.finish(t)
	if txSupported {
		t.Skip("the test Mongo is a replica set")
	}

	for _, allow := range []bool{false, true} {
		conf := new(config.Config)
		conf.Mongo = config.Mongo{Connect: testMongoURI(), Database: testDbName}
		conf.Cache = config.Cache{AllowUnsafeStandalone: allow}
		err := startCache(t, conf)
		if allow {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, "replica set")
		}
	}
}

// the unique index on the name: an existing one is accepted (prod: name_1), a missing one is
// created, a non-unique one stops the node (it is never dropped)
func TestCacheService_NameIndexAtStart(t *testing.T) {
	const dbName = "any-ns-test-cache-index"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI()))
	require.NoError(t, err)
	defer func() { _ = client.Disconnect(ctx) }()
	coll := client.Database(dbName).Collection("cache")

	conf := new(config.Config)
	conf.Mongo = config.Mongo{Connect: testMongoURI(), Database: dbName}
	conf.Cache = config.Cache{AllowUnsafeStandalone: true}

	indexes := func() map[string]bool {
		specs, err := coll.Indexes().ListSpecifications(ctx)
		require.NoError(t, err)
		out := map[string]bool{}
		for _, s := range specs {
			out[s.Name] = s.Unique != nil && *s.Unique
		}
		return out
	}
	reset := func() {
		require.NoError(t, client.Database(dbName).Drop(ctx))
	}

	t.Run("an existing unique index is accepted whatever its name", func(t *testing.T) {
		for _, name := range []string{"name_1", "my_name_index"} {
			reset()
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true).SetName(name)})
			require.NoError(t, err)
			require.NoError(t, startCache(t, conf))
			require.Equal(t, map[string]bool{"_id_": false, "repair_at": false, "name_ci": false, "canon_ci": false, name: true}, indexes())
		}
	})

	t.Run("a missing one is created", func(t *testing.T) {
		reset()
		require.NoError(t, startCache(t, conf))
		require.Equal(t, map[string]bool{"_id_": false, "repair_at": false, "name_ci": false, "canon_ci": false, "name_1": true}, indexes())
	})

	t.Run("a non-unique one stops it, nothing is dropped", func(t *testing.T) {
		reset()
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "name", Value: 1}}})
		require.NoError(t, err)
		err = startCache(t, conf)
		require.ErrorIs(t, err, ErrNameIndexNotUnique)
		require.ErrorContains(t, err, "name_1")
		require.Equal(t, map[string]bool{"_id_": false, "name_1": false}, indexes())
	})
	reset()
}

// every contract read and the cache key use the canonical spelling of the name; a registry owner
// without a registrar expiry is a failure, never a lapse
func TestCacheService_CanonicalNames(t *testing.T) {
	t.Run("a mixed-case name is read and cached as the canonical one", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		// (the mocks expect testFullName, "test.any")
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: "Test.any"}))
		require.Equal(t, testEoa, cachedItem(t, fx).OwnerEthAddress)
		require.EqualValues(t, 1, countRecords(t, fx))
	})

	t.Run("a legacy record under a non-canonical spelling is never removed while the name is registered", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		_, err := fx.itemColl.InsertOne(ctx, bson.M{"name": "Test.any", "owner_eth_address": testEoa,
			"owner_scw_eth_address": testScw, "owner_any_address": testAnyID, "name_expires": lapsed})
		require.NoError(t, err)

		// the backfill (apply) refreshes the canonical name
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		stats, err := fx.RefreshAll(ctx, true, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 1, Updated: 1, NonCanonical: 1}, stats)

		// the reverse lookup answers the canonical record (fresh: nothing is handed to the background)
		res, err := fx.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{AnyAddress: testAnyID})
		require.NoError(t, err)
		require.Equal(t, testFullName, res.Name)
		require.Zero(t, fx.runQueued())

		var legacy NameDataItem
		require.NoError(t, fx.itemColl.FindOne(ctx, bson.M{"name": "Test.any"}).Decode(&legacy))
		require.False(t, legacy.Removed)
		require.Equal(t, testEoa, legacy.OwnerEthAddress)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
	})

	t.Run("a registry owner without a registrar expiry: a failure, nothing removed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).AnyTimes()
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(0), nil).AnyTimes()

		err := fx.refreshName()
		require.ErrorIs(t, err, errInconsistentRegistry)
		require.NotErrorIs(t, err, ErrNameNotRegistered)

		stats, err := fx.RefreshAll(ctx, true, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 1, Failed: 1}, stats)

		require.True(t, fx.queue.push(refreshRequest{name: testFullName}))
		require.Equal(t, 1, fx.runQueued())
		item := cachedItem(t, fx)
		require.False(t, item.Removed)
		require.Greater(t, item.RefreshNextAt, time.Now().UnixMilli(), "backing off")
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), inGrace)
	})
}
