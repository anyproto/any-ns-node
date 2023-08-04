package queue

import (
	"time"

	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
)

type QueueItemStatus int32

// when adding new status, don't forget to update these function:
// 1. StatusToState
// 2. NameRegisterMoveStateNext
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

func StatusToState(status QueueItemStatus) as.OperationState {
	switch status {
	case OperationStatus_Initial:
		return as.OperationState_Pending
	case OperationStatus_CommitSent:
		return as.OperationState_Pending
	case OperationStatus_RegisterSent:
		return as.OperationState_Pending
	case OperationStatus_Completed:
		return as.OperationState_Completed

	case OperationStatus_CommitError:
		return as.OperationState_Error
	case OperationStatus_RegisterError:
		return as.OperationState_Error
	case OperationStatus_Error:
		return as.OperationState_Error
	default:
		return as.OperationState_Pending
	}
}

// this structure is saved to mem queue and to DB
type QueueItem struct {
	Index           int64  `bson:"index"`
	FullName        string `bson:"fullName"`
	OwnerAnyAddress string `bson:"ownerAnyAddress"`
	OwnerEthAddress string `bson:"ownerEthAddress"`
	SpaceId         string `bson:"spaceId"`
	// aux fields
	SecretBase64   string          `bson:"secretBase64"`
	Status         QueueItemStatus `bson:"status"`
	TxCommitHash   string          `bson:"txCommitHash"`
	TxRegisterHash string          `bson:"txRegisterHash"`
	DateCreated    int64           `bson:"dateCreated"`
	DateModified   int64           `bson:"dateModified"`
}

// convert item to in-memory queue struct from initial dRPC request struct
func queueItemFromNameRegisterRequest(req *as.NameRegisterRequest, count int64) QueueItem {
	currTime := time.Now().Unix()

	return QueueItem{
		Index:           count,
		FullName:        req.FullName,
		OwnerAnyAddress: req.OwnerAnyAddress,
		OwnerEthAddress: req.OwnerEthAddress,
		SpaceId:         req.SpaceId,
		Status:          OperationStatus_Initial,

		DateCreated:  currTime,
		DateModified: currTime,
	}
}

// TODO: remove this
func nameRegisterRequestFromQueueItem(item QueueItem) *as.NameRegisterRequest {
	req := as.NameRegisterRequest{
		FullName:        item.FullName,
		OwnerAnyAddress: item.OwnerAnyAddress,
		OwnerEthAddress: item.OwnerEthAddress,
		SpaceId:         item.SpaceId,
	}
	return &req
}
