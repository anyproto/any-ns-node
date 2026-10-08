package cache

import (
	"errors"
	"math/big"
	"slices"
	"testing"
	"time"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
)

// a: the registrar decides, even without a registry owner (a registrant can reclaim the name to
// address(0): it stays reserved until it lapses)
func TestCacheService_ReservedWithoutRegistryOwner(t *testing.T) {
	noOwner := func(fx *fixture, expires int64, times int) {
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(times)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(expires), nil).Times(times)
	}

	t.Run("cached, the owner reclaimed to zero, unexpired: stays taken, never a tombstone", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, 0)

		noOwner(fx, notExpired, 1)
		require.NoError(t, fx.refreshName())
		item := cachedItem(t, fx)
		require.False(t, item.Removed)
		require.False(t, item.Incomplete)
		require.Empty(t, item.OwnerEthAddress)
		require.Equal(t, common.Address{}.Hex(), common.HexToAddress(item.RegistryOwner).Hex())
		out := isNameAvailable(t, fx.cacheService)
		require.False(t, out.Available)
		require.Equal(t, notExpired, out.NameExpires)
	})

	t.Run("not cached, reserved in the grace period: cached as taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		noOwner(fx, inGrace, 1)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
	})

	t.Run("no owner and lapsed (or never registered): not registered", func(t *testing.T) {
		for _, expires := range []int64{lapsed, 0} {
			fx := newFixture(t)
			noOwner(fx, expires, 1)
			require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameNotRegistered)
			require.Nil(t, cachedItem(t, fx))
			fx.finish(t)
		}
	})
}

// b: a failed refresh is retried at the backoff deadline, not at a distant cached expiry
func TestCacheService_FailedRefreshRetriedAtTheBackoff(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)
	expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
	require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
	_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$unset": bson.M{"rereads": ""}, "$set": bson.M{"repair_at": notExpired * 1000}})
	require.NoError(t, err)

	// the background: the registry can not be read
	fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
	start := time.Now()
	require.True(t, fx.refreshLeased(ctx, testFullName))

	item := cachedItem(t, fx)
	require.Equal(t, 1, item.RefreshFailures)
	require.GreaterOrEqual(t, item.RepairAt, start.Add(refreshFailureBackoff).UnixMilli())
	require.LessOrEqual(t, item.RepairAt, time.Now().Add(refreshLease).UnixMilli(), "not at the cached expiry")
}

// b2: the backoff of a record that keeps failing doubles with every failure in a row (up to a
// day); a successful write resets it
func TestCacheService_FailureBackoffGrowsAndResets(t *testing.T) {
	t.Run("failed background refreshes", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, 0)
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"refresh_needed": true, "repair_at": int64(1)}})
		require.NoError(t, err)
		// a short lease: what holds the record is the backoff
		old := refreshLease
		refreshLease = time.Millisecond
		defer func() { refreshLease = old }()

		at := time.Now()
		for i, want := range []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute} {
			fx.now = func() time.Time { return at }
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))
			require.Equal(t, 1, repairRound(t, fx), "failure %d", i+1)
			item := cachedItem(t, fx)
			require.Equal(t, i+1, item.RefreshFailures)
			require.Equal(t, at.Add(want).UnixMilli(), item.RefreshNextAt, "failure %d", i+1)
			require.Equal(t, item.RefreshNextAt, item.RepairAt)
			// not before the backoff
			fx.now = func() time.Time { return at.Add(want - time.Second) }
			require.Zero(t, repairRound(t, fx))
			at = at.Add(want + time.Second)
		}

		// a success resets it
		fx.now = func() time.Time { return at }
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.Equal(t, 1, repairRound(t, fx))
		item := cachedItem(t, fx)
		require.Zero(t, item.RefreshFailures)
		require.False(t, item.RefreshNeeded)
		_, ok := rawItem(t, fx, testFullName)["refresh_failures"]
		require.False(t, ok)
	})

	t.Run("failed refreshes after operations (markRefreshNeeded), capped at a day", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 10, 0)
		_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"refresh_failures": 20}})
		require.NoError(t, err)
		at := time.Now()
		fx.now = func() time.Time { return at }
		require.NoError(t, fx.markRefreshNeeded(ctx, testFullName, nil))
		item := cachedItem(t, fx)
		require.Equal(t, 21, item.RefreshFailures)
		require.Equal(t, at.Add(maxFailureBackoff).UnixMilli(), item.RefreshNextAt)
	})

	t.Run("incomplete reads in a row back off too", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		for i, want := range []time.Duration{time.Minute, 2 * time.Minute, 4 * time.Minute} {
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
			fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))
			require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameDataIncomplete)
			item := cachedItem(t, fx)
			require.Equal(t, i+1, item.RefreshFailures)
			require.Equal(t, item.ObservedAt+want.Milliseconds(), item.RefreshNextAt)
		}
	})
}

