package alchemysdk

import (
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/ethereum/go-ethereum/common"

	asdk "github.com/anyproto/alchemy-aa-sdk/alchemysdk"
)

const CName = "any-ns.alchemysdk"

var log = logger.NewNamed(CName)

type alchemysdk struct {
}

// A simple wrapper around github.com/anyproto/alchemy-aa-sdk/alchemysdk
// We need it to mock the library
type AlchemyAAService interface {
	// if factoryAddr is non-null -> will set init code
	CreateRequestGasAndPaymasterData(callData []byte, sender common.Address, senderScw common.Address, nonce uint64, policyID string, entryPointAddr common.Address, factoryAddr common.Address, id int) (asdk.JSONRPCRequestGasAndPaymaster, error)
	CreateRequestAndSign(callData []byte, rgap asdk.JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, senderScw common.Address, nonce uint64, id int, myPK string, factoryAddr common.Address, appendEntryPoint bool) ([]byte, error)

	// can be used to send any type of request to Alchemy
	SendRequest(apiKey string, jsonDATA []byte) ([]byte, error)
	DecodeResponseSendRequest(response []byte) (opHash string, err error)

	CreateRequestGetUserOperationReceipt(operationHash string, id int) ([]byte, error)
	DecodeResponseGetUserOperationReceipt(response []byte) (ret *asdk.JSONRPCResponseGetOp, err error)

	//CreateRequestGetUserOperationByHash(operationHash string, id int) ([]byte, error)
	//DecodeResponseGetUserOperationByHash(response []byte) (ret *JSONRPCResponseGetUserOpByHash, err error)

	// creates a UserOperation and data to sign with user's private key
	CreateRequestStep1(callData []byte, rgap asdk.JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, nonce uint64) (dataToSign []byte, uo asdk.UserOperation, err error)
	// adds signature to UserOperation and creates final JSONRPCRequest that can be sent with 'SendRequest'
	CreateRequestStep2(alchemyRequestId int, signedByUserData []byte, uo asdk.UserOperation, entryPointAddr common.Address) ([]byte, error)

	app.Component
}

func New() app.Component {
	return &alchemysdk{}
}

func (aa *alchemysdk) Init(a *app.App) (err error) {

	return nil
}

func (aa *alchemysdk) Name() (name string) {
	return CName
}

// should create a GasAndPaymentStruct
func (aa *alchemysdk) CreateRequestGasAndPaymasterData(callData []byte, sender common.Address, senderScw common.Address, nonce uint64, policyID string, entryPointAddr common.Address, factoryAddr common.Address, id int) (asdk.JSONRPCRequestGasAndPaymaster, error) {
	return asdk.CreateRequestGasAndPaymasterData(callData, sender, senderScw, nonce, policyID, entryPointAddr, factoryAddr, id)
}

// creates a JSONRPCRequest with "eth_sendUserOperation" formatted data
func (aa *alchemysdk) CreateRequestAndSign(callData []byte, rgap asdk.JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, senderScw common.Address, nonce uint64, id int, myPK string, factoryAddr common.Address, appendEntryPoint bool) ([]byte, error) {
	return asdk.CreateRequestAndSign(callData, rgap, chainID, entryPointAddr, sender, senderScw, nonce, id, myPK, factoryAddr, appendEntryPoint)
}

// creates a JSONRPCRequest with "eth_getUserOperationReceipt" formatted data
func (aa *alchemysdk) CreateRequestGetUserOperationReceipt(operationHash string, id int) ([]byte, error) {
	return asdk.CreateRequestGetUserOperationReceipt(operationHash, id)
}

func (aa *alchemysdk) DecodeResponseSendRequest(response []byte) (opHash string, err error) {
	return asdk.DecodeResponseSendRequest(response)
}

func (aa *alchemysdk) DecodeResponseGetUserOperationReceipt(response []byte) (ret *asdk.JSONRPCResponseGetOp, err error) {
	return asdk.DecodeResponseGetUserOperationReceipt(response)
}

// creates data to sign with UserOperation
func (aa *alchemysdk) CreateRequestStep1(callData []byte, rgap asdk.JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, nonce uint64) (dataToSign []byte, uo asdk.UserOperation, err error) {
	return asdk.CreateRequestStep1(callData, rgap, chainID, entryPointAddr, sender, nonce)
}

func (aa *alchemysdk) CreateRequestStep2(alchemyRequestId int, signedByUserData []byte, uo asdk.UserOperation, entryPointAddr common.Address) ([]byte, error) {
	return asdk.CreateRequestStep2(alchemyRequestId, signedByUserData, uo, entryPointAddr)
}

// TODO: unused
// TODO: test
func (aa *alchemysdk) CreateRequestGetUserOperationByHash(operationHash string, id int) ([]byte, error) {
	return asdk.CreateRequestGetUserOperationByHash(operationHash, id)
}

// TODO: unused
// TODO: test
func (aa *alchemysdk) DecodeResponseGetUserOperationByHash(response []byte) (ret *asdk.JSONRPCResponseGetUserOpByHash, err error) {
	return asdk.DecodeResponseGetUserOperationByHash(response)
}
