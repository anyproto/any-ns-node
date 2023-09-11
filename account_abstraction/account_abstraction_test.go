package accountabstraction

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"

	as "github.com/anyproto/any-ns-node/pb/anyns_api"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var ctx = context.Background()

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	*anynsAA
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsAA: New().(*anynsAA),
	}

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}
	fx.config.Account.PeerId = "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
	fx.config.Account.PeerKey = "psqF8Rj52Ci6gsUl5ttwBVhINTP8Yowc2hea73MeFm4Ek9AxedYSB4+r7DYCclDL4WmLggj2caNapFUmsMtn5Q=="
	fx.config.Account.SigningKey = "3MFdA66xRw9PbCWlfa620980P4QccXehFlABnyJ/tfwHbtBVHt+KWuXOfyWSF63Ngi70m+gcWtPAcW5fxCwgVg=="

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin(gomock.Any()).MaxTimes(2)
	fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().ConnectToController(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().WaitForTxToStartMining(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	fx.a.Register(fx.ts).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.anynsAA)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestAAS_GetSmartWalletAddress(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if can not connect to smart contract", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		const factoryAddress = "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"
		factoryAddressBytes := common.HexToAddress(factoryAddress).Bytes()
		assert.Equal(t, factoryAddressBytes[0], byte(0x5F))

		fa := common.BytesToAddress(factoryAddressBytes)
		assert.Equal(t, factoryAddress, fa.String())

		fx.contracts.EXPECT().CallContract(gomock.Any(), gomock.Any()).DoAndReturn(func(tokenAddress interface{}, scw interface{}) ([]byte, error) {
			return nil, errors.New("fail")
		})

		pctx := context.Background()
		_, err := fx.GetSmartWalletAddress(pctx, common.HexToAddress("0xE34230c1f916e9d628D5F9863Eb3F5667D8FcB37"))
		assert.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CallContract(gomock.Any(), gomock.Any()).DoAndReturn(func(tokenAddress interface{}, scw interface{}) ([]byte, error) {
			byteArr := common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a").Bytes()
			return byteArr, nil
		})

		pctx := context.Background()
		swa, err := fx.GetSmartWalletAddress(pctx, common.HexToAddress("0xE34230c1f916e9d628D5F9863Eb3F5667D8FcB37"))

		assert.NoError(t, err)
		assert.Equal(t, common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"), swa)
	})
}

func TestAAS_GetNonceForSmartWalletAddress(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if can not connect to smart contract", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CallContract(gomock.Any(), gomock.Any()).DoAndReturn(func(tokenAddress interface{}, scw interface{}) ([]byte, error) {
			return nil, errors.New("fail")
		})

		pctx := context.Background()
		_, err := fx.GetNonceForSmartWalletAddress(pctx, common.HexToAddress("0xE34230c1f916e9d628D5F9863Eb3F5667D8FcB37"))
		assert.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().CallContract(gomock.Any(), gomock.Any()).DoAndReturn(func(tokenAddress interface{}, scw interface{}) ([]byte, error) {
			out := big.NewInt(6)

			byteArr := out.Bytes()
			return byteArr, nil
		})

		pctx := context.Background()
		nonce, err := fx.GetNonceForSmartWalletAddress(pctx, common.HexToAddress("0xE34230c1f916e9d628D5F9863Eb3F5667D8FcB37"))

		assert.NoError(t, err)
		six := big.NewInt(6)
		assert.Equal(t, nonce, six)
	})
}

