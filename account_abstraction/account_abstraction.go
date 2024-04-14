package accountabstraction

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"strings"

	"github.com/anyproto/any-ns-node/alchemysdk"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	asdk "github.com/anyproto/alchemy-aa-sdk/alchemysdk"
)

const CName = "any-ns.aa"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsAA{}
}

type anynsAA struct {
	confAccount   accountservice.Config
	aaConfig      config.AA
	confContracts config.Contracts
	contracts     contracts.ContractsService
	alchemy       alchemysdk.AlchemyAAService
}

type OperationInfo struct {
	OperationState nsp.OperationState
}

type AccountAbstractionService interface {
	GetOperation(ctx context.Context, operationID string) (info *OperationInfo, err error)

	// each EOA has an associated smart wallet address
	// even if it is not deployed yet - we can determine it
	GetSmartWalletAddress(ctx context.Context, eoa common.Address) (address common.Address, err error)
	IsScwDeployed(ctx context.Context, scw common.Address) (bool, error)
	GetNamesCountLeft(ctx context.Context, scw common.Address) (count uint64, err error)

	// will mint + approve tokens to the specified smart wallet
	AdminMintAccessTokens(ctx context.Context, scw common.Address, amount *big.Int) (operationID string, err error)
	// use it to register a name on behalf of a user
	AdminNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (operationID string, err error)

	// get data to sign with your PK:
	GetDataNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (dataOut []byte, contextData []byte, err error)
	GetDataNameRegisterForSpace(ctx context.Context, in *nsp.NameRegisterForSpaceRequest) (dataOut []byte, contextData []byte, err error)

	// after data is signed - now you are ready to send it
	// contextData was received from functions like GetDataNameRegister and should be left intact
	SendUserOperation(ctx context.Context, contextData []byte, signedByUserData []byte) (operationID string, err error)

	app.Component
}

func (aa *anynsAA) Init(a *app.App) (err error) {
	aa.confAccount = a.MustComponent(config.CName).(*config.Config).GetAccount()
	aa.aaConfig = a.MustComponent(config.CName).(*config.Config).GetAA()
	aa.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	aa.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	aa.alchemy = a.MustComponent(alchemysdk.CName).(alchemysdk.AlchemyAAService)

	return nil
}

func (aa *anynsAA) Name() (name string) {
	return CName
}

func (aa *anynsAA) getNextAlchemyRequestID() int {
	// TODO: return real operation ID
	return 1
}

func (aa *anynsAA) GetSmartWalletAddress(ctx context.Context, eoa common.Address) (address common.Address, err error) {
	parsedABI, err := abi.JSON(strings.NewReader(factoryContractABI))
	if err != nil {
		return common.Address{}, err
	}

	input, err := parsedABI.Pack("getAddress", eoa, big.NewInt(0))
	if err != nil {
		return common.Address{}, err
	}

	addr := common.HexToAddress(aa.aaConfig.AccountFactory)
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: input,
	}

	res, err := aa.contracts.CallContract(ctx, callMsg)
	if err != nil {
		log.Error("failed to call getAddress", zap.Error(err))
		return common.Address{}, err
	}

	out := common.BytesToAddress(res)

	log.Info("SCW address is", zap.String("address", out.Hex()))
	if out.Hex() == "0x0000000000000000000000000000000000000000" {
		return common.Address{}, errors.New("can not get SCW address")
	}

	return out, nil
}

func (aa *anynsAA) IsScwDeployed(ctx context.Context, scwa common.Address) (bool, error) {
	return aa.contracts.IsContractDeployed(ctx, scwa)
}

func (aa *anynsAA) getNonceForSmartWalletAddress(ctx context.Context, scw common.Address) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(entryPointJSON))
	if err != nil {
		return nil, err
	}

	input, err := parsedABI.Pack("getNonce", scw, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(aa.aaConfig.EntryPoint)
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: input,
	}

	res, err := aa.contracts.CallContract(ctx, callMsg)
	if err != nil {
		log.Error("failed to call getNonce", zap.Error(err))
		return nil, err
	}

	out := big.NewInt(0)
	out.SetBytes(res)
	return out, nil
}

