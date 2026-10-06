package cache

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"
)

// repairRound runs the periodic scan once, at the local time
func repairRound(t *testing.T, fx *fixture) int {
	t.Helper()
	n, err := fx.repairOnce(ctx, repairBatch)
	require.NoError(t, err)
	return n
}

// rawItem: the stored document as it is
func rawItem(t *testing.T, fx *fixture, name string) bson.M {
	t.Helper()
	var doc bson.M
	require.NoError(t, fx.itemColl.FindOne(ctx, bson.M{"name": name}).Decode(&doc))
	return doc
}

func TestNeedsRefresh(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)

	// not expired -> never
	require.False(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() + 1}, now))
	// expired, never read (an old record) -> yes
	require.True(t, needsRefresh(&NameDataItem{NameExpires: now.Unix()}, now))
	// expired, read recently -> no
	require.False(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() - 1, ObservedAt: now.Add(-time.Minute).UnixMilli()}, now))
	// expired, read long ago -> yes
	require.True(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() - 1, ObservedAt: now.Add(-expiredRefreshInterval).UnixMilli()}, now))
	// lapsed, read recently -> yes: only the contracts can make it available
	require.True(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() - gracePeriodSec - 1, ObservedAt: now.UnixMilli()}, now))
	// marked -> yes, even if not expired
	require.True(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() + 1, RefreshNeeded: true}, now))
	// a tombstone: only an old one (the name could have been registered since)
	require.False(t, needsRefresh(&NameDataItem{Removed: true, ObservedAt: now.Add(-time.Minute).UnixMilli()}, now))
	require.True(t, needsRefresh(&NameDataItem{Removed: true, ObservedAt: now.Add(-expiredRefreshInterval).UnixMilli()}, now))
	// leased or backing off -> no
	require.False(t, needsRefresh(&NameDataItem{NameExpires: now.Unix() - gracePeriodSec - 1, RefreshNeeded: true, RefreshNextAt: now.UnixMilli() + 1}, now))
}

func TestRepairAt(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	ms := now.UnixMilli()

	// incomplete: now, after its backoff
	require.Equal(t, int64(1), repairAt(&NameDataItem{RefreshNeeded: true, NameExpires: now.Unix() + 100}))
	require.Equal(t, ms+5, repairAt(&NameDataItem{RefreshNeeded: true, RefreshNextAt: ms + 5}))
	// a registration: when it expires
	require.Equal(t, (now.Unix()+100)*1000, repairAt(&NameDataItem{NameExpires: now.Unix() + 100, ObservedAt: ms}))
	// expired at the read, lapses in an hour: when it lapses
	exp := now.Unix() - gracePeriodSec + 3600
	require.Equal(t, (exp+gracePeriodSec+1)*1000, repairAt(&NameDataItem{NameExpires: exp, ObservedAt: ms}))
	// expired at the read, lapses in 30 days: once per expiredRepairInterval
	require.Equal(t, ms+expiredRepairInterval.Milliseconds(), repairAt(&NameDataItem{NameExpires: now.Unix() - 60, ObservedAt: ms}))
	// a tombstone, a legacy record without an expiry: never
	require.Zero(t, repairAt(&NameDataItem{Removed: true, ObservedAt: ms}))
	require.Zero(t, repairAt(&NameDataItem{}))
}

// with readFromCache a lookup serves the cache only: no contract read, no Mongo write. a record
// that needs a refresh is handed to the background worker through the in-memory queue
func TestCacheService_LookupsNeverRefresh(t *testing.T) {
	t.Run("an expired record is served as taken and queued, nothing is read or written", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)
		before := rawItem(t, fx, testFullName)

		// (no contracts expectations: any call fails the test)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), inGrace)
		res, err := fx.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{AnyAddress: testAnyID})
		require.NoError(t, err)
		require.True(t, res.Found)
		res, err = fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: testScw})
		require.NoError(t, err)
		require.True(t, res.Found)

		require.Equal(t, before, rawItem(t, fx, testFullName))
		// queued once
		require.Len(t, fx.queue.ch, 1)
	})

	t.Run("a full queue drops the request, the lookup is served", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 10, 0)
		fx.queue = newRefreshQueue(1)
		require.True(t, fx.queue.push(refreshRequest{name: "other.any"}))

		within(t, "IsNameAvailable", func() {
			requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), inGrace)
		})
		require.Len(t, fx.queue.ch, 1)
		require.False(t, fx.queue.queued[refreshRequest{name: testFullName}])
	})

	t.Run("a record that does not need a refresh is not queued", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, time.Now().UnixMilli())

		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
		require.Zero(t, fx.runQueued())
	})
}

