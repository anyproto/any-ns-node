package nonce_manager

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"
)

var ctx = context.Background()

func TestNonceManager_GetCurrentNonce(t *testing.T) {
	t.Run("get nonce from DB if present", func(t *testing.T) {
		fx := newFixture(t, 0)
		defer fx.finish(t)

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("nonce")

		dbItem := &NonceDbItem{
			Address: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			Nonce:   12,
		}

		// save to DB first
		optns := options.Replace().SetUpsert(true)
		_, err = coll.ReplaceOne(ctx, findNonceByAddress{Address: dbItem.Address}, dbItem, optns)
		require.NoError(t, err)

		// get from DB
		nonce, err := fx.GetCurrentNonce(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		require.NoError(t, err)
		require.Equal(t, uint64(12), nonce)
	})

	t.Run("get nonce from network if not present in DB", func(t *testing.T) {
		fx := newFixture(t, 0)
		defer fx.finish(t)

		fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*big.Int, uint64, error) {
			// return "nonce from network"
			return nil, 15, nil
		})

		// get from DB
		nonce, err := fx.GetCurrentNonce(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		require.NoError(t, err)
		require.Equal(t, uint64(15), nonce)
	})

	t.Run("get nonce from config if override param is present", func(t *testing.T) {
		fx := newFixture(t, 19)
		defer fx.finish(t)

		// get from DB
		nonce, err := fx.GetCurrentNonce(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		require.NoError(t, err)
		require.Equal(t, uint64(19), nonce)
	})
}

func TestNonceManager_GetCurrentNonceFromNetwork(t *testing.T) {
	t.Run("should get nonce from network", func(t *testing.T) {
		fx := newFixture(t, 20)
		defer fx.finish(t)

		fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*big.Int, uint64, error) {
			// return "nonce from network"
			return nil, 15, nil
		})

		// get from DB
		nonce, err := fx.GetCurrentNonceFromNetwork(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		require.NoError(t, err)
		require.Equal(t, uint64(15), nonce)
	})
}

func TestNonceManager_SaveNonce(t *testing.T) {
	t.Run("should save to DB", func(t *testing.T) {
		fx := newFixture(t, 0)
		defer fx.finish(t)

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("nonce")

		// get from DB
		newNonce, err := fx.SaveNonce(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"), 18)
		require.NoError(t, err)
		require.Equal(t, uint64(18), newNonce)

		// get from DB
		// convert string to ethcommon.Address
		nonce, err := fx.GetCurrentNonce(common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		require.NoError(t, err)
		require.Equal(t, uint64(18), nonce)

		// check in DB
		ctx := context.Background()
		adminAddr := "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"

		dbItem := &NonceDbItem{}

		err = coll.FindOne(ctx, findNonceByAddress{Address: adminAddr}).Decode(&dbItem)
		require.NoError(t, err)
		require.Equal(t, dbItem.Nonce, int64(18))
	})
}

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	*anynsNonceService
}

func newFixture(t *testing.T, nonceOverride uint64) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsNonceService: New().(*anynsNonceService),
	}

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns",
	}

	fx.config.Nonce = config.Nonce{
		NonceOverride: nonceOverride,
	}

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin().MaxTimes(2)
	fx.contracts.EXPECT().ConnectToPrivateController().AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().WaitForTxToStartMining(gomock.Any(), gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "xxx",
	}

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.contracts).
		Register(fx.anynsNonceService)

	require.NoError(t, fx.a.Start(ctx))

	// TODO: mock Mongo!
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	// drop database any-ns
	err = client.Database("any-ns").Drop(ctx)
	if err != nil {
		// sleep 1 second
		time.Sleep(1 * time.Second)
	}

	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
