package queue

import (
	"context"
	"fmt"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cheggaaa/mb/v3"
)

const CName = "any-ns.queue"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsQueue{}
}

type QueueService interface {
	// add new request to the queue and process it in the background
	ProcessRequest(ctx context.Context, req *as.NameRegisterRequest) (operationId int64, err error)

	GetRequestStatus(ctx context.Context, operationId int64) (status as.OperationState, err error)

	NameRegister(ctx context.Context, in *as.NameRegisterRequest) error

	app.ComponentRunnable
}

type anynsQueue struct {
	q        *mb.MB[int64]
	conf     config.Mongo
	itemColl *mongo.Collection
	done     chan bool

	contractsConfig config.Contracts
	contracts       contracts.ContractsService
}

func (aqueue *anynsQueue) Name() (name string) {
	return CName
}

func (aqueue *anynsQueue) Init(a *app.App) (err error) {
	aqueue.conf = a.MustComponent(config.CName).(*config.Config).Mongo
	aqueue.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	aqueue.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)

	aqueue.done = make(chan bool)
	aqueue.q = mb.New[int64](10) // TODO: queue size -> config

	return nil
}

func (aqueue *anynsQueue) Run(ctx context.Context) (err error) {
	uri := aqueue.conf.Connect
	dbName := aqueue.conf.Database
	collectionName := aqueue.conf.Collection

	// 1 - connect to DB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	aqueue.itemColl = client.Database(dbName).Collection(collectionName)

	// 2 - start one worker
	go aqueue.worker(ctx, aqueue.itemColl, aqueue.q, aqueue.done)
	return nil
}

func (aqueue *anynsQueue) Close(ctx context.Context) (err error) {
	if aqueue.itemColl != nil {
		err = aqueue.itemColl.Database().Client().Disconnect(ctx)
		aqueue.itemColl = nil
	}
	return
}

func (aqueue *anynsQueue) ProcessRequest(ctx context.Context, req *as.NameRegisterRequest) (operationId int64, err error) {
	// find current item count in the queue
	count, err := aqueue.itemColl.CountDocuments(ctx, nil)
	if err != nil {
		return 0, err
	}

	// 1 - insert into Mongo
	item := queueItemFromNameRegisterRequest(req, count)

	res, err := aqueue.itemColl.InsertOne(ctx, item)
	if err != nil {
		return 0, err
	}
	log.Info("inserted pending operation into DB", zap.String("ObjectID", res.InsertedID.(string)), zap.Int64("Item Index", item.Index))

	// 2 - insert into in-memory queue
	err = aqueue.q.Add(ctx, item.Index)
	if err != nil {
		// TODO: the record in DB will be never processed
		return 0, err
	}

	operationId = item.Index
	return operationId, nil
}

type findItemQuery struct {
	Index int64 `bson:"index"`
}

func (aqueue *anynsQueue) GetRequestStatus(ctx context.Context, operationId int64) (status as.OperationState, err error) {
	// get status from the queue
	var item QueueItem
	result := aqueue.itemColl.FindOne(ctx, findItemQuery{Index: operationId}).Decode(&item)
	if result == mongo.ErrNoDocuments {
		return 0, errors.New("item not found")
	}

	return item.Status, nil
}

func (aqueue *anynsQueue) worker(ctx context.Context, coll *mongo.Collection, queue *mb.MB[int64], done chan bool) {
	log.Info("worker started")

	for {
		// getting items from in-memory queue
		items, err := queue.Wait(ctx)
		if err != nil {
			break
		}

		for _, item := range items {
			// in case of error - do not stop processing queue
			aqueue.ProcessSingleItemInQueue(ctx, coll, item)
		}
	}

	log.Info("worker stopped")
	done <- true
}

func (aqueue *anynsQueue) ProcessSingleItemInQueue(ctx context.Context, coll *mongo.Collection, item int64) error {
	// 1 - get item from DB
	// each item in in-memory queue is an index of item in DB
	// so please get them from DB
	var queueItem QueueItem
	err := coll.FindOne(ctx, findItemQuery{Index: item}).Decode(&queueItem)
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}

	// 2 - process item
	log.Info("processing item", zap.Int64("Item Index", queueItem.Index))

	req := nameRegisterRequestFromQueueItem(queueItem)
	err = aqueue.NameRegister(ctx, req)

	// 3 - update item in DB
	if err != nil {
		log.Error("failed to process item. move state to ERROR", zap.Error(err))
		queueItem.Status = as.OperationState_Error
	} else {
		log.Info("item processed without error. move state to COMPLETED")
		queueItem.Status = as.OperationState_Completed
	}

	res, err := coll.ReplaceOne(ctx, findItemQuery{Index: item}, queueItem)
	if res.MatchedCount == 0 {
		log.Error("failed to update item in DB", zap.Error(err))
		return errors.New("failed to update item in DB")
	}
	return err
}

func (aqueue *anynsQueue) NameRegister(ctx context.Context, in *as.NameRegisterRequest) error {
	var registrantAccount common.Address = common.HexToAddress(in.OwnerEthAddress)

	fmt.Println("Contracts: ", aqueue.contracts)

	// 1 - connect to geth
	conn, err := aqueue.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	controller, err := aqueue.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return err
	}

	// 2 - get a name's first part
	// TODO: normalize string
	nameFirstPart := contracts.RemoveTLD(in.FullName)

	// 3 - calculate a commitment
	secret, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Error("can not generate random secret", zap.Error(err))
		return err
	}

	commitment, err := aqueue.contracts.MakeCommitment(
		nameFirstPart,
		registrantAccount,
		secret,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	if err != nil {
		log.Error("can not calculate a commitment", zap.Error(err))
		return err
	}

	authOpts, err := aqueue.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 4 - commit from Admin
	tx, err := aqueue.contracts.Commit(
		authOpts,
		commitment,
		controller)

	// TODO: check if tx is nil?
	if err != nil {
		log.Error("can not Commit tx", zap.Error(err))
		return err
	}

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for commit tx", zap.Error(err))
		return err
	}
	if !txRes {
		// new error
		return errors.New("commit tx not mined")
	}

	// update nonce again...
	authOpts, err = aqueue.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 5 - register
	tx, err = aqueue.contracts.Register(
		authOpts,
		nameFirstPart,
		registrantAccount,
		secret,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	// TODO: check if tx is nil?
	if err != nil {
		log.Error("can not Commit tx", zap.Error(err))
		return err
	}

	// wait for tx to be mined
	txRes, err = aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return err
	}
	if !txRes {
		// new error
		return errors.New("register tx failed")
	}

	log.Info("operation succeeded!")
	return nil
}
