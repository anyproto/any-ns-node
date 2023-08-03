package queue

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"

	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var ctx = context.Background()

func TestAnynsQueue_RegisterName(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("send commit", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		})

		pctx := context.Background()
		err, newState := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Status:          OperationStatus_Initial,
			},
			nil,
			nil,
		)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_CommitSent, newState)
	})

	mt.Run("commit failed", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			return false, nil
		}).AnyTimes()

		pctx := context.Background()
		err, newState := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxCommitHash:    "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// should just wait for tx
				Status: OperationStatus_CommitSent,
			},
			nil,
			nil,
		)
		require.Error(t, err)
		require.Equal(t, OperationStatus_CommitError, newState)
	})

	t.Run("register tx failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// error
			return nil, errors.New("error")
		})

		pctx := context.Background()

		err, newState := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				// should send register tx
				Status: OperationStatus_CommitDone,
			},
			nil,
			nil,
		)

		require.Error(t, err)
		require.Equal(t, OperationStatus_RegisterError, newState)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			return true, nil
		}).AnyTimes()

		pctx := context.Background()
		err, newState := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxRegisterHash:  "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// wait for register tx
				Status: OperationStatus_RegisterSent,
			},
			nil,
			nil,
		)

		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, newState)
	})

	t.Run("success with SpaceID", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			return true, nil
		}).AnyTimes()

		pctx := context.Background()
		err, newState := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxRegisterHash:  "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// wait for register tx
				Status: OperationStatus_RegisterSent,
				// also, SpaceID is attached to
				SpaceId: "bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjze",
			},
			nil,
			nil,
		)

		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, newState)
	})
}

type fixture struct {
	a         *app.App
	ctrl      *gomock.Controller
	ts        *rpctest.TestServer
	config    *config.Config
	contracts *mock_contracts.MockContractsService

	*anynsQueue
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:          new(app.App),
		ctrl:       gomock.NewController(t),
		ts:         rpctest.NewTestServer(),
		config:     new(config.Config),
		anynsQueue: New().(*anynsQueue),
	}

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin(gomock.Any()).MaxTimes(2)
	fx.contracts.EXPECT().ConnectToController(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f",
	}

	fx.config.Mongo = config.Mongo{
		Connect:    "mongodb://localhost:27017",
		Database:   "any-ns",
		Collection: "queue",
	}

	fx.a.Register(fx.ts).
		Register(fx.contracts).
		Register(fx.config).
		Register(fx.anynsQueue)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
