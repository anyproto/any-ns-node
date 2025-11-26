package anynsaarpc

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync/nodeconf/mock_nodeconf"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	mock_accountabstraction "github.com/anyproto/any-ns-node/account_abstraction/mock"
	"github.com/anyproto/any-ns-node/cache"
	mock_cache "github.com/anyproto/any-ns-node/cache/mock"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	db_service "github.com/anyproto/any-ns-node/db"
	mock_db_service "github.com/anyproto/any-ns-node/db/mock"
	"github.com/anyproto/any-ns-node/verification"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	nodeConf  *mock_nodeconf.MockService
	contracts *mock_contracts.MockContractsService
	aa        *mock_accountabstraction.MockAccountAbstractionService
	cache     *mock_cache.MockCacheService
	db        *mock_db_service.MockDbService

	*anynsAARpc
}

func newFixture(t *testing.T, adminSignKey string) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsAARpc: New().(*anynsAARpc),
	}

	fx.nodeConf = mock_nodeconf.NewMockService(fx.ctrl)
	fx.nodeConf.EXPECT().Name().Return(nodeconf.CName).AnyTimes()
	fx.nodeConf.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Close(gomock.Any()).AnyTimes()

	// NodeTypes(nodeId string) []NodeType
	fx.nodeConf.EXPECT().NodeTypes(gomock.Any()).Return([]nodeconf.NodeType{nodeconf.NodeTypeConsensus}).AnyTimes()

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().IsContractDeployed(gomock.Any(), gomock.Any()).AnyTimes()

	fx.cache = mock_cache.NewMockCacheService(fx.ctrl)
	fx.cache.EXPECT().Name().Return(cache.CName).AnyTimes()
	fx.cache.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.db = mock_db_service.NewMockDbService(fx.ctrl)
	fx.db.EXPECT().Name().Return(db_service.CName).AnyTimes()
	fx.db.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "xxx",
	}

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns-test",
	}

	fx.config.Account = accountservice.Config{
		SigningKey: adminSignKey,
		PeerKey:    "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q==",
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
		Register(fx.nodeConf).
		Register(fx.db).
		Register(fx.anynsAARpc)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestIsValidAnyAddress(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		len := len("A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4")
		assert.Equal(t, len, 48)

		valid := []string{
			"A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4", // Anytype address
		}

		for _, address := range valid {
			res := verification.IsValidAnyAddress(address)
			assert.Equal(t, res, true)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		invalid := []string{
			"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",         // Legacy Bitcoin address
			"3FZbgi29cpjq2GjdwV8eyHuJJnkLtktZc5",         // Segwit Bitcoin address
			"bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", // Bech32 Bitcoin address
			"invalidaddress",                             // Invalid address
			"",                                           // Empty address
			"0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", // Ethereum address
			"bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjze", // CID
		}

		for _, address := range invalid {
			res := verification.IsValidAnyAddress(address)
			assert.Equal(t, res, false)
		}
	})
}

func TestAnynsRpc_GetUserAccount(t *testing.T) {
	t.Run("return not found error if no such account", func(t *testing.T) {
		fx := newFixture(t, "")
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
		fx := newFixture(t, "")
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().GetNamesCountLeft(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, param interface{}) (count uint64, err error) {
			return uint64(10), nil
		})

		pctx := context.Background()

		fx.db.EXPECT().GetUserOperationsCount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, owner common.Address, ownerAnyID string) (uint64, error) {
			return 21, nil
		})

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
		PeerID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		realSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		fx := newFixture(t, realSignKey)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().AdminMintAccessTokens(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, scw common.Address, count *big.Int) (op string, err error) {
			// no error
			return "123", nil
		})

		// create payload
		var in nsp.AdminFundUserAccountRequestSigned

		// pack
		nrr := nsp.AdminFundUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),

			// add 0 name calls
			NamesCount: 0,
		}

		marshalled, err := nrr.MarshalVT()
		require.NoError(t, err)

		signKey, err := crypto.DecodeKeyFromString(
			realSignKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		require.NoError(t, err)

		in.Payload = marshalled
		in.Signature, err = signKey.Sign(in.Payload)
		require.NoError(t, err)

		pctx := peer.CtxWithPeerId(context.Background(), PeerID)
		resp, err := fx.AdminFundUserAccount(pctx, &in)

		// should return "pending" operation
		require.NoError(t, err)
		require.Equal(t, resp.OperationId, "123")
		require.Equal(t, resp.OperationState, nsp.OperationState_Pending)
	})
}

