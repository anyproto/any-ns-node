package anynsrpc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-sync/nodeconf/mock_nodeconf"

	"github.com/anyproto/any-ns-node/cache"
	mock_cache "github.com/anyproto/any-ns-node/cache/mock"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	db_service "github.com/anyproto/any-ns-node/db"
	mock_db "github.com/anyproto/any-ns-node/db/mock"
	"github.com/anyproto/any-ns-node/queue"
	mock_queue "github.com/anyproto/any-ns-node/queue/mock"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	mock_accountabstraction "github.com/anyproto/any-ns-node/account_abstraction/mock"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	nodeConf  *mock_nodeconf.MockService
	contracts *mock_contracts.MockContractsService
	cache     *mock_cache.MockCacheService
	queue     *mock_queue.MockQueueService
	aa        *mock_accountabstraction.MockAccountAbstractionService
	db        *mock_db.MockDbService

	*anynsRpc
}

func newFixture(t *testing.T, readFromCache bool) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsRpc: New().(*anynsRpc),
	}

	fx.nodeConf = mock_nodeconf.NewMockService(fx.ctrl)
	fx.nodeConf.EXPECT().Name().Return(nodeconf.CName).AnyTimes()
	fx.nodeConf.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().Close(gomock.Any()).AnyTimes()
	fx.nodeConf.EXPECT().NodeTypes(gomock.Any()).Return([]nodeconf.NodeType{nodeconf.NodeTypeConsensus}).AnyTimes()

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.cache = mock_cache.NewMockCacheService(fx.ctrl)
	fx.cache.EXPECT().Name().Return(cache.CName).AnyTimes()
	fx.cache.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.queue = mock_queue.NewMockQueueService(fx.ctrl)
	fx.queue.EXPECT().Name().Return(queue.CName).AnyTimes()
	fx.queue.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().Close(gomock.Any()).AnyTimes()

	fx.aa = mock_accountabstraction.NewMockAccountAbstractionService(fx.ctrl)
	fx.aa.EXPECT().Name().Return(accountabstraction.CName).AnyTimes()
	fx.aa.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.db = mock_db.NewMockDbService(fx.ctrl)
	fx.db.EXPECT().Name().Return(db_service.CName).AnyTimes()
	fx.db.EXPECT().Init(gomock.Any()).AnyTimes()

	// read only from cache (mongo)
	// by default should be true
	fx.config.ReadFromCache = readFromCache

	adminSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

	fx.config.Account = accountservice.Config{
		SigningKey: adminSignKey,
		PeerKey:    "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q==",
	}

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.contracts).
		Register(fx.config).
		Register(fx.cache).
		Register(fx.queue).
		Register(fx.nodeConf).
		Register(fx.aa).
		Register(fx.db).
		Register(fx.anynsRpc)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestAnynsRpc_IsNameAvailable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameAvailableResponse, error) {
			return &nsp.NameAvailableResponse{
				// free
				Available: true,
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Available)
	})

	t.Run("success if reading from smart contracts", func(t *testing.T) {
		// see here >
		readFromCache := false

		fx := newFixture(t, readFromCache)
		defer fx.finish(t)

		fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (err error) {
			return nil
		})

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameAvailableResponse, error) {
			return &nsp.NameAvailableResponse{
				// free
				Available: true,
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Available)
	})

	t.Run("fail if reading from smart contracts failed", func(t *testing.T) {
		// see here >
		readFromCache := false
		fx := newFixture(t, readFromCache)
		defer fx.finish(t)

		fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, nar *nsp.NameAvailableRequest) (err error) {
			// see here >
			return errors.New("failed to update in cache")
		})

		pctx := context.Background()
		_, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.Error(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().UpdateInCache(gomock.Any(), gomock.Any()).MaxTimes(0)

		fx.cache.EXPECT().IsNameAvailable(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameAvailableResponse, error) {
			return &nsp.NameAvailableResponse{
				// occupied
				Available: false,
				// owner
				OwnerScwEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress:    "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Available)
		assert.Equal(t, resp.OwnerScwEthAddress, "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		assert.Equal(t, resp.OwnerAnyAddress, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, resp.SpaceId, "")
	})
}

func TestAnynsRpc_GetNameByAddress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return &nsp.NameByAddressResponse{
				Found: true,
				Name:  "hello.any",
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("success even if Eth Address is not in lowercase", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return &nsp.NameByAddressResponse{
				Found: true,
				Name:  "hello.any",
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("success with no cache (direct request to smart contracts)", func(t *testing.T) {
		// see here >
		readFromCache := false

		fx := newFixture(t, readFromCache)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any()).DoAndReturn(func(owner interface{}) (string, error) {
			return "hello.any", nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("fail if smart contract call failed", func(t *testing.T) {
		// see here >
		readFromCache := false

		fx := newFixture(t, readFromCache)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any()).DoAndReturn(func(owner interface{}) (string, error) {
			return "", errors.New("failed to get name by address")
		})

		pctx := context.Background()
		_, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.Error(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		/*
			fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
			fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
				return "", nil
			})
		*/
		fx.cache.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return &nsp.NameByAddressResponse{
				Found: false,
				Name:  "",
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerScwEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Found)
		assert.Equal(t, resp.Name, "")
	})
}

func TestAnynsRpc_GetNameByAnyId(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().GetNameByAnyId(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return &nsp.NameByAddressResponse{
				Found: true,
				Name:  "hello.any",
			}, nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAnyId(pctx, &nsp.NameByAnyIdRequest{
			AnyAddress: "A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("fail if DB call failed", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().GetNameByAnyId(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return nil, errors.New("failed to get item from DB")
		})

		pctx := context.Background()
		_, err := fx.GetNameByAnyId(pctx, &nsp.NameByAnyIdRequest{
			AnyAddress: "A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4",
		})

		require.Error(t, err)
	})

	t.Run("return false if not in DB", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.cache.EXPECT().GetNameByAnyId(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (*nsp.NameByAddressResponse, error) {
			return &nsp.NameByAddressResponse{Found: false}, nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAnyId(pctx, &nsp.NameByAnyIdRequest{
			AnyAddress: "A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Found)
		assert.Equal(t, resp.Name, "")
	})
}

func TestAnynsRpc_AdminNameRegisterSigned(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, true)
		defer fx.finish(t)

		fx.aa.EXPECT().AdminNameRegister(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, in interface{}) (string, error) {
			return "operation-id", nil
		}).MinTimes(1)

		fx.db.EXPECT().SaveOperation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, operationID string, operation nsp.CreateUserOperationRequest) error {
			return nil
		}).MinTimes(1)

		// OwnerAnyID
		AnytypeID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"
		PeerID := "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		realSignKey := "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

		decodedPeerKey, err := crypto.DecodeKeyFromString(
			realSignKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		assert.NoError(t, err)

		// OwnerAnyID in string format
		req := nsp.NameRegisterRequest{}
		req.OwnerAnyAddress = AnytypeID
		req.FullName = "hello.any"
		req.OwnerEthAddress = "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"

		nrrs := nsp.NameRegisterRequestSigned{}
		nrrs.Payload, err = req.MarshalVT()
		require.NoError(t, err)

		nrrs.Signature, err = decodedPeerKey.Sign(nrrs.Payload)
		assert.NoError(t, err)

		// call it
		pctx := peer.CtxWithPeerId(context.Background(), PeerID)
		resp, err := fx.AdminNameRegisterSigned(pctx, &nrrs)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "operation-id", resp.OperationId)
	})
}