func (aa *anynsAA) GetNamesCountLeft(ctx context.Context, scw common.Address) (count uint64, err error) {
	tokenAddress := common.HexToAddress(aa.confContracts.AddrToken)

	balance, err := aa.contracts.GetBalanceOf(ctx, tokenAddress, scw)
	if err != nil {
		log.Error("failed to get balance of", zap.Error(err), zap.String("scw", scw.String()), zap.String("tokenAddress", tokenAddress.String()))
		return 0, err
	}

	// N tokens per name (current testnet settings)
	weiPerToken := big.NewInt(1).Exp(big.NewInt(10), big.NewInt(int64(aa.confContracts.TokenDecimals)), nil)
	oneNamePriceWei := weiPerToken.Mul(big.NewInt(int64(aa.aaConfig.NameTokensPerName)), weiPerToken)

	count = balance.Div(balance, oneNamePriceWei).Uint64()

	log.Info("got token balance of SCW",
		zap.String("scw", scw.String()),
		zap.Uint64("balance", balance.Uint64()),
		zap.Uint64("name count left", count),
	)

	return count, nil
}

// Admin sends transaction to mint tokens to the specified smart wallet
func (aa *anynsAA) AdminMintAccessTokens(ctx context.Context, userScwAddress common.Address, namesCount *big.Int) (operationID string, err error) {
	// settings from config:
	entryPointAddr := common.HexToAddress(aa.aaConfig.EntryPoint)
	erc20tokenAddr := common.HexToAddress(aa.confContracts.AddrToken)

	registrarController := common.HexToAddress(aa.confContracts.AddrRegistrarConroller)
	alchemyApiKey := aa.aaConfig.AlchemyApiKey
	policyID := aa.aaConfig.GasPolicyId

	adminAddress := common.HexToAddress(aa.confContracts.AddrAdmin)
	adminPK := aa.confContracts.AdminPk

	var chainID int64 = int64(aa.aaConfig.ChainID)

	// 0 - check params
	if namesCount.Cmp(big.NewInt(0)) == 0 {
		return "", errors.New("names count is 0")
	}

	// TODO: optimize, cache it or move to settings
	// 1 - determine admin's SCW
	adminScw, err := aa.GetSmartWalletAddress(ctx, adminAddress)
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return "", err
	}

	// 2 - get nonce (from admin's SCW)
	nonce, err := aa.getNonceForSmartWalletAddress(ctx, adminScw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return "", err
	}
	log.Info("got nonce for admin", zap.String("adminScw", adminScw.String()), zap.Int64("nonce", nonce.Int64()))

	// 3 - create user operation
	// N tokens per each name (was 10 during our tests)
	tokensToMint := namesCount.Mul(namesCount, big.NewInt(int64(aa.aaConfig.NameTokensPerName)))
	tokenDecimals := aa.confContracts.TokenDecimals

	callDataOriginal, err := getCallDataForMint(userScwAddress, tokensToMint, tokenDecimals)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return "", err
	}
	log.Debug("prepared original call data", zap.String("callDataOriginal", hex.EncodeToString(callDataOriginal)))

	callDataOriginal2, err := getCallDataForAprove(userScwAddress, registrarController, tokensToMint, tokenDecimals)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return "", err
	}
	log.Debug("prepared original call data 2", zap.String("callDataOriginal2", hex.EncodeToString(callDataOriginal2)))

	// create array of call data
	targets := []common.Address{erc20tokenAddr, erc20tokenAddr}
	callDataOriginals := [][]byte{callDataOriginal, callDataOriginal2}

	// 4 - wrap it into "execute" call
	callData, err := getCallDataForBatchExecute(targets, callDataOriginals)
	if err != nil {
		log.Error("failed to get call data", zap.Error(err))
		return "", err
	}
	log.Info("prepared call data", zap.String("callData", hex.EncodeToString(callData)))

	id := aa.getNextAlchemyRequestID()

	// only specify factoryAddr if you need to instanitate a new SCW
	factoryAddr := common.Address{}
	deployed, err := aa.IsScwDeployed(ctx, adminScw)
	if err != nil {
		log.Error("failed to check if SCW is deployed", zap.Error(err))
		return "", err
	}
	if !deployed {
		factoryAddr = common.HexToAddress(aa.aaConfig.AccountFactory)
	}

	rgapd, err := aa.alchemy.CreateRequestGasAndPaymasterData(callData, adminAddress, adminScw, uint64(nonce.Int64()), policyID, entryPointAddr, factoryAddr, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return "", err
	}

	jsonDATAPre, err := json.Marshal(rgapd)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return "", err
	}

	log.Info("jsonDataPre is ready", zap.String("jsonDataPre", string(jsonDATAPre)))

	// 5 - send it
	response, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return "", err
	}

	// parse response
	responseStruct := asdk.JSONRPCResponseGasAndPaymaster{}
	err = json.Unmarshal(response, &responseStruct)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return "", err
	}
	// TODO: handle "Error code": -32500  "AA25 invalid account nonce" error
	// TODO: handle "Error code": -32500, "AA20 account not deployed"
	// TODO: handle "Error code": -32500,	"AA10 sender already constructed"
	if responseStruct.Error.Code != 0 {
		log.Error("GasAndPaymaster call failed",
			zap.Int("Error code", responseStruct.Error.Code),
			zap.String("Error message", responseStruct.Error.Message),
		)
		return "", errors.New(responseStruct.Error.Message)
	}

	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 6 - now create new transaction
	appendEntryPoint := true
	jsonDATA, err := aa.alchemy.CreateRequestAndSign(callData, responseStruct, chainID, entryPointAddr, adminAddress, adminScw, uint64(nonce.Int64()), id+1, adminPK, factoryAddr, appendEntryPoint)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return "", err
	}

	log.Info("created eth_sendUserOperation request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return "", err
	}

	log.Info("eth_sendUserOperation got response", zap.Any("response", response))

	// 7 - get op hash
	// returns err if error is in the response
	opHash, err := aa.alchemy.DecodeResponseSendRequest(response)
	if err != nil {
		log.Error("failed to decode response or error", zap.Error(err))
		return "", err
	}
	log.Info("decoded response", zap.String("opHash", opHash))

	// TODO: loop
	// GetOperation()

	log.Info("eth_getUserOperationReceipt got response", zap.Any("r", response), zap.String("opHash", opHash))
	return opHash, nil
}

