package queue

import (
	"context"
	b64 "encoding/base64"
	"math/big"
	"time"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/config"
	contracts "github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-ns-node/nonce_manager"
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
	// 1 - new name registration request
	AddNewRequest(ctx context.Context, req *as.NameRegisterRequest) (operationId int64, err error)
	GetRequestStatus(ctx context.Context, operationId int64) (status as.OperationState, err error)
	// 2 - name renew request
	AddRenewRequest(ctx context.Context, req *as.NameRenewRequest) (operationId int64, err error)

	// Internal methods (public for tests):
	// read all "pending" items from DB and try to process em during startup
	FindAndProcessAllItemsInDb(ctx context.Context)
	FindAndProcessAllItemsInDbWithStatus(ctx context.Context, status QueueItemStatus)

	// process 1 item and update its state in the DB
	ProcessItem(ctx context.Context, queueItem *QueueItem) error
	SaveItemToDb(ctx context.Context, queueItem *QueueItem) error

	// NameRegister functions and states:
	// TODO: refactor - move to separate file
	NameRegister(ctx context.Context, queueItem *QueueItem) error
	NameRegisterMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) (err error, newState QueueItemStatus)
	IsStopProcessing(err error, prevState QueueItemStatus, newState QueueItemStatus) bool

	NameRegister_InitialState(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error
	NameRegister_CommitSent(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error
	NameRegister_RegisterWaiting(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error

	RecoverLowNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error
	RecoverHighNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error

	// NameRenew functions and states:
	NameRenew(ctx context.Context, queueItem *QueueItem) error
	NameRenewMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error

	app.ComponentRunnable
}

type anynsQueue struct {
	q        *mb.MB[int64]
	itemColl *mongo.Collection
	done     chan bool

	confMongo     config.Mongo
	confContracts config.Contracts
	confQueue     config.Queue

	contracts    contracts.ContractsService
	nonceManager nonce_manager.NonceService
}

func (aqueue *anynsQueue) Name() (name string) {
	return CName
}

func (aqueue *anynsQueue) Init(a *app.App) (err error) {
	aqueue.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	aqueue.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	aqueue.confQueue = a.MustComponent(config.CName).(*config.Config).GetQueue()

	aqueue.nonceManager = a.MustComponent(nonce_manager.CName).(nonce_manager.NonceService)
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
	if !aqueue.confQueue.SkipExistingItemsInDB {
		aqueue.FindAndProcessAllItemsInDb(ctx)
	}

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

	// calculate new secret
	secret, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Error("can not generate random secret", zap.Error(err))
		return 0, err
	}
	// convert [32]byte to base64 string
	item.SecretBase64 = b64.StdEncoding.EncodeToString(secret[:])

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

			// 1 - get item from DB
			// each item in in-memory queue is an index of item in DB
			// so please get them from DB
			var queueItem QueueItem

			// TODO: add to index
			err := coll.FindOne(ctx, findItemByIndexQuery{Index: itemIndex}).Decode(&queueItem)
			if err != nil {
				log.Warn("failed to get item from DB by index from Queue", zap.Error(err), zap.Any("Item Index", itemIndex))
				// in case of error - do not stop processing queue
			}

			err = aqueue.ProcessItem(ctx, &queueItem)
			if err != nil {
				log.Warn("failed to process single item from Queue", zap.Error(err), zap.Any("Item Index", itemIndex))
				// in case of error - do not stop processing queue
			}
		}
	}

	log.Info("worker stopped")
	done <- true
}

func (aqueue *anynsQueue) FindAndProcessAllItemsInDb(ctx context.Context) {
	aqueue.FindAndProcessAllItemsInDbWithStatus(ctx, OperationStatus_Initial)
	aqueue.FindAndProcessAllItemsInDbWithStatus(ctx, OperationStatus_CommitSent)
	aqueue.FindAndProcessAllItemsInDbWithStatus(ctx, OperationStatus_CommitDone)
	aqueue.FindAndProcessAllItemsInDbWithStatus(ctx, OperationStatus_RegisterSent)
}

