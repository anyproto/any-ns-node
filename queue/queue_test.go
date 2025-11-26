package queue

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

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
	"github.com/anyproto/any-ns-node/nonce_manager"
	mock_nonce_manager "github.com/anyproto/any-ns-node/nonce_manager/mock"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()

func TestAnynsQueue_NameRegisterMoveStateNext(t *testing.T) {
	t.Run("send commit", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		})

		fx.itemColl = nil

		pctx := context.Background()
		newState, err := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				ItemType:        ItemType_NameRegister,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				Status:          OperationStatus_Initial,
			},
		)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_CommitSent, newState)
	})

	t.Run("commit failed", func(t *testing.T) {
		//mt.AddMockResponses(mtest.CreateSuccessResponse())
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
			return false, nil
		}).AnyTimes()

		pctx := context.Background()
		newState, err := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				ItemType:        ItemType_NameRegister,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxCommitHash:    "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// should just wait for tx
				Status: OperationStatus_CommitSent,
			},
		)
		require.Error(t, err)
		require.Equal(t, OperationStatus_CommitError, newState)
	})

	t.Run("register tx failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// error
			return nil, errors.New("error")
		})

		pctx := context.Background()

		newState, err := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				ItemType:        ItemType_NameRegister,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				// should send register tx
				Status: OperationStatus_CommitDone,
			},
		)

		require.Error(t, err)
		require.Equal(t, OperationStatus_RegisterError, newState)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
			return true, nil
		}).AnyTimes()

		fx.itemColl = nil

		pctx := context.Background()
		newState, err := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				ItemType:        ItemType_NameRegister,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxRegisterHash:  "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// wait for register tx
				Status: OperationStatus_RegisterSent,
			},
		)

		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, newState)
	})

	t.Run("success with SpaceID", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
			return true, nil
		}).AnyTimes()

		fx.itemColl = nil

		pctx := context.Background()
		newState, err := fx.NameRegisterMoveStateNext(pctx,
			&QueueItem{
				FullName:        "hello.any",
				ItemType:        ItemType_NameRegister,
				OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
				OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
				TxRegisterHash:  "0x4a8e76e2739c2214eca73b0cfa05d0eb64dcfad0a27c027bf2ecf0ce00110963",
				// wait for register tx
				Status: OperationStatus_RegisterSent,
				// also, SpaceID is attached to
				SpaceId: "bafybeiaysi4s6lnjev27ln5icwm6tueaw2vdykrtjkwiphwekaywqhcjze",
			},
		)

		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, newState)
	})
}

func TestAnynsQueue_SaveItemToDb(t *testing.T) {
	t.Run("fail if item not found", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		// bad index here
		_, err := fx.GetRequestStatus(pctx, 3)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		item := QueueItem{
			Index:           1,
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// save it
		item.Status = OperationStatus_CommitSent
		err = fx.SaveItemToDb(pctx, &item)
		require.NoError(t, err)

		// read status
		s, err := fx.GetRequestStatus(pctx, item.Index)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Pending, s)
	})
}

func TestAnynsQueue_NameRegister(t *testing.T) {
	t.Run("commit failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
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
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.ProcessItem(pctx, &item)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_CommitError, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, nsp.OperationState_Error)
	})

	t.Run("register failed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// error
			return nil, errors.New("error")
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
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
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.ProcessItem(pctx, &item)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_RegisterError, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, nsp.OperationState_Error)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
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
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		_, err = coll.InsertOne(ctx, item)
		require.NoError(t, err)

		// should move through first states
		err = fx.ProcessItem(pctx, &item)
		require.NoError(t, err)

		// read state from DB
		var itemOut QueueItem
		err = coll.FindOne(ctx, findItemByIndexQuery{Index: 1}).Decode(&itemOut)
		require.NoError(t, err)
		require.Equal(t, OperationStatus_Completed, itemOut.Status)

		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, s, nsp.OperationState_Completed)
	})
}

