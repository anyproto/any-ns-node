package cache

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
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
		// a short lease: what holds the record afterwards is the backoff
		old := refreshLease
		refreshLease = time.Second
		defer func() { refreshLease = old }()

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
			ObservedAt: time.Now().Add(-expiredRefreshInterval - time.Minute).UnixMilli()}, refreshOpts{})
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
			ObservedAt: time.Now().Add(-expiredRefreshInterval - time.Minute).UnixMilli()}, refreshOpts{})
		require.NoError(t, err)

		require.True(t, isNameAvailable(t, fx.cacheService).Available)
		fx.setHead(530, time.Now())
		fx.setFinalized(520, time.Now())
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(0), nil).Times(2)
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
			ObservedBlock: 500, ObservedBlockHash: blockHash(500).Hex(), ObservedAt: time.Now().UnixMilli()}, refreshOpts{})
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
		// a new record: its re-reads first (+5, +30 min), then its expiry
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$unset": bson.M{"rereads": ""}, "$set": bson.M{"repair_at": expires * 1000}})
		require.NoError(t, err)
		require.Zero(t, repairRound(t, fx))

		// an hour later: renewed elsewhere (a change: its re-reads, then the new expiry)
		fx.now = func() time.Time { return time.Now().Add(time.Hour + time.Second) }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		require.Len(t, item.Rereads, 2)
		require.Equal(t, item.Rereads[0], item.RepairAt)
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

		// a short lease: what moves the failing records back is their backoff
		old := refreshLease
		refreshLease = time.Millisecond
		defer func() { refreshLease = old }()

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

func TestMergeRereads(t *testing.T) {
	m := time.Minute.Milliseconds()
	// union, sorted
	require.Equal(t, []int64{10 * m, 20 * m, 30 * m}, mergeRereads([]int64{20 * m}, []int64{30 * m, 10 * m}, 0))
	// the ones at or before the read are done
	require.Equal(t, []int64{30 * m}, mergeRereads([]int64{10 * m, 20 * m}, []int64{30 * m}, 20*m))
	// close ones coalesce (repeated polls)
	require.Equal(t, []int64{10 * m, 35 * m}, mergeRereads([]int64{10 * m, 35 * m}, []int64{10*m + 1000, 35*m + 2000}, 0))
	// bounded
	var many []int64
	for i := int64(1); i <= 20; i++ {
		many = append(many, i*10*m)
	}
	require.Len(t, mergeRereads(many, nil, 0), maxRereads)
	require.Nil(t, mergeRereads(nil, nil, 0))
}