// c: completeness (what orders two reads of one block) is separate from the retry flag; an
// incomplete read carries the owner fields over only from the same registry owner / wallet
func TestCacheService_IncompleteReads(t *testing.T) {
	infoFails := func(fx *fixture, registryOwner string) {
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(registryOwner), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))
	}

	t.Run("complete, marked for a retry, then an incomplete read of the same block: the data stays", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		fx.setHead(700, time.Now())
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.NoError(t, fx.markRefreshNeeded(ctx, testFullName, nil))
		require.True(t, cachedItem(t, fx).RefreshNeeded)

		infoFails(fx, nameWrapper)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		item := cachedItem(t, fx)
		require.False(t, item.Incomplete)
		require.Equal(t, testEoa, item.OwnerEthAddress)
		require.Equal(t, testAnyID, item.OwnerAnyAddress)
		require.True(t, byAnyID(t, fx, testAnyID).Found)
	})

	t.Run("a wrapped name transferred, the enrichment fails: nothing of the old owner is carried over", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		// the registry owner is the NameWrapper before and after the transfer
		infoFails(fx, nameWrapper)
		require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameDataIncomplete)
		item := cachedItem(t, fx)
		require.True(t, item.Incomplete)
		require.True(t, item.RefreshNeeded)
		require.Empty(t, item.OwnerEthAddress)
		require.Empty(t, item.OwnerScwEthAddress)
		require.Empty(t, item.OwnerAnyAddress)
		require.False(t, byAnyID(t, fx, testAnyID).Found, "the old owner does not resolve")
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
	})

	t.Run("another registry owner (or none known): nothing is carried over", func(t *testing.T) {
		for _, legacy := range []bool{false, true} {
			fx := newFixture(t)
			if legacy {
				seedItem(t, fx, notExpired, 10, 0)
			} else {
				expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
				require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
			}
			owner := otherEoa
			if legacy {
				owner = nameWrapper
			}
			infoFails(fx, owner)
			require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameDataIncomplete)
			item := cachedItem(t, fx)
			require.True(t, item.Incomplete)
			require.Empty(t, item.OwnerEthAddress)
			require.Empty(t, item.OwnerAnyAddress)
			require.False(t, isNameAvailable(t, fx.cacheService).Available)
			fx.finish(t)
		}
	})
}

// d: a background refresh that finds a changed registration reads it again later
func TestCacheService_ChangedStateSchedulesRereads(t *testing.T) {
	fx := newFixture(t)
	defer fx.finish(t)
	// a 0.7.1 tombstone
	insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(500),
		"observed_block_hash": blockHash(500).Hex(), "observed_at": time.Now().Add(-time.Hour).UnixMilli()})
	require.Empty(t, cachedItem(t, fx).Rereads)

	start := time.Now()
	fx.setHead(520, time.Now())
	expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
	require.True(t, fx.refreshLeased(ctx, testFullName))
	item := cachedItem(t, fx)
	require.False(t, item.Removed)
	require.Len(t, item.Rereads, 2)
	require.GreaterOrEqual(t, item.Rereads[0], start.Add(5*time.Minute).UnixMilli())
	require.GreaterOrEqual(t, item.Rereads[1], start.Add(30*time.Minute).UnixMilli())
	require.Equal(t, item.Rereads[0], item.RepairAt)

	// an unchanged read schedules nothing new
	later := time.Now().Add(5*time.Minute + time.Second)
	fx.now = func() time.Time { return later }
	expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
	require.Equal(t, 1, repairRound(t, fx))
	require.Len(t, cachedItem(t, fx).Rereads, 1)
}

// e: aliases are never deleted automatically next to a lapsed record or a tombstone: they keep
// the name taken and are reported
func TestCacheService_AliasesNextToTombstonesStay(t *testing.T) {
	t.Run("a lapse leaves a legacy alias, the name stays taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 10, 0)
		insertRaw(t, fx, aliasDoc(testAnyID, 0))

		fx.setHead(300, time.Now())
		expectLapsed(fx.contracts)
		require.NoError(t, fx.refreshName())
		require.True(t, cachedItem(t, fx).Lapsed)
		require.EqualValues(t, 1, aliasCount(t, fx))
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
	})

	t.Run("dedupe keeps a legacy alias next to a tombstone and reports it", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 0))
		insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(10)})
		require.Equal(t, AliasStats{Aliases: 1, Kept: []string{testAlias}, Canon: 2}, migrate(t, fx, true))
		require.EqualValues(t, 1, aliasCount(t, fx))
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
	})
}

