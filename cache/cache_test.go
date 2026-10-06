package cache

import (
	"context"
	"errors"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	"github.com/anyproto/any-sync/app"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"
)

var ctx = context.Background()

// not shared with the other packages: they drop their databases while these tests run
const testDbName = "any-ns-test-cache"

// the test Mongo. transactions need a replica set, e.g.:
//
// (enableTestCommands: the fail point tests stall Mongo commands, they skip without it)
//
//	docker run -d --name any-ns-test-rs -p 27018:27018 mongo:7 --replSet rs0 --port 27018 --bind_ip_all \
//	  --setParameter enableTestCommands=1
//	mongosh --port 27018 --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27018"}]})'
//	ANY_NS_TEST_MONGO="mongodb://localhost:27018/?replicaSet=rs0" go test ./cache/
func testMongoURI() string {
	if uri := os.Getenv("ANY_NS_TEST_MONGO"); uri != "" {
		return uri
	}
	return "mongodb://localhost:27017"
}

// a name that is registered and not expired
var notExpired = time.Now().Add(365 * 24 * time.Hour).Unix()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	// what LatestBlock and FinalizedBlock return. by default a new block on every call of
	// LatestBlock, at the local time, and the finalized block is the latest one
	headMu    sync.Mutex
	head      func() *contracts.Block
	finalized func() *contracts.Block
	// what FinalizedBlock fails with
	finalizedErr error
	block        int64

	*cacheService
}

// the hash of a test block: one fork, unless a test says otherwise
func blockHash(n int64) common.Hash {
	return common.BigToHash(big.NewInt(n))
}

func testBlock(n int64, at time.Time) *contracts.Block {
	return &contracts.Block{Number: n, Hash: blockHash(n), Time: uint64(at.Unix())}
}

func (fx *fixture) latestBlock() *contracts.Block {
	fx.headMu.Lock()
	defer fx.headMu.Unlock()
	if fx.head != nil {
		return fx.head()
	}
	fx.block++
	return testBlock(fx.block, time.Now())
}

func (fx *fixture) finalizedBlock() (*contracts.Block, error) {
	fx.headMu.Lock()
	defer fx.headMu.Unlock()
	switch {
	case fx.finalizedErr != nil:
		return nil, fx.finalizedErr
	case fx.finalized != nil:
		return fx.finalized(), nil
	case fx.head != nil:
		return fx.head(), nil
	}
	return testBlock(fx.block, time.Now()), nil
}

// setHead makes LatestBlock return this block from now on
func (fx *fixture) setHead(block int64, at time.Time) {
	fx.setHeadBlock(testBlock(block, at))
}

func (fx *fixture) setHeadBlock(b *contracts.Block) {
	fx.setHeadFunc(func() *contracts.Block { return b })
}

// setHeadFunc: LatestBlock returns what head returns from now on
func (fx *fixture) setHeadFunc(head func() *contracts.Block) {
	fx.headMu.Lock()
	defer fx.headMu.Unlock()
	fx.head = head
}

// setFinalized makes FinalizedBlock return this block from now on
func (fx *fixture) setFinalized(block int64, at time.Time) {
	fx.headMu.Lock()
	defer fx.headMu.Unlock()
	b := testBlock(block, at)
	fx.finalized = func() *contracts.Block { return b }
}

