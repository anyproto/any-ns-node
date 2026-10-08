package cache

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
)

func TestCacheService_RefreshAll(t *testing.T) {
	seed := func(t *testing.T, fx *fixture) {
		for _, n := range []struct {
			name    string
			expires int64
		}{{"renewed.any", inGrace}, {"lapsed.any", inGrace}, {"broken.any", inGrace}, {"ghost.any", notExpired}} {
			require.NoError(t, fx.setNameData(ctx, &NameDataItem{
				FullName: n.name, OwnerEthAddress: testEoa, OwnerScwEthAddress: testScw, OwnerAnyAddress: testAnyID,
				NameExpires: n.expires,
			}))
		}
	}

	// one read per name, at the latest block: no finalized read, no enrichment for a lapse
	ghost, err := contracts.NameHash("ghost.any")
	require.NoError(t, err)
	// never registered on chain (first: the other owner reads match any name)
	expectGhost := func(fx *fixture) {
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), ghost, gomock.Any()).Return(common.Address{}, nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "ghost.any", gomock.Any()).Return(big.NewInt(0), nil)
	}
	expectContracts := func(fx *fixture) {
		expectGhost(fx)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(3)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "renewed.any", gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), "renewed.any", gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(testEoa), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "lapsed.any", gomock.Any()).Return(big.NewInt(lapsed), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "broken.any", gomock.Any()).Return(nil, errors.New("rpc is down"))
	}

	get := func(t *testing.T, fx *fixture, name string) *NameDataItem {
		item, err := fx.getNameData(ctx, name)
		require.NoError(t, err)
		return item
	}

	t.Run("dry run: the same decisions, nothing written", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		expectContracts(fx)

		stats, err := fx.RefreshAll(ctx, false, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 4, Updated: 1, Lapsed: 1, NotOnChain: 1, Failed: 1}, stats)

		require.Equal(t, inGrace, get(t, fx, "renewed.any").NameExpires)
		require.False(t, get(t, fx, "lapsed.any").Lapsed)
		require.Zero(t, get(t, fx, "renewed.any").ObservedBlock)
	})

	t.Run("apply: stale records updated, lapsed ones marked (the owner stays), nothing removed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		expectContracts(fx)

		stats, err := fx.RefreshAll(ctx, true, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 4, Updated: 1, Lapsed: 1, NotOnChain: 1, Failed: 1}, stats)

		renewed := get(t, fx, "renewed.any")
		require.Equal(t, notExpired, renewed.NameExpires)
		require.NotZero(t, renewed.ObservedBlock)
		// a changed one: no re-reads (the live nodes would run them on their provider), not due
		require.Empty(t, renewed.Rereads)
		require.Zero(t, renewed.RepairAt)
		l := get(t, fx, "lapsed.any")
		require.True(t, l.Lapsed)
		require.Equal(t, lapsed, l.NameExpires)
		require.Equal(t, testAnyID, l.OwnerAnyAddress)
		require.Zero(t, l.RepairAt)
		// a failed read, a name the chain never had: the records as they were
		require.Equal(t, inGrace, get(t, fx, "broken.any").NameExpires)
		ghost := get(t, fx, "ghost.any")
		require.Equal(t, testAnyID, ghost.OwnerAnyAddress)
		require.Zero(t, ghost.RepairAt)
		n, err := fx.itemColl.CountDocuments(ctx, bson.M{})
		require.NoError(t, err)
		require.EqualValues(t, 4, n)

		// a second run: the lapsed one is kept lapsed, the rest unchanged
		expectGhost(fx)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(3)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "renewed.any", gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), "renewed.any", gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(testEoa), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "lapsed.any", gomock.Any()).Return(big.NewInt(lapsed), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "broken.any", gomock.Any()).Return(big.NewInt(inGrace), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), "broken.any", gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(testEoa), nil)
		stats, err = fx.RefreshAll(ctx, true, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 4, Unchanged: 2, Lapsed: 1, NotOnChain: 1}, stats)
	})

	for _, apply := range []bool{false, true} {
		t.Run(fmt.Sprintf("apply=%v: the owner can not be read: failed", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			seedItem(t, fx, inGrace, 10, 0)
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
			fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))

			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, Failed: 1}, stats)
			require.Equal(t, apply, cachedItem(t, fx).RefreshNeeded)
		})
	}
}