func (aa *anynsAA) GetDataNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (dataOut []byte, contextData []byte, err error) {
	// do not attach SpaceID here, instead use GetDataNameRegisterForSpace method
	spaceID := ""

	// always update reverse record!
	// Example:
	// 1 - register alice.any -> ANYID123123
	// 2 - register xxxx.any -> ANYID123123
	// reverse resolve ANYID123123 will return xxxx.any
	isReverseRecordUpdate := true

	return aa.getDataNameRegister(ctx, in.FullName, in.OwnerAnyAddress, in.OwnerEthAddress, spaceID, isReverseRecordUpdate, in.RegisterPeriodMonths)
}

func (aa *anynsAA) GetDataNameRegisterForSpace(ctx context.Context, in *nsp.NameRegisterForSpaceRequest) (dataOut []byte, contextData []byte, err error) {
	// Registering for space should not update reverse record for owner
	isReverseRecordUpdate := false

	return aa.getDataNameRegister(ctx, in.FullName, in.OwnerAnyAddress, in.OwnerEthAddress, in.SpaceId, isReverseRecordUpdate, in.RegisterPeriodMonths)
}

func (aa *anynsAA) getDataNameRegister(ctx context.Context, fullName string, ownerAnyAddress string, ownerEthAddress string, spaceID string, isReverseRecordUpdate bool, registerPeriodMonths uint32) (dataOut []byte, contextData []byte, err error) {
	// settings from config:
	entryPointAddr := common.HexToAddress(aa.aaConfig.EntryPoint)
	alchemyApiKey := aa.aaConfig.AlchemyApiKey
	policyID := aa.aaConfig.GasPolicyId

	var chainID int64 = int64(aa.aaConfig.ChainID)
	var id int = aa.getNextAlchemyRequestID()

	// overwrites fullName
	fullName, err = contracts.Normalize(fullName)
	if err != nil {
		log.Error("failed to normalize name", zap.Error(err))
		return nil, nil, err
	}

	// 0 - determine users's SCW
	addr := common.HexToAddress(ownerEthAddress)
	scw, err := aa.GetSmartWalletAddress(ctx, addr)
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return nil, nil, err
	}

	// specify only if you need to instanitate a new SCW
	factoryAddr := common.Address{}

	deployed, err := aa.IsScwDeployed(ctx, scw)
	if err != nil {
		log.Error("failed to check if SCW is deployed", zap.Error(err))
		return nil, nil, err
	}
	if !deployed {
		factoryAddr = common.HexToAddress(aa.aaConfig.AccountFactory)
	}

	// 1 - get nonce
	nonce, err := aa.getNonceForSmartWalletAddress(ctx, scw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return nil, nil, err
	}
	log.Info("got nonce", zap.String("scw", scw.String()), zap.Int64("nonce", nonce.Int64()))

	// 2 - create user operation
	callData, err := aa.getCallDataForNameRegister(fullName, ownerAnyAddress, ownerEthAddress, spaceID, isReverseRecordUpdate, registerPeriodMonths)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, nil, err
	}

	rgapd, err := aa.alchemy.CreateRequestGasAndPaymasterData(callData, addr, scw, uint64(nonce.Int64()), policyID, entryPointAddr, factoryAddr, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	jsonDATAPre, err := json.Marshal(rgapd)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, nil, err
	}

	log.Info("jsonDataPre is ready", zap.String("jsonDataPre", string(jsonDATAPre)))

	// 3 - send it
	response, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return nil, nil, err
	}
	// parse response
	responseStruct := asdk.JSONRPCResponseGasAndPaymaster{}
	err = json.Unmarshal(response, &responseStruct)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return nil, nil, err
	}
	// TODO: handle "Error code": -32500  "AA25 invalid account nonce" error
	// TODO: handle "Error code": -32500, "AA20 account not deployed"
	// TODO: handle "Error code": -32500,	"AA10 sender already constructed"
	if responseStruct.Error.Code != 0 {
		log.Error("GasAndPaymaster call failed",
			zap.Int("Error code", responseStruct.Error.Code),
			zap.String("Error message", responseStruct.Error.Message),
		)
		return nil, nil, errors.New(responseStruct.Error.Message)
	}

	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 4 - get data to sign
	jsonData, uo, err := aa.alchemy.CreateRequestStep1(callData, responseStruct, chainID, entryPointAddr, scw, uint64(nonce.Int64()))
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	// serialize UserOperation to contextData
	contextData, err = json.Marshal(uo)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, nil, err
	}

	return jsonData, contextData, nil
}

