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
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()

func TestAnynsQueue_NameRegisterMoveStateNext(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("send commit", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

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

func TestAnynsQueue_SaveItemToDb(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if item not found", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		pctx := context.Background()

		// bad index here
		_, err := fx.GetRequestStatus(pctx, 3)
		require.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// save it
		item.Status = OperationStatus_CommitSent
		err = fx.SaveItemToDb(pctx, coll,
			&item,
		)
		require.NoError(t, err)

		// read status
		s, err := fx.GetRequestStatus(pctx, item.Index)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Pending, s)
	})
}

func TestAnynsQueue_NameRegister(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("commit failed", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			return false, nil
		}).AnyTimes()

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.NameRegister(pctx, &item, coll)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_CommitError, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, as.OperationState_Error)
	})

	mt.Run("register failed", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// error
			return nil, errors.New("error")
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			// good
			return true, nil
		}).AnyTimes()

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.NameRegister(pctx, &item, coll)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_RegisterError, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, as.OperationState_Error)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			// good
			return true, nil
		}).AnyTimes()

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.NameRegister(pctx, &item, coll)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, as.OperationState_Completed)
	})
}

func TestAnynsQueue_FindAndProcessAllItemsInDb(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should process all items that stuck in DB", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}, interface{}) (bool, error) {
			// fail
			return false, nil
		}).AnyTimes()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		item2 := QueueItem{
			Index:           2,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_CommitSent,
		}

		item3 := QueueItem{
			Index:           3,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_CommitDone,
		}

		item4 := QueueItem{
			Index:           4,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_RegisterSent,
		}

		// should not process it
		item5 := QueueItem{
			Index:           5,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_CommitError,
		}

		// create array of items
		items := []interface{}{item, item2, item3, item4, item5}

		// add all items to DB
		_, err = coll.InsertMany(ctx, items)
		require.NoError(t, err)

		// save it
		item.Status = OperationStatus_CommitSent
		err = fx.SaveItemToDb(pctx, coll,
			&item,
		)
		require.NoError(t, err)

		// process
		fx.FindAndProcessAllItemsInDb(pctx, coll)

		// read status
		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 2)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 3)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 4)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 5)
		require.NoError(t, err)
		require.Equal(t, as.OperationState_Error, s)
	})
}

func TestAnynsQueue_AddNewRequest(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should add new item", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		operationId, err := fx.AddNewRequest(pctx, &as.NameRegisterRequest{
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		require.Equal(t, int64(0), operationId)

		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 0}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, int64(0), itemOut.Index)
		require.Equal(t, OperationStatus_Initial, itemOut.Status)
		require.NotEmpty(t, itemOut.SecretBase64)

		// should add another one too
		operationId, err = fx.AddNewRequest(pctx, &as.NameRegisterRequest{
			FullName:        "hello222.any",
			OwnerEthAddress: "0x2225B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		require.Equal(t, int64(1), operationId)
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

	fx.config.Queue = config.Queue{
		SkipProcessing:          false,
		SkipExistingItemsInDB:   true,
		SkipBackroundProcessing: true,
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

	// TODO: mock Mongo!
	uri := "mongodb://localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	// drop database any-ns
	err = client.Database("any-ns").Drop(ctx)
	require.NoError(t, err)

	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}
