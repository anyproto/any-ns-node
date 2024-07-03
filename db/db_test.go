package mongo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/net/rpc/rpctest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/mock/gomock"
)

var ctx = context.Background()

type fixture struct {
	a      *app.App
	ctrl   *gomock.Controller
	ts     *rpctest.TestServer
	config *config.Config

	*anynsDb
}

func newFixture(t *testing.T, adminSignKey string) *fixture {
	fx := &fixture{
		a:      new(app.App),
		ctrl:   gomock.NewController(t),
		ts:     rpctest.NewTestServer(),
		config: new(config.Config),

		anynsDb: New().(*anynsDb),
	}

	fx.config.Mongo = config.Mongo{
		Connect:  "mongodb://localhost:27017",
		Database: "any-ns-test",
	}

	fx.config.Account = accountservice.Config{
		SigningKey: adminSignKey,
		PeerKey:    "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
	}

	// drop everything
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fx.config.Mongo.Connect))
	require.NoError(t, err)

	err = client.Database(fx.config.Mongo.Database).Drop(ctx)
	if err != nil {
		// sleep 1 second
		time.Sleep(1 * time.Second)
	}

	fx.a.Register(fx.ts).
		// this generates new random account every Init
		// Register(&accounttest.AccountTestService{}).
		Register(fx.config).
		Register(fx.anynsDb)

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestAnynsRpc_MongoAddUserToTheWhitelist(t *testing.T) {
	t.Run("fail if wrong operations count", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 0)
		assert.Error(t, err)
	})

	t.Run("success if user never existed", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// TODO: mock!
		// read from Mongo and check
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-users"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUser{}
		err = itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.Address, owner.Hex())
		assert.Equal(t, item.OperationsCount, uint64(1))
	})

	t.Run("fail if user has existed before but Any ID is different", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// again but with different Any ID
		anyIDDifferent := "12D3KaaXB5EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
		err = fx.AddUserToTheWhitelist(pctx, owner, anyIDDifferent, 1)
		assert.Error(t, err)
	})

	t.Run("success if user has existed before", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// again!
		err = fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// TODO: mock!
		// read from Mongo and check
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-users"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUser{}
		err = itemColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.Address, owner.Hex())
		assert.Equal(t, item.OperationsCount, uint64(2))
	})
}

func TestAnynsRpc_MongoGetUserOperationsCount(t *testing.T) {
	t.Run("fail if nothing is found operations count", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		ops, err := fx.GetUserOperationsCount(pctx, owner, "")
		assert.Error(t, err)
		assert.Equal(t, ops, uint64(0))
	})

	t.Run("fail if wrong Any ID", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.AddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.GetUserOperationsCount(pctx, owner, "1111KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Error(t, err)
		assert.Equal(t, ops, uint64(0))
	})

	t.Run("success if AnyID is empty (do not compare it!)", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.AddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.GetUserOperationsCount(pctx, owner, "")
		assert.NoError(t, err)
		assert.Equal(t, ops, uint64(1))
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.AddUserToTheWhitelist(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS", 1)
		assert.NoError(t, err)

		ops, err := fx.GetUserOperationsCount(pctx, owner, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.NoError(t, err)
		assert.Equal(t, ops, uint64(1))
	})
}

func TestAnynsRpc_MongoDecreaseUserOperationsCount(t *testing.T) {
	t.Run("fail if user is not found", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")

		err := fx.DecreaseUserOperationsCount(pctx, owner)
		assert.Error(t, err)
	})

	t.Run("fail if user has no operations already", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		// 1st time - should work
		err = fx.DecreaseUserOperationsCount(pctx, owner)
		assert.NoError(t, err)

		// 2nd time - should fail
		err = fx.DecreaseUserOperationsCount(pctx, owner)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()
		owner := common.HexToAddress("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51")
		anyID := "A5k2d9sFZw84yisTxRnz2bPRd1YPfVfhxqymZ6yESprFTG65"

		err := fx.AddUserToTheWhitelist(pctx, owner, anyID, 1)
		assert.NoError(t, err)

		err = fx.DecreaseUserOperationsCount(pctx, owner)
		assert.NoError(t, err)
	})
}

func TestAnynsRpc_MongoSaveOperation(t *testing.T) {
	t.Run("should create new item", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()

		err := fx.SaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		assert.NoError(t, err)

		// check in the DB
		uri := fx.config.Mongo.Connect
		dbName := fx.config.Mongo.Database
		collectionName := "aa-operations"

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		assert.NoError(t, err)

		var itemColl *mongo.Collection = client.Database(dbName).Collection(collectionName)
		item := &AAUserOperation{}
		err = itemColl.FindOne(ctx, findUserOperationByID{OperationID: "123"}).Decode(&item)
		assert.NoError(t, err)

		assert.Equal(t, item.OperationID, "123")
		assert.Equal(t, item.OwnerEthAddress, strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, item.OwnerAnyID, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, item.Data, []byte("data"))
		assert.Equal(t, item.FullName, "hello.any")
	})

	t.Run("should not update existing item", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		pctx := context.Background()

		err := fx.SaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
		})
		assert.NoError(t, err)

		err = fx.SaveOperation(pctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
		})
		assert.Error(t, err)
	})
}

func TestAnynsRpc_MongoGetOpertaion(t *testing.T) {
	t.Run("should return error if not found", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		_, err := fx.GetOperation(ctx, "123")
		assert.Error(t, err)
	})

	t.Run("should return item", func(t *testing.T) {
		fx := newFixture(t, "")
		defer fx.finish(t)

		err := fx.SaveOperation(ctx, "123", nsp.CreateUserOperationRequest{
			// should be converted to lower case
			OwnerEthAddress: "0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51",
			OwnerAnyID:      "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS",
			Data:            []byte("data"),
			SignedData:      []byte("signed_data"),
			Context:         []byte("context"),
			FullName:        "hello.any",
		})
		assert.NoError(t, err)

		item, err := fx.GetOperation(ctx, "123")
		assert.NoError(t, err)
		assert.Equal(t, item.OperationID, "123")
		assert.Equal(t, item.OwnerEthAddress, strings.ToLower("0x10d5B0e279E5E4c1d1Df5F57DFB7E84813920a51"))
		assert.Equal(t, item.OwnerAnyID, "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS")
		assert.Equal(t, item.Data, []byte("data"))
		assert.Equal(t, item.FullName, "hello.any")
	})
}