// the background worker refreshes what lookups queue: under the lease, a removal only when the
// finalized block confirms it
func TestCacheService_BackgroundRefresh(t *testing.T) {
	t.Run("a lapsed record: served as taken until the background refresh confirmed the lapse", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 10, 0)

		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)
		expectLapsed(fx.contracts)
		require.Equal(t, 1, fx.runQueued())

		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		item := cachedItem(t, fx)
		require.True(t, item.Removed)
		require.Zero(t, item.RepairAt)
	})

	t.Run("a lapsed record not confirmed by the finalized block stays taken and backs off", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 10, 0)

		fx.setFinalizedErr(errors.New("unknown block tag"))
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(lapsed), nil)

		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)
		require.Equal(t, 1, fx.runQueued())
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)

		item := cachedItem(t, fx)
		require.False(t, item.Removed)
		require.Greater(t, item.RefreshNextAt, time.Now().Add(refreshFailureBackoff-10*time.Second).UnixMilli())
		// backing off: the next lookup does not queue it, the scan does not take it
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)
		require.Zero(t, fx.runQueued())
		require.Zero(t, repairRound(t, fx))
	})

	t.Run("an old tombstone is re-read: a registration replaces it", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		_, err := fx.applyObservation(ctx, &NameDataItem{FullName: testFullName, Removed: true,
			ObservedBlock: 500, ObservedBlockHash: blockHash(500).Hex(),
			ObservedAt: time.Now().Add(-expiredRefreshInterval - time.Minute).UnixMilli()})
		require.NoError(t, err)

		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		fx.setHead(520, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, fx.runQueued())
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
	})

	t.Run("an old tombstone of a free name is confirmed again at a newer finalized block", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		_, err := fx.applyObservation(ctx, &NameDataItem{FullName: testFullName, Removed: true,
			ObservedBlock: 500, ObservedBlockHash: blockHash(500).Hex(),
			ObservedAt: time.Now().Add(-expiredRefreshInterval - time.Minute).UnixMilli()})
		require.NoError(t, err)

		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		fx.setHead(530, time.Now())
		fx.setFinalized(520, time.Now())
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)
		require.Equal(t, 1, fx.runQueued())

		item := cachedItem(t, fx)
		require.True(t, item.Removed)
		require.Equal(t, int64(520), item.ObservedBlock)
		// read now: the next lookups do not queue it
		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		require.Zero(t, fx.runQueued())
	})

	t.Run("a recent tombstone is served without reading the contracts", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		_, err := fx.applyObservation(ctx, &NameDataItem{FullName: testFullName, Removed: true,
			ObservedBlock: 500, ObservedBlockHash: blockHash(500).Hex(), ObservedAt: time.Now().UnixMilli()})
		require.NoError(t, err)

		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		require.Zero(t, fx.runQueued())
		require.Zero(t, repairRound(t, fx))
	})

	t.Run("a request is skipped while another node holds the lease", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 10, 0)

		other, _ := otherCacheService(t, fx, 1, time.Now())
		claimed, err := other.claimRefresh(ctx, testFullName, time.Now())
		require.NoError(t, err)
		require.True(t, claimed)

		require.True(t, fx.queue.push(refreshRequest{name: testFullName}))
		// (no contracts expectations: a refresh would fail the test)
		require.Equal(t, 1, fx.runQueued())
		require.False(t, cachedItem(t, fx).Removed)
	})
}

func TestCacheService_Lease(t *testing.T) {
	t.Run("one claim wins, whatever the clocks of the nodes", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 0, 0)

		a, _ := otherCacheService(t, fx, 1, time.Now())
		b, _ := otherCacheService(t, fx, 1, time.Now())

		for round := 0; round < 20; round++ {
			_, err := fx.itemColl.UpdateMany(ctx, bson.M{"name": testFullName}, bson.M{"$unset": bson.M{"refresh_next_at": ""}})
			require.NoError(t, err)

			now := time.Now()
			start := make(chan struct{})
			results := make(chan bool, 2)
			for i, cs := range []*cacheService{a, b} {
				go func(cs *cacheService, now time.Time) {
					<-start
					claimed, err := cs.claimRefresh(ctx, testFullName, now)
					results <- claimed && err == nil
				}(cs, now.Add(time.Duration(i)*time.Millisecond))
			}
			close(start)

			winners := 0
			for i := 0; i < 2; i++ {
				select {
				case claimed := <-results:
					if claimed {
						winners++
					}
				case <-time.After(watchdog):
					t.Fatal("a claim did not return")
				}
			}
			require.Equal(t, 1, winners, "round %d", round)
		}
	})

	t.Run("the lease ends: the name can be claimed again", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 0, 0)

		now := time.Now()
		claimed, err := fx.claimRefresh(ctx, testFullName, now)
		require.NoError(t, err)
		require.True(t, claimed)
		claimed, err = fx.claimRefresh(ctx, testFullName, now.Add(refreshLease-time.Second))
		require.NoError(t, err)
		require.False(t, claimed)
		claimed, err = fx.claimRefresh(ctx, testFullName, now.Add(refreshLease))
		require.NoError(t, err)
		require.True(t, claimed)
	})
}