// setFinalizedErr makes FinalizedBlock fail with err from now on
func (fx *fixture) setFinalizedErr(err error) {
	fx.headMu.Lock()
	defer fx.headMu.Unlock()
	fx.finalizedErr = err
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		cacheService: New().(*cacheService),
	}

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin().MaxTimes(2)
	fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().ConnectToPrivateController().AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().WaitForTxToStartMining(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().IsContractDeployed(gomock.Any(), gomock.Any()).AnyTimes()
	fx.block = 1000
	fx.contracts.EXPECT().LatestBlock(gomock.Any()).DoAndReturn(func(context.Context) (*contracts.Block, error) {
		return fx.latestBlock(), nil
	}).AnyTimes()
	fx.contracts.EXPECT().FinalizedBlock(gomock.Any()).DoAndReturn(func(context.Context) (*contracts.Block, error) {
		return fx.finalizedBlock()
	}).AnyTimes()
	fx.config.Mongo = config.Mongo{
		Connect:  testMongoURI(),
		Database: testDbName,
	}
	// the tests run on a standalone Mongo too (see testMongoURI), and they run the repair by hand
	fx.config.Cache = config.Cache{AllowUnsafeStandalone: true, RepairIntervalSec: -1}

	fx.a.Register(fx.ts).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.cacheService)

	require.NoError(t, fx.a.Start(ctx))
	// the tests run the background refresh by hand (see runQueued): the requests stay queued
	fx.stopBackground()

	// TODO: mock Mongo!
	uri := testMongoURI()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	defer func() { _ = client.Disconnect(ctx) }()

	// drop the test database
	err = client.Database(testDbName).Drop(ctx)
	if err != nil {
		// sleep 1 second
		time.Sleep(1 * time.Second)
	}

	// like prod: a unique index on the name (default name, name_1)
	for i := 0; ; i++ {
		_, err = fx.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)})
		if err == nil || i == 20 {
			break
		}
		// the drop can still be in progress
		time.Sleep(100 * time.Millisecond)
	}
	require.NoError(t, err)

	return fx
}

// stopBackground stops the background worker: what lookups request stays in the queue
func (fx *fixture) stopBackground() {
	if fx.stopWorker != nil {
		fx.stopWorker()
		<-fx.workerDone
		fx.stopWorker = nil
	}
}

// runQueued runs what is in the queue of the background worker, as it would. returns how many
func (fx *fixture) runQueued() int {
	n := 0
	for {
		select {
		case r := <-fx.queue.ch:
			fx.queue.done(r)
			fx.handle(ctx, r)
			n++
		default:
			return n
		}
	}
}

// setNameData writes the record as it is (an upsert by the name), like a node older than GO-7567
// did. the refresh paths of the cache never do: they write in order (see applyObservation)
func (fx *fixture) setNameData(ctx context.Context, in *NameDataItem) error {
	in.OwnerScwEthAddress = strings.ToLower(in.OwnerScwEthAddress)
	in.OwnerEthAddress = strings.ToLower(in.OwnerEthAddress)
	_, err := fx.itemColl.ReplaceOne(ctx, findNameDataByName{FullName: in.FullName}, in, options.Replace().SetUpsert(true))
	return err
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestCacheService_IsNameAvailable(t *testing.T) {
	t.Run("find nothing", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "test"})
		require.NoError(t, err)

		assert.True(t, out.Available)
	})

	t.Run("find nothing if capitalization is different", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: "owner_scw",
			OwnerAnyAddress:    "anyid",
		})
		require.NoError(t, err)

		// 2 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "TEST.any"})
		require.NoError(t, err)
		assert.True(t, out.Available)
	})

	t.Run("find one", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: "owner_scw",
			OwnerAnyAddress:    "anyid",
			NameExpires:        notExpired,
		})
		require.NoError(t, err)

		// 2 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "test.any"})
		require.NoError(t, err)

		assert.False(t, out.Available)
		assert.Equal(t, "owner", out.OwnerEthAddress)
		assert.Equal(t, "owner_scw", out.OwnerScwEthAddress)
		assert.Equal(t, "anyid", out.OwnerAnyAddress)
	})
}

func TestCacheService_GetNameByAddress(t *testing.T) {
	t.Run("find nothing", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: "owner_scw",
			OwnerAnyAddress:    "anyid",
		})
		require.NoError(t, err)

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: "owner_scw"})
		require.NoError(t, err)

		require.True(t, out.Found)
		require.Equal(t, "test.any", out.Name)
	})

	t.Run("find one if address is in lowercase", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName: "test.any",
			// should be always stored in lower case
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
			OwnerAnyAddress:    "anyid",
		})
		require.NoError(t, err)

		lower := strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: lower})
		require.NoError(t, err)

		require.True(t, out.Found)
		require.Equal(t, "test.any", out.Name)
	})

	t.Run("find one if address is not in lower", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			// should be always stored in lower case
			OwnerScwEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
			OwnerAnyAddress:    "anyid",
		})
		require.NoError(t, err)

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerScwEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"})
		require.NoError(t, err)

		require.True(t, out.Found)
		require.Equal(t, "test.any", out.Name)
	})
}

