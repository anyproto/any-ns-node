package anynsaarpc

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	mock_accountabstraction "github.com/anyproto/any-ns-node/account_abstraction/mock"
	"github.com/anyproto/any-ns-node/cache"
	mock_cache "github.com/anyproto/any-ns-node/cache/mock"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService
	aa        *mock_accountabstraction.MockAccountAbstractionService
	cache     *mock_cache.MockCacheService

	*anynsAARpc
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsAARpc: New().(*anynsAARpc),
	}

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().IsContractDeployed(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	fx.cache = mock_cache.NewMockCacheService(fx.ctrl)
	fx.cache.EXPECT().Name().Return(cache.CName).AnyTimes()
	fx.cache.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns-tst",
	}

	// drop everything
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fx.config.Mongo.Connect))
	require.NoError(t, err)

	err = client.Database(fx.config.Mongo.Database).Drop(ctx)
	require.NoError(t, err)

	fx.aa = mock_accountabstraction.NewMockAccountAbstractionService(fx.ctrl)
	fx.aa.EXPECT().Name().Return(accountabstraction.CName).AnyTimes()
	fx.aa.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.aa).
		Register(fx.cache).
		Register(fx.anynsAARpc)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestAnynsRpc_GetUserAccount(t *testing.T) {
	t.Run("return not found error if no such account", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
			return common.Address{}, errors.New("not found")
		})

		pctx := context.Background()
		resp, err := fx.GetUserAccount(pctx, &nsp.GetUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.Error(t, err, "not found")
		assert.Nil(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().GetNamesCountLeft(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, param interface{}) (count uint64, err error) {
			return uint64(10), nil
		})

		pctx := context.Background()

		err := fx.mongoAddUserToTheWhitelist(pctx,
			common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
			"12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			21,
		)
		require.NoError(t, err)

		resp, err := fx.GetUserAccount(pctx, &nsp.GetUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})
		assert.NoError(t, err)

		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress(resp.OwnerEthAddress), common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, common.HexToAddress(resp.OwnerSmartContracWalletAddress), common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))
		assert.Equal(t, resp.NamesCountLeft, uint64(10))
		assert.Equal(t, resp.OperationsCountLeft, uint64(21))
	})
}

func TestAnynsRpc_AdminFundUserAccount(t *testing.T) {

	t.Run("success when asked to add 0 additional name requests", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().AdminVerifyIdentity(gomock.Any(), gomock.Any()).DoAndReturn(func(payload []byte, signature []byte) (err error) {
			// no error, IT IS THE ADMIN!
			return nil
		})

		fx.aa.EXPECT().AdminMintAccessTokens(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, scw common.Address, count *big.Int) (op string, err error) {
			// no error
			return "123", nil
		})

		// create payload
		var in nsp.AdminFundUserAccountRequestSigned

		accountKeys, err := accountdata.NewRandom()
		require.NoError(t, err)

		/*
			fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
				return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
			})*/

		// pack
		nrr := nsp.AdminFundUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),

			// add 0 name calls
			NamesCount: 0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		in.Payload = marshalled
		in.Signature, err = accountKeys.SignKey.Sign(in.Payload)
		require.NoError(t, err)

		pctx := context.Background()
		resp, err := fx.AdminFundUserAccount(pctx, &in)

		// should return "pending" operation
		require.NoError(t, err)
		require.Equal(t, resp.OperationId, "123")
		require.Equal(t, resp.OperationState, nsp.OperationState_Pending)
	})
}

func TestAnynsRpc_GetOperation(t *testing.T) {
	t.Run("fail if no operation is in the DB", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		_, err := fx.GetOperation(pctx, &gosr)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Pending,
			}, nil
		})

		// create operation in Mongo first
		err := fx.mongoSaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		require.NoError(t, err)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		resp, err := fx.GetOperation(pctx, &gosr)

		require.NoError(t, err)
		require.Equal(t, resp.OperationId, "123")
		// not found
		require.Equal(t, resp.OperationState, nsp.OperationState_Pending)
	})

	t.Run("opertaion completed - check in cache", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Completed,
			}, nil
		})

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error) {
			// name already in cache!
			return &nsp.NameAvailableResponse{}, nil
		})

		// create operation in Mongo first
		err := fx.mongoSaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		require.NoError(t, err)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		resp, err := fx.GetOperation(pctx, &gosr)

		require.NoError(t, err)
		require.Equal(t, resp.OperationId, "123")
		// not found
		require.Equal(t, resp.OperationState, nsp.OperationState_Completed)
	})

	t.Run("opertaion completed - update in cache", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Completed,
			}, nil
		})

		fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).MinTimes(1)

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error) {
			// name is not in cache!
			// should call UpdateCache
			return nil, errors.New("not found")
		})

		// create operation in Mongo first
		err := fx.mongoSaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		require.NoError(t, err)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		resp, err := fx.GetOperation(pctx, &gosr)

		require.NoError(t, err)
		require.Equal(t, resp.OperationId, "123")
		// not found
		require.Equal(t, resp.OperationState, nsp.OperationState_Completed)
	})

	t.Run("should return error if update in cache failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Completed,
			}, nil
		})

		fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (err error) {
			return errors.New("failed to update in cache")
		})

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error) {
			// name is not in cache!
			// should call UpdateCache
			return nil, errors.New("not found")
		})

		// create operation in Mongo first
		err := fx.mongoSaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		require.NoError(t, err)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		_, err = fx.GetOperation(pctx, &gosr)

		require.Error(t, err)
	})
}