func (aa *anynsAA) getCallDataForNameRegister(fullName string, ownerAnyAddress string, ownerEthAddress string, spaceID string, isReverseRecordUpdate bool, registerPeriodMonths uint32) ([]byte, error) {
	registrarControllerPrivate := common.HexToAddress(aa.confContracts.AddrRegistrarPrivateController)

	resolverAddress := common.HexToAddress(aa.confContracts.AddrResolver)
	registrantAccount := common.HexToAddress(ownerEthAddress)

	var nameFirstPart string = contracts.RemoveTLD(fullName)
	var regTime = contracts.PeriodMonthsToTimestamp(registerPeriodMonths)

	var ownerControlledFuses uint16 = 0

	// 1 - get new random secret
	secret32, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Error("can not generate random secret", zap.Error(err))
		return nil, err
	}

	// 2 - create a commitment
	controller, err := aa.contracts.ConnectToPrivateController()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	commitment, err := aa.contracts.MakeCommitment(&contracts.MakeCommitmentParams{
		NameFirstPart:         nameFirstPart,
		RegistrantAccount:     registrantAccount,
		Secret:                secret32,
		Controller:            controller,
		FullName:              fullName,
		OwnerAnyAddr:          ownerAnyAddress,
		SpaceId:               spaceID,
		IsReverseRecordUpdate: isReverseRecordUpdate,
		RegisterPeriodMonths:  registerPeriodMonths})

	if err != nil {
		log.Error("can not calculate a commitment", zap.Error(err))
		return nil, err
	}

	// 2 - generate original callData (that will set name)
	callData, err := contracts.PrepareCallData_SetContentHashSpaceID(fullName, ownerAnyAddress, spaceID)
	if err != nil {
		log.Error("can not prepare call data", zap.Error(err))
		return nil, err
	}

	// 4 - now prepare 2 operations
	callDataOriginal1, err := getCallDataForCommit(commitment)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, err
	}

	callDataOriginal2, err := getCallDataForRegister(
		nameFirstPart,
		registrantAccount,
		regTime,
		secret32,
		resolverAddress,
		callData,
		isReverseRecordUpdate,
		ownerControlledFuses)

	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, err
	}

	// create array of call data
	targets := []common.Address{registrarControllerPrivate, registrarControllerPrivate}
	callDataOriginals := [][]byte{callDataOriginal1, callDataOriginal2}

	// 4 - wrap it into "execute" call
	executeCallDataOut, err := getCallDataForBatchExecute(targets, callDataOriginals)
	if err != nil {
		log.Error("failed to get call data", zap.Error(err))
		return nil, err
	}

	return executeCallDataOut, nil
}

