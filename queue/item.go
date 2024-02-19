package queue

import (
	"time"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"go.uber.org/zap"
)

type QueueItemType int32
type QueueItemStatus int32

// when adding new status, don't forget to update these function:
// 1. StatusToState
// 2. NameRegisterMoveStateNext (IsStopProcessing in some rare cases)
const (
	OperationStatus_Initial    QueueItemStatus = 0
	OperationStatus_CommitSent QueueItemStatus = 1
	OperationStatus_CommitDone QueueItemStatus = 2

	OperationStatus_RegisterSent QueueItemStatus = 3
	OperationStatus_Completed    QueueItemStatus = 4

	OperationStatus_CommitError   QueueItemStatus = 5
	OperationStatus_RegisterError QueueItemStatus = 6

	OperationStatus_Error QueueItemStatus = 7
)

const (
	ItemType_NameRegister QueueItemType = 1
	ItemType_NameRenew    QueueItemType = 2
)

func StatusToState(status QueueItemStatus) nsp.OperationState {
	switch status {
	case OperationStatus_Initial, OperationStatus_CommitSent, OperationStatus_CommitDone, OperationStatus_RegisterSent:
		return nsp.OperationState_Pending

	case OperationStatus_CommitError, OperationStatus_RegisterError, OperationStatus_Error:
		return nsp.OperationState_Error

	case OperationStatus_Completed:
		return nsp.OperationState_Completed
	}

	log.Fatal("unknown status: ", zap.Any("status", status))
	return nsp.OperationState_Error
}

// this structure is saved to mem queue and to DB
type QueueItem struct {
	Index           int64         `bson:"index"`
	ItemType        QueueItemType `bson:"itemType"`
	FullName        string        `bson:"fullName"`
	OwnerAnyAddress string        `bson:"ownerAnyAddress"`
	OwnerEthAddress string        `bson:"ownerEthAddress"`
	SpaceId         string        `bson:"spaceId"`
	// aux fields
	SecretBase64    string          `bson:"secretBase64"`
	Status          QueueItemStatus `bson:"status"`
	TxCommitHash    string          `bson:"txCommitHash"`
	TxCommitNonce   uint64          `bson:"txCommitNonce"`
	TxRegisterHash  string          `bson:"txRegisterHash"`
	TxRegisterNonce uint64          `bson:"txRegisterNonce"`
	DateCreated     int64           `bson:"dateCreated"`
	DateModified    int64           `bson:"dateModified"`

	// for ItemType_NameRenew
	NameRenewDurationSec uint64 `bson:"nameRenewDurationSec"`
	TxRenewCommitHash    string `bson:"txRenewCommitHash"`

	TxCurrentNonce uint64 `bson:"currentTxNonce"`
	TxCurrentRetry uint   `bson:"currentTxRetry"`
}

// convert item to in-memory queue struct from initial dRPC request struct
func queueItemFromNameRegisterRequest(req *nsp.NameRegisterRequest, count int64) QueueItem {
	currTime := time.Now().Unix()

	return QueueItem{
		Index:           count,
		ItemType:        ItemType_NameRegister,
		FullName:        req.FullName,
		OwnerAnyAddress: req.OwnerAnyAddress,
		OwnerEthAddress: req.OwnerEthAddress,

		// WARNING: current interface doesn't support/have SpaceId
		SpaceId: "",
		//SpaceId:         req.SpaceId,
		Status: OperationStatus_Initial,

		DateCreated:  currTime,
		DateModified: currTime,
	}
}

/*
func queueItemFromNameRenewRequest(req *nsp.NameRenewRequest, count int64) QueueItem {
	currTime := time.Now().Unix()

	return QueueItem{
		Index:                count,
		ItemType:             ItemType_NameRenew,
		FullName:             req.FullName,
		NameRenewDurationSec: req.DurationSeconds,
		Status:               OperationStatus_Initial,
		DateCreated:          currTime,
		DateModified:         currTime,
	}
}*/

// TODO: remove this
func nameRegisterRequestFromQueueItem(item QueueItem) *nsp.NameRegisterRequest {
	req := nsp.NameRegisterRequest{
		FullName:        item.FullName,
		OwnerAnyAddress: item.OwnerAnyAddress,
		OwnerEthAddress: item.OwnerEthAddress,

		// WARNING: current interface doesn't support/have SpaceId
		//SpaceId:         item.SpaceId,
	}
	return &req
}
