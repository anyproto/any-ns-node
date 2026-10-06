package cache

import (
	"errors"
	"fmt"
	"math/big"
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
		}{{"renewed.any", inGrace}, {"lapsed.any", lapsed}, {"broken.any", inGrace}} {
			require.NoError(t, fx.setNameData(ctx, &NameDataItem{
				FullName: n.name, OwnerEthAddress: testEoa, OwnerScwEthAddress: testScw, OwnerAnyAddress: testAnyID,
				NameExpires: n.expires,
			}))
		}
	}

	expectContracts := func(fx *fixture) {
		// the lapsed name is read again at the finalized block before it is removed (or before
		// the dry run says it would be)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(4)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "renewed.any", gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), "renewed.any", gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(testEoa), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), "lapsed.any", gomock.Any()).Return(big.NewInt(lapsed), nil).Times(2)
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
		require.Equal(t, RefreshStats{Total: 3, Updated: 1, Removed: 1, Failed: 1}, stats)

		require.Equal(t, inGrace, get(t, fx, "renewed.any").NameExpires)
		require.False(t, get(t, fx, "lapsed.any").Removed)
		require.Zero(t, get(t, fx, "renewed.any").ObservedBlock)
	})

	t.Run("apply: stale records updated, lapsed ones become tombstones", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		expectContracts(fx)

		stats, err := fx.RefreshAll(ctx, true, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, RefreshStats{Total: 3, Updated: 1, Removed: 1, Failed: 1}, stats)

		renewed := get(t, fx, "renewed.any")
		require.Equal(t, notExpired, renewed.NameExpires)
		require.NotZero(t, renewed.ObservedBlock)
		require.Equal(t, notExpired*1000, renewed.RepairAt, "the backfill schedules the scan of legacy records")
		require.True(t, get(t, fx, "lapsed.any").Removed)
		// a failed read leaves the record as it was
		require.Equal(t, inGrace, get(t, fx, "broken.any").NameExpires)

		// a second run: the tombstone is unchanged (no finalized read for it)
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
		require.Equal(t, RefreshStats{Total: 3, Unchanged: 3}, stats)
	})

	for _, apply := range []bool{false, true} {
		t.Run(fmt.Sprintf("apply=%v: lapsed at the latest block, registered at the finalized one: not final, kept", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			seedItem(t, fx, inGrace, 10, 0)

			fx.setHead(300, time.Now())
			fx.setFinalized(290, time.Now())
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(2)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(300)).Return(big.NewInt(lapsed), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(290)).Return(big.NewInt(inGrace), nil)

			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, NotFinal: 1}, stats)
			require.False(t, cachedItem(t, fx).Removed)
		})

		t.Run(fmt.Sprintf("apply=%v: a not-final legacy record is retried by the periodic scan (apply only)", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			// a legacy record: no repair_at
			_, err := fx.itemColl.InsertOne(ctx, bson.M{"name": testFullName, "owner_eth_address": testEoa, "name_expires": lapsed})
			require.NoError(t, err)

			fx.setHead(300, time.Now())
			fx.setFinalized(290, time.Now())
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), blockHash(300)).Return(common.Address{}, nil)
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), blockHash(290)).Return(common.HexToAddress(nameWrapper), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(290)).Return(big.NewInt(notExpired), nil)
			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, NotFinal: 1}, stats)
			item := cachedItem(t, fx)
			require.Equal(t, testEoa, item.OwnerEthAddress, "the data stays")
			if !apply {
				require.Zero(t, item.RepairAt)
				return
			}
			require.Greater(t, item.RepairAt, time.Now().UnixMilli())
			require.Equal(t, item.RefreshNextAt, item.RepairAt, "backing off: no lookup or scan takes it before")
			require.Zero(t, repairRound(t, fx))

			// finality catches up: the scan (no lookup) removes it after the backoff
			later := time.Now().Add(refreshFailureBackoff + time.Second)
			fx.now = func() time.Time { return later }
			fx.setHead(320, time.Now())
			fx.setFinalized(310, time.Now())
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)
			require.Equal(t, 1, repairRound(t, fx))
			require.True(t, cachedItem(t, fx).Removed)
		})

		t.Run(fmt.Sprintf("apply=%v: a not-final record that was already overdue backs off too", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			seedItem(t, fx, inGrace, 10, 0)
			_, err := fx.itemColl.UpdateOne(ctx, bson.M{"name": testFullName}, bson.M{"$set": bson.M{"repair_at": int64(1)}})
			require.NoError(t, err)

			fx.setHead(300, time.Now())
			fx.setFinalized(290, time.Now())
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(2)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(300)).Return(big.NewInt(lapsed), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, blockHash(290)).Return(big.NewInt(inGrace), nil)
			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, NotFinal: 1}, stats)

			item := cachedItem(t, fx)
			if !apply {
				require.Equal(t, int64(1), item.RepairAt)
				return
			}
			require.Greater(t, item.RefreshNextAt, time.Now().Add(refreshFailureBackoff-10*time.Second).UnixMilli())
			require.Equal(t, item.RefreshNextAt, item.RepairAt)
			require.Zero(t, repairRound(t, fx), "not before the backoff")
			require.False(t, needsRefresh(item, time.Now()), "nor a lookup")
		})

		t.Run(fmt.Sprintf("apply=%v: lapsed at the latest block, the finalized block can not be read: failed, not not-final", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			seedItem(t, fx, inGrace, 10, 0)

			fx.setFinalizedErr(errors.New("unknown block tag"))
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
			fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(lapsed), nil)

			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, Failed: 1}, stats)
			require.False(t, cachedItem(t, fx).Removed)
		})

		t.Run(fmt.Sprintf("apply=%v: a lagging provider: a removal older than the cached registration is not final", apply), func(t *testing.T) {
			fx := newFixture(t)
			defer fx.finish(t)
			seedItem(t, fx, notExpired, 200, 0)
			fx.setHead(150, time.Now())
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).Times(2)

			stats, err := fx.RefreshAll(ctx, apply, time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, RefreshStats{Total: 1, NotFinal: 1}, stats)
			require.Equal(t, int64(200), cachedItem(t, fx).ObservedBlock)
		})

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