// after data is signed - now you are ready to send it
func (aa *anynsAA) SendUserOperation(ctx context.Context, contextData []byte, signedByUserData []byte) (operationID string, err error) {
	entryPointAddr := common.HexToAddress(aa.aaConfig.EntryPoint)
	requestId := aa.getNextAlchemyRequestID()

	// 1 - Unmarshal UserOperations from contextData
	var uo asdk.UserOperation
	err = json.Unmarshal(contextData, &uo)
	if err != nil {
		log.Error("can not unmarshal JSON", zap.Error(err))
		return "", err
	}

	// TODO: check that data is signed by user correctly with Eth private key?

	data, err := aa.alchemy.CreateRequestStep2(requestId, signedByUserData, uo, entryPointAddr)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return "", err
	}

	// 2 - send it
	response, err := aa.alchemy.SendRequest(aa.aaConfig.AlchemyApiKey, data)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return "", err
	}

	// 3 - get op hash
	// will handle error in response
	opHash, err := aa.alchemy.DecodeResponseSendRequest(response)
	if err != nil {
		log.Error("failed to decode response", zap.Error(err))
		return "", err
	}
	log.Info("decoded response", zap.String("opHash", opHash))

	// TODO: loop
	//_, _ = aa.GetOperation(ctx, opHash)

	return opHash, nil
}

func (aa *anynsAA) GetOperation(ctx context.Context, operationID string) (*OperationInfo, error) {
	alchemyApiKey := aa.aaConfig.AlchemyApiKey

	var out OperationInfo

	// 1 - eth_getUserOperationReceipt
	//returns null if operation is PENDING or NOT FOUND
	//   right now there are no any means to check if operation is pending
	// 	 or not found, so we always return PENDING
	//returns success==true if COMPLETED
	//returns success==false if FAILED

	// ID is zero
	id := 0
	jsonDATA, err := aa.alchemy.CreateRequestGetUserOperationReceipt(operationID, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return nil, err
	}

	log.Info("created eth_getUserOperationReceipt request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	res, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return nil, err
	}

	// if operation is pending or not found ->
	// {"jsonrpc":"2.0","id":1,"result":null}
	uoRes, err := aa.alchemy.DecodeResponseGetUserOperationReceipt(res)

	if err != nil || (uoRes.Error.Code != 0) {
		log.Info("can not decode operation response", zap.String("operation", operationID), zap.Error(err))

		// additional check - does not work too :-(
		//return aa.getUserOperationByHash(ctx, operationID)

		out.OperationState = nsp.OperationState_Error
		return &out, nil
	}

	if uoRes.Result.UserOpHash == "" {
		log.Info("operation is not found", zap.String("operation", operationID), zap.Error(err))

		// additional check - does not work too :-(
		//return aa.getUserOperationByHash(ctx, operationID)

		out.OperationState = nsp.OperationState_PendingOrNotFound
		return &out, nil
	}

	// return results
	if uoRes.Result.Success {
		out.OperationState = nsp.OperationState_Completed
	} else {
		out.OperationState = nsp.OperationState_Error
	}

	return &out, nil
}

/*
func (aa *anynsAA) getUserOperationByHash(ctx context.Context, operationID string) (*OperationInfo, error) {
	alchemyApiKey := aa.aaConfig.AlchemyApiKey

	var out OperationInfo

	id := 0
	req, err := aa.alchemy.CreateRequestGetUserOperationByHash(operationID, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return nil, err
	}

	jsonDATA, err := json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return nil, err
	}

	log.Info("created eth_getUserOperationReceipt request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	res, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return nil, err
	}

	// if operation is pending or not found ->
	// {"jsonrpc":"2.0","id":1,"result":null}
	uoRes, err := aa.alchemy.DecodeResponseGetUserOperationByHash(res)

	if err != nil || (uoRes.Error.Code != 0) {
		log.Info("can not decode user operation response.", zap.String("operation", operationID), zap.Error(err))
		out.OperationState = nsp.OperationState_Error
		return &out, nil
	}

	if uoRes.Result.UserOperation == "" {
		log.Info("user operation is not found", zap.String("operation", operationID), zap.Error(err))
		out.OperationState = nsp.OperationState_Error
		return &out, nil
	}

	out.OperationState = nsp.OperationState_Success
	return &out, nil
}
*/