func TestAnynsRpc_GetOperation(t *testing.T) {
	t.Run("fail if Mongo returns error", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{}, errors.New("not found")
		})

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		_, err := fx.GetOperation(pctx, &gosr)
		require.Error(t, err)
	})

	t.Run("success even if no operation is in the DB", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Pending,
			}, nil
		})

		pctx := context.Background()

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{}, mongo.ErrNoDocuments
		})

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		resp, err := fx.GetOperation(pctx, &gosr)
		require.NoError(t, err)
		require.Equal(t, resp.OperationState, nsp.OperationState_Pending)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Pending,
			}, nil
		})

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{
				OperationID:     "123",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Data:            []byte("data"),
				SignedData:      []byte("signed_data"),
				Context:         []byte("context"),
				FullName:        "hello.any",
			}, nil
		})

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

	t.Run("opertaion completed - already in cache", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		fx.aa.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (status *accountabstraction.OperationInfo, err error) {
			return &accountabstraction.OperationInfo{
				OperationState: nsp.OperationState_Completed,
			}, nil
		})

		//fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).MinTimes(1)

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error) {
			// name already in cache!
			return &nsp.NameAvailableResponse{}, nil
		})

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{
				OperationID: "123",
				// should be converted to lower case
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Data:            []byte("data"),
				SignedData:      []byte("signed_data"),
				Context:         []byte("context"),
				FullName:        "hello.any",
			}, nil
		}).MinTimes(1)

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

	t.Run("opertaion completed - cash miss", func(t *testing.T) {
		fx := newFixture(t, "")
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

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{
				OperationID: "123",
				// should be converted to lower case
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Data:            []byte("data"),
				SignedData:      []byte("signed_data"),
				Context:         []byte("context"),
				FullName:        "hello.any",
			}, nil
		}).MinTimes(1)

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
		fx := newFixture(t, "")
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

		fx.db.EXPECT().GetOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opID string) (op db_service.AAUserOperation, err error) {
			return db_service.AAUserOperation{
				OperationID: "123",
				// should be converted to lower case
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Data:            []byte("data"),
				SignedData:      []byte("signed_data"),
				Context:         []byte("context"),
				FullName:        "hello.any",
			}, nil
		}).MinTimes(1)

		pctx := context.Background()

		gosr := nsp.GetOperationStatusRequest{
			OperationId: "123",
		}
		_, err := fx.GetOperation(pctx, &gosr)

		require.Error(t, err)
	})
}