// the periodic scan: incomplete records, expired and lapsed ones, through the repair index,
// the longest waiting first
func TestCacheService_Repair(t *testing.T) {
	t.Run("a refreshed registration is due when it expires", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		expires := time.Now().Add(time.Hour).Unix()
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, expires)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.Equal(t, expires*1000, cachedItem(t, fx).RepairAt)
		require.Zero(t, repairRound(t, fx))

		// an hour later: renewed elsewhere
		fx.now = func() time.Time { return time.Now().Add(time.Hour + time.Second) }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		require.Equal(t, notExpired*1000, item.RepairAt)
	})

	t.Run("an incomplete record is repaired after its backoff", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))
		require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameDataIncomplete)
		require.Zero(t, repairRound(t, fx), "backing off")

		fx.now = func() time.Time { return time.Now().Add(refreshFailureBackoff + time.Second) }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.False(t, item.RefreshNeeded)
		require.Equal(t, testEoa, item.OwnerEthAddress)
	})

	t.Run("a failing record backs off and can not starve the others", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// repairBatch failing records, due first; then one that works
		for i := 0; i < repairBatch; i++ {
			require.NoError(t, fx.setNameData(ctx, &NameDataItem{FullName: fmt.Sprintf("bad%02d.any", i), NameExpires: notExpired,
				RefreshNeeded: true, ObservedBlock: 1, RepairAt: int64(1 + i)}))
		}
		require.NoError(t, fx.setNameData(ctx, &NameDataItem{FullName: testFullName, NameExpires: notExpired,
			RefreshNeeded: true, ObservedBlock: 1, RepairAt: 1000}))

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(common.Address{}, errors.New("rpc is down")).Times(repairBatch)
		require.Equal(t, repairBatch, repairRound(t, fx))

		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		require.False(t, cachedItem(t, fx).RefreshNeeded)
	})

	t.Run("a legacy record without repair_at is not scanned (the backfill and the lookups find it)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		_, err := fx.itemColl.InsertOne(ctx, bson.M{"name": testFullName, "owner_eth_address": testEoa, "name_expires": lapsed})
		require.NoError(t, err)
		require.Zero(t, repairRound(t, fx))
	})
}

// the periodic scan reads the repair index only: no collection scan, no in-memory sort
func TestCacheService_RepairScanUsesTheIndex(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)

	var docs []interface{}
	for i := 0; i < 50; i++ {
		d := NameDataItem{FullName: fmt.Sprintf("n%02d.any", i), NameExpires: notExpired}
		if i%3 == 0 {
			d.RefreshNeeded, d.RepairAt = true, int64(i+1)
		}
		docs = append(docs, d)
	}
	_, err := fx.itemColl.InsertMany(ctx, docs)
	require.NoError(t, err)
	// Init created it, but the fixture drops the database after the start
	fx.ensureRepairIndex(ctx)

	var plan bson.M
	err = fx.itemColl.Database().RunCommand(ctx, bson.D{
		{Key: "explain", Value: bson.D{
			{Key: "find", Value: fx.itemColl.Name()},
			{Key: "filter", Value: repairDue(time.Now())},
			{Key: "sort", Value: repairOrder},
			{Key: "limit", Value: repairBatch},
		}},
		{Key: "verbosity", Value: "queryPlanner"},
	}).Decode(&plan)
	require.NoError(t, err)
	winning, err := bson.MarshalExtJSON(plan["queryPlanner"].(bson.M)["winningPlan"], false, false)
	require.NoError(t, err)

	require.Contains(t, string(winning), `"IXSCAN"`)
	require.Contains(t, string(winning), `"indexName":"`+repairIndexName+`"`)
	require.NotContains(t, string(winning), `"SORT"`)
	require.NotContains(t, string(winning), `"COLLSCAN"`)

	n, err := fx.itemColl.CountDocuments(ctx, repairDue(time.Now()))
	require.NoError(t, err)
	require.EqualValues(t, 17, n)
}

// the background worker takes the refresh lease and stores a backoff after a failure: a stalled
// write of either is bounded (the worker moves on)
func TestCacheService_StalledLeaseWritesAreBounded(t *testing.T) {
	for _, c := range []struct {
		name string
		// failCommand mode: the lease is the first update, the backoff the second
		mode bson.M
	}{
		{"lease", bson.M{"times": 1}},
		{"backoff", bson.M{"skip": 1}},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			requireReplicaSet(t, fx)
			seedItem(t, fx, inGrace, 0, 0)
			_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"repair_at": 1}})
			require.NoError(t, err)

			old := storeTimeout
			storeTimeout = 300 * time.Millisecond
			defer func() { storeTimeout = old }()

			const appName = "go7567-stall-lease"
			cs, m := serviceWithClient(t, fx, options.Client().SetAppName(appName))
			if c.name == "backoff" {
				m.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
			}
			failCommand(t, bson.M{
				"failCommands":    []string{"update"},
				"appName":         appName,
				"blockConnection": true,
				"blockTimeMS":     stallBlock.Milliseconds(),
			}, c.mode)

			start := time.Now()
			within(t, "repairOnce", func() {
				_, _ = cs.repairOnce(ctx, repairBatch)
			})
			require.Less(t, time.Since(start), stallBlock-time.Second, "the write was waited for")
		})
	}
}