func (aa *anynsAA) AdminNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (operationID string, err error) {
	// settings from config:
	entryPointAddr := common.HexToAddress(aa.aaConfig.EntryPoint)

	alchemyApiKey := aa.aaConfig.AlchemyApiKey
	policyID := aa.aaConfig.GasPolicyId

	adminAddress := common.HexToAddress(aa.confContracts.AddrAdmin)
	adminPK := aa.confContracts.AdminPk

	var chainID int64 = int64(aa.aaConfig.ChainID)

	// overwrites in.FullName!
	in.FullName, err = contracts.Normalize(in.FullName)
	if err != nil {
		log.Error("failed to normalize name", zap.Error(err))
		return "", err
	}

	// 0 - use SCW?
	nameOwnerEthAddress := in.OwnerEthAddress

	if in.RegisterToSmartContractWallet {
		// get SCW of the in.OwnerEthAddress
		addr, err := aa.GetSmartWalletAddress(ctx, common.HexToAddress(in.OwnerEthAddress))
		if err != nil {
			log.Error("failed to get smart wallet address", zap.Error(err))
			return "", err
		}
		nameOwnerEthAddress = addr.String()
		log.Info("RegisterToSmartContractWallet was true. Using SCW to register name",
			zap.String("FullName", in.FullName),
			zap.String("OwnerEthAddress", in.OwnerEthAddress),
			zap.String("SCW", nameOwnerEthAddress),
		)
	}

	// TODO: optimize, cache it or move to settings
	// 1 - determine admin's SCW
	adminScw, err := aa.GetSmartWalletAddress(ctx, adminAddress)
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return "", err
	}

	// 2 - get nonce (from admin's SCW)
	nonce, err := aa.getNonceForSmartWalletAddress(ctx, adminScw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return "", err
	}
	log.Info("got nonce for admin", zap.String("adminScw", adminScw.String()), zap.Int64("nonce", nonce.Int64()))

	// 3 - create user operation
	spaceID := ""
	isReverseRecordUpdate := true

	callData, err := aa.getCallDataForNameRegister(in.FullName, in.OwnerAnyAddress, nameOwnerEthAddress, spaceID, isReverseRecordUpdate, in.RegisterPeriodMonths)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return "", err
	}

	log.Debug("prepared call data", zap.String("callData", hex.EncodeToString(callData)))
	id := aa.getNextAlchemyRequestID()

	// only specify factoryAddr if you need to instanitate a new SCW
	factoryAddr := common.Address{}
	deployed, err := aa.IsScwDeployed(ctx, adminScw)
	if err != nil {
		log.Error("failed to check if SCW is deployed", zap.Error(err))
		return "", err
	}
	if !deployed {
		factoryAddr = common.HexToAddress(aa.aaConfig.AccountFactory)
	}

	rgapd, err := aa.alchemy.CreateRequestGasAndPaymasterData(callData, adminAddress, adminScw, uint64(nonce.Int64()), policyID, entryPointAddr, factoryAddr, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return "", err
	}

	jsonDATAPre, err := json.Marshal(rgapd)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return "", err
	}

	log.Debug("jsonDataPre is ready", zap.String("jsonDataPre", string(jsonDATAPre)))

	// 5 - send it
	response, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return "", err
	}

	// parse response
	responseStruct := asdk.JSONRPCResponseGasAndPaymaster{}
	err = json.Unmarshal(response, &responseStruct)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return "", err
	}
	// TODO: handle "Error code": -32500  "AA25 invalid account nonce" error
	// TODO: handle "Error code": -32500, "AA20 account not deployed"
	// TODO: handle "Error code": -32500,	"AA10 sender already constructed"
	if responseStruct.Error.Code != 0 {
		log.Error("GasAndPaymaster call failed",
			zap.Int("Error code", responseStruct.Error.Code),
			zap.String("Error message", responseStruct.Error.Message),
		)
		return "", errors.New(responseStruct.Error.Message)
	}

	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 6 - now create new transaction
	appendEntryPoint := true
	jsonDATA, err := aa.alchemy.CreateRequestAndSign(callData, responseStruct, chainID, entryPointAddr, adminAddress, adminScw, uint64(nonce.Int64()), id+1, adminPK, factoryAddr, appendEntryPoint)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return "", err
	}

	log.Info("created eth_sendUserOperation request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return "", err
	}

	log.Info("eth_sendUserOperation got response", zap.Any("response", response))

	// 7 - get op hash
	// returns err if error is in the response
	opHash, err := aa.alchemy.DecodeResponseSendRequest(response)
	if err != nil {
		log.Error("failed to decode response or error", zap.Error(err))
		return "", err
	}
	log.Info("decoded response", zap.String("opHash", opHash))

	// TODO: loop
	// GetOperation()
	log.Info("eth_getUserOperationReceipt got response", zap.Any("r", response), zap.String("opHash", opHash))
	return opHash, nil
}
