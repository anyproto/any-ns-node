package mongo

import (
	"context"
	"errors"
	"strings"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const CName = "any-ns.db"

var log = logger.NewNamed(CName)

// TODO: index it
type AAUser struct {
	Address         string `bson:"address"`
	AnyID           string `bson:"any_id"`
	OperationsCount uint64 `bson:"operations"`
}

type findAAUserByAddress struct {
	Address string `bson:"address"`
}

// see nameserviceprotoюCreateUserOperationRequest
type AAUserOperation struct {
	OperationID string `bson:"operation_id"`

	Data       []byte `bson:"data"`
	SignedData []byte `bson:"signed_data"`
	Context    []byte `bson:"context"`

	OwnerEthAddress string `bson:"owner_eth_address"`
	OwnerAnyID      string `bson:"owner_any_id"`
	FullName        string `bson:"full_name"`
}

type findUserOperationByID struct {
	OperationID string `bson:"operation_id"`
}

func New() app.Component {
	return &anynsDb{}
}

type DbService interface {
	AddUserToTheWhitelist(ctx context.Context, owner common.Address, ownerAnyID string, newOperations uint64) (err error)

	GetUserOperationsCount(ctx context.Context, owner common.Address, ownerAnyID string) (operations uint64, err error)
	DecreaseUserOperationsCount(ctx context.Context, owner common.Address) (err error)

	SaveOperation(ctx context.Context, opID string, cuor nsp.CreateUserOperationRequest) error
	GetOperation(ctx context.Context, opID string) (op AAUserOperation, err error)

	app.Component
}

type anynsDb struct {
	confMongo config.Mongo

	usersColl *mongo.Collection
	opColl    *mongo.Collection
}

func (arpc *anynsDb) Name() (name string) {
	return CName
}

func (arpc *anynsDb) Init(a *app.App) (err error) {
	arpc.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo

	// connect to mongo
	uri := arpc.confMongo.Connect
	dbName := arpc.confMongo.Database

	// 1 - connect to DB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	arpc.usersColl = client.Database(dbName).Collection("aa-users")
	if arpc.usersColl == nil {
		return errors.New("failed to connect to MongoDB")
	}
	arpc.opColl = client.Database(dbName).Collection("aa-operations")
	if arpc.opColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("mongo connected!")
	return nil
}

func (arpc *anynsDb) Close(ctx context.Context) (err error) {
	if arpc.usersColl != nil {
		err = arpc.usersColl.Database().Client().Disconnect(ctx)
		arpc.usersColl = nil
	}
	if arpc.opColl != nil {
		err = arpc.opColl.Database().Client().Disconnect(ctx)
		arpc.opColl = nil
	}
	return
}

func (arpc *anynsDb) AddUserToTheWhitelist(ctx context.Context, owner common.Address, ownerAnyID string, newOperations uint64) (err error) {
	// TODO: rewrite atomically

	// 1 - verify parameters
	if newOperations == 0 {
		return errors.New("wrong operations count")
	}

	// 2 - get item from mongo
	item := &AAUser{}
	err = arpc.usersColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		// 3.1 - if not found - create new
		if err == mongo.ErrNoDocuments {
			_, err = arpc.usersColl.InsertOne(ctx, AAUser{
				Address:         owner.Hex(),
				AnyID:           ownerAnyID,
				OperationsCount: newOperations,
			})
			if err != nil {
				log.Error("failed to insert item to DB", zap.Error(err))
				return err
			}
			log.Info("added new user to the whitelist", zap.String("owner", owner.Hex()))
			return nil
		}

		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}

	// 3.2 - update item in mongo
	// but first check if Any ID is the same as was passed above
	if ownerAnyID != item.AnyID {
		log.Error("AnyID does not match", zap.String("any_id", ownerAnyID), zap.String("item.AnyID", item.AnyID))
		return errors.New("AnyID does not match")
	}

	log.Debug("increasing operations count in the whitelist", zap.String("owner", owner.Hex()))

	optns := options.Replace().SetUpsert(true)
	item.OperationsCount += newOperations // update operations count

	// 4 - write it back to DB
	_, err = arpc.usersColl.ReplaceOne(ctx, findAAUserByAddress{Address: owner.Hex()}, item, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return err
	}

	log.Info("updated whitelist", zap.String("owner", owner.Hex()))
	return nil
}

// will check if ownerAnyID matches AnyID in the DB (was set by Admin before)
// if ownerAnyID is empty -> do not check it
func (arpc *anynsDb) GetUserOperationsCount(ctx context.Context, owner common.Address, ownerAnyID string) (operations uint64, err error) {
	item := &AAUser{}
	err = arpc.usersColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return 0, err
	}

	// check if AnyID is correct
	// this should be in the format of PeerID - 12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS
	if (ownerAnyID != "") && (item.AnyID != ownerAnyID) {
		log.Error("AnyID does not match", zap.String("any_id", ownerAnyID))
		return 0, errors.New("AnyID does not match")
	}

	return item.OperationsCount, nil
}

func (arpc *anynsDb) DecreaseUserOperationsCount(ctx context.Context, owner common.Address) (err error) {
	// TODO: rewrite atomically

	// 1 - get item from mongo
	item := &AAUser{}
	err = arpc.usersColl.FindOne(ctx, findAAUserByAddress{Address: owner.Hex()}).Decode(&item)

	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}
	if item.OperationsCount == 0 {
		log.Error("operations count is already 0", zap.String("owner", owner.Hex()))
		return errors.New("operations count is already 0")
	}

	// 2 - update item in mongo
	log.Debug("decreasing operations count in the whitelist", zap.String("owner", owner.Hex()))

	optns := options.Replace().SetUpsert(false)
	// update operations count
	item.OperationsCount -= 1

	_, err = arpc.usersColl.ReplaceOne(ctx, findAAUserByAddress{Address: owner.Hex()}, item, optns)
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return err
	}

	log.Info("decreased op count in the whitelist", zap.String("owner", owner.Hex()))
	return nil
}

func (arpc *anynsDb) SaveOperation(ctx context.Context, opID string, cuor nsp.CreateUserOperationRequest) error {
	// 1 - check if operation with this ID already exists
	_, err := arpc.GetOperation(ctx, opID)
	if err == nil {
		log.Error("operation with this ID already exists", zap.String("opID", opID))
		return errors.New("operation with this ID already exists")
	}

	// 2 - create a new AAUserOperation object
	op := &AAUserOperation{
		OperationID: opID,

		Data:            cuor.Data,
		SignedData:      cuor.SignedData,
		Context:         cuor.Context,
		OwnerEthAddress: strings.ToLower(cuor.OwnerEthAddress),
		OwnerAnyID:      cuor.OwnerAnyID,
		FullName:        cuor.FullName,
	}

	_, err = arpc.opColl.InsertOne(ctx, op)
	if err != nil {
		log.Error("failed to save operation to DB", zap.String("opID", opID), zap.Error(err))
		return err
	}

	log.Info("saved operation to DB", zap.String("opID", opID))
	return nil
}

func (arpc *anynsDb) GetOperation(ctx context.Context, opID string) (op AAUserOperation, err error) {
	err = arpc.opColl.FindOne(ctx, findUserOperationByID{OperationID: opID}).Decode(&op)

	if err != nil {
		log.Debug("failed to get operation from DB", zap.String("opID", opID), zap.Error(err))
		return AAUserOperation{}, err
	}

	return op, nil
}
