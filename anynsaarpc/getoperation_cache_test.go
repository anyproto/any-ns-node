package anynsaarpc

import (
	"context"
	"errors"
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

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	db_service "github.com/anyproto/any-ns-node/db"
)

const (
	realCacheDb = "any-ns-test-aarpc-cache"
	// the real cache's Mongo client, for the fail points
	realCacheApp = "go7567-aarpc"

	opName      = "hello.any"
	nameWrapper = "0x0000000000000000000000000000000000000abc"
	oldScw      = "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51"
	oldEoa      = "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5"
	oldAnyID    = "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
	newScw      = "0xaab27b150451726ec7738aa1d0a94505c8729bd1"
	newEoa      = "0xe595e2ba3f0ce990d8037e07250c5c78ce40f8ff"
	newAnyID    = "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio"

	watchdog = 10 * time.Second
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
// new block on every read of the latest one. the background worker runs (the periodic scan does
// not). a separate database, so that it does not race with the other packages' tests
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
	cm.EXPECT().FinalizedBlock(gomock.Any()).DoAndReturn(func(context.Context) (*contracts.Block, error) {
		n := 1000 + block.Load()
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

func completedOp(fx *fixture) {
	fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).Return(&accountabstraction.OperationInfo{
		OperationState: nsp.OperationState_Completed,
	}, nil)
	fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).Return(db_service.AAUserOperation{
		OperationID: "123",
		FullName:    opName,
	}, nil)
}

func getOperation(t *testing.T, fx *fixture) (*nsp.OperationResponse, error) {
	var (
		resp *nsp.OperationResponse
		err  error
	)
	done := make(chan struct{})
	go func() {
		defer close(done)
		resp, err = fx.GetOperation(ctx, &nsp.GetOperationStatusRequest{OperationId: "123"})
	}()
	select {
	case <-done:
	case <-time.After(watchdog):
		t.Fatal("GetOperation did not return")
	}
	return resp, err
}

func cachedItem(t *testing.T, coll *mongo.Collection) *cache.NameDataItem {
	t.Helper()
	var item cache.NameDataItem
	err := coll.FindOne(ctx, bson.M{"name": opName}).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	require.NoError(t, err)
	return &item
}

// eventually waits until the background (the worker) made cond true
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(watchdog)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("%s: not within %s", what, watchdog)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func expectRegistered(cm *mock_contracts.MockContractsService, scw, eoa, anyID string, expires int64) {
	cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
	cm.EXPECT().GetNameExpires(gomock.Any(), opName, gomock.Any()).Return(big.NewInt(expires), nil)
	cm.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), opName, gomock.Any()).Return(scw, anyID, "", nil)
	cm.EXPECT().GetScwOwner(gomock.Any(), common.HexToAddress(scw), gomock.Any()).Return(common.HexToAddress(eoa), nil)
}

// the re-reads scheduled by a completed operation: about now + rereadDelays (5 and 30 minutes)
func requireRereads(t *testing.T, item *cache.NameDataItem, since time.Time) {
	t.Helper()
	require.Len(t, item.Rereads, 2)
	for i, d := range []time.Duration{5 * time.Minute, 30 * time.Minute} {
		require.GreaterOrEqual(t, item.Rereads[i], since.Add(d).UnixMilli())
		require.LessOrEqual(t, item.Rereads[i], time.Now().Add(d).UnixMilli())
	}
	// the periodic scan takes the record at the first one
	require.Equal(t, item.Rereads[0], item.RepairAt)
}