func (aqueue *anynsQueue) FindAndProcessAllItemsInDbWithStatus(ctx context.Context, status QueueItemStatus) {
	type findItemByStatusQuery struct {
		Status QueueItemStatus `bson:"status"`
	}

	log.Info("Process all items in DB with state", zap.Any("Status", status))

	for {
		// 1 - get item from DB that has INITIAL status (not processed yet)
		var queueItem QueueItem
		// TODO: add to index
		err := aqueue.itemColl.FindOne(ctx, findItemByStatusQuery{Status: status}).Decode(&queueItem)
		if err == mongo.ErrNoDocuments {
			log.Info("no more items in the DB with such state", zap.Any("Status", status))
			return
		}

		if err != nil {
			log.Warn("failed to get item from DB", zap.Error(err))
			// in case of error - do not stop processing queue
		}

		err = aqueue.ProcessItem(ctx, &queueItem)
		if err != nil {
			log.Warn("failed to process item from DB. continue", zap.Error(err))
			// in case of error - do not stop processing queue
		}
	}
}

func (aqueue *anynsQueue) ProcessItem(ctx context.Context, queueItem *QueueItem) error {
	log.Info("Found item in state", zap.Any("Item", queueItem), zap.Any("Status", queueItem.Status))

	if aqueue.confQueue.SkipProcessing {
		log.Info("skipping processing item in DB. mark item as completed", zap.Any("Item Index", queueItem.Index))
		queueItem.Status = OperationStatus_Completed
		return aqueue.SaveItemToDb(ctx, queueItem)
	}

	log.Info("processing item from DB", zap.Int64("Item Index", queueItem.Index))

	switch queueItem.ItemType {
	case ItemType_NameRegister:
		return aqueue.NameRegister(ctx, queueItem)
	case ItemType_NameRenew:
		return aqueue.NameRenew(ctx, queueItem)
	}

	log.Fatal("unknown item type", zap.Any("Item", queueItem))
	return errors.New("unknown item type")
}

func (aqueue *anynsQueue) SaveItemToDb(ctx context.Context, queueItem *QueueItem) error {
	if aqueue.itemColl == nil {
		// TODO: mock mongo and remove this weird logics please
		// no error, required for some tests!
		return nil
	}

	queueItem.DateModified = time.Now().Unix()

	res, err := aqueue.itemColl.ReplaceOne(ctx, findItemByIndexQuery{Index: queueItem.Index}, queueItem)
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

func (aqueue *anynsQueue) UpdateItemStatus(ctx context.Context, itemIndex int64, newStatus QueueItemStatus) error {
	// 1 - find item
	var queueItem QueueItem

	// TODO: add to index
	err := aqueue.itemColl.FindOne(ctx, findItemByIndexQuery{Index: itemIndex}).Decode(&queueItem)
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return err
	}

	// 2 - update status and save
	queueItem.Status = newStatus

	return aqueue.SaveItemToDb(ctx, &queueItem)
}

func (aqueue *anynsQueue) RecoverLowNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	retryCount := queueItem.TxCurrentRetry
	nonce := queueItem.TxCurrentNonce

	if retryCount >= aqueue.confQueue.NonceRetryCount {
		return errors.New("NONCE IS TOO LOW but RETRY COUNT IS TOO BIG, STOP...")
	}

	log.Error("NONCE IS TOO LOW!!! Retrying with new nonce...", zap.Any("retry", retryCount))

	// update nonce in the DB immediately, even if TX is still not sent
	_, err := aqueue.nonceManager.SaveNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin), nonce+1)
	if err != nil {
		log.Error("can not update nonce in DB!", zap.Error(err))
		return err
	}

	// update nonce in the item
	queueItem.TxCurrentNonce = nonce + 1
	queueItem.TxCurrentRetry = retryCount + 1

	// save item to DB
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save item to DB!", zap.Error(err))
		return err
	}

	// continue!
	return nil
}

func (aqueue *anynsQueue) RecoverHighNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	retryCount := queueItem.TxCurrentRetry

	// do not give more than 2 tries
	if retryCount >= 2 {
		return errors.New("NONCE IS probably TOO HIGH but RETRY COUNT IS TOO BIG, STOP...")
	}

	log.Error("NONCE IS probably TOO HIGH!!! Retrying with new nonce...", zap.Any("retry", retryCount))

	// - get new nonce from network
	newNonce, err := aqueue.nonceManager.GetCurrentNonceFromNetwork(common.HexToAddress(aqueue.confContracts.AddrAdmin))
	if err != nil {
		log.Error("can not get new nonce from network!", zap.Error(err))
		return err
	}

	// update nonce in the DB immediately, even if TX is still not sent
	_, err = aqueue.nonceManager.SaveNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin), newNonce)
	if err != nil {
		log.Error("can not update nonce in DB!", zap.Error(err))
		return err
	}

	// update nonce in the item
	queueItem.TxCurrentNonce = newNonce
	queueItem.TxCurrentRetry = retryCount + 1

	// save item to DB
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save item to DB!", zap.Error(err))
		return err
	}

	// continue!
	return nil
}

