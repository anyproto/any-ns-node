package queue

import (
	"context"
	b64 "encoding/base64"
	"math/big"
	"strings"
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
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cheggaaa/mb/v3"
)

const CName = "any-ns.queue"

var log = logger.NewNamed(CName)

type findItemByIndexQuery struct {
	Index int64 `bson:"index"`
}

func New() app.ComponentRunnable {
	return &anynsQueue{}
}

type QueueService interface {
	// 1 - new name registration request
	AddNewRequest(ctx context.Context, req *nsp.NameRegisterRequest) (operationId int64, err error)
	GetRequestStatus(ctx context.Context, operationId int64) (status nsp.OperationState, err error)

	// Internal methods (public for tests):
	// read all "pending" items from DB and try to process em during startup
	FindAndProcessAllItemsInDb(ctx context.Context)
	FindAndProcessAllItemsInDbWithStatus(ctx context.Context, status QueueItemStatus)
	SaveItemToDb(ctx context.Context, queueItem *QueueItem) error

	// process 1 item and update its state in the DB
	ProcessItem(ctx context.Context, queueItem *QueueItem) error

	// NameRegister functions and states:
	// TODO: refactor - move to separate file
	NameRegisterMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) (newState QueueItemStatus, err error)

	app.ComponentRunnable
}