func TestAnynsRpc_MongoAddUserToTheWhitelist(t *testing.T) {
	t.Run("fail if wrong operations count", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 0)
		assert.Error(t, err)
	})

	t.Run("success if user never existed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// TODO: mock!
		// read from Mongo and check
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-users"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUser{}
		err = itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.Address, owner.Hex())
		assert.Equal(t, item.OperationsCount, uint64(1))
	})

	t.Run("fail if user has existed before but Any ID is different", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// again but with different Any ID
		anyIDDifferent := "12D3KaaXB5EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		err = fx.mongoAddUserToTheWhitelist(pctx, owner, anyIDDifferent, 1)
		assert.Error(t, err)
	})

	t.Run("success if user has existed before", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// again!
		err = fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// TODO: mock!
		// read from Mongo and check
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-users"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUser{}
		err = itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.Address, owner.Hex())
		assert.Equal(t, item.OperationsCount, uint64(2))
	})
}

func TestAnynsRpc_MongoGetUserOperationsCount(t *testing.T) {
	t.Run("fail if nothing is found operations count", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		ops, err := fx.mongoGetUserOperationsCount(pctx, owner, "")
		assert.Error(t, err)
		assert.Equal(t, ops, uint64(0))
	})

	t.Run("fail if wrong Any ID", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.mongoGetUserOperationsCount(pctx, owner, "1111KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Error(t, err)
		assert.Equal(t, ops, uint64(0))
	})

	t.Run("success if AnyID is empty (do not compare it!)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.mongoGetUserOperationsCount(pctx, owner, "")
		assert.NoError(t, err)
		assert.Equal(t, ops, uint64(1))
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.mongoGetUserOperationsCount(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.NoError(t, err)
		assert.Equal(t, ops, uint64(1))
	})
}

func TestAnynsRpc_MongoDecreaseUserOperationsCount(t *testing.T) {
	t.Run("fail if user is not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.mongoDecreaseUserOperationsCount(pctx, owner)
		assert.Error(t, err)
	})

	t.Run("fail if user has no operations already", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// 1st time - should work
		err = fx.mongoDecreaseUserOperationsCount(pctx, owner)
		assert.NoError(t, err)

		// 2nd time - should fail
		err = fx.mongoDecreaseUserOperationsCount(pctx, owner)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		err = fx.mongoDecreaseUserOperationsCount(pctx, owner)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_MongoSaveOperation(t *testing.T) {
	t.Run("should create new item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		err := fx.mongoSaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		assert.NoError(t, err)

		// check in the DB
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-operations"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUserOperation{}
		err = itemColl.FindOne(ctx, findUserOperationByID{OperationID: "123"}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.OperationID, "123")
		assert.Equal(t, item.OwnerEthAddress, strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, item.OwnerAnyID, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, item.Data, []byte("data"))
		assert.Equal(t, item.FullName, "hello.any")
	})

	t.Run("should not update existing item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		err := fx.mongoSaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
		})
		assert.NoError(t, err)

		err = fx.mongoSaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
		})
		assert.Error(t, err)
	})
}

func TestAnynsRpc_MongoGetOpertaion(t *testing.T) {
	t.Run("should return error if not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		_, err := fx.mongoGetOperation(ctx, "123")
		assert.Error(t, err)
	})

	t.Run("should return item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		err := fx.mongoSaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		assert.NoError(t, err)

		item, err := fx.mongoGetOperation(ctx, "123")
		assert.NoError(t, err)
		assert.Equal(t, item.OperationID, "123")
		assert.Equal(t, item.OwnerEthAddress, strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, item.OwnerAnyID, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, item.Data, []byte("data"))
		assert.Equal(t, item.FullName, "hello.any")
	})
}

