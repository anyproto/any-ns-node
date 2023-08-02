package queue

import (
	"context"
	"fmt"
	"time"

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

type findItemByIndexQuery struct {
	Index int64 `bson:"index"`
}

func New() app.Component {
	return &anynsQueue{}
}

type QueueService interface {
	// add new request to the queue and process it in the background
	AddNewRequest(ctx context.Context, req *as.NameRegisterRequest) (operationId int64, err error)
	GetRequestStatus(ctx context.Context, operationId int64) (status as.OperationState, err error)

	// Internal methods (public for tests):
	// process single "name registration" request, will update the status in the DB
	// with each tx sent
	NameRegister(ctx context.Context, queueItem *QueueItem, coll *mongo.Collection) error

	// read all "pending" items from DB and try to process em during startup
	ProcessAllItemsInDb(ctx context.Context, coll *mongo.Collection) error
	// process 1 item and update its state in the DB
	ProcessSingleItemInDb(ctx context.Context, coll *mongo.Collection, queueItem *QueueItem) error

	// read one item from the DB and process it, but do not remove it
	ProcessSingleItemInQueue(ctx context.Context, coll *mongo.Collection, itemIndex int64) error
	// just update item status in the DB
	SaveItemToDb(ctx context.Context, coll *mongo.Collection, queueItem *QueueItem) error

	app.ComponentRunnable
}

type anynsQueue struct {
	q        *mb.MB[int64]
	itemColl *mongo.Collection
	done     chan bool

	confMongo     config.Mongo
	confContracts config.Contracts
	confQueue     config.Queue

	contracts contracts.ContractsService
}

func (aqueue *anynsQueue) Name() (name string) {
	return CName
}

func (aqueue *anynsQueue) Init(a *app.App) (err error) {
	aqueue.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	aqueue.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	aqueue.confQueue = a.MustComponent(config.CName).(*config.Config).GetQueue()

	aqueue.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	aqueue.done = make(chan bool)
	aqueue.q = mb.New[int64](10) // TODO: queue size -> config

	return nil
}

func (aqueue *anynsQueue) Run(ctx context.Context) (err error) {
	uri := aqueue.confMongo.Connect
	dbName := aqueue.confMongo.Database
	collectionName := aqueue.confMongo.Collection

	// 1 - connect to DB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	aqueue.itemColl = client.Database(dbName).Collection(collectionName)
	if aqueue.itemColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("mongo connected!")

	// 2 - try to process all items in the DB
	aqueue.ProcessAllItemsInDb(ctx, aqueue.itemColl)

	// 3 - start one worker
	if !aqueue.confQueue.SkipBackroundProcessing {
		go aqueue.worker(ctx, aqueue.itemColl, aqueue.q, aqueue.done)
	}
	return nil
}

func (aqueue *anynsQueue) Close(ctx context.Context) (err error) {
	if aqueue.itemColl != nil {
		err = aqueue.itemColl.Database().Client().Disconnect(ctx)
		aqueue.itemColl = nil
	}
	return
}

func (aqueue *anynsQueue) AddNewRequest(ctx context.Context, req *as.NameRegisterRequest) (operationId int64, err error) {
	// count all documents in the collection (filter can not be nil)
	type countAllItemsQuery struct {
	}

	// find current item count in the queue
	count, err := aqueue.itemColl.CountDocuments(ctx, countAllItemsQuery{})
	if err != nil {
		return 0, err
	}

	// 1 - insert into Mongo
	item := queueItemFromNameRegisterRequest(req, count)

	_, err = aqueue.itemColl.InsertOne(ctx, item)
	if err != nil {
		return 0, err
	}
	log.Info("inserted pending operation into DB", zap.Int64("Item Index", item.Index))

	// 2 - insert into in-memory queue
	err = aqueue.q.Add(ctx, item.Index)
	if err != nil {
		// TODO: the record in DB will be never processed
		return 0, err
	}

	operationId = item.Index
	return operationId, nil
}

func (aqueue *anynsQueue) GetRequestStatus(ctx context.Context, operationId int64) (status as.OperationState, err error) {
	// get status from the queue
	var item QueueItem
	result := aqueue.itemColl.FindOne(ctx, findItemByIndexQuery{Index: operationId}).Decode(&item)
	if result == mongo.ErrNoDocuments {
		return 0, errors.New("item not found")
	}

	return StatusToState(item.Status), nil
}

// runs only if "SkipBackroundProcessing" is not set
func (aqueue *anynsQueue) worker(ctx context.Context, coll *mongo.Collection, queue *mb.MB[int64], done chan bool) {
	log.Info("worker started")

	// process items from in-memory queue
	for {
		items, err := queue.Wait(ctx)
		if err != nil {
			break
		}

		for _, itemIndex := range items {
			// in case of error - do not stop processing queue
			err := aqueue.ProcessSingleItemInQueue(ctx, coll, itemIndex)
			if err != nil {
				log.Error("failed to process item. continue", zap.Error(err))
			}
		}
	}

	log.Info("worker stopped")
	done <- true
}

func (aqueue *anynsQueue) ProcessAllItemsInDb(ctx context.Context, coll *mongo.Collection) error {
	type findItemByStatusQuery struct {
		Status QueueItemStatus `bson:"status"`
	}

	log.Info("searching for items in INITIAL state in the DB")

	for {
		// 1 - get item from DB that has INITIAL status (not processed yet)
		var queueItem QueueItem
		// TODO: add to index
		err := coll.FindOne(ctx, findItemByStatusQuery{Status: OperationStatus_Initial}).Decode(&queueItem)
		if err == mongo.ErrNoDocuments {
			log.Info("no more PENDING items in the DB")
			return nil
		}
		if err != nil {
			log.Error("failed to get PENDING item from DB", zap.Error(err))
		}

		// in case of error - do not stop processing queue
		err = aqueue.ProcessSingleItemInDb(ctx, coll, &queueItem)
		if err != nil {
			log.Error("failed to process item from DB. continue", zap.Error(err))
		}
	}
}

func (aqueue *anynsQueue) ProcessSingleItemInDb(ctx context.Context, coll *mongo.Collection, queueItem *QueueItem) error {
	if queueItem.Status != OperationStatus_Initial {
		log.Warn("item has BAD STATUS. skipping it", zap.Int64("Item Index", queueItem.Index), zap.Any("Status", queueItem.Status))
		return errors.New("item has BAD STATUS. skipping it")
	}

	log.Info("Found item", zap.Any("Item", queueItem))

	var err error = nil
	if !aqueue.confQueue.SkipProcessing {
		// 2 - process item
		log.Info("processing item from DB", zap.Int64("Item Index", queueItem.Index))
		err = aqueue.NameRegister(ctx, queueItem, coll)
	} else {
		log.Info("skipping processing item in DB. setting that it was processed", zap.Any("Item Index", queueItem.Index))
	}

	// 3 - update item in DB
	queueItem.Status = nameRegisterErrToStatus(err)

	log.Info("Save item to DB", zap.Any("Item", queueItem))
	return aqueue.SaveItemToDb(ctx, coll, queueItem)
}

func (aqueue *anynsQueue) SaveItemToDb(ctx context.Context, coll *mongo.Collection, queueItem *QueueItem) error {
	queueItem.DateModified = time.Now().Unix()

	res, err := coll.ReplaceOne(ctx, findItemByIndexQuery{Index: queueItem.Index}, queueItem)
	if res.MatchedCount == 0 {
		log.Error("failed to update item in DB", zap.Error(err))
		return errors.New("failed to update item in DB")
	}
	if err != nil {
		log.Error("failed to update item in DB", zap.Error(err))
		return err
	}
	return nil
}

func (aqueue *anynsQueue) ProcessSingleItemInQueue(ctx context.Context, coll *mongo.Collection, itemIndex int64) error {
	// 1 - get item from DB
	// each item in in-memory queue is an index of item in DB
	// so please get them from DB
	var queueItem QueueItem

	// TODO: add to index
	err := coll.FindOne(ctx, findItemByIndexQuery{Index: itemIndex}).Decode(&queueItem)
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}

	return aqueue.ProcessSingleItemInDb(ctx, coll, &queueItem)
}

