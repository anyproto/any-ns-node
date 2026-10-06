package contracts

import (
	"errors"
)

var (
	ErrNonceTooLow  = errors.New("nonce too low")
	ErrNonceTooHigh = errors.New("nonce too high")

	// ErrNotAContract is returned by GetScwOwner when there is no code at the address:
	// it is either an EOA or a smart contract wallet that is not deployed yet
	ErrNotAContract = errors.New("address is not a smart contract")
)