func TestAnynsRpc_GetDataNameRegister(t *testing.T) {
	t.Run("fail if name is invalid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if eth address is invalid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if Any address is invalid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "oWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if space ID is invalid", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "baxxxybeibs62gqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		fx.aa.EXPECT().GetDataNameRegister(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (dataOut []byte, contextData []byte, err error) {
			return []byte("data"), []byte("context"), nil
		})

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_VerifyAnyIdentity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - enable user
		PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		//SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="
		PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="

		// OwnerAnyID
		decodedPeerKey, err := crypto.DecodeKeyFromString(
			PeerKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		marshalled, err := hex.DecodeString("123445ffff")
		assert.NoError(t, err)

		signature, err := decodedPeerKey.Sign(marshalled)
		assert.NoError(t, err)

		// Identity here is in the marshalled format
		x := decodedPeerKey.GetPublic()
		identityMarshalled, err := decodedPeerKey.GetPublic().Marshall()
		assert.NoError(t, err)

		// convert PeerId to marashalled PubKey
		pid, err := crypto.DecodePeerId(PeerId)
		assert.NoError(t, err)

		// compare 2 PubKeys (should be same)
		assert.Equal(t, x, pid)

		identityMarshalled2, err := pid.Marshall()
		assert.NoError(t, err)

		// compare identityMarshalled with identity
		assert.Equal(t, identityMarshalled, identityMarshalled2)

		// VerifyAnyIdentity used marshalled version before!
		//err = fx.VerifyAnyIdentity(string(identityMarshalled2), marshalled, signature)
		//assert.NoError(t, err)

		err = fx.VerifyAnyIdentity(PeerId, marshalled, signature)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_CreateUserOperation(t *testing.T) {
	t.Run("fail if wrong signature", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - enable user
		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		//PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, PeerId, 1)
		assert.NoError(t, err)

		var cuor nsp.CreateUserOperationRequest
		// from string to []byte
		cuor.Data, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.SignedData, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.Context, err = hex.DecodeString("1234")
		assert.NoError(t, err)

		cuor.OwnerEthAddress = owner.Hex()

		// OwnerAnyID in string format
		cuor.OwnerAnyID = PeerId

		marshalled, err := cuor.Marshal()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		// sign with WRONG key here
		SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="
		wrongKey, err := crypto.DecodeKeyFromString(
			SignKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		cuor_signed.Signature, err = wrongKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		// let's go
		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.Error(t, err)
	})

	t.Run("fail if SendUserOperation failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - enable user
		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
		//SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, PeerId, 1)
		assert.NoError(t, err)

		var cuor nsp.CreateUserOperationRequest
		// from string to []byte
		cuor.Data, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.SignedData, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.Context, err = hex.DecodeString("1234")
		assert.NoError(t, err)

		cuor.OwnerEthAddress = owner.Hex()

		// OwnerAnyID
		decodedPeerKey, err := crypto.DecodeKeyFromString(
			PeerKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		//identity, err := decodedPeerKey.GetPublic().Marshall()
		//assert.NoError(t, err)

		// OwnerAnyID here is in the marshalled format
		//cuor.OwnerAnyID = string(identity)

		// OwnerAnyID in string format
		cuor.OwnerAnyID = PeerId

		marshalled, err := cuor.Marshal()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		cuor_signed.Signature, err = decodedPeerKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		fx.aa.EXPECT().SendUserOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, x interface{}, y interface{}) (operationID string, err error) {
			return "", errors.New("Bad error. Youre doomed")
		})

		// let's go
		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - enable user
		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
		//SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		err := fx.mongoAddUserToTheWhitelist(pctx, owner, PeerId, 1)
		assert.NoError(t, err)

		var cuor nsp.CreateUserOperationRequest
		// from string to []byte
		cuor.Data, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.SignedData, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.Context, err = hex.DecodeString("1234")
		assert.NoError(t, err)

		cuor.OwnerEthAddress = owner.Hex()

		// OwnerAnyID
		decodedPeerKey, err := crypto.DecodeKeyFromString(
			PeerKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		//identity, err := decodedPeerKey.GetPublic().Marshall()
		//assert.NoError(t, err)

		// OwnerAnyID here is in the marshalled format
		//cuor.OwnerAnyID = string(identity)

		// OwnerAnyID in string format
		cuor.OwnerAnyID = PeerId

		marshalled, err := cuor.Marshal()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		cuor_signed.Signature, err = decodedPeerKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		fx.aa.EXPECT().SendUserOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, x interface{}, y interface{}) (operationID string, err error) {
			return "123", nil
		})

		// let's go
		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.NoError(t, err)
	})
}
