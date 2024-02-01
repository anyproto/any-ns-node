package anynsrpc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

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
	cache     *mock_cache.MockCacheService

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

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.cache = mock_cache.NewMockCacheService(fx.ctrl)
	fx.cache.EXPECT().Name().Return(cache.CName).AnyTimes()
	fx.cache.EXPECT().Init(gomock.Any()).AnyTimes()

	// read only from cache (mongo)
	// by default should be true
	fx.config.ReadFromCache = readFromCache

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.contracts).
		Register(fx.config).
		Register(fx.cache).
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
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
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
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
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