func TestAnynsRpc_GetDataNameRegister(t *testing.T) {
	t.Run("fail if name is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			//SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if eth address is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			//SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if Any address is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "oWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			//SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterRequest = nsp.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			//SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		fx.aa.EXPECT().GetDataNameRegister(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (dataOut []byte, contextData []byte, err error) {
			return []byte("data"), []byte("context"), nil
		})

		pctx := context.Background()
		_, err := fx.GetDataNameRegister(pctx, &req)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_GetDataNameRegisterForSpace(t *testing.T) {
	t.Run("fail if name is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterForSpaceRequest = nsp.NameRegisterForSpaceRequest{
			FullName:        "hello",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegisterForSpace(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if eth address is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterForSpaceRequest = nsp.NameRegisterForSpaceRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegisterForSpace(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("fail if Any address is invalid", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterForSpaceRequest = nsp.NameRegisterForSpaceRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "oWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		pctx := context.Background()
		_, err := fx.GetDataNameRegisterForSpace(pctx, &req)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		var req nsp.NameRegisterForSpaceRequest = nsp.NameRegisterForSpaceRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF",
			OwnerAnyAddress: "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65",
			SpaceId:         "bafybeibs62gqtignuckfqlcr7lhhihgzh2vorxtmc5afm6uxh4zdcmuwuu",
		}

		fx.aa.EXPECT().GetDataNameRegisterForSpace(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (dataOut []byte, contextData []byte, err error) {
			return []byte("data"), []byte("context"), nil
		})

		pctx := context.Background()
		_, err := fx.GetDataNameRegisterForSpace(pctx, &req)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_VerifyAnyIdentity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		// 1 - enable user
		AnytypeID := "A5pn5zo7MDspEvR1qHtZdnP8KG4nLmVaj9emFW2r7Kaho1Ny"
		SignKey := "yUxvxE5ugUTXOfX6wsYBO0ROw4Na/p9InOsgcuuuiJYIMxqQloJhelUH6Kyold890pHIHtw4mEfbeKgBVnzkfg=="

		// OwnerAnyID
		decodedSignKey, err := crypto.DecodeKeyFromString(
			SignKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		marshalled, err := hex.DecodeString("123445ffff")
		assert.NoError(t, err)

		signature, err := decodedSignKey.Sign(marshalled)
		assert.NoError(t, err)

		// Identity here is in the marshalled format
		x := decodedSignKey.GetPublic()
		identityMarshalled, err := decodedSignKey.GetPublic().Marshall()
		assert.NoError(t, err)

		// convert AnytypeID to marashalled PubKey
		pid, err := crypto.DecodeAccountAddress(AnytypeID)
		assert.NoError(t, err)

		// compare 2 PubKeys (should be same)
		assert.Equal(t, x, pid)

		identityMarshalled2, err := pid.Marshall()
		assert.NoError(t, err)

		// compare identityMarshalled with identity
		assert.Equal(t, identityMarshalled, identityMarshalled2)

		err = verification.VerifyAnyIdentity(AnytypeID, marshalled, signature)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_CreateUserOperation(t *testing.T) {
	t.Run("fail if wrong signature", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		// 1 - enable user
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		AnytypeID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		// sign with WRONG key here
		SignKeyBad := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		var cuor nsp.CreateUserOperationRequest

		// from string to []byte
		data, err := hex.DecodeString("1234")
		cuor.Data = data
		assert.NoError(t, err)
		cuor.SignedData, err = hex.DecodeString("1234")
		assert.NoError(t, err)
		cuor.Context, err = hex.DecodeString("1234")
		assert.NoError(t, err)

		cuor.OwnerEthAddress = owner.Hex()

		// OwnerAnyID in string format
		cuor.OwnerAnyID = AnytypeID

		marshalled, err := cuor.MarshalVT()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		wrongKey, err := crypto.DecodeKeyFromString(
			SignKeyBad,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		cuor_signed.Signature, err = wrongKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		// let's go
		identityStr := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		identityBytes, err := crypto.DecodeBytesFromString(identityStr)
		assert.NoError(t, err)
		pctx := peer.CtxWithIdentity(context.Background(), identityBytes)

		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.Error(t, err)
	})

	t.Run("fail if SendUserOperation failed", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		// 1 - enable user
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		AnytypeID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		//PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
		//SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		fx.db.EXPECT().GetUserOperationsCount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, owner common.Address, ownerAnyID string) (uint64, error) {
			return 1, nil
		})

		var cuor nsp.CreateUserOperationRequest
		// from string to []byte
		data, err := hex.DecodeString("1234")
		cuor.Data = data
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

		// OwnerAnyID in string format
		cuor.OwnerAnyID = AnytypeID

		marshalled, err := cuor.MarshalVT()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		cuor_signed.Signature, err = decodedPeerKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		fx.aa.EXPECT().SendUserOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, x interface{}, y interface{}) (operationID string, err error) {
			return "", errors.New("Bad error. Youre doomed")
		})

		// let's go
		// let's go
		identityStr := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		identityBytes, err := crypto.DecodeBytesFromString(identityStr)
		assert.NoError(t, err)
		pctx := peer.CtxWithIdentity(context.Background(), identityBytes)

		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		// 1 - enable user
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		AnytypeID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		//PeerId := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		PeerKey := "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
		//SignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		fx.db.EXPECT().GetUserOperationsCount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, owner common.Address, ownerAnyID string) (uint64, error) {
			return 1, nil
		}).MinTimes(1)

		fx.db.EXPECT().DecreaseUserOperationsCount(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, owner common.Address) error {
			return nil
		}).MinTimes(1)

		fx.db.EXPECT().SaveOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, operationID string, operation nsp.CreateUserOperationRequest) error {
			return nil
		}).MinTimes(1)

		var cuor nsp.CreateUserOperationRequest
		// from string to []byte
		data, err := hex.DecodeString("1234")
		cuor.Data = data
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
		cuor.OwnerAnyID = AnytypeID

		marshalled, err := cuor.MarshalVT()
		assert.NoError(t, err)

		var cuor_signed nsp.CreateUserOperationRequestSigned
		cuor_signed.Payload = marshalled

		cuor_signed.Signature, err = decodedPeerKey.Sign(cuor_signed.Payload)
		assert.NoError(t, err)

		fx.aa.EXPECT().SendUserOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, x interface{}, y interface{}) (operationID string, err error) {
			return "123", nil
		})

		// let's go
		identityStr := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		identityBytes, err := crypto.DecodeBytesFromString(identityStr)
		assert.NoError(t, err)
		pctx := peer.CtxWithIdentity(context.Background(), identityBytes)

		_, err = fx.CreateUserOperation(pctx, &cuor_signed)
		assert.NoError(t, err)
	})
}