func TestAnynsQueue_FindAndProcessAllItemsInDb(t *testing.T) {
	t.Run("should process all items that stuck in DB", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		fx.contracts.EXPECT().Commit(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			// no error
			var tx = types.NewTransaction(
				0,
				common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
				big.NewInt(0), 0, big.NewInt(0),
				nil,
			)
			return tx, nil
		}).AnyTimes()

		fx.contracts.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (*types.Transaction, error) {
			return nil, errors.New("some error")
		}).AnyTimes()

		fx.contracts.EXPECT().WaitMined(gomock.Any(), gomock.Any()).DoAndReturn(func(interface{}, interface{}) (bool, error) {
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
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_Initial,
		}

		item2 := QueueItem{
			Index:           2,
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_CommitSent,
		}

		item3 := QueueItem{
			Index:           3,
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_CommitDone,
		}

		item4 := QueueItem{
			Index:           4,
			ItemType:        ItemType_NameRegister,
			FullName:        "hello.any",
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyAddress: "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Status:          OperationStatus_RegisterSent,
		}

		// should not process it
		item5 := QueueItem{
			Index:           5,
			ItemType:        ItemType_NameRegister,
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
		err = fx.SaveItemToDb(pctx, &item)
		require.NoError(t, err)

		// process
		fx.FindAndProcessAllItemsInDb(pctx)

		// read status
		s, err := fx.GetRequestStatus(pctx, 1)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 2)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 3)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 4)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Error, s)

		s, err = fx.GetRequestStatus(pctx, 5)
		require.NoError(t, err)
		require.Equal(t, nsp.OperationState_Error, s)
	})
}

func TestAnynsQueue_AddNewRequest(t *testing.T) {
	t.Run("should add new item", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		pctx := context.Background()

		// TODO: mock Mongo!
		uri := "mongodb://localhost:27017"
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		require.NoError(t, err)
		coll := client.Database("any-ns").Collection("queue")

		operationId, err := fx.AddNewRequest(pctx, &nsp.NameRegisterRequest{
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
		operationId, err = fx.AddNewRequest(pctx, &nsp.NameRegisterRequest{
			FullName:        "hello222.any",
			OwnerEthAddress: "0x2225B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		})

		require.NoError(t, err)
		require.Equal(t, int64(1), operationId)
	})
}

type fixture struct {
	a            *app.App
	ctrl         *gomock.Controller
	ts           *rpctest.TestServer
	config       *config.Config
	contracts    *mock_contracts.MockContractsService
	nonceManager *mock_nonce_manager.MockNonceService

	*anynsQueue
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsQueue: New().(*anynsQueue),
	}

	fx.contracts = mock_contracts.NewMockContractsService(fx.ctrl)
	fx.contracts.EXPECT().Name().Return(contracts.CName).AnyTimes()
	fx.contracts.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().CreateEthConnection().AnyTimes()
	fx.contracts.EXPECT().GenerateAuthOptsForAdmin().MaxTimes(2)
	fx.contracts.EXPECT().CalculateTxParams(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().ConnectToPrivateController().AnyTimes()
	fx.contracts.EXPECT().TxByHash(gomock.Any(), gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().MakeCommitment(gomock.Any()).AnyTimes()
	fx.contracts.EXPECT().WaitForTxToStartMining(gomock.Any(), gomock.Any()).AnyTimes()

	fx.nonceManager = mock_nonce_manager.NewMockNonceService(fx.ctrl)
	fx.nonceManager.EXPECT().Init(gomock.Any()).AnyTimes()
	fx.nonceManager.EXPECT().Name().Return(nonce_manager.CName).AnyTimes()

	fx.nonceManager.EXPECT().GetCurrentNonce(gomock.Any()).DoAndReturn(func(interface{}) (uint64, error) {
		return 0, nil
	}).AnyTimes()

	fx.nonceManager.EXPECT().GetCurrentNonceFromNetwork(gomock.Any()).DoAndReturn(func(interface{}) (uint64, error) {
		return 0, nil
	}).AnyTimes()
	fx.nonceManager.EXPECT().SaveNonce(gomock.Any(), gomock.Any()).AnyTimes()

	fx.config.Contracts = config.Contracts{
		AddrAdmin: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
		GethUrl:   "xxx",
	}

	fx.config.Queue = config.Queue{
		SkipProcessing:          false,
		SkipExistingItemsInDB:   true,
		SkipBackroundProcessing: true,
	}

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns",
	}

	fx.a.Register(fx.ts).
		Register(fx.contracts).
		Register(fx.config).
		Register(fx.nonceManager).
		Register(fx.anynsQueue)

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