func (aqueue *anynsQueue) NameRegister(ctx context.Context, queueItem *QueueItem) error {
	conn, err := aqueue.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	// 1 - init nonce
	// get nonce (from DB, config file or network)
	nonce, err := aqueue.nonceManager.GetCurrentNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin))
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return err
	}
	queueItem.TxCurrentNonce = nonce
	queueItem.TxCurrentRetry = 0 // reset retries counter

	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("failed to save item to DB", zap.Error(err))
		return err
	}

	// 2 - move states
	for {
		prevState := queueItem.Status

		// OperationStatus_Initial -> OperationStatus_CommitSent
		// OperationStatus_CommitSent -> OperationStatus_CommitDone
		// OperationStatus_CommitDone -> OperationStatus_RegisterSent
		// OperationStatus_RegisterSent -> OperationStatus_Completed
		err, newState := aqueue.NameRegisterMoveStateNext(ctx, queueItem, conn)

		// 3 - try to recover from errors
		if err != nil {
			if err == contracts.ErrNonceTooLow {
				// if we got "nonce too low" error the tx is immediately rejected. to fix it:
				// - get nonce from network
				// - send this tx again with +1 nonce
				aqueue.RecoverLowNonce(ctx, queueItem, conn)

				newState = prevState // try again with the same state
				err = nil
			} else if err == contracts.ErrNonceTooHigh {
				// if nonce is higher than needed - tx will be rejected by the network with "not found" error immediately
				// in this case we:
				// - wait for N minutes for all TXs to settle
				// - get new nonce from network
				// - retry sending this tx with new nonce
				aqueue.RecoverHighNonce(ctx, queueItem, conn)

				newState = prevState // try again with the same state
				err = nil
			}
		}

		// 4 - update state in DB
		if newState != prevState {
			err2 := aqueue.UpdateItemStatus(ctx, queueItem.Index, newState)
			if err2 != nil {
				log.Error("failed to update item status in DB", zap.Error(err), zap.Any("prev state", prevState), zap.Any("new state", newState))
				return err2
			}
		}

		// 5 - check if stop?
		isStopProcessing := aqueue.IsStopProcessing(err, prevState, newState)
		if isStopProcessing {
			log.Info("state machine: stop processing item", zap.Any("Item", queueItem))
			return nil
		}
	}
}

func (aqueue *anynsQueue) IsStopProcessing(err error, prevState QueueItemStatus, newState QueueItemStatus) bool {
	if err != nil {
		// TODO: retry logic?
		// always stop in case of error
		return true
	}

	state := StatusToState(newState)
	switch state {
	case as.OperationState_Pending:
		return false
	case as.OperationState_Completed, as.OperationState_Error:
		return true
	}

	// TODO: retry logic?
	// in case of errors/unknown state -> do not RETRY, just halt
	log.Fatal("unknown state", zap.Any("prev state", prevState), zap.Any("new state", newState))
	return true
}

func (aqueue *anynsQueue) NameRegisterMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) (error, QueueItemStatus) {
	currentStatus := queueItem.Status

	switch currentStatus {
	case OperationStatus_Initial:
		err := aqueue.NameRegister_InitialState(ctx, queueItem, conn)
		// TODO: assert that item.TxCommitHash should not be null here

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return err, OperationStatus_CommitError
		}

		return err, OperationStatus_CommitSent
	case OperationStatus_CommitSent:
		err := aqueue.NameRegister_CommitSent(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return err, OperationStatus_CommitError
		}
		return err, OperationStatus_CommitDone
	case OperationStatus_CommitDone:
		err := aqueue.NameRegister_CommitDone(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return err, OperationStatus_RegisterError
		}
		return err, OperationStatus_RegisterSent
	case OperationStatus_RegisterSent:
		err := aqueue.NameRegister_RegisterWaiting(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return err, OperationStatus_Error
		}
		return err, OperationStatus_Completed
	case OperationStatus_Completed:
		// Success
		return nil, OperationStatus_Completed
	case OperationStatus_CommitError, OperationStatus_RegisterError, OperationStatus_Error:
		// no state transition in case of ERRORS
		return nil, queueItem.Status
	}

	log.Fatal("unknown state", zap.Any("state", queueItem.Status))
	return nil, queueItem.Status
}

