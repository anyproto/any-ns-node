package anynsrpc

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

var ctx = context.Background()

func TestIsValidAnyAddress(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		len := len("12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy")
		assert.Equal(t, len, 52)

		valid := []string{
			"12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy", // Anytype address
		}

		for _, address := range valid {
			res := isValidAnyAddress(address)
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
			res := isValidAnyAddress(address)
			assert.Equal(t, res, false)
		}
	})
}

func TestAnynsRpc_IsNameAvailable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		// if this return empty address -> it means address is free
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, client interface{}, namehash interface{}) (common.Address, error) {
			return common.Address{}, nil
		})

		pctx := context.Background()
		resp, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Available)
	})

	t.Run("fail", func(t *testing.T) {
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

		pctx := context.Background()
		resp, err := fx.IsNameAvailable(pctx, &nsp.NameAvailableRequest{
			FullName: "hello.any",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Available)
		assert.Equal(t, resp.OwnerEthAddress, "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		assert.Equal(t, resp.OwnerAnyAddress, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, resp.SpaceId, "")
	})
}

func TestAnynsRpc_GetNameByAddress(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
			return "hello.any", nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("success even if Eth Address is not in lowercase", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
			return "hello.any", nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, resp.Name, "hello.any")
	})

	t.Run("fail", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GetNameByAddress(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, owner interface{}) (string, error) {
			return "", nil
		})

		pctx := context.Background()
		resp, err := fx.GetNameByAddress(pctx, &nsp.NameByAddressRequest{
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Found)
		assert.Equal(t, resp.Name, "")
	})
}

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	*anynsRpc
}

func newFixture(t *testing.T) *fixture {
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

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.anynsRpc)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
