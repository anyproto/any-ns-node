package cache

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	"github.com/anyproto/any-sync/app"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	*cacheService
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
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin(gomock.Any()).MaxTimes(2)
	fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().ConnectToController(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().WaitForTxToStartMining(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().IsContractDeployed(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns-test",
	}

	fx.a.Register(fx.ts).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.cacheService)

	require.NoError(t, fx.a.Start(ctx))

	// TODO: mock Mongo!
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	// drop database any-ns-test
	err = client.Database("any-ns-test").Drop(ctx)
	require.NoError(t, err)

	return fx
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
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			OwnerAnyAddress: "anyid",
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
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		// 2 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "test.any"})
		require.NoError(t, err)

		assert.False(t, out.Available)
		assert.Equal(t, "owner", out.OwnerEthAddress)
		assert.Equal(t, "anyid", out.OwnerAnyAddress)
	})
}

func TestCacheService_GetNameByAddress(t *testing.T) {
	t.Run("find nothing", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerEthAddress: "owner"})
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
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		lower := strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerEthAddress: lower})
		require.NoError(t, err)

		require.True(t, out.Found)
		require.Equal(t, "test.any", out.Name)
	})

	t.Run("find one if address is not in lower", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		_, err := fx.itemColl.InsertOne(ctx, NameDataItem{
			FullName: "test.any",
			// should be always stored in lower case
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		// 2 - call GetNameByAddress
		out, err := fx.GetNameByAddress(ctx, &nsp.NameByAddressRequest{OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"})
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
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		require.Equal(t, "owner", item.OwnerEthAddress)
		require.Equal(t, "anyid", item.OwnerAnyAddress)

		// 3 - call IsNameAvailable
		out, err := fx.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: "test.any"})
		require.NoError(t, err)

		assert.False(t, out.Available)
		assert.Equal(t, "owner", out.OwnerEthAddress)
		assert.Equal(t, "anyid", out.OwnerAnyAddress)
	})

	t.Run("update item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - insert item to DB
		err := fx.setNameData(ctx, &NameDataItem{
			FullName:        "test.any",
			OwnerEthAddress: "owner",
			OwnerAnyAddress: "anyid",
		})
		require.NoError(t, err)

		// 2 - insert item to DB
		err = fx.setNameData(ctx, &NameDataItem{
			FullName:        "test.any",
			OwnerEthAddress: "owner2",
			OwnerAnyAddress: "anyid2",
		})
		require.NoError(t, err)

		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.NoError(t, err)

		require.Equal(t, "test.any", item.FullName)
		require.Equal(t, "owner2", item.OwnerEthAddress)
		require.Equal(t, "anyid2", item.OwnerAnyAddress)
	})
}

func TestCacheService_UpdateInCache(t *testing.T) {

	t.Run("return error if CreateEthConnection fails", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().Return(nil, errors.New("failed to connect"))

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
	})

	t.Run("return error if GetOwnerForNamehash fails", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
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

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return notEmptyAddr, nil
		})
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}, owner interface{}) (string, string, string, *big.Int, error) {
			return "", "", "", big.NewInt(0), errors.New("SOME BIG ERROR")
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
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
			return common.Address{}, errors.New("not found")
		})

		// call it
		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: "test.any",
		})
		require.Error(t, err)
		require.Equal(t, "not found", err.Error())

		// it should create new item in Mongo
		// 2 - check if item is in DB
		item := &NameDataItem{}
		err = fx.itemColl.FindOne(ctx, findNameDataByName{FullName: "test.any"}).Decode(&item)
		require.Error(t, err)
	})

	t.Run("create new item if found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// if this returns some address -> it means name is taken
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return notEmptyAddr, nil
		})

		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}, owner interface{}) (string, string, string, *big.Int, error) {
			return "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", big.NewInt(12390243), nil
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
		require.Equal(t, "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51", item.OwnerEthAddress)
		require.Equal(t, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", item.OwnerAnyAddress)
	})

	t.Run("update item if found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - create item in DB first
		err := fx.setNameData(ctx, &NameDataItem{
			FullName:        "test.any",
			OwnerEthAddress: "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
		})
		require.NoError(t, err)

		// 2 - call it
		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()

		// if this returns some address -> it means name is taken
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
			// this was changed!
			anotherAddr := common.HexToAddress("0xAAB27b150451726EC7738aa1d0A94505c8729bd1")
			return anotherAddr, nil
		})

		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}, owner interface{}) (string, string, string, *big.Int, error) {
			return "0xAAB27b150451726EC7738aa1d0A94505c8729bd1", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", big.NewInt(12390243), nil
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
		require.Equal(t, "0xaab27b150451726ec7738aa1d0a94505c8729bd1", item.OwnerEthAddress)
		require.Equal(t, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", item.OwnerAnyAddress)
	})
}