func TestCacheService_PurgeTombstones(t *testing.T) {
	seed := func(t *testing.T, fx *fixture) {
		_, err := fx.itemColl.InsertMany(ctx, []interface{}{
			NameDataItem{FullName: "a.any", Removed: true, ObservedBlock: 10},
			NameDataItem{FullName: "b.any", Removed: true, ObservedBlock: 11},
			NameDataItem{FullName: "c.any", OwnerEthAddress: testEoa, NameExpires: notExpired, ObservedBlock: 12},
			// marked after a failed refresh, its data is complete: an old node serves it fine
			NameDataItem{FullName: "d.any", OwnerEthAddress: testEoa, NameExpires: notExpired, RefreshNeeded: true},
		})
		require.NoError(t, err)
	}
	count := func(t *testing.T, fx *fixture, filter bson.M) int64 {
		n, err := fx.itemColl.CountDocuments(ctx, filter)
		require.NoError(t, err)
		return n
	}

	t.Run("dry run counts, apply deletes the tombstones only", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)

		stats, err := fx.PurgeTombstones(ctx, false)
		require.NoError(t, err)
		require.Equal(t, PurgeStats{Tombstones: 2}, stats)
		require.EqualValues(t, 4, count(t, fx, bson.M{}))

		stats, err = fx.PurgeTombstones(ctx, true)
		require.NoError(t, err)
		require.Equal(t, PurgeStats{Tombstones: 2}, stats)
		require.EqualValues(t, 2, count(t, fx, bson.M{}))
		require.Zero(t, count(t, fx, bson.M{"removed": true}))
	})

	t.Run("refused while a record has no owner (an old node would serve it so forever)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seed(t, fx)
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{FullName: "e.any", OwnerScwEthAddress: testScw, NameExpires: notExpired, RefreshNeeded: true})
		require.NoError(t, err)

		for _, apply := range []bool{false, true} {
			stats, err := fx.PurgeTombstones(ctx, apply)
			require.ErrorIs(t, err, ErrIncompleteRecords)
			require.Equal(t, []string{"e.any"}, stats.Incomplete)
		}
		require.EqualValues(t, 2, count(t, fx, bson.M{"removed": true}))
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
