package contracts

import (
	"errors"
)

var (
	ErrNonceTooLow  = errors.New("nonce too low")
	ErrNonceTooHigh = errors.New("nonce too high")
)
