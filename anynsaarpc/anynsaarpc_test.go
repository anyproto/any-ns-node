package anynsaarpc

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	mock_accountabstraction "github.com/anyproto/any-ns-node/account_abstraction/mock"
	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	as "github.com/anyproto/any-ns-node/pb/anyns_api"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService
	aa        *mock_accountabstraction.MockAccountAbstractionService

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

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}

	fx.aa = mock_accountabstraction.NewMockAccountAbstractionService(fx.ctrl)
	fx.aa.EXPECT().Name().Return(accountabstraction.CName).AnyTimes()
	fx.aa.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.aa).
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

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any()).DoAndReturn(func(ctx interface{}) (address common.Address, err error) {
			return common.Address{}, errors.New("not found")
		})

		pctx := context.Background()
		resp, err := fx.GetUserAccount(pctx, &as.GetUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.Error(t, err, "not found")
		assert.Nil(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any()).DoAndReturn(func(ctx interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().GetNamesCountLeft(gomock.Any()).DoAndReturn(func(ctx interface{}) (count uint64, err error) {
			return uint64(10), nil
		})

		fx.aa.EXPECT().GetOperationsCountLeft(gomock.Any()).DoAndReturn(func(ctx interface{}) (count uint64, err error) {
			return uint64(20), nil
		})

		pctx := context.Background()
		resp, err := fx.GetUserAccount(pctx, &as.GetUserAccountRequest{
			OwnerEthAddress: strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"),
		})

		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress(resp.OwnerEthAddress), common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, common.HexToAddress(resp.OwnerSmartContracWalletAddress), common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))
		assert.Equal(t, resp.NamesCountLeft, uint64(10))
		assert.Equal(t, resp.OperationsCountLeft, uint64(20))
	})
}

func TestAnynsRpc_AdminFundUserAccount(t *testing.T) {

	t.Run("success when asked to add 0 additional name requests", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.aa.EXPECT().GetSmartWalletAddress(gomock.Any()).DoAndReturn(func(ctx interface{}) (address common.Address, err error) {
			return common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), nil
		})

		fx.aa.EXPECT().GetNamesCountLeft(gomock.Any()).DoAndReturn(func(ctx interface{}) (count uint64, err error) {
			return uint64(10), nil
		})

		fx.aa.EXPECT().GetOperationsCountLeft(gomock.Any()).DoAndReturn(func(ctx interface{}) (count uint64, err error) {
			return uint64(20), nil
		})

		fx.aa.EXPECT().VerifyAdminIdentity(gomock.Any(), gomock.Any()).DoAndReturn(func(payload []byte, signature []byte) (err error) {
			// no error, IT IS THE ADMIN!
			return nil
		})

		fx.aa.EXPECT().AdminMintAccessTokens(gomock.Any(), gomock.Any()).DoAndReturn(func(scw common.Address, count *big.Int) (err error) {
			// no error
			return nil
		})

		// create payload
		var in as.AdminFundUserAccountRequestSigned

		accountKeys, err := accountdata.NewRandom()
		require.NoError(t, err)

		// pack
		nrr := as.AdminFundUserAccountRequest{
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

		// should return previous data
		require.NoError(t, err)
		assert.Equal(t, common.HexToAddress(resp.OwnerEthAddress), common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, common.HexToAddress(resp.OwnerSmartContracWalletAddress), common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))
		assert.Equal(t, resp.NamesCountLeft, uint64(10))
		assert.Equal(t, resp.OperationsCountLeft, uint64(20))
	})
}
