package cache

import (
	"testing"
	"time"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
)

// an alias: testFullName ("test.any") cached as "Test.any" by an old node
const testAlias = "Test.any"

func insertRaw(t *testing.T, fx *fixture, doc bson.M) {
	t.Helper()
	_, err := fx.itemColl.InsertOne(ctx, doc)
	require.NoError(t, err)
}

func aliasDoc(anyID string, block int64) bson.M {
	return bson.M{"name": testAlias, "owner_eth_address": testEoa, "owner_scw_eth_address": testScw,
		"owner_any_address": anyID, "name_expires": notExpired, "observed_block": block}
}

func canonicalDoc(anyID string, block int64) bson.M {
	return bson.M{"name": testFullName, "owner_eth_address": otherEoa, "owner_scw_eth_address": otherScw,
		"owner_any_address": anyID, "name_expires": notExpired, "observed_block": block}
}

func byAnyID(t *testing.T, fx *fixture, anyID string) *nsp.NameByAddressResponse {
	t.Helper()
	res, err := fx.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{AnyAddress: anyID})
	require.NoError(t, err)
	return res
}

func migrate(t *testing.T, fx *fixture, apply bool) AliasStats {
	t.Helper()
	stats, err := fx.MigrateAliases(ctx, apply)
	require.NoError(t, err)
	return stats
}

func aliasCount(t *testing.T, fx *fixture) int64 {
	t.Helper()
	n, err := fx.itemColl.CountDocuments(ctx, bson.M{"name": testAlias})
	require.NoError(t, err)
	return n
}

func TestCacheService_Aliases(t *testing.T) {
	t.Run("alias only: the name stays taken, reverse lookups answer the canonical spelling; migration renames it", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 0))
		fx.loadAliases(ctx)
		require.Equal(t, map[string][]string{testFullName: {testAlias}}, fx.aliases.byCanonical)

		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
		require.Equal(t, &nsp.NameByAddressResponse{Found: true, Name: testFullName}, byAnyID(t, fx, testAnyID))

		// dry run: decided, nothing changed
		require.Equal(t, AliasStats{Aliases: 1, Renamed: 1}, migrate(t, fx, false))
		require.EqualValues(t, 1, aliasCount(t, fx))

		require.Equal(t, AliasStats{Aliases: 1, Renamed: 1}, migrate(t, fx, true))
		require.Zero(t, aliasCount(t, fx))
		item := cachedItem(t, fx)
		require.Equal(t, testEoa, item.OwnerEthAddress, "the data is kept")
		require.True(t, item.RefreshNeeded)
		require.Equal(t, int64(1), item.RepairAt)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)

		// the scan refreshes it under its canonical name
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired+1)
		require.Equal(t, 1, repairRound(t, fx))
		require.False(t, cachedItem(t, fx).RefreshNeeded)
	})

	t.Run("alias and an older canonical tombstone: taken until an operator decides, migration keeps both", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 20))
		insertRaw(t, fx, bson.M{"name": testFullName, "removed": true, "observed_block": 10})
		fx.loadAliases(ctx)

		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
		require.Equal(t, AliasStats{Aliases: 1, Kept: []string{testAlias}}, migrate(t, fx, true))
		require.EqualValues(t, 1, aliasCount(t, fx))
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
	})

	t.Run("both spellings, different owners: reverse lookups ignore the alias, migration deletes it", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 5))
		insertRaw(t, fx, canonicalDoc(otherAnyID, 5))
		fx.loadAliases(ctx)

		// the alias' old owner does not resolve, the canonical owner does
		require.False(t, byAnyID(t, fx, testAnyID).Found)
		require.Equal(t, &nsp.NameByAddressResponse{Found: true, Name: testFullName}, byAnyID(t, fx, otherAnyID))
		require.Equal(t, otherEoa, isNameAvailable(t, fx.cacheService).OwnerEthAddress)

		require.Equal(t, AliasStats{Aliases: 1, Deleted: 1}, migrate(t, fx, true))
		require.Zero(t, aliasCount(t, fx))
		require.Equal(t, &nsp.NameByAddressResponse{Found: true, Name: testFullName}, byAnyID(t, fx, otherAnyID))
	})

	t.Run("both spellings, the same owner: the canonical name is answered", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		insertRaw(t, fx, aliasDoc(testAnyID, 5))
		insertRaw(t, fx, canonicalDoc(testAnyID, 5))
		for i := 0; i < 5; i++ {
			require.Equal(t, &nsp.NameByAddressResponse{Found: true, Name: testFullName}, byAnyID(t, fx, testAnyID))
		}
	})

	t.Run("migration keeps the alias when the canonical record is incomplete or older", func(t *testing.T) {
		for _, canonical := range []bson.M{
			{"name": testFullName, "observed_block": int64(10), "refresh_needed": true, "name_expires": notExpired},
			{"name": testFullName, "observed_block": int64(4), "owner_eth_address": otherEoa, "name_expires": notExpired},
		} {
			fx := newFixture(t)
			insertRaw(t, fx, aliasDoc(testAnyID, 5))
			insertRaw(t, fx, canonical)
			require.Equal(t, AliasStats{Aliases: 1, Kept: []string{testAlias}}, migrate(t, fx, true))
			require.EqualValues(t, 1, aliasCount(t, fx))
			fx.finish(t)
		}
	})

	t.Run("the worker refreshes the canonical name and stops reselecting the alias", func(t *testing.T) {
		for _, registered := range []bool{true, false} {
			fx := newFixture(t)
			doc := aliasDoc(testAnyID, 0)
			doc["repair_at"] = int64(1)
			insertRaw(t, fx, doc)
			fx.loadAliases(ctx)

			if registered {
				expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)
			} else {
				fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil)
			}
			require.Equal(t, 1, repairRound(t, fx))
			// (no more contract expectations: another refresh would fail the test)
			require.Zero(t, repairRound(t, fx))
			later := time.Now().Add(time.Hour)
			fx.now = func() time.Time { return later }
			require.Zero(t, repairRound(t, fx))

			var alias NameDataItem
			require.NoError(t, fx.itemColl.FindOne(ctx, bson.M{"name": testAlias}).Decode(&alias))
			require.Zero(t, alias.RepairAt)
			require.Equal(t, registered, cachedItem(t, fx) != nil)
			// a lookup does not hand it to the background again right away
			fx.now = time.Now
			requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), notExpired)
			require.Zero(t, fx.runQueued())
			fx.finish(t)
		}
	})
}