// a completed operation schedules re-reads of the name (rereadDelays): to pick up a lagging
// provider and reorgs. the periodic scan runs them; a read after one is done with it
func TestCacheService_Rereads(t *testing.T) {
	t.Run("scheduled in the write of the refresh, run by the scan at their time, then done", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		start := time.Now()
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCacheAfterOperation(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		item := cachedItem(t, fx)
		require.Len(t, item.Rereads, 2)
		require.GreaterOrEqual(t, item.Rereads[0], start.Add(5*time.Minute).UnixMilli())
		require.GreaterOrEqual(t, item.Rereads[1], start.Add(30*time.Minute).UnixMilli())
		require.Equal(t, item.Rereads[0], item.RepairAt)
		require.Zero(t, repairRound(t, fx))

		// 5 minutes later: the first one. the lagging provider has caught up with a renewal (a
		// change: it schedules re-reads of its own)
		fx.now = func() time.Time { return time.Now().Add(5*time.Minute + time.Second) }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired+100)
		require.Equal(t, 1, repairRound(t, fx))
		item = cachedItem(t, fx)
		require.Equal(t, notExpired+100, item.NameExpires)
		require.Equal(t, slices.Min(item.Rereads), item.RepairAt)
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"rereads": bson.A{item.Rereads[0]}}})
		require.NoError(t, err)

		// 30 minutes later: the second one, nothing changed
		fx.now = func() time.Time { return time.Now().Add(30*time.Minute + time.Second) }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired+100)
		require.Equal(t, 1, repairRound(t, fx))
		item = cachedItem(t, fx)
		require.Empty(t, item.Rereads)
		require.Equal(t, (notExpired+100)*1000, item.RepairAt, "back to the expiry")
		require.Zero(t, repairRound(t, fx))
	})

	t.Run("a re-read that fails backs off and stays", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCacheAfterOperation(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		later := time.Now().Add(5*time.Minute + time.Second)
		fx.now = func() time.Time { return later }
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.Len(t, item.Rereads, 2)
		require.Equal(t, later.Add(refreshFailureBackoff).UnixMilli(), item.RepairAt)
		require.Zero(t, repairRound(t, fx))
	})

	t.Run("a stale observation (older block) does not do the due re-reads, a read of the stored state does", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, 0)
		r1, r2 := 10*time.Minute.Milliseconds(), 50*time.Minute.Milliseconds()
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"rereads": bson.A{r1, r2}, "repair_at": r1}})
		require.NoError(t, err)

		stored, err := fx.applyObservation(ctx, obsAt(400, blockHash(400).Hex(), r2+1), refreshOpts{})
		require.NoError(t, err)
		require.Equal(t, int64(500), stored.ObservedBlock)
		item := cachedItem(t, fx)
		require.Equal(t, []int64{r1, r2}, item.Rereads)
		require.Equal(t, r1, item.RepairAt)

		// another fork at the same height, read before the stored one: stale too
		_, err = fx.applyObservation(ctx, obsAt(500, blockHash(499).Hex(), -1), refreshOpts{})
		require.NoError(t, err)
		require.Equal(t, []int64{r1, r2}, cachedItem(t, fx).Rereads)

		// the stored state read again (same block and hash) at 2000: the first one is done
		same := obsAt(500, blockHash(500).Hex(), r1+1)
		same.OwnerAnyAddress, same.SpaceId = testAnyID, "space"
		_, err = fx.applyObservation(ctx, same, refreshOpts{})
		require.NoError(t, err)
		item = cachedItem(t, fx)
		require.Equal(t, []int64{r2}, item.Rereads)
		require.Equal(t, r2, item.RepairAt)
	})

	t.Run("re-reads against a lagging head stay due until the head moves, then reconcile", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// a renewal cached at block 500 (it can be on a fork that is reorged away)
		fx.setHead(500, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired+100)
		require.NoError(t, fx.UpdateInCacheAfterOperation(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		// both re-reads hit a backend at block 499 (the old expiry): stale, they stay due
		fx.setHead(499, time.Now())
		for _, at := range []time.Duration{5 * time.Minute, 30 * time.Minute} {
			later := time.Now().Add(at + time.Second)
			fx.now = func() time.Time { return later }
			expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
			require.Equal(t, 1, repairRound(t, fx))
			item := cachedItem(t, fx)
			require.Len(t, item.Rereads, 2)
			require.Equal(t, notExpired+100, item.NameExpires)
			// retried after the lease, not at the cached expiry
			require.LessOrEqual(t, item.RepairAt, later.Add(refreshLease).UnixMilli())
		}

		// the chain moved on (block 510): the renewal was reorged away, the old expiry is final
		later := time.Now().Add(31*time.Minute + refreshLease + time.Second)
		fx.now = func() time.Time { return later }
		fx.setHead(510, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		// the due ones are done; the change schedules its own
		require.Len(t, item.Rereads, 2)
		require.Greater(t, item.Rereads[0], later.UnixMilli())
		require.Equal(t, item.Rereads[0], item.RepairAt)
	})

	t.Run("a removal refused against a newer registration keeps the re-reads", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.setHead(500, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCacheAfterOperation(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		// at the re-read: no owner at the latest block, and the finalized block (490) is behind
		// the cached registration: the tombstone loses, not final
		later := time.Now().Add(5*time.Minute + time.Second)
		fx.now = func() time.Time { return later }
		fx.setHead(505, time.Now())
		fx.setFinalized(490, time.Now())
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(0), nil).Times(2)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.False(t, item.Removed)
		require.Len(t, item.Rereads, 2)
		require.Equal(t, later.Add(refreshFailureBackoff).UnixMilli(), item.RepairAt)

		// finality catches up: removed
		later2 := later.Add(refreshFailureBackoff + time.Second)
		fx.now = func() time.Time { return later2 }
		fx.setHead(520, time.Now())
		fx.setFinalized(510, time.Now())
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(0), nil).Times(2)
		require.Equal(t, 1, repairRound(t, fx))
		require.True(t, cachedItem(t, fx).Removed)
	})

	t.Run("a failed refresh after an operation marks the record (data kept) with the re-reads", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
		fx.RefreshAfterOperation(testFullName)
		// the schedule first (a standalone Mongo does not order concurrent writes)
		fx.async.Wait()
		require.Equal(t, 1, fx.runQueued())

		item := cachedItem(t, fx)
		require.True(t, item.RefreshNeeded)
		require.Len(t, item.Rereads, 2)
		require.Greater(t, item.RefreshNextAt, time.Now().UnixMilli())
		require.Equal(t, item.RefreshNextAt, item.RepairAt)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
	})

	t.Run("after an operation the background leaves a leased record to its holder, the re-reads are stored anyway", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		other, _ := otherCacheService(t, fx, 1, time.Now())
		claimed, err := other.claimRefresh(ctx, testFullName, time.Now())
		require.NoError(t, err)
		require.True(t, claimed)

		// (no contracts expectations: a refresh would fail the test)
		fx.RefreshAfterOperation(testFullName)
		fx.async.Wait()
		require.Equal(t, 1, fx.runQueued())
		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		require.Len(t, item.Rereads, 2)
	})

	t.Run("after an operation the background respects a failure backoff", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
		fx.RefreshAfterOperation(testFullName)
		fx.async.Wait()
		require.Equal(t, 1, fx.runQueued())
		require.True(t, cachedItem(t, fx).RefreshNeeded)

		// polled again during the incident: no read before the backoff ends
		for i := 0; i < 5; i++ {
			fx.RefreshAfterOperation(testFullName)
			fx.async.Wait()
			require.Equal(t, 1, fx.runQueued())
		}
	})

	t.Run("the re-reads of a cached name survive a dropped request and a restart", func(t *testing.T) {
		for _, c := range []string{"queue overflow", "restart"} {
			t.Run(c, func(t *testing.T) {
				fx := newFixture(t)
				defer fx.finish(t)
				// an unexpired renewal: nothing else would refresh it before the old expiry
				seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

				start := time.Now()
				require.NoError(t, fx.scheduleRereads(ctx, testFullName))
				if c == "queue overflow" {
					fx.queue = newRefreshQueue(1)
					require.True(t, fx.queue.push(refreshRequest{name: "other.any"}))
					fx.RefreshAfterOperation(testFullName)
					require.Len(t, fx.queue.ch, 1, "dropped")
				}
				item := cachedItem(t, fx)
				require.Len(t, item.Rereads, 2)
				require.GreaterOrEqual(t, item.RepairAt, start.Add(5*time.Minute).UnixMilli())
				require.Less(t, item.RepairAt, notExpired*1000)

				// "restart": another service on the same collection, its queue empty
				cs := fx.cacheService
				var m = fx.contracts
				if c == "restart" {
					cs, m = otherCacheService(t, fx, 600, time.Now())
				}
				later := time.Now().Add(5*time.Minute + time.Second)
				cs.now = func() time.Time { return later }
				expectRegistered(m, testScw, testEoa, testAnyID, notExpired+100)
				n, err := cs.repairOnce(ctx, repairBatch)
				require.NoError(t, err)
				require.Equal(t, 1, n)
				item = cachedItem(t, fx)
				require.Equal(t, notExpired+100, item.NameExpires)
				// the +30 min one stays (the renewal adds its own, coalesced into the same window)
				require.Greater(t, slices.Min(item.Rereads), later.UnixMilli())
				require.LessOrEqual(t, len(item.Rereads), 3)
			})
		}
	})

	t.Run("scheduling respects the record's lease, and a repeated schedule coalesces on the next write", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())
		now := time.Now()
		claimed, err := fx.claimRefresh(ctx, testFullName, now.Add(10*time.Minute))
		require.NoError(t, err)
		require.True(t, claimed)

		require.NoError(t, fx.scheduleRereads(ctx, testFullName))
		require.NoError(t, fx.scheduleRereads(ctx, testFullName))
		item := cachedItem(t, fx)
		require.Equal(t, item.RefreshNextAt, item.RepairAt, "not before the lease")
		require.Len(t, item.Rereads, 2, "coalesced")
	})

	t.Run("sustained polling with the background held: the stored schedule stays bounded, the poll never waits", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		// the worker is stopped (the fixture): nothing processes the queue. a poll every 2 minutes
		// for 2 hours, then 50 more within a second
		start := time.Now()
		for i := 0; i < 60+50; i++ {
			at := start.Add(time.Duration(min(i, 60)) * 2 * time.Minute)
			fx.now = func() time.Time { return at }
			within(t, "RefreshAfterOperation", func() { fx.RefreshAfterOperation(testFullName) })
			fx.async.Wait()
		}
		item := cachedItem(t, fx)
		require.LessOrEqual(t, len(item.Rereads), maxRereads)
		require.NotEmpty(t, item.Rereads)
		require.Equal(t, slices.Min(item.Rereads), item.RepairAt)
	})

	t.Run("a stalled schedule write is not waited for, Close waits for it (bounded)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		requireReplicaSet(t, fx)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())

		old := scheduleTimeout
		scheduleTimeout = 300 * time.Millisecond
		defer func() { scheduleTimeout = old }()
		failCommand(t, bson.M{
			"failCommands":    []string{"find", "update", "commitTransaction"},
			"blockConnection": true,
			"blockTimeMS":     stallBlock.Milliseconds(),
		}, bson.M{"times": 3})

		start := time.Now()
		fx.RefreshAfterOperation(testFullName)
		require.Less(t, time.Since(start), 100*time.Millisecond)
		within(t, "the schedule", fx.async.Wait)
		require.Less(t, time.Since(start), stallBlock-time.Second, "bounded by scheduleTimeout")
	})

	t.Run("the scheduling cap holds for the transactions themselves (stalled commits)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		requireReplicaSet(t, fx)
		const names, slots = 12, 2
		for i := 0; i < names; i++ {
			insertRaw(t, fx, bson.M{"name": fmt.Sprintf("n%02d.any", i), "owner_eth_address": testEoa, "name_expires": notExpired, "observed_block": int64(1)})
		}

		oldStore, oldSchedule := storeTimeout, scheduleTimeout
		storeTimeout, scheduleTimeout = 200*time.Millisecond, 200*time.Millisecond
		defer func() { storeTimeout, scheduleTimeout = oldStore, oldSchedule }()
		fx.scheduling = make(chan struct{}, slots)
		// every commit stalls well past the callers' waits
		failCommand(t, bson.M{
			"failCommands":    []string{"commitTransaction"},
			"blockConnection": true,
			"blockTimeMS":     1500,
		}, bson.M{"times": names})

		var peak int64
		deadline := time.Now().Add(4 * time.Second)
		for i := 0; time.Now().Before(deadline); i++ {
			if i < names*10 {
				fx.RefreshAfterOperation(fmt.Sprintf("n%02d.any", i%names))
			}
			peak = max(peak, fx.txActive.Load())
			time.Sleep(10 * time.Millisecond)
		}
		within(t, "the schedules", fx.async.Wait)
		require.LessOrEqual(t, peak, int64(slots))
		require.Positive(t, peak)
	})
}