func TestCacheService_VerifyNameIndex(t *testing.T) {
	indexes := func(t *testing.T, fx *fixture) map[string]bool {
		specs, err := fx.itemColl.Indexes().ListSpecifications(ctx)
		require.NoError(t, err)
		out := map[string]bool{}
		for _, s := range specs {
			out[s.Name] = s.Unique != nil && *s.Unique
		}
		return out
	}
	dropNameIndex := func(t *testing.T, fx *fixture) {
		_, err := fx.itemColl.Indexes().DropOne(ctx, "name_1")
		require.NoError(t, err)
	}

	t.Run("the existing unique name_1 (prod) or any other name is accepted, nothing changes", func(t *testing.T) {
		for _, name := range []string{"name_1", "my_name_index"} {
			fx := newFixture(t)
			if name != "name_1" {
				dropNameIndex(t, fx)
				_, err := fx.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true).SetName(name)})
				require.NoError(t, err)
			}
			before := indexes(t, fx)
			for _, apply := range []bool{false, true} {
				stats, err := fx.VerifyNameIndex(ctx, apply)
				require.NoError(t, err)
				require.Equal(t, NameIndexStats{NameIndex: "exists"}, stats)
			}
			require.Equal(t, before, indexes(t, fx))
			fx.finish(t)
		}
	})

	t.Run("a non-unique index is an error, it is never dropped", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		dropNameIndex(t, fx)
		_, err := fx.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "name", Value: 1}}})
		require.NoError(t, err)
		before := indexes(t, fx)

		for _, apply := range []bool{false, true} {
			_, err = fx.VerifyNameIndex(ctx, apply)
			require.ErrorIs(t, err, ErrNameIndexNotUnique)
			require.ErrorContains(t, err, "name_1")
		}
		require.Equal(t, before, indexes(t, fx))
	})

	t.Run("missing: a dry run reports it, apply creates it", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		dropNameIndex(t, fx)

		stats, err := fx.VerifyNameIndex(ctx, false)
		require.NoError(t, err)
		require.Equal(t, NameIndexStats{NameIndex: "missing"}, stats)
		require.NotContains(t, indexes(t, fx), "name_1")

		stats, err = fx.VerifyNameIndex(ctx, true)
		require.NoError(t, err)
		require.Equal(t, "created", stats.NameIndex)
		require.True(t, indexes(t, fx)["name_1"])
	})

	t.Run("missing with duplicates: an error, nothing is removed or created", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		dropNameIndex(t, fx)
		_, err := fx.itemColl.InsertMany(ctx, []interface{}{
			NameDataItem{FullName: testFullName, ObservedBlock: 1}, NameDataItem{FullName: testFullName, ObservedBlock: 2}})
		require.NoError(t, err)

		for _, apply := range []bool{false, true} {
			stats, err := fx.VerifyNameIndex(ctx, apply)
			require.ErrorIs(t, err, ErrDuplicateNames)
			require.Equal(t, NameIndexStats{NameIndex: "missing", DuplicateNames: 1}, stats)
		}
		require.EqualValues(t, 2, countRecords(t, fx))
		require.NotContains(t, indexes(t, fx), "name_1")
	})
}

func TestCacheService_ReleaseName(t *testing.T) {
	seed := func(t *testing.T, fx *fixture) {
		seedItem(t, fx, lapsed, 10, 0)
		insertRaw(t, fx, bson.M{"name": "other.any", "owner_any_address": otherAnyID, "name_expires": notExpired})
		insertRaw(t, fx, aliasDoc(otherAnyID, 0))
	}
	count := func(t *testing.T, fx *fixture) int64 {
		n, err := fx.itemColl.CountDocuments(ctx, bson.M{})
		require.NoError(t, err)
		return n
	}

	t.Run("dry run: the record and the aliases are reported, nothing is deleted", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)

		stats, err := fx.ReleaseName(ctx, "TEST.any", false)
		require.NoError(t, err)
		require.Equal(t, testFullName, stats.Name)
		require.Equal(t, testAnyID, stats.Record.OwnerAnyAddress)
		require.Equal(t, []string{testAlias}, stats.Aliases)
		require.False(t, stats.Deleted)
		require.EqualValues(t, 3, count(t, fx))
	})

	t.Run("apply deletes exactly that record: the name can be registered again", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)

		stats, err := fx.ReleaseName(ctx, testFullName, true)
		require.NoError(t, err)
		require.True(t, stats.Deleted)
		require.Nil(t, cachedItem(t, fx))
		require.EqualValues(t, 2, count(t, fx))
		require.EqualValues(t, 1, aliasCount(t, fx), "an alias is never deleted")
		// (the alias still keeps the name taken: for the operator)
		require.False(t, isNameAvailable(t, fx.cacheService).Available)

		require.NoError(t, fx.itemColl.Database().Collection(fx.itemColl.Name()).FindOneAndDelete(ctx, bson.M{"name": testAlias}).Err())
		require.True(t, isNameAvailable(t, fx.cacheService).Available)
	})

	t.Run("not cached: ErrNameNotCached, nothing deleted", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		stats, err := fx.ReleaseName(ctx, "nobody.any", true)
		require.ErrorIs(t, err, ErrNameNotCached)
		require.Nil(t, stats.Record)
		require.EqualValues(t, 3, count(t, fx))
	})
}

