package queue

import (
	"errors"

	"go.uber.org/zap"
)

var (
	ErrCommitFailed   = errors.New("failed to commit name")
	ErrRegisterFailed = errors.New("failed to register name")
)

func nameRegisterErrToStatus(err error) QueueItemStatus {
	if err != nil {
		log.Error("failed to process item. move state to ERROR", zap.Error(err))

		if err == ErrCommitFailed {
			return OperationStatus_CommitError
		} else if err == ErrRegisterFailed {
			return OperationStatus_RegisterError
		} else {
			return OperationStatus_Error
		}
	}

	log.Info("item processed without error. move state to COMPLETED")
	return OperationStatus_Completed
}
