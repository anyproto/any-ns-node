package anynsrpc

import (
	"context"
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

	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
	"github.com/anyproto/any-ns-node/queue"
	mock_queue "github.com/anyproto/any-ns-node/queue/mock"
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

		fx.queue.EXPECT().GetRequestStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, operationId interface{}) (as.OperationState, error) {
			return as.OperationState_Completed, nil
		})

		pctx := context.Background()
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
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}) (common.Address, error) {
			return common.Address{}, nil
		})

		pctx := context.Background()
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
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}) (common.Address, error) {
			notEmptyAddr := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
			return notEmptyAddr, nil
		})

		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(client interface{}, namehash interface{}, owner interface{}) (string, string, string, *big.Int, error) {
			return "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51", "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", "", big.NewInt(12390243), nil
		})

		pctx := context.Background()
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

	t.Run("bad names", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.queue.EXPECT().AddNewRequest(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, req interface{}) (int64, error) {
			return 1, nil
		}).AnyTimes()

		// 1 - bad names
		var arrayOfBadNames = []string{
			"hello",          // no extension
			"somename/hello", // multi-level is not allowed
			"xx.any",         // too short
			"somename.hello", // bad TLD
		}

		// for-each this array
		for _, badName := range arrayOfBadNames {
			pctx := context.Background()
			_, err := fx.NameRegister(pctx, &as.NameRegisterRequest{
				FullName:        badName,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			})

			require.Error(t, err)
		}
	})

	t.Run("bad eth address", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var arrayOfBadEthAddresses = []string{
			"", // no address
			"0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a5",   // too short
			"0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a511", // too long
			"0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a5g",  // bad symbol
		}

		for _, badEthAddress := range arrayOfBadEthAddresses {
			pctx := context.Background()
			_, err := fx.NameRegister(pctx, &as.NameRegisterRequest{
				FullName:        "coolName.any",
				OwnerEthAddress: badEthAddress,
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			})

			require.Error(t, err)
		}
	})

	t.Run("bad any address", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 3 - bad Any address
		var arrayOfBadAnyAddresses = []string{
			"", // no address
			"12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSu",   // too short
			"12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuSS", // too long
			"12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuSg", // bad symbol
		}

		// for-each this array
		for _, badAnyAddress := range arrayOfBadAnyAddresses {
			pctx := context.Background()
			_, err := fx.NameRegister(pctx, &as.NameRegisterRequest{
				FullName:        "coolName.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: badAnyAddress,
			})

			require.Error(t, err)
		}
	})

	t.Run("bad space ID", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var arrayOfBadSpaces = []string{
			"bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjz",   // too short
			"bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjzee", // too long
			"АВФbafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhc",  // bad symbols
		}

		// for-each this array
		for _, badSpace := range arrayOfBadSpaces {
			pctx := context.Background()
			_, err := fx.NameRegister(pctx, &as.NameRegisterRequest{
				FullName:        "coolName.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				SpaceId:         badSpace,
			})

			require.Error(t, err)
		}
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
		resp, err := fx.GetNameByAddress(pctx, &as.NameByAddressRequest{
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
		resp, err := fx.GetNameByAddress(pctx, &as.NameByAddressRequest{
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
		resp, err := fx.GetNameByAddress(pctx, &as.NameByAddressRequest{
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
	queue     *mock_queue.MockQueueService

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

	fx.queue = mock_queue.NewMockQueueService(fx.ctrl)
	fx.queue.EXPECT().Name().Return(queue.CName).AnyTimes()
	fx.queue.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().Run(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().Close(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().FindAndProcessAllItemsInDb(gomock.Any()).AnyTimes()
	fx.queue.EXPECT().FindAndProcessAllItemsInDbWithStatus(gomock.Any(), gomock.Any()).AnyTimes()
	fx.queue.EXPECT().ProcessItem(gomock.Any(), gomock.Any()).AnyTimes()
	fx.queue.EXPECT().SaveItemToDb(gomock.Any(), gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.queue).
		Register(fx.anynsRpc)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
