package anynsrpc

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync/atomic"
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
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
)

const (
	realCacheDb = "any-ns-test-rpc-cache"
	// the real cache's Mongo client, for the fail points
	realCacheApp = "go7567-rpc"

	nameWrapper = "0x0000000000000000000000000000000000000abc"
	testEoa     = "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5"
	testScw     = "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51"
	testAnyID   = "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
)

func testMongoURI() string {
	// ANY_NS_TEST_MONGO: see testMongoURI in the cache package
	if uri := os.Getenv("ANY_NS_TEST_MONGO"); uri != "" {
		return uri
	}
	return "mongodb://localhost:27017"
}

func withAppName(uri, app string) string {
	if strings.Contains(uri, "?") {
		return uri + "&appName=" + app
	}
	return uri + "/?appName=" + app
}

// a real cache (local test Mongo, its client named realCacheApp) on top of mocked contracts: a
// new block on every read of the latest one. the background worker runs (the periodic scan
// does not). a separate database, so that it does not
// race with the other packages' tests
func newRealCache(t *testing.T) (cache.CacheService, *mock_contracts.MockContractsService, *mongo.Collection) {
	ctrl := gomock.NewController(t)

	conf := new(config.Config)
	conf.Mongo = config.Mongo{Connect: withAppName(testMongoURI(), realCacheApp), Database: realCacheDb}
	conf.Cache = config.Cache{AllowUnsafeStandalone: true, RepairIntervalSec: -1}

	cm := mock_contracts.NewMockContractsService(ctrl)
	cm.EXPECT().Name().Return(contracts.CName).AnyTimes()
	cm.EXPECT().Init(gomock.Any()).AnyTimes()
	var block atomic.Int64
	cm.EXPECT().LatestBlock(gomock.Any()).DoAndReturn(func(context.Context) (*contracts.Block, error) {
		n := 1000 + block.Add(1)
		return &contracts.Block{Number: n, Hash: common.BigToHash(big.NewInt(n)), Time: uint64(time.Now().Unix())}, nil
	}).AnyTimes()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI()))
	require.NoError(t, err)
	require.NoError(t, client.Database(realCacheDb).Drop(ctx))

	a := new(app.App)
	cs := cache.New()
	a.Register(rpctest.NewTestServer()).Register(conf).Register(cm).Register(cs)
	require.NoError(t, a.Start(ctx))
	t.Cleanup(func() {
		_ = a.Close(ctx)
		_ = client.Disconnect(ctx)
	})

	return cs.(cache.CacheService), cm, client.Database(realCacheDb).Collection("cache")
}

// failCommand sets the failCommand fail point of the test replica set (enableTestCommands),
// skips the test if there is none
func failCommand(t *testing.T, data bson.M, mode bson.M) {
	t.Helper()
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
		t.Skipf("SKIPPED %s: needs the failCommand fail point (a test replica set with enableTestCommands, see the README)", t.Name())
	}
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = admin.RunCommand(ctx, bson.D{{Key: "configureFailPoint", Value: "failCommand"}, {Key: "mode", Value: "off"}}).Err()
	})
}

func TestAnynsRpc_IsNameAvailable_LapsedName(t *testing.T) {
	const fullName = "hello.any"
	// expired and past the 90 days grace period
	lapsed := time.Now().Add(-100 * 24 * time.Hour).Unix()

	for _, readFromCache := range []bool{true, false} {
		t.Run(fmt.Sprintf("readFromCache=%v: a lapsed name stays taken by its owner, no background refresh", readFromCache), func(t *testing.T) {
			fx := newFixture(t, readFromCache)
			defer fx.finish(t)

			cs, cm, coll := newRealCache(t)
			fx.anynsRpc.cache = cs

			_, err := coll.InsertOne(ctx, cache.NameDataItem{
				FullName: fullName, OwnerEthAddress: testEoa, OwnerScwEthAddress: testScw, OwnerAnyAddress: testAnyID,
				NameExpires: lapsed,
			})
			require.NoError(t, err)

			// readFromCache: false reads the registry at the latest block, once per request (the
			// lapse is stored on the record). nothing else reads the chain
			if !readFromCache {
				cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(2)
				cm.EXPECT().GetNameExpires(gomock.Any(), fullName, gomock.Any()).Return(big.NewInt(lapsed), nil).Times(2)
			}
			for i := 0; i < 2; i++ {
				resp, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: fullName})
				require.NoError(t, err)
				require.False(t, resp.Available)
				require.Equal(t, testEoa, resp.OwnerEthAddress)
				require.Equal(t, testAnyID, resp.OwnerAnyAddress)
				require.Equal(t, lapsed, resp.NameExpires)
			}
			// the background had the time to refresh it, if anything had asked it to
			time.Sleep(200 * time.Millisecond)

			var item cache.NameDataItem
			require.NoError(t, coll.FindOne(ctx, bson.M{"name": fullName}).Decode(&item))
			require.False(t, item.Removed)
			require.Equal(t, !readFromCache, item.Lapsed)
			require.Equal(t, testAnyID, item.OwnerAnyAddress)
			byID, err := fx.GetNameByAnyId(ctx, &nsp.NameByAnyIdRequest{AnyAddress: testAnyID})
			require.NoError(t, err)
			require.True(t, byID.Found)
		})
	}
}

// readFromCache: a batch is answered from Mongo whatever the contracts and the Mongo writes do
// (here: every contract read stalls until its context ends, every write of the cache's client
// stalls). the refreshes run in the background
func TestAnynsRpc_Batch_NeverBlocks(t *testing.T) {
	for _, stalledWrites := range []bool{false, true} {
		t.Run(fmt.Sprintf("stalled writes: %v", stalledWrites), func(t *testing.T) {
			fx := newFixture(t, true)
			defer fx.finish(t)

			cs, cm, coll := newRealCache(t)
			fx.anynsRpc.cache = cs

			inGrace := time.Now().Add(-30 * 24 * time.Hour).Unix()
			var names, ids []string
			for i := 0; i < 20; i++ {
				name := fmt.Sprintf("expired%02d.any", i)
				names = append(names, name)
				ids = append(ids, fmt.Sprintf("any-id-%02d", i))
				_, err := coll.InsertOne(ctx, cache.NameDataItem{FullName: name, OwnerEthAddress: testEoa,
					OwnerAnyAddress: ids[i], NameExpires: inGrace})
				require.NoError(t, err)
			}
			if stalledWrites {
				failCommand(t, bson.M{
					"failCommands":    []string{"update", "insert", "findAndModify", "delete", "commitTransaction"},
					"appName":         realCacheApp,
					"blockConnection": true,
					"blockTimeMS":     5000,
				}, bson.M{"times": 1000})
			}
			cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, _ [32]byte, _ common.Hash) (common.Address, error) {
					<-ctx.Done()
					return common.Address{}, ctx.Err()
				}).AnyTimes()

			start := time.Now()
			resp, err := fx.BatchIsNameAvailable(ctx, &nsp.BatchNameAvailableRequest{FullNames: names})
			require.NoError(t, err)
			byID, err := fx.BatchGetNameByAnyId(ctx, &nsp.BatchNameByAnyIdRequest{AnyAddresses: ids})
			require.NoError(t, err)
			require.Less(t, time.Since(start), 2*time.Second)

			for i, r := range resp.Results {
				require.False(t, r.Available, names[i])
				require.Equal(t, inGrace, r.NameExpires, names[i])
				require.True(t, byID.Results[i].Found, ids[i])
				require.Equal(t, names[i], byID.Results[i].Name)
			}
		})
	}
}