// GetOperation keeps the semantics of GO-7482 (PR #59): Completed once the registry confirmed the
// name (or the cache already had it), Pending while the registry does not have it, an error when
// the refresh failed. the cache is refreshed after it, best effort, and read again later
func TestAnynsRpc_GetOperation_RealCache(t *testing.T) {
	// not expired: nothing but the operation would refresh it
	oldExpires := time.Now().Add(10 * 24 * time.Hour).Unix()
	newExpires := time.Now().Add(355 * 24 * time.Hour).Unix()

	seed := func(t *testing.T, coll *mongo.Collection) {
		_, err := coll.InsertOne(ctx, cache.NameDataItem{
			FullName: opName, OwnerEthAddress: oldEoa, OwnerScwEthAddress: oldScw, OwnerAnyAddress: oldAnyID,
			NameExpires: oldExpires, ObservedBlock: 10, ObservedAt: time.Now().UnixMilli(),
		})
		require.NoError(t, err)
	}

	t.Run("not cached, registered: Completed, the name is cached with its re-reads in the same write", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		completedOp(fx)
		expectRegistered(cm, newScw, newEoa, newAnyID, newExpires)

		start := time.Now()
		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Completed, resp.OperationState)

		item := cachedItem(t, coll)
		require.Equal(t, newEoa, item.OwnerEthAddress)
		require.Equal(t, newExpires, item.NameExpires)
		requireRereads(t, item, start)
	})

	t.Run("not cached, the registry does not have it yet: Pending, nothing is cached", func(t *testing.T) {
		old := updateCacheRetryDelay
		updateCacheRetryDelay = time.Millisecond
		defer func() { updateCacheRetryDelay = old }()

		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		completedOp(fx)
		// the poll's retries, and the background's try
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil).MinTimes(updateCacheRetryCount)

		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_PendingOrNotFound, resp.OperationState)
		require.Nil(t, cachedItem(t, coll))
	})

	t.Run("not cached, the owner can not be read: Completed, the name is cached as taken, marked", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		completedOp(fx)
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		cm.EXPECT().GetNameExpires(gomock.Any(), opName, gomock.Any()).Return(big.NewInt(newExpires), nil)
		cm.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), opName, gomock.Any()).Return("", "", "", errors.New("rpc is down"))

		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Completed, resp.OperationState)

		item := cachedItem(t, coll)
		require.True(t, item.RefreshNeeded)
		require.Len(t, item.Rereads, 2)
		out, err := cs.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: opName})
		require.NoError(t, err)
		require.False(t, out.Available)
	})

	t.Run("not cached, the registry can not be read: an error (as before), nothing is cached", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		completedOp(fx)
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down")).MinTimes(1)

		_, err := getOperation(t, fx)
		require.Error(t, err)
		require.Nil(t, cachedItem(t, coll))
	})

	t.Run("cached (a renewal): Completed at once, the background refreshes it and schedules the re-reads", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		seed(t, coll)
		completedOp(fx)

		// the poll does not wait for the refresh: it is blocked until the poll returned
		release := make(chan struct{})
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(context.Context, [32]byte, common.Hash) (common.Address, error) {
				<-release
				return common.HexToAddress(nameWrapper), nil
			})
		cm.EXPECT().GetNameExpires(gomock.Any(), opName, gomock.Any()).Return(big.NewInt(newExpires), nil)
		cm.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), opName, gomock.Any()).Return(oldScw, oldAnyID, "", nil)
		cm.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(oldEoa), nil)

		start := time.Now()
		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Completed, resp.OperationState)
		require.Equal(t, oldExpires, cachedItem(t, coll).NameExpires)
		close(release)

		eventually(t, "the background refresh", func() bool { return cachedItem(t, coll).NameExpires == newExpires })
		item := cachedItem(t, coll)
		require.Equal(t, oldEoa, item.OwnerEthAddress)
		requireRereads(t, item, start)
	})

	t.Run("cached, the background refresh fails: Completed, the record is marked with the re-reads", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		seed(t, coll)
		completedOp(fx)
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))

		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Completed, resp.OperationState)

		eventually(t, "the mark", func() bool { return cachedItem(t, coll).RefreshNeeded })
		item := cachedItem(t, coll)
		require.Len(t, item.Rereads, 2)
		require.Equal(t, oldEoa, item.OwnerEthAddress, "the data stays, the name is still taken")
		require.Greater(t, item.RefreshNextAt, time.Now().UnixMilli(), "after a backoff")
		require.Equal(t, item.RefreshNextAt, item.RepairAt)
	})

	t.Run("cached, the latest block says not registered (lagging provider): Completed, nothing is removed", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)
		cs, cm, coll := newRealCache(t)
		fx.anynsAARpc.cache = cs
		seed(t, coll)
		completedOp(fx)
		// no owner at the latest block (a lagging backend), registered at the finalized one
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, nil)
		cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		cm.EXPECT().GetNameExpires(gomock.Any(), opName, gomock.Any()).Return(big.NewInt(oldExpires), nil)

		resp, err := getOperation(t, fx)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Completed, resp.OperationState)

		// not final: the record stays as it is, marked, with the re-reads
		eventually(t, "the mark", func() bool { return cachedItem(t, coll).RefreshNeeded })
		item := cachedItem(t, coll)
		require.False(t, item.Removed)
		require.Equal(t, oldEoa, item.OwnerEthAddress)
		require.Len(t, item.Rereads, 2)
		out, err := cs.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: opName})
		require.NoError(t, err)
		require.False(t, out.Available)
	})
}

// the cached path of a completed operation never waits for Mongo writes or the contracts
func TestAnynsRpc_GetOperation_CachedNeverWaits(t *testing.T) {
	fx := newFixture(t, "")
	defer fx.finish(t)
	cs, cm, coll := newRealCache(t)
	fx.anynsAARpc.cache = cs
	_, err := coll.InsertOne(ctx, cache.NameDataItem{
		FullName: opName, OwnerEthAddress: oldEoa, OwnerScwEthAddress: oldScw, OwnerAnyAddress: oldAnyID,
		NameExpires: time.Now().Add(-time.Hour).Unix(), ObservedBlock: 10,
	})
	require.NoError(t, err)

	// every write of the cache's client stalls, and so does every contract read
	failCommand(t, bson.M{
		"failCommands":    []string{"update", "insert", "findAndModify", "commitTransaction"},
		"appName":         realCacheApp,
		"blockConnection": true,
		"blockTimeMS":     5000,
	}, bson.M{"times": 100})
	cm.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, _ [32]byte, _ common.Hash) (common.Address, error) {
			<-ctx.Done()
			return common.Address{}, ctx.Err()
		}).AnyTimes()
	completedOp(fx)

	start := time.Now()
	resp, err := getOperation(t, fx)
	require.NoError(t, err)
	require.Equal(t, nsp.OperationState_Completed, resp.OperationState)
	require.Less(t, time.Since(start), time.Second)
}
