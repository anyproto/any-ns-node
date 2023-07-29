package queue

import (
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
)

type QueueItem struct {
	Index           int64             `bson:"index"`
	FullName        string            `bson:"fullName"`
	OwnerAnyAddress string            `bson:"ownerAnyAddress"`
	OwnerEthAddress string            `bson:"ownerEthAddress"`
	SpaceId         string            `bson:"spaceId"`
	Status          as.OperationState `bson:"status"`
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
		Status:          as.OperationState_Pending,
	}
}