func TestAAS_GetNamesCountLeft(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if can not get token balance", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, tokenAddress interface{}, scw interface{}) (*big.Int, error) {
			return big.NewInt(0), errors.New("failed to get balance")
		})

		count, err := fx.GetNamesCountLeft(common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))

		assert.Error(t, err)
		assert.Equal(t, uint64(0), count)
	})

	mt.Run("success if no tokens", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, tokenAddress interface{}, scw interface{}) (*big.Int, error) {
			return big.NewInt(0), nil
		})

		count, err := fx.GetNamesCountLeft(common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))

		assert.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})

	mt.Run("success if not enough tokens", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, tokenAddress interface{}, scw interface{}) (*big.Int, error) {
			// $20 USD per name (current testnet settings)
			oneNamePriceWei := big.NewInt(20 * 1000000)

			// divide oneNamePriceWei /2 to get less than 1 name
			out := big.NewInt(0).Div(oneNamePriceWei, big.NewInt(2))
			return out, nil
		})

		count, err := fx.GetNamesCountLeft(common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))

		assert.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})

	mt.Run("success if got N tokens", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetBalanceOf(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx interface{}, tokenAddress interface{}, scw interface{}) (*big.Int, error) {
			oneNamePriceWei := big.NewInt(20 * 1000000)

			// multiply by 12
			out := big.NewInt(0).Mul(oneNamePriceWei, big.NewInt(12))
			return out, nil
		})

		count, err := fx.GetNamesCountLeft(common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))

		assert.NoError(t, err)
		assert.Equal(t, uint64(12), count)
	})
}

func TestAAS_GetOperationsCountLeft(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		count, err := fx.GetOperationsCountLeft(common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a"))

		assert.NoError(t, err)
		assert.Equal(t, uint64(0), count)
	})
}

func TestAAS_VerifyAdminIdentity(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		// 0 - garbage data test
		err := fx.VerifyAdminIdentity([]byte("payload"), []byte("signature"))
		assert.Error(t, err)

		// 1 - pack some structure
		nrr := as.AdminFundUserAccountRequest{
			OwnerEthAddress: "",
			NamesCount:      0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		// 2 - sign it with some random (wrong) key
		accountKeys, err := accountdata.NewRandom()
		require.NoError(t, err)

		sig, err := accountKeys.SignKey.Sign(marshalled)
		require.NoError(t, err)

		err = fx.VerifyAdminIdentity(marshalled, sig)
		assert.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// 1 - pack some structure
		nrr := as.AdminFundUserAccountRequest{
			OwnerEthAddress: "",
			NamesCount:      0,
		}

		marshalled, err := nrr.Marshal()
		require.NoError(t, err)

		// 2 - sign it
		signKey, err := crypto.DecodeKeyFromString(
			fx.config.Account.PeerKey,
			crypto.UnmarshalEd25519PrivateKey,
			nil)
		require.NoError(t, err)

		sig, err := signKey.Sign(marshalled)
		require.NoError(t, err)

		// get associated pub key
		//pubKey := signKey.GetPublic()
		// identity str
		//identityStr := pubKey.Account()
		// A5ommzwhpR5ngp11q9q1P2MMzhUE46Hi421RJbPqswALyoyr
		//log.Info("identity", zap.String("identity", identityStr))

		err = fx.VerifyAdminIdentity(marshalled, sig)
		assert.NoError(t, err)
	})
}

func TestAAS_MintAccessTokens(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if names count is ZERO", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// already deployed
		scw := common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a")
		err := fx.AdminMintAccessTokens(scw, big.NewInt(0))
		assert.Error(t, err)
	})

	/*
		mt.Run("success", func(mt *mtest.T) {
			fx := newFixture(t)
			defer fx.finish(t)

			// nonce is 5
			fx.contracts.EXPECT().CallContract(gomock.Any(), gomock.Any()).DoAndReturn(func(tokenAddress interface{}, scw interface{}) ([]byte, error) {
				out := big.NewInt(5)

				byteArr := out.Bytes()
				return byteArr, nil
			})

			// TODO: mock alchemyaa.SendRequest and uncomment this test

			// already deployed
			scw := common.HexToAddress("0x77d454b313e9D1Acb8cD0cFa140A27544aEC483a")
			err := fx.AdminMintAccessTokens(scw, big.NewInt(5))
			assert.NoError(t, err)
		})
	*/
}