type anynsQueue struct {
	q    *mb.MB[int64]
	done chan bool

	confMongo     config.Mongo
	confContracts config.Contracts
	confQueue     config.Queue

	itemColl     *mongo.Collection
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
	collectionName := "queue"

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

func (aqueue *anynsQueue) AddNewRequest(ctx context.Context, req *nsp.NameRegisterRequest) (operationId int64, err error) {
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

func (aqueue *anynsQueue) GetRequestStatus(ctx context.Context, operationId int64) (status nsp.OperationState, err error) {
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

	// 1 - init item - reset retry count, get new nonce
	err := aqueue.initNonce(ctx, queueItem)
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	conn, err := aqueue.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	// 2 - move states
	for {
		prevState := queueItem.Status

		var err error
		var newState QueueItemStatus

		// ItemType_NameRegister:
		// 	OperationStatus_Initial -> OperationStatus_CommitSent
		// 	OperationStatus_CommitSent -> OperationStatus_CommitDone
		// 	OperationStatus_CommitDone -> OperationStatus_RegisterSent
		// 	OperationStatus_RegisterSent -> OperationStatus_Completed
		//
		// ItemType_NameRenew:
		// 	OperationStatus_Initial ->OperationStatus_Completed
		switch queueItem.ItemType {
		case ItemType_NameRegister:
			newState, err = aqueue.NameRegisterMoveStateNext(ctx, queueItem, conn)
		case ItemType_NameRenew:
			newState, err = aqueue.nameRenewMoveStateNext(ctx, queueItem, conn)
		}

		// 3 - handle nonce errors
		newState, err = aqueue.handleNonceErrors(ctx, err, prevState, newState, queueItem, conn)

		// 4 - update state in DB
		if newState != prevState {
			err2 := aqueue.updateItemStatus(ctx, queueItem.Index, newState)
			if err2 != nil {
				log.Error("failed to update item status in DB", zap.Error(err), zap.Any("prev state", prevState), zap.Any("new state", newState))
				return err2
			}
		}

		// 5 - check if stop?
		isStopProcessing := aqueue.isStopProcessing(err, prevState, newState)
		if isStopProcessing {
			log.Info("state machine: stop processing item", zap.Any("Item", queueItem))
			return nil
		}
	}

	return nil
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

func (aqueue *anynsQueue) updateItemStatus(ctx context.Context, itemIndex int64, newStatus QueueItemStatus) error {
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

func (aqueue *anynsQueue) recoverLowNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	retryCount := queueItem.TxCurrentRetry
	nonce := queueItem.TxCurrentNonce

	if retryCount >= aqueue.confQueue.LowNonceRetryCount {
		return errors.New("NONCE IS TOO LOW but RETRY COUNT IS TOO BIG, STOP...")
	}

	log.Warn("NONCE IS TOO LOW!!! Retrying with new nonce...", zap.Any("retry", retryCount))

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

func (aqueue *anynsQueue) recoverHighNonce(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	retryCount := queueItem.TxCurrentRetry

	// do not give more than N tries
	if retryCount >= aqueue.confQueue.HighNonceRetryCount {
		return errors.New("NONCE IS probably TOO HIGH but RETRY COUNT IS TOO BIG, STOP...")
	}

	log.Warn("NONCE IS probably TOO HIGH!!! Retrying with new nonce...", zap.Any("retry", retryCount))

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

func (aqueue *anynsQueue) initNonce(ctx context.Context, queueItem *QueueItem) error {
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

	return nil
}

func (aqueue *anynsQueue) handleNonceErrors(ctx context.Context, err error, prevState QueueItemStatus, newState QueueItemStatus, queueItem *QueueItem, conn *ethclient.Client) (newStatusOut QueueItemStatus, errOut error) {
	// try to recover from nonoce errors
	if err != nil {
		if err == contracts.ErrNonceTooLow {
			// if we got "nonce too low" error the tx is immediately rejected. to fix it:
			// - get nonce from network
			// - send this tx again with +1 nonce
			aqueue.recoverLowNonce(ctx, queueItem, conn)

			newState = prevState // try again with the same state
			err = nil
		} else if err == contracts.ErrNonceTooHigh {
			// if nonce is higher than needed - tx will be rejected by the network with "not found" error immediately
			// in this case we:
			// - wait for N minutes for all TXs to settle
			// - get new nonce from network
			// - retry sending this tx with new nonce
			aqueue.recoverHighNonce(ctx, queueItem, conn)

			newState = prevState // try again with the same state
			err = nil
		}
	}

	return newState, err
}

func (aqueue *anynsQueue) isStopProcessing(err error, prevState QueueItemStatus, newState QueueItemStatus) bool {
	if err != nil {
		// TODO: retry logic?
		// always stop in case of error
		return true
	}

	state := StatusToState(newState)
	switch state {
	case nsp.OperationState_Pending:
		return false
	case nsp.OperationState_Completed, nsp.OperationState_Error:
		return true
	}

	// TODO: retry logic?
	// in case of errors/unknown state -> do not RETRY, just halt
	log.Fatal("unknown state", zap.Any("prev state", prevState), zap.Any("new state", newState))
	return true
}

func (aqueue *anynsQueue) NameRegisterMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) (QueueItemStatus, error) {
	currentStatus := queueItem.Status

	switch currentStatus {
	case OperationStatus_Initial:
		err := aqueue.nameRegister_InitialState(ctx, queueItem, conn)
		// TODO: assert that item.TxCommitHash should not be null here

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return OperationStatus_CommitError, err
		}

		return OperationStatus_CommitSent, err
	case OperationStatus_CommitSent:
		err := aqueue.nameRegister_CommitSent(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return OperationStatus_CommitError, err
		}
		return OperationStatus_CommitDone, err
	case OperationStatus_CommitDone:
		err := aqueue.nameRegister_CommitDone(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return OperationStatus_RegisterError, err
		}
		return OperationStatus_RegisterSent, err
	case OperationStatus_RegisterSent:
		err := aqueue.nameRegister_RegisterWaiting(ctx, queueItem, conn)

		// in case of failed tx -> save error to DB and stop processing it next time
		if err != nil {
			// save to DB
			return OperationStatus_Error, err
		}
		return OperationStatus_Completed, err
	case OperationStatus_Completed:
		// Success
		return OperationStatus_Completed, nil
	case OperationStatus_CommitError, OperationStatus_RegisterError, OperationStatus_Error:
		// no state transition in case of ERRORS
		return queueItem.Status, nil
	}

	log.Fatal("unknown state", zap.Any("state", queueItem.Status))
	return queueItem.Status, nil
}

func (aqueue *anynsQueue) nameRegister_InitialState(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	nonce := queueItem.TxCurrentNonce

	controller, err := aqueue.contracts.ConnectToPrivateController(conn)
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

	// Currently by default this is always true!
	// NameRegisterRequest has no field for this
	//
	// it means that we are going to update reverse record:
	// mike.any -> 0x1234. resolving 0x1234 will return mike.any
	// john.any -> 0x1234. now resolving 0x1234 will return john.any
	//
	// so basically 0x1234 controls both names, but only the LAST ONE is reverse resolved
	// in some cases it can be a problem
	isReverseRecordUpdate := true

	// 1 - make a commitment
	commitment, err := aqueue.contracts.MakeCommitment(
		nameFirstPart,
		registrantAccount,
		secret32,
		controller,
		in.GetFullName(),
		in.GetOwnerAnyAddress(),
		// WARNING: current interface doesn't support/have SpaceId
		// in.GetSpaceId()
		"",
		isReverseRecordUpdate,
		in.RegisterPeriodMonths,
	)

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
		return err
	}

	return nil
}

// wait for commit tx
func (aqueue *anynsQueue) nameRegister_CommitSent(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	if len(queueItem.TxCommitHash) == 0 {
		return errors.New("tx hash is empty")
	}

	log.Info("waiting for commit tx", zap.String("tx hash", queueItem.TxCommitHash), zap.Any("Item", queueItem))
	txHash := common.HexToHash(queueItem.TxCommitHash)

	// 0 - try to wait for TX first and handle "nonce too high" error
	// wait until TX is "seen" by the network (N times)
	// can return ErrNonceTooHigh or just error
	err := aqueue.contracts.WaitForTxToStartMining(ctx, conn, txHash)
	if err != nil {
		log.Error("can not wait for Commit tx, can not start", zap.Error(err))
		return err
	}

	// 1
	tx, err := aqueue.contracts.TxByHash(ctx, conn, txHash)
	if err != nil {
		// TODO: handle it and retry
		log.Error("failed to fetch transaction details:", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return err
	}

	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for commit tx", zap.Error(err))
		return err
	}
	if !txRes {
		// new error
		log.Warn("tx finished with ERROR result", zap.String("tx hash", queueItem.TxCommitHash))
		return errors.New("WaitMined - tx not found")
	}

	// 2 - update in DB
	queueItem.Status = OperationStatus_CommitDone

	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save Register tx", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return err
	}

	return nil
}

// generate new register tx
func (aqueue *anynsQueue) nameRegister_CommitDone(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	nonce := queueItem.TxCurrentNonce

	controller, err := aqueue.contracts.ConnectToPrivateController(conn)
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

	// Currently by default this is always true!
	// NameRegisterRequest has no field for this
	isReverseRecordUpdate := true

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
		// WARNING: current interface doesn't support/have SpaceId
		//in.GetSpaceId(),
		"",
		isReverseRecordUpdate,
		in.RegisterPeriodMonths,
	)

	// can return ErrNonceTooLow error
	// can return ErrNonceTooHigh error
	if err != nil {
		log.Error("can not Regsiter tx", zap.Error(err))
		return err
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
		return err
	}

	return nil
}