func TestCacheService_RestoreTombstones(t *testing.T) {
	// a mongoexport of the cache before 0.7.1 wiped the tombstones: canonical and relaxed
	// Extended JSON, an owner in another case, a record without an owner, a tombstone already
	export := strings.Join([]string{
		`{"_id":{"$oid":"6700000000000000000000a1"},"name":"a.any","owner_eth_address":"0xAA","owner_scw_eth_address":"0xBB","owner_any_address":"A-owner","space_id":"space-a","name_expires":{"$numberLong":"` + fmt.Sprint(lapsed) + `"}}`,
		`{"_id":{"$oid":"6700000000000000000000a2"},"name":"b.any","owner_eth_address":"0xcc","owner_scw_eth_address":"","owner_any_address":"B-owner","space_id":"","name_expires":` + fmt.Sprint(notExpired) + `}`,
		``,
		`{"_id":{"$oid":"6700000000000000000000a3"},"name":"c.any","owner_eth_address":"","owner_any_address":"","name_expires":{"$numberLong":"0"}}`,
		`{"_id":{"$oid":"6700000000000000000000a4"},"name":"d.any","removed":true,"observed_block":{"$numberLong":"5"}}`,
		`{"_id":{"$oid":"6700000000000000000000a5"},"name":"e.any","owner_eth_address":"0xee","owner_any_address":"E-owner","name_expires":{"$numberLong":"` + fmt.Sprint(lapsed) + `"}}`,
	}, "\n")
	seed := func(t *testing.T, fx *fixture) {
		for _, name := range []string{"a.any", "b.any", "c.any", "d.any", "missing.any"} {
			insertRaw(t, fx, bson.M{"name": name, "removed": true, "observed_block": int64(700),
				"observed_block_hash": blockHash(700).Hex(), "owner_eth_address": "", "owner_any_address": "", "name_expires": int64(0)})
		}
		// not a tombstone (any more): never touched
		insertRaw(t, fx, bson.M{"name": "e.any", "owner_any_address": "E-new", "name_expires": notExpired, "observed_block": int64(800)})
	}
	get := func(t *testing.T, fx *fixture, name string) *NameDataItem {
		item, err := fx.getNameData(ctx, name)
		require.NoError(t, err)
		return item
	}
	wantMissing := []string{"c.any", "d.any", "missing.any"}

	t.Run("dry run: counts, writes nothing", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		var before []bson.M
		cur, err := fx.itemColl.Find(ctx, bson.M{})
		require.NoError(t, err)
		require.NoError(t, cur.All(ctx, &before))

		stats, err := fx.RestoreTombstones(ctx, strings.NewReader(export), false)
		require.NoError(t, err)
		require.Equal(t, RestoreStats{Tombstones: 5, Restored: 2, Missing: wantMissing}, stats)

		var after []bson.M
		cur, err = fx.itemColl.Find(ctx, bson.M{})
		require.NoError(t, err)
		require.NoError(t, cur.All(ctx, &after))
		require.Equal(t, before, after)
	})

	t.Run("apply: owners restored (lapsed or marked for one read), missing ones listed and left", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)

		stats, err := fx.RestoreTombstones(ctx, strings.NewReader(export), true)
		require.NoError(t, err)
		require.Equal(t, RestoreStats{Tombstones: 5, Restored: 2, Missing: wantMissing}, stats)

		a := get(t, fx, "a.any")
		require.False(t, a.Removed)
		require.True(t, a.Lapsed)
		require.False(t, a.RefreshNeeded)
		require.Equal(t, "0xaa", a.OwnerEthAddress)
		require.Equal(t, "0xbb", a.OwnerScwEthAddress)
		require.Equal(t, "A-owner", a.OwnerAnyAddress)
		require.Equal(t, "space-a", a.SpaceId)
		require.Equal(t, lapsed, a.NameExpires)
		require.Equal(t, "a.any", a.Canon)
		require.Equal(t, int64(700), a.ObservedBlock, "the block of the tombstone stays")
		require.Zero(t, a.RepairAt, "lapsed: no chain read")
		require.True(t, byAnyID(t, fx, "A-owner").Found)

		// not lapsed: renewed since? the background reads it once
		b := get(t, fx, "b.any")
		require.False(t, b.Removed)
		require.False(t, b.Lapsed)
		require.True(t, b.RefreshNeeded)
		require.Equal(t, int64(1), b.RepairAt)
		require.Equal(t, "B-owner", b.OwnerAnyAddress)
		require.Equal(t, notExpired, b.NameExpires)

		for _, name := range []string{"c.any", "d.any", "missing.any"} {
			item := get(t, fx, name)
			require.True(t, item.Removed, name)
			require.Empty(t, item.OwnerAnyAddress, name)
		}
		e := get(t, fx, "e.any")
		require.Equal(t, "E-new", e.OwnerAnyAddress)
		require.Equal(t, notExpired, e.NameExpires)

		// a second run: nothing left to restore
		stats, err = fx.RestoreTombstones(ctx, strings.NewReader(export), true)
		require.NoError(t, err)
		require.Equal(t, RestoreStats{Tombstones: 3, Missing: wantMissing}, stats)
	})

	t.Run("a record that is not removed any more when it is written is not touched", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		// a registration read replaces the tombstone between the scan of the tombstones and the write
		hookBeforeRestore = func(name string) {
			if name == "a.any" {
				_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": "a.any"}, bson.M{
					"$set": bson.M{"owner_any_address": "A-chain"}, "$unset": bson.M{"removed": ""}})
				require.NoError(t, err)
			}
		}
		defer func() { hookBeforeRestore = nil }()

		stats, err := fx.RestoreTombstones(ctx, strings.NewReader(export), true)
		require.NoError(t, err)
		require.Equal(t, RestoreStats{Tombstones: 5, Restored: 1, Missing: wantMissing}, stats)
		a := get(t, fx, "a.any")
		require.Equal(t, "A-chain", a.OwnerAnyAddress)
		require.False(t, a.Lapsed)
	})

	t.Run("an export that can not be parsed: an error, nothing written", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		_, err := fx.RestoreTombstones(ctx, strings.NewReader(export+"\n{not json"), true)
		require.Error(t, err)
		require.True(t, get(t, fx, "a.any").Removed)
	})
}

