package queue

import (
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
)

type QueueItemStatus int32

const (
	OperationStatus_Initial         QueueItemStatus = 0
	OperationStatus_CommitWaiting   QueueItemStatus = 1
	OperationStatus_RegisterWaiting QueueItemStatus = 2
	OperationStatus_Completed       QueueItemStatus = 3

	OperationStatus_CommitError   QueueItemStatus = 4
	OperationStatus_RegisterError QueueItemStatus = 5
	OperationStatus_Error         QueueItemStatus = 6
)

func StatusToState(status QueueItemStatus) as.OperationState {
	switch status {
	case OperationStatus_Initial:
		return as.OperationState_Pending
	case OperationStatus_CommitWaiting:
		return as.OperationState_Pending
	case OperationStatus_RegisterWaiting:
		return as.OperationState_Pending
	case OperationStatus_Completed:
		return as.OperationState_Completed
	case OperationStatus_Error:
		return as.OperationState_Error
	default:
		return as.OperationState_Pending
	}
}

type QueueItem struct {
	Index           int64           `bson:"index"`
	FullName        string          `bson:"fullName"`
	OwnerAnyAddress string          `bson:"ownerAnyAddress"`
	OwnerEthAddress string          `bson:"ownerEthAddress"`
	SpaceId         string          `bson:"spaceId"`
	Status          QueueItemStatus `bson:"status"`
	TxCommitHash    string          `bson:"txCommitHash"`
	TxRegisterHash  string          `bson:"txRegisterHash"`
}

func nameRegisterRequestFromQueueItem(item QueueItem) *as.NameRegisterRequest {
	req := as.NameRegisterRequest{
		FullName:        item.FullName,
		OwnerAnyAddress: item.OwnerAnyAddress,
		OwnerEthAddress: item.OwnerEthAddress,
		SpaceId:         item.SpaceId,
	}
	return &req
}

func queueItemFromNameRegisterRequest(req *as.NameRegisterRequest, count int64) QueueItem {
	return QueueItem{
		Index:           count,
		FullName:        req.FullName,
		OwnerAnyAddress: req.OwnerAnyAddress,
		OwnerEthAddress: req.OwnerEthAddress,
		SpaceId:         req.SpaceId,
		Status:          OperationStatus_Initial,
	}
}