// f: the alias check by the canonical spelling (canon) covers what the collation can not
func TestCacheService_CanonField(t *testing.T) {
	const puny = "xn--tst-bma.any" // normalizes to "tést.any"
	canonical, err := (&cacheService{}).canonical(puny)
	require.NoError(t, err)
	require.Equal(t, "tést.any", canonical)
	lookup := func(fx *fixture) *nsp.NameAvailableResponse {
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: canonical})
		require.NoError(t, err)
		return out
	}

	t.Run("a migration that gives the alias its canon between the lookup's miss and the alias check: taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(10), "canon": testFullName})
		insertRaw(t, fx, aliasDoc(testAnyID, 20))
		hookBeforeAliasCheck = func() {
			stats := migrate(t, fx, true)
			require.Equal(t, []string{testAlias}, stats.Kept)
		}
		defer func() { hookBeforeAliasCheck = nil }()
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
		var alias NameDataItem
		require.NoError(t, fx.itemColl.FindOne(ctx, bson.M{"name": testAlias}).Decode(&alias))
		require.Equal(t, testFullName, alias.Canon)
	})

	t.Run("every write stores canon", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.Equal(t, testFullName, cachedItem(t, fx).Canon)
	})

	t.Run("a punycode alias: the collation misses it, migration gives it canon, then it keeps the name taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		// the canonical name was removed at block 10; the alias was read at a later block 20
		insertRaw(t, fx, bson.M{"name": canonical, "removed": true, "observed_block": int64(10), "canon": canonical})
		insertRaw(t, fx, bson.M{"name": puny, "owner_eth_address": testEoa, "name_expires": notExpired, "observed_block": int64(20)})
		// no canon yet: the collation fallback can not see the alias, the tombstone keeps the name
		// taken (without an owner)
		out := lookup(fx)
		require.False(t, out.Available)
		require.Empty(t, out.OwnerEthAddress)

		stats := migrate(t, fx, true)
		require.Equal(t, []string{puny}, stats.Kept)
		out = lookup(fx)
		require.False(t, out.Available)
		require.Equal(t, testEoa, out.OwnerEthAddress)
	})
}

// g: the cap of the re-reads keeps the latest one (the finality follow-up of the newest change)
func TestCacheService_RereadCapKeepsTheLatest(t *testing.T) {
	m := time.Minute.Milliseconds()
	var all []int64
	// polls at minutes 0..7 (each: +5, +30), the worker behind
	for i := int64(0); i < 8; i++ {
		all = mergeRereads(all, []int64{(i + 5) * m, (i + 30) * m}, 0)
	}
	require.Len(t, all, maxRereads)
	require.Equal(t, 5*m, all[0])
	require.Equal(t, 37*m, all[len(all)-1], "a finality follow-up is kept")
	// a later renewal at minute 120
	all = mergeRereads(all, []int64{125 * m, 150 * m}, 0)
	require.Len(t, all, maxRereads)
	require.Equal(t, 150*m, slices.Max(all))

	t.Run("restart: the scan still runs the latest one", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, notExpired, 500, time.Now().UnixMilli())
		start := time.Now()
		for i := 0; i < 8; i++ {
			at := start.Add(time.Duration(i) * time.Minute)
			fx.now = func() time.Time { return at }
			require.NoError(t, fx.scheduleRereads(ctx, testFullName))
		}
		item := cachedItem(t, fx)
		require.Len(t, item.Rereads, maxRereads)
		require.GreaterOrEqual(t, slices.Max(item.Rereads), start.Add(37*time.Minute).UnixMilli())

		// another node (a restart): the early ones hit a stale backend, the latest still runs
		cs, cm := otherCacheService(t, fx, 400, time.Now())
		late := start.Add(38 * time.Minute)
		cs.now = func() time.Time { return late }
		expectRegistered(cm, testScw, testEoa, testAnyID, notExpired)
		n, err := cs.repairOnce(ctx, repairBatch)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.NotEmpty(t, cachedItem(t, fx).Rereads, "a stale read does not do them")
	})
}