// a maintenance run changes no index at start (a dry -dedupe-cache must not create one) and runs
// no background refresh
func TestNewMaintainer(t *testing.T) {
	const dbName = "any-ns-test-cache-maintainer"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI()))
	require.NoError(t, err)
	defer func() { _ = client.Disconnect(ctx) }()
	require.NoError(t, client.Database(dbName).Drop(ctx))
	coll := client.Database(dbName).Collection("cache")
	_, err = coll.InsertOne(ctx, NameDataItem{FullName: testFullName})
	require.NoError(t, err)

	conf := new(config.Config)
	conf.Mongo = config.Mongo{Connect: testMongoURI(), Database: dbName}
	conf.Cache = config.Cache{AllowUnsafeStandalone: true}
	m := mock_contracts.NewMockContractsService(gomock.NewController(t))
	m.EXPECT().Name().Return(contracts.CName).AnyTimes()
	m.EXPECT().Init(gomock.Any()).AnyTimes()

	a := new(app.App)
	a.Register(rpctest.NewTestServer()).Register(conf).Register(m).Register(NewMaintainer())
	require.NoError(t, a.Start(ctx))
	defer func() { _ = a.Close(ctx) }()
	cs := a.MustComponent(CName).(*cacheService)
	require.Nil(t, cs.stopWorker)

	stats, err := a.MustComponent(CName).(Maintainer).VerifyNameIndex(ctx, false)
	require.NoError(t, err)
	require.Equal(t, "missing", stats.NameIndex)
	specs, err := coll.Indexes().ListSpecifications(ctx)
	require.NoError(t, err)
	require.Len(t, specs, 1, "only _id")
	require.NoError(t, client.Database(dbName).Drop(ctx))
}