func (aqueue *anynsQueue) NameRegister(ctx context.Context, queueItem *QueueItem, coll *mongo.Collection) error {
	in := nameRegisterRequestFromQueueItem(*queueItem)

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
		return ErrCommitFailed
	}

	// save tx hash to DB (optional)
	if coll != nil {
		queueItem.TxCommitHash = tx.Hash().String()
		queueItem.Status = OperationStatus_CommitWaiting

		err = aqueue.SaveItemToDb(ctx, coll, queueItem)
		if err != nil {
			log.Error("can not save Commit tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
			return ErrCommitFailed
		}
	}

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for commit tx", zap.Error(err))
		return ErrCommitFailed
	}
	if !txRes {
		// new error
		return ErrCommitFailed
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
		return ErrRegisterFailed
	}

	// save tx hash to DB (optional)
	if coll != nil {
		queueItem.TxRegisterHash = tx.Hash().String()
		queueItem.Status = OperationStatus_RegisterWaiting

		err = aqueue.SaveItemToDb(ctx, coll, queueItem)
		if err != nil {
			log.Error("can not save Register tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
			return ErrRegisterFailed
		}
	}

	// wait for tx to be mined
	txRes, err = aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return ErrRegisterFailed
	}
	if !txRes {
		return ErrRegisterFailed
	}

	log.Info("operation succeeded!")
	return nil
}
