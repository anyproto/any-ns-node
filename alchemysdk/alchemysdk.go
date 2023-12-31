package alchemysdk

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const CName = "any-ns.alchemysdk"

var log = logger.NewNamed(CName)

type EntryPointAddress interface{}

type UserOperation struct {
	Sender               string `json:"sender,omitempty"`
	Nonce                string `json:"nonce,omitempty"`
	InitCode             string `json:"initCode,omitempty"`
	CallData             string `json:"callData,omitempty"`
	Signature            string `json:"signature,omitempty"`
	CallGasLimit         string `json:"callGasLimit,omitempty"`
	VerificationGasLimit string `json:"verificationGasLimit,omitempty"`
	PreVerificationGas   string `json:"preVerificationGas,omitempty"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
	PaymasterAndData     string `json:"paymasterAndData,omitempty"`
}

type GasAndPaymentStruct struct {
	PolicyID       string        `json:"policyId"`
	EntryPoint     string        `json:"entryPoint"`
	UserOperation  UserOperation `json:"userOperation"`
	DummySignature string        `json:"dummySignature"`
}

type JSONRPCRequestGasAndPaymaster struct {
	ID      int                   `json:"id"`
	JSONRPC string                `json:"jsonrpc"`
	Method  string                `json:"method"`
	Params  []GasAndPaymentStruct `json:"params"`
}

type JSONRPCResponseGasAndPaymaster struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`

	Error struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`

	Result struct {
		PreVerificationGas   string `json:"preVerificationGas,omitempty"`
		CallGasLimit         string `json:"callGasLimit,omitempty"`
		VerificationGasLimit string `json:"verificationGasLimit,omitempty"`
		PaymasterAndData     string `json:"paymasterAndData,omitempty"`
		MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
		MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
	} `json:"result"`
}

type JSONRPCRequest struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  []UserOperation `json:"params"`
}

type JSONRPCResponseUserOpHash struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`

	Error struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`

	Result string `json:"result"`
}

type JSONRPCResponseGetOp struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`

	Error struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`

	Result struct {
		UserOpHash string `json:"userOpHash,omitempty"`
		Success    bool   `json:"success,omitempty"`
	} `json:"result"`
}

type JSONRPCResponseGetUserOpByHash struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`

	Error struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`

	Result struct {
		UserOperation   string `json:"userOperation,omitempty"`
		BlockNumber     string `json:"blockNumber,omitempty"`
		BlockHash       string `json:"blockNumber,omitempty"`
		TransactionHash string `json:"blockNumber,omitempty"`
	} `json:"result"`
}

type JSONRPCRequestGetUserOperationReceipt struct {
	ID      int      `json:"id"`
	JSONRPC string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Hashes  []string `json:"params"`
}

type alchemysdk struct {
}

type AlchemyAAService interface {
	// if factoryAddr is non-null -> will set init code
	CreateRequestGasAndPaymasterData(callData []byte, sender common.Address, senderScw common.Address, nonce uint64, policyID string, entryPointAddr common.Address, factoryAddr common.Address, id int) (JSONRPCRequestGasAndPaymaster, error)
	CreateRequestAndSign(callData []byte, rgap JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, senderScw common.Address, nonce uint64, id int, myPK string, factoryAddr common.Address, appendEntryPoint bool) ([]byte, error)
	// can be used to send any type of request to Alchemy
	SendRequest(apiKey string, jsonDATA []byte) ([]byte, error)
	DecodeResponseSendRequest(response []byte) (opHash string, err error)

	CreateRequestGetUserOperationReceipt(operationHash string, id int) ([]byte, error)
	DecodeResponseGetUserOperationReceipt(response []byte) (ret *JSONRPCResponseGetOp, err error)

	//CreateRequestGetUserOperationByHash(operationHash string, id int) ([]byte, error)
	//DecodeResponseGetUserOperationByHash(response []byte) (ret *JSONRPCResponseGetUserOpByHash, err error)

	// creates a UserOperation and data to sign with user's private key
	CreateRequestStep1(callData []byte, rgap JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, nonce uint64) (dataToSign []byte, uo UserOperation, err error)
	// adds signature to UserOperation and creates final JSONRPCRequest that can be sent with 'SendRequest'
	CreateRequestStep2(alchemyRequestId int, signedByUserData []byte, uo UserOperation, entryPointAddr common.Address) ([]byte, error)

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
func (aa *alchemysdk) CreateRequestGasAndPaymasterData(callData []byte, sender common.Address, senderScw common.Address, nonce uint64, policyID string, entryPointAddr common.Address, factoryAddr common.Address, id int) (JSONRPCRequestGasAndPaymaster, error) {
	var req JSONRPCRequestGasAndPaymaster
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "alchemy_requestGasAndPaymasterAndData"

	var gaps GasAndPaymentStruct
	gaps.PolicyID = policyID
	gaps.EntryPoint = entryPointAddr.String()
	gaps.DummySignature = "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c"
	gaps.UserOperation.Sender = senderScw.String()

	// set InitCode
	if (factoryAddr != common.Address{}) {
		log.Debug("factoryAddr is not null. Initializing SCW")

		code, err := getAccountInitCode(sender, factoryAddr)
		if err != nil {
			log.Error("failed to get init code", zap.Error(err))
			return req, err
		}

		gaps.UserOperation.InitCode = "0x" + hex.EncodeToString(code)
	} else {
		gaps.UserOperation.InitCode = "0x"
	}

	nonceHexStr := fmt.Sprintf("0x%x", nonce)

	gaps.UserOperation.Nonce = nonceHexStr
	gaps.UserOperation.CallData = "0x" + hex.EncodeToString(callData)
	gaps.UserOperation.Signature = "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c"
	gaps.UserOperation.PaymasterAndData = "0x"
	gaps.UserOperation.MaxFeePerGas = "0x0"
	gaps.UserOperation.MaxPriorityFeePerGas = "0x0"
	gaps.UserOperation.CallGasLimit = "0x0"
	gaps.UserOperation.PreVerificationGas = "0x0"
	gaps.UserOperation.VerificationGasLimit = "0x0"

	// Add our UserOperation to the list
	req.Params = append(req.Params, gaps)
	return req, nil
}

// creates a JSONRPCRequest with "eth_sendUserOperation" formatted data
func (aa *alchemysdk) CreateRequestAndSign(callData []byte, rgap JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, senderScw common.Address, nonce uint64, id int, myPK string, factoryAddr common.Address, appendEntryPoint bool) ([]byte, error) {
	var req JSONRPCRequest
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "eth_sendUserOperation"

	var uo UserOperation
	uo.Sender = senderScw.String()
	uo.CallData = "0x" + hex.EncodeToString(callData)

	// convert nonce to hex string
	nonceHexStr := fmt.Sprintf("0x%x", nonce)
	uo.Nonce = nonceHexStr

	// set InitCode
	if (factoryAddr != common.Address{}) {
		log.Debug("factoryAddr is not null. Initializing SCW")

		code, err := getAccountInitCode(sender, factoryAddr)
		if err != nil {
			log.Error("failed to get init code", zap.Error(err))
			return nil, err
		}

		uo.InitCode = "0x" + hex.EncodeToString(code)
	} else {
		uo.InitCode = "0x"
	}

	uo.CallGasLimit = rgap.Result.CallGasLimit
	uo.VerificationGasLimit = rgap.Result.VerificationGasLimit
	uo.PreVerificationGas = rgap.Result.PreVerificationGas
	uo.MaxFeePerGas = rgap.Result.MaxFeePerGas
	uo.MaxPriorityFeePerGas = rgap.Result.MaxPriorityFeePerGas
	uo.PaymasterAndData = rgap.Result.PaymasterAndData

	dataToSign, err := getUserOperationHash(uo, chainID, entryPointAddr)
	if err != nil {
		log.Error("failed to pack UserOperation", zap.Error(err))
		return nil, err
	}
	log.Debug("dataToSign: ", zap.String("hash", hex.EncodeToString(dataToSign)))

	sig, err := signDataWithEthereumPrivateKey(dataToSign, myPK)
	if err != nil {
		log.Error("failed to sign", zap.Error(err))
		return nil, err
	}
	log.Debug("signed: ", zap.String("sig", hex.EncodeToString(sig)))

	uo.Signature = "0x" + hex.EncodeToString(sig)

	// Add our UserOperation to the list
	req.Params = append(req.Params, uo)

	// 2 - convert struct to json
	jsonDATA, err := json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, err
	}

	// add entryPointAddr
	if appendEntryPoint {
		jsonDATA, err = appendEntryPointAddress(jsonDATA, entryPointAddr)

		if err != nil {
			log.Error("can not append entry point", zap.Error(err))
			return nil, err
		}
	}

	return jsonDATA, nil
}

// creates a JSONRPCRequest with "eth_getUserOperationReceipt" formatted data
func (aa *alchemysdk) CreateRequestGetUserOperationReceipt(operationHash string, id int) ([]byte, error) {
	// {"jsonrpc":"2.0","id":11,"method":"eth_getUserOperationReceipt","params":["0x5fad93d239e4e7a7dd634822513b27f04e57ed8ea1be7b3e74df177eefd8beb8"]}
	var req JSONRPCRequestGetUserOperationReceipt
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "eth_getUserOperationReceipt"
	req.Hashes = append(req.Hashes, operationHash)

	// 2 - convert struct to json
	jsonDATA, err := json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, err
	}

	return jsonDATA, nil
}

func (aa *alchemysdk) SendRequest(apiKey string, jsonDATA []byte) ([]byte, error) {
	payload := strings.NewReader(string(jsonDATA))

	url := "https://eth-sepolia.g.alchemy.com/v2/" + apiKey
	r, _ := http.NewRequest("POST", url, payload)

	r.Header.Add("accept", "application/json")
	r.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error("failed to send data", zap.Error(err))
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("failed to read response", zap.Error(err))
		return nil, err
	}

	log.Debug("sent Alchemy request", zap.String("response", string(body)))
	return body, nil
}

func (aa *alchemysdk) DecodeResponseSendRequest(response []byte) (opHash string, err error) {
	// {"jsonrpc":"2.0","id":2,"result":"0x31b09cc37a91866b493ee9a31980e90b94b09195a85599f5e6d6a246c9e20186"}
	// 1 - parse JSON
	var responseStruct2 JSONRPCResponseUserOpHash
	err = json.Unmarshal(response, &responseStruct2)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return "", err
	}

	if responseStruct2.Error.Code != 0 {
		strErr := fmt.Sprintf("Error: %v - %v", responseStruct2.Error.Code, responseStruct2.Error.Message)
		return "", errors.New(strErr)
	}

	return responseStruct2.Result, nil
}

func (aa *alchemysdk) DecodeResponseGetUserOperationReceipt(response []byte) (ret *JSONRPCResponseGetOp, err error) {
	// {"jsonrpc":"2.0","id":2,"result":{"success": true}}
	// 1 - parse JSON
	var responseStruct2 JSONRPCResponseGetOp
	err = json.Unmarshal(response, &responseStruct2)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	if responseStruct2.Error.Code != 0 {
		strErr := fmt.Sprintf("Error: %v - %v", responseStruct2.Error.Code, responseStruct2.Error.Message)
		return nil, errors.New(strErr)
	}

	return &responseStruct2, nil
}

// creates data to sign with UserOperation
func (aa *alchemysdk) CreateRequestStep1(callData []byte, rgap JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddr common.Address, sender common.Address, nonce uint64) (dataToSign []byte, uo UserOperation, err error) {
	uo = UserOperation{}

	uo.Sender = sender.String()
	uo.CallData = "0x" + hex.EncodeToString(callData)

	// convert nonce to hex string
	nonceHexStr := fmt.Sprintf("0x%x", nonce)
	uo.Nonce = nonceHexStr
	uo.InitCode = "0x"
	uo.CallGasLimit = rgap.Result.CallGasLimit
	uo.VerificationGasLimit = rgap.Result.VerificationGasLimit
	uo.PreVerificationGas = rgap.Result.PreVerificationGas
	uo.MaxFeePerGas = rgap.Result.MaxFeePerGas
	uo.MaxPriorityFeePerGas = rgap.Result.MaxPriorityFeePerGas
	uo.PaymasterAndData = rgap.Result.PaymasterAndData

	// data should be signed and then set in CreateRequestStep2
	// uo.Signature =

	dataToSign, err = getUserOperationHash(uo, chainID, entryPointAddr)
	if err != nil {
		log.Error("failed to pack UserOperation", zap.Error(err))
		return nil, uo, err
	}
	log.Debug("dataToSign: ", zap.String("hash", hex.EncodeToString(dataToSign)))

	// user now should sign that data with his PK
	return dataToSign, uo, nil
}

func (aa *alchemysdk) CreateRequestStep2(alchemyRequestId int, signedByUserData []byte, uo UserOperation, entryPointAddr common.Address) ([]byte, error) {
	var req JSONRPCRequest
	req.ID = alchemyRequestId
	req.JSONRPC = "2.0"
	req.Method = "eth_sendUserOperation"

	uo.Signature = "0x" + hex.EncodeToString(signedByUserData)

	// add our UserOperation to the list
	req.Params = append(req.Params, uo)

	// convert struct to json
	jsonDATA, err := json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, err
	}

	// add entryPointAddr
	jsonDATA, err = appendEntryPointAddress(jsonDATA, entryPointAddr)
	if err != nil {
		log.Error("can not append entry point", zap.Error(err))
		return nil, err
	}

	return jsonDATA, nil
}

// TODO: unused
// TODO: test
func (aa *alchemysdk) CreateRequestGetUserOperationByHash(operationHash string, id int) ([]byte, error) {
	// {"jsonrpc":"2.0","id":11,"method":"eth_getUserOperationByHash","params":["0x5fad93d239e4e7a7dd634822513b27f04e57ed8ea1be7b3e74df177eefd8beb8"]}
	var req JSONRPCRequestGetUserOperationReceipt
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "eth_getUserOperationByHash"
	req.Hashes = append(req.Hashes, operationHash)

	// 2 - convert struct to json
	jsonDATA, err := json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, err
	}

	return jsonDATA, nil
}

// TODO: unused
// TODO: test
func (aa *alchemysdk) DecodeResponseGetUserOperationByHash(response []byte) (ret *JSONRPCResponseGetUserOpByHash, err error) {
	// {"jsonrpc":"2.0","id":2,"result":{"success": true}}
	// 1 - parse JSON
	var responseStruct2 JSONRPCResponseGetUserOpByHash
	err = json.Unmarshal(response, &responseStruct2)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return nil, err
	}

	if responseStruct2.Error.Code != 0 {
		strErr := fmt.Sprintf("Error: %v - %v", responseStruct2.Error.Code, responseStruct2.Error.Message)
		return nil, errors.New(strErr)
	}

	return &responseStruct2, nil
}
