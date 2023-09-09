package alchemyaa

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anyproto/any-sync/app/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const CName = "any-ns.alchemyaa"

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
		Message string `json:"message, omitempty"`
	} `json:"error, omitempty"`

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
	} `json:"error, omitempty"`

	Result string `json:"result"`
}

type JSONRPCRequestGetUserOperationReceipt struct {
	ID      int      `json:"id"`
	JSONRPC string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Hashes  []string `json:"params"`
}

// should create a GasAndPaymentStruct
func CreateRequestGasAndPaymasterData(callData []byte, sender common.Address, nonce uint64, policyID string, entryPointAddr common.Address, id int) (JSONRPCRequestGasAndPaymaster, error) {
	var req JSONRPCRequestGasAndPaymaster
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "alchemy_requestGasAndPaymasterAndData"

	var gaps GasAndPaymentStruct
	gaps.PolicyID = policyID
	gaps.EntryPoint = entryPointAddr.String()
	gaps.DummySignature = "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c"
	gaps.UserOperation.Sender = sender.String()
	gaps.UserOperation.InitCode = "0x"

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
func CreateRequest(callData []byte, rgap JSONRPCResponseGasAndPaymaster, chainID int64, entryPointAddress common.Address, sender common.Address, nonce uint64, id int, myPK string, appendEntryPoint bool) ([]byte, error) {
	var req JSONRPCRequest
	req.ID = id
	req.JSONRPC = "2.0"
	req.Method = "eth_sendUserOperation"

	var uo UserOperation
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

	dataToSign, err := GetUserOperationHash(uo, chainID, entryPointAddress)
	if err != nil {
		log.Error("failed to pack UserOperation", zap.Error(err))
		return nil, err
	}
	log.Debug("dataToSign: ", zap.String("hash", hex.EncodeToString(dataToSign)))

	sig, err := SignDataWithEthereumPrivateKey(dataToSign, myPK)
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

	// add entryPointAddress
	if appendEntryPoint {
		err, jsonDATA = AppendEntryPointAddress(jsonDATA, entryPointAddress)

		if err != nil {
			log.Error("can not append entry point", zap.Error(err))
			return nil, err
		}
	}

	return jsonDATA, nil
}

func CreateRequestGetUserOperation(operationHash string, id int) ([]byte, error) {
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

func SendRequest(apiKey string, jsonDATA []byte) ([]byte, error) {
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