// wait for register tx
func (aqueue *anynsQueue) nameRegister_RegisterWaiting(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) error {
	if len(queueItem.TxRegisterHash) == 0 {
		return errors.New("tx hash is empty")
	}

	log.Info("waiting for register tx", zap.String("tx hash", queueItem.TxRegisterHash), zap.Any("Item", queueItem))
	txHash := common.HexToHash(queueItem.TxRegisterHash)

	// 0 - try to wait for TX first and handle "nonce too high" error
	// wait until TX is "seen" by the network (N times)
	// can return ErrNonceTooHigh or just error
	err := aqueue.contracts.WaitForTxToStartMining(ctx, conn, txHash)
	if err != nil {
		log.Error("can not wait for Register tx, can not start", zap.Error(err))
		return err
	}

	tx, err := aqueue.contracts.TxByHash(ctx, conn, txHash)
	if err != nil {
		// TODO: handle it and retry
		log.Error("failed to fetch transaction details:", zap.Error(err), zap.String("tx hash", queueItem.TxRegisterHash))
		return err
	}

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return err
	}
	if !txRes {
		log.Warn("tx finished with ERROR result", zap.String("tx hash", queueItem.TxRegisterHash))
		return err
	}

	// update item in DB
	queueItem.Status = OperationStatus_Completed
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save last update", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return err
	}

	log.Info("operation succeeded!")
	return nil
}

func (aqueue *anynsQueue) nameRenewMoveStateNext(ctx context.Context, queueItem *QueueItem, conn *ethclient.Client) (newState QueueItemStatus, err error) {
	controller, err := aqueue.contracts.ConnectToPrivateController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return OperationStatus_Error, err
	}

	// 1 - get proper nonce (from DB, config file or network)
	nonce, err := aqueue.nonceManager.GetCurrentNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin))
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return OperationStatus_Error, err
	}

	authOpts, err := aqueue.contracts.GenerateAuthOptsForAdmin(conn)
	if err != nil {
		log.Error("can not get auth params for admin", zap.Error(err))
		return OperationStatus_Error, err
	}
	if authOpts != nil {
		authOpts.Nonce = big.NewInt(int64(nonce))
	}
	log.Info("Nonce is", zap.Any("Nonce", nonce))

	//
	parts := strings.Split(queueItem.FullName, ".")
	if len(parts) != 2 {
		return OperationStatus_Error, errors.New("invalid name")
	}
	firstPart := parts[0]

	// 2 - send tx to the network
	tx, err := aqueue.contracts.RenewName(
		ctx,
		conn,
		authOpts,
		firstPart,
		queueItem.NameRenewDurationSec,
		controller)

	// can return ErrNonceTooLow error
	// can return ErrNonceTooHigh error
	if err != nil {
		log.Error("can not Renew tx", zap.Error(err))
		return OperationStatus_Error, err
	}

	// update nonce in DB
	_, err = aqueue.nonceManager.SaveNonce(common.HexToAddress(aqueue.confContracts.AddrAdmin), nonce+1)
	if err != nil {
		log.Error("can not update nonce in DB!", zap.Error(err))
		return OperationStatus_Error, err
	}

	// wait for tx to be mined
	txRes, err := aqueue.contracts.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("can not wait for register tx", zap.Error(err))
		return OperationStatus_Error, err
	}
	if !txRes {
		log.Warn("tx finished with ERROR result", zap.String("tx hash", tx.Hash().String()))
		return OperationStatus_Error, err
	}

	// update item in DB
	queueItem.Status = OperationStatus_Completed
	err = aqueue.SaveItemToDb(ctx, queueItem)
	if err != nil {
		log.Error("can not save last update", zap.Error(err), zap.String("tx hash", queueItem.TxCommitHash))
		return OperationStatus_Error, err
	}

	log.Info("renew operation succeeded!")
	return OperationStatus_Completed, nil
}