func (aqueue *anynsQueue) NameRegister_InitialState(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	nonce := queueItem.TxCurrentNonce

	controller, err := aqueue.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return err
	}

	// TODO: normalize string
	in := nameRegisterRequestFromQueueItem(*queueItem)
	var registrantAccount common.Address = common.HexToAddress(in.OwnerEthAddress)
	nameFirstPart := contracts.RemoveTLD(in.FullName)
	secret, err := b64.StdEncoding.DecodeString(queueItem.SecretBase64)

	if err != nil {
		log.Error("can not decode base64 secret", zap.Error(err), zap.Any("secret", queueItem.SecretBase64))
		return err
	}

	var secret32 [32]byte
	copy(secret32[:], secret)

	// 1 - make a commitment
	commitment, err := aqueue.contracts.MakeCommitment(
		nameFirstPart,
		registrantAccount,
		secret32,
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
	if authOpts != nil {
		authOpts.Nonce = big.NewInt(int64(nonce))
	}
	log.Info("Nonce is", zap.Any("Nonce", nonce))

	// 2 - commit
	tx, err := aqueue.contracts.Commit(
		ctx,
		conn,
		authOpts,
		commitment,
		controller)

	// can return ErrNonceTooLow error
	// can return ErrNonceTooHigh error
	if err != nil {
		log.Error("can not Commit tx", zap.Error(err), zap.Any("tx", tx))
		return err
	}

	// 3 - update nonce and item in DB
	_, err = aqueue.nonceManager.SaveNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin), nonce+1)
	if err != nil {
		log.Error("can not update nonce in DB!", zap.Error(err))
		return err
	}

	queueItem.TxCommitHash = tx.Hash().String()
	queueItem.TxCommitNonce = nonce
	queueItem.TxCurrentNonce = nonce + 1
	queueItem.TxCurrentRetry = 0
	queueItem.Status = OperationStatus_CommitSent

	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save Commit tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrCommitFailed
	}

	return nil
}

// wait for commit tx
func (aqueue *anynsQueue) NameRegister_CommitSent(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	if len(queueItem.TxCommitHash) == 0 {
		return errors.New("tx hash is empty")
	}

	log.Info("waiting for commit tx", zap.String("tx hash", queueItem.TxCommitHash), zap.Any("Item", queueItem))

	// 1
	txHash := common.HexToHash(queueItem.TxCommitHash)
	tx, err := aqueue.contracts.TxByHash(ctx, conn, txHash)
	if err != nil {
		log.Error("failed to fetch transaction details:", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrCommitFailed
	}

	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for commit tx", zap.Error(err))
		return ErrCommitFailed
	}
	if !txRes {
		// new error
		log.Warn("tx finished with ERROR result", zap.String("tx hash", queueItem.TxCommitHash))
		return ErrCommitFailed
	}

	// 2 - update in DB
	queueItem.Status = OperationStatus_CommitDone

	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save Register tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrCommitFailed
	}

	return nil
}

// generate new register tx
func (aqueue *anynsQueue) NameRegister_CommitDone(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	// get proper nonce (from DB, config file or network)
	nonce, err := aqueue.nonceManager.GetCurrentNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin))
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return err
	}

	controller, err := aqueue.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return err
	}

	// get new nonce
	authOpts, err := aqueue.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}
	if authOpts != nil {
		authOpts.Nonce = big.NewInt(int64(nonce))
	}

	log.Info("Nonce is", zap.Any("Nonce", nonce))

	// register
	// TODO: normalize string
	in := nameRegisterRequestFromQueueItem(*queueItem)
	var registrantAccount common.Address = common.HexToAddress(in.OwnerEthAddress)
	nameFirstPart := contracts.RemoveTLD(in.FullName)
	secret, err := b64.StdEncoding.DecodeString(queueItem.SecretBase64)

	if err != nil {
		log.Error("can not decode base64 secret", zap.Error(err), zap.Any("secret", queueItem.SecretBase64))
		return err
	}

	var secret32 [32]byte
	copy(secret32[:], secret)

	tx, err := aqueue.contracts.Register(
		ctx,
		conn,
		authOpts,
		nameFirstPart,
		registrantAccount,
		secret32,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		in.GetSpaceId())

	// can return ErrNonceTooLow error
	// can return ErrNonceTooHigh error
	if err != nil {
		log.Error("can not Regsiter tx", zap.Error(err))
		return ErrRegisterFailed
	}

	// update nonce in DB
	_, err = aqueue.nonceManager.SaveNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin), nonce+1)
	if err != nil {
		log.Error("can not update nonce in DB!", zap.Error(err))
		return err
	}

	// update item in DB
	queueItem.TxRegisterHash = tx.Hash().String()
	queueItem.TxRegisterNonce = nonce
	queueItem.TxCurrentNonce = nonce + 1
	queueItem.TxCurrentRetry = 0
	queueItem.Status = OperationStatus_RegisterSent

	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save Register tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrRegisterFailed
	}

	return nil
}