func TestCacheService_setNameData(t *testing.T) {
	t.Run("create new item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		err := fx.setNameData(ctx, &NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: "owner_scw",
			OwnerAnyAddress:    "anyid",
			NameExpires:        notExpired,
		})
		require.NoError(t, err)

		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		require.Equal(t, "owner", item.OwnerEthAddress)
		require.Equal(t, "owner_scw", item.OwnerScwEthAddress)
		require.Equal(t, "anyid", item.OwnerAnyAddress)

		// 3 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "test.any"})
		require.NoError(t, err)

		assert.False(t, out.Available)
		assert.Equal(t, "owner", out.OwnerEthAddress)
		assert.Equal(t, "owner_scw", out.OwnerScwEthAddress)
		assert.Equal(t, "anyid", out.OwnerAnyAddress)
	})

	t.Run("update item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		err := fx.setNameData(ctx, &NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner",
			OwnerScwEthAddress: "owner_scw",
			OwnerAnyAddress:    "anyid",
		})
		require.NoError(t, err)

		// 2 - insert item to DB
		err = fx.setNameData(ctx, &NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner2",
			OwnerScwEthAddress: "owner_scw2",
			OwnerAnyAddress:    "anyid2",
		})
		require.NoError(t, err)

		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		require.Equal(t, "owner2", item.OwnerEthAddress)
		require.Equal(t, "owner_scw2", item.OwnerScwEthAddress)
		require.Equal(t, "anyid2", item.OwnerAnyAddress)
	})
}

func TestCacheService_UpdateInCache(t *testing.T) {
	t.Run("return error if GetOwnerForNamehash fails", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			return common.Address{}, errors.New("SOME BIG ERROR")
		})

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
	})

	t.Run("return error if GetAdditionalNameInfo fails", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return notEmptyAddr, nil
		})
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}, _ interface{}) (string, string, string, error) {
			return "", "", "", errors.New("SOME BIG ERROR")
		})

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
	})

	t.Run("do not create new item if not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// if this returns some address -> it means name is taken
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			return common.Address{}, errors.New("not found")
		})

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNameNotRegistered)

		// it should not create an item in Mongo: there is nothing to remove either
		require.EqualValues(t, 0, countRecords(t, fx))
	})

	t.Run("create new item if found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// if this returns some address -> it means name is taken
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return notEmptyAddr, nil
		})

		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}, _ interface{}) (string, string, string, error) {
			return "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", nil
		})

		// >>> see this:
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			return common.HexToAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"), nil
		})

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.NoError(t, err)

		// it should create new item in Mongo
		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		// should be in lower case
		require.Equal(t, "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5", item.OwnerEthAddress)
		require.Equal(t, "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51", item.OwnerScwEthAddress)
		require.Equal(t, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", item.OwnerAnyAddress)
	})

	t.Run("update item if found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - create item in DB first
		err := fx.setNameData(ctx, &NameDataItem{
			FullName:           "test.any",
			OwnerEthAddress:    "owner111",
			OwnerScwEthAddress: "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51",
			OwnerAnyAddress:    "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
		})
		require.NoError(t, err)

		// 2 - call it
		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// if this returns some address -> it means name is taken
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			// this was changed!
			anotherAddr := common.HexToAddress("0xAAB27b150451726EC7738aa1d0A94505c8729bd1")
			return anotherAddr, nil
		})

		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), gomock.Any(), gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}, _ interface{}) (string, string, string, error) {
			return "0xAAB27b150451726EC7738aa1d0A94505c8729bd1", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", nil
		})

		// >>> see this:
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			return common.HexToAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"), nil
		})

		err = fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.NoError(t, err)

		// it should create new item in Mongo
		// 3 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		// should be in lower case
		require.Equal(t, "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5", item.OwnerEthAddress)
		require.Equal(t, "0xaab27b150451726ec7738aa1d0a94505c8729bd1", item.OwnerScwEthAddress)
		require.Equal(t, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", item.OwnerAnyAddress)
	})

	t.Run("return ErrNameNotRegistered if owner is a zero address", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// zero address -> name is not in the registry
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ interface{}, _ interface{}) (common.Address, error) {
			return common.Address{}, nil
		}).Times(1)

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNameNotRegistered)

		// no record should be written to Mongo
		require.EqualValues(t, 0, countRecords(t, fx))
	})
}
