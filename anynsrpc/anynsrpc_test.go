package anynsrpc

import (
	"context"
	"math/big"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/anyproto/any-sync/testutil/accounttest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/anyns-node/config"
	contracts "github.com/anyproto/anyns-node/contracts"
	mock_contracts "github.com/anyproto/anyns-node/contracts/mock"
	as "github.com/anyproto/anyns-node/pb/anyns_api_server"
)

var ctx = context.Background()

func TestVerifyIdentity_IdentityIsOK(t *testing.T) {
	var in as.NameRegisterSignedRequest

	accountKeys, err := accountdata.NewRandom()
	require.NoError(t, err)

	identity, err := accountKeys.SignKey.GetPublic().Marshall()
	require.NoError(t, err)

	// pack
	nrr := as.NameRegisterRequest{
		OwnerAnyAddress: string(identity),
		OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		FullName:        "hello.any",
		SpaceId:         "",
	}

	marshalled, err := nrr.Marshal()
	require.NoError(t, err)

	in.Payload = marshalled
	in.Signature, err = accountKeys.SignKey.Sign(in.Payload)
	require.NoError(t, err)

	// run
	err = VerifyIdentity(&in, nrr.OwnerAnyAddress)
	require.NoError(t, err)
}

func TestVerifyIdentity_IdentityIsBad(t *testing.T) {
	var in as.NameRegisterSignedRequest

	accountKeys, err := accountdata.NewRandom()
	require.NoError(t, err)

	accountKeys2, err := accountdata.NewRandom()
	require.NoError(t, err)

	identity2, err := accountKeys2.SignKey.GetPublic().Marshall()
	require.NoError(t, err)

	// pack
	nrr := as.NameRegisterRequest{
		// DIFFERENT!
		OwnerAnyAddress: string(identity2),
		OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		FullName:        "hello.any",
		SpaceId:         "",
	}

	marshalled, err := nrr.Marshal()
	require.NoError(t, err)

	in.Payload = marshalled
	in.Signature, err = accountKeys.SignKey.Sign(in.Payload)
	require.NoError(t, err)

	// run
	err = VerifyIdentity(&in, nrr.OwnerAnyAddress)
	require.Error(t, err)
}

func TestAnynsRpc_GetOperationStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := peer.CtxWithPeerId(ctx, "peerId")
		resp, err := fx.GetOperationStatus(pctx, &as.GetOperationStatusRequest{
			OperationId: 1,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)

		// this always returns completed even if operation was never created
		assert.Equal(t, resp.OperationId, uint64(1))
		assert.Equal(t, resp.OperationState, as.OperationState_Completed)
	})
}

func TestAnynsRpc_IsNameAvailable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		// if this return empty address -> it means address is free
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}) (*common.Address, error) {
			return &common.Address{}, nil
		})

		pctx := peer.CtxWithPeerId(ctx, "peerId")
		resp, err := fx.IsNameAvailable(pctx, &as.NameAvailableRequest{
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
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}) (*common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return &notEmptyAddr, nil
		})
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}, owner interface{}) (string, string, string, error) {
			return "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", nil
		})

		pctx := peer.CtxWithPeerId(ctx, "peerId")
		resp, err := fx.IsNameAvailable(pctx, &as.NameAvailableRequest{
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

func TestAnynsRpc_RegisterName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
		fx.contracts.EXPECT().GenerateAuthOptsForAdmin(gomock.Any()).MaxTimes(2)
		fx.contracts.EXPECT().ConnectToController(gomock.Any()).AnyTimes()
		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			var emptyTx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)

			return emptyTx, nil
		})

		fx.contracts.EXPECT().CheckTransactionReceipt(gomock.Any(), gomock.Any()).AnyTimes()
		fx.contracts.EXPECT().MakeCommitment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		/*
			// if this returns some address -> it means name is taken
			fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}) (*common.Address, error) {
				notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
				return &notEmptyAddr, nil
			})
			fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}, owner interface{}) (string, string, string, error) {
				return "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", nil
			})
		*/

		pctx := peer.CtxWithPeerId(ctx, "peerId")
		resp, err := fx.NameRegister(pctx, &as.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(t, resp.OperationId, uint64(1))
		assert.Equal(t, resp.OperationState, as.OperationState_Completed)
	})
}

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockService

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

	fx.contracts = mock_contracts.NewMockService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
		// TODO:
	}

	fx.a.Register(fx.ts).
		Register(&accounttest.AccountTestService{}).
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