// wait for register tx
func (aqueue *anynsQueue) NameRegister_RegisterWaiting(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	if len(queueItem.TxRegisterHash) == 0 {
		return errors.New("tx hash is empty")
	}

	log.Info("waiting for register tx", zap.String("tx hash", queueItem.TxRegisterHash), zap.Any("Item", queueItem))
	txHash := common.HexToHash(queueItem.TxRegisterHash)
	tx, err := aqueue.contracts.TxByHash(ctx, conn, txHash)
	if err != nil {
		log.Error("failed to fetch transaction details:", zap.Error(err), zap.String("tx hash", queueItem.TxRegisterHash))
		return ErrRegisterFailed
	}

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return ErrRegisterFailed
	}
	if !txRes {
		log.Warn("tx finished with ERROR result", zap.String("tx hash", queueItem.TxRegisterHash))
		return ErrRegisterFailed
	}

	// update item in DB
	queueItem.Status = OperationStatus_Completed
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save last update", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrRegisterFailed
	}

	log.Info("operation succeeded!")
	return nil
}

func (aqueue *anynsQueue) AddRenewRequest(ctx context.Context, req *as.NameRenewRequest) (operationId int64, err error) {
	// count all documents in the collection (filter can not be nil)
	type countAllItemsQuery struct {
	}

	// find current item count in the queue
	count, err := aqueue.itemColl.CountDocuments(ctx, countAllItemsQuery{})
	if err != nil {
		return 0, err
	}

	// 1 - insert into Mongo
	item := queueItemFromNameRenewRequest(req, count)

	_, err = aqueue.itemColl.InsertOne(ctx, item)
	if err != nil {
		return 0, err
	}
	log.Info("inserted pending renew operation into DB", zap.Int64("Item Index", item.Index))

	// 2 - insert into in-memory queue
	err = aqueue.q.Add(ctx, item.Index)
	if err != nil {
		// TODO: the record in DB will be never processed
		return 0, err
	}

	operationId = item.Index
	return operationId, nil
}

func (aqueue *anynsQueue) NameRenew(ctx context.Context, queueItem *QueueItem) error {
	conn, err := aqueue.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	// just try to renew and finish with this item
	err = aqueue.NameRenewMoveStateNext(ctx, queueItem, conn)
	newState := OperationStatus_Completed // success

	if err != nil {
		newState = OperationStatus_Error
	}
	err2 := aqueue.UpdateItemStatus(ctx, queueItem.Index, newState)
	if err2 != nil {
		log.Error("failed to update item status in DB", zap.Error(err))
		return err2
	}

	return nil
}

func (aqueue *anynsQueue) NameRenewMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	controller, err := aqueue.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return err
	}

	// 1 - get proper nonce (from DB, config file or network)
	nonce, err := aqueue.nonceManager.GetCurrentNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin))
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return err
	}

	authOpts, err := aqueue.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return err
	}
	if authOpts != nil {
		authOpts.Nonce = big.NewInt(int64(nonce))
	}
	log.Info("Nonce is", zap.Any("Nonce", nonce))

	// 2 - send tx to the network
	tx, err := aqueue.contracts.RenewName(
		ctx,
		conn,
		authOpts,
		queueItem.FullName,
		queueItem.NameRenewDurationSec,
		controller)

	if err != nil {
		if err == contracts.ErrNonceTooLow {
			// TODO:
		}

		log.Error("can not Regsiter tx", zap.Error(err))
		return ErrRenewFailed
	}

	// TODO: check if started to mine?

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return ErrRenewFailed
	}
	if !txRes {
		log.Warn("tx finished with ERROR result", zap.String("tx hash", queueItem.TxRegisterHash))
		return ErrRenewFailed
	}

	// TODO: recover from low,high nonce

	// update item in DB
	queueItem.Status = OperationStatus_Completed
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save last update", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return ErrRenewFailed
	}

	log.Info("renew operation succeeded!")
	return nil
}
