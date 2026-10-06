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
		require.NoError(t, fx.updateConfirmed())
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
	// a renewal (a year) that was reorged away: cached, its scan at the expiry
	expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
	require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
	_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$unset": bson.M{"rereads": ""}, "$set": bson.M{"repair_at": notExpired * 1000}})
	require.NoError(t, err)

	// the background: no owner at the latest block, the finalized block is behind: not final
	fx.setHead(2000, time.Now())
	fx.setFinalized(1000, time.Now())
	fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), blockHash(2000)).Return(common.Address{}, nil)
	fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(2000)).Return(big.NewInt(0), nil)
	fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), blockHash(1000)).Return(common.HexToAddress(nameWrapper), nil)
	fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(1000)).Return(big.NewInt(notExpired), nil)
	start := time.Now()
	require.True(t, fx.refreshLeased(ctx, testFullName))

	item := cachedItem(t, fx)
	require.False(t, item.Removed)
	require.GreaterOrEqual(t, item.RepairAt, start.Add(refreshFailureBackoff).UnixMilli())
	require.LessOrEqual(t, item.RepairAt, time.Now().Add(refreshLease).UnixMilli(), "not at the cached expiry")
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

	t.Run("a newer block, the enrichment fails, the same registry owner: the fields are carried over", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		infoFails(fx, nameWrapper)
		require.ErrorIs(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}), ErrNameDataIncomplete)
		item := cachedItem(t, fx)
		require.True(t, item.Incomplete)
		require.True(t, item.RefreshNeeded)
		require.Equal(t, testEoa, item.OwnerEthAddress)
		require.Equal(t, testScw, item.OwnerScwEthAddress)
		require.Equal(t, testAnyID, item.OwnerAnyAddress)
		require.True(t, byAnyID(t, fx, testAnyID).Found)
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
	_, err := fx.applyObservation(ctx, &NameDataItem{FullName: testFullName, Removed: true,
		ObservedBlock: 500, ObservedBlockHash: blockHash(500).Hex(), ObservedAt: time.Now().Add(-time.Hour).UnixMilli()}, refreshOpts{})
	require.NoError(t, err)
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

// e: a removal confirmed at a finalized block retires the legacy aliases of the name
func TestCacheService_RemovalRetiresLegacyAliases(t *testing.T) {
	t.Run("in the refresh transaction: a legacy alias goes, an alias with a newer block stays", func(t *testing.T) {
		for _, aliasBlock := range []int64{0, 5000} {
			fx := newFixture(t)
			seedItem(t, fx, lapsed, 10, 0)
			insertRaw(t, fx, aliasDoc(testAnyID, aliasBlock))

			fx.setHead(300, time.Now())
			fx.setFinalized(290, time.Now())
			expectLapsed(fx.contracts)
			require.ErrorIs(t, fx.updateConfirmed(), ErrNameNotRegistered)
			require.True(t, cachedItem(t, fx).Removed)
			if aliasBlock == 0 {
				require.Zero(t, aliasCount(t, fx))
				require.True(t, isNameAvailable(t, fx.cacheService).Available)
			} else {
				require.EqualValues(t, 1, aliasCount(t, fx))
				require.False(t, isNameAvailable(t, fx.cacheService).Available)
			}
			fx.finish(t)
		}
	})

	t.Run("in the migration: a canonical tombstone retires a legacy alias", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 0))
		insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": int64(10)})
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
		require.Equal(t, AliasStats{Aliases: 1, Deleted: 1, Canon: 1}, migrate(t, fx, true))
		require.True(t, isNameAvailable(t, fx.cacheService).Available)
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
		require.True(t, lookup(fx).Available, "no canon yet: the collation fallback can not see it")

		stats := migrate(t, fx, true)
		require.Equal(t, []string{puny}, stats.Kept)
		out := lookup(fx)
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
