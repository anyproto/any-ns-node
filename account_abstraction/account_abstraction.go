package accountabstraction

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"strings"

	"github.com/anyproto/any-ns-node/alchemysdk"
	"github.com/anyproto/any-ns-node/anynsrpc"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	as "github.com/anyproto/any-ns-node/pb/anyns_api"
	commonaccount "github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const CName = "any-ns.aa"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsAA{}
}

type anynsAA struct {
	accountConfig   commonaccount.Config
	aaConfig        config.AA
	contractsConfig config.Contracts
	contracts       contracts.ContractsService
	alchemy         alchemysdk.AlchemyAAService
}

type AccountAbstractionService interface {
	// each EOA has an associated smart wallet address
	// even if it is not deployed yet - we can determine it
	GetSmartWalletAddress(ctx context.Context, eoa common.Address) (address common.Address, err error)
	GetNamesCountLeft(ctx context.Context, scw common.Address) (count uint64, err error)
	GetOperationsCountLeft(ctx context.Context, scw common.Address) (count uint64, err error)

	// TODO: implement
	//SendUserOperation(ctx context.Context, uo UserOperation, signedByUserData []byte) (err error)

	// will return error if signature is invalid
	AdminVerifyIdentity(payload []byte, signature []byte) (err error)
	// will mint + approve tokens to the specified smart wallet
	AdminMintAccessTokens(ctx context.Context, scw common.Address, amount *big.Int) (err error)

	// get data to sign
	GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (dataOut []byte, contextData []byte, err error)

	app.Component
}

func (aa *anynsAA) Init(a *app.App) (err error) {
	aa.accountConfig = a.MustComponent(config.CName).(*config.Config).GetAccount()
	aa.aaConfig = a.MustComponent(config.CName).(*config.Config).GetAA()
	aa.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
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
	factoryContractABI := `
		[
			{
				"inputs": [
					{
						"internalType": "address",
						"name": "owner",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "salt",
						"type": "uint256"
					}
				],
				"name": "getAddress",
				"outputs": [
					{
						"internalType": "address",
						"name": "",
						"type": "address"
					}
				],
				"stateMutability": "view",
				"type": "function"
			}
		]
	`

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

func (aa *anynsAA) getNonceForSmartWalletAddress(ctx context.Context, scw common.Address) (*big.Int, error) {

	entryPointJSON := `
		[
			{
				"inputs": [
					{
						"internalType": "address",
						"name": "sender",
						"type": "address"
					},
					{
						"internalType": "uint192",
						"name": "key",
						"type": "uint192"
					}
				],
				"name": "getNonce",
				"outputs": [
					{
						"internalType": "uint256",
						"name": "nonce",
						"type": "uint256"
					}
				],
				"stateMutability": "view",
				"type": "function"
			}
		]
		`

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
	tokenAddress := common.HexToAddress(aa.contractsConfig.AddrToken)

	client, err := aa.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create eth connection", zap.Error(err))
		return 0, err
	}

	balance, err := aa.contracts.GetBalanceOf(ctx, client, tokenAddress, scw)
	if err != nil {
		log.Error("failed to get balance of", zap.Error(err), zap.String("scw", scw.String()), zap.String("tokenAddress", tokenAddress.String()))
		return 0, err
	}

	// TODO: remove hardcode and move to pricing methods
	// $20 USD per name (current testnet settings)
	// $1 USD = 10^6
	oneNamePriceWei := big.NewInt(20 * 1000000)

	count = balance.Div(balance, oneNamePriceWei).Uint64()

	log.Info("got token balance of SCW",
		zap.String("scw", scw.String()),
		zap.Uint64("balance", balance.Uint64()),
		zap.Uint64("name count left", count),
	)

	return count, nil
}

func (aa *anynsAA) GetOperationsCountLeft(ctx context.Context, scw common.Address) (count uint64, err error) {
	// TODO: implement

	return 0, nil
}

func (aa *anynsAA) AdminVerifyIdentity(payload []byte, signature []byte) (err error) {
	// 1 - load public key of admin
	ownerAnyIdentity, err := crypto.DecodePeerId(aa.accountConfig.PeerId)

	if err != nil {
		log.Error("failed to unmarshal public key", zap.Error(err))
		return err
	}

	// 2 - verify signature
	res, err := ownerAnyIdentity.Verify(payload, signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// success
	return nil
}

// Admin sends transaction to mint tokens to the specified smart wallet
func (aa *anynsAA) AdminMintAccessTokens(ctx context.Context, userScwAddress common.Address, namesCount *big.Int) (err error) {
	// settings from config:
	entryPointAddress := common.HexToAddress(aa.aaConfig.EntryPoint)
	erc20tokenAddr := common.HexToAddress(aa.contractsConfig.AddrToken)
	registrarController := common.HexToAddress(aa.contractsConfig.AddrRegistrar)
	alchemyApiKey := aa.aaConfig.AlchemyApiKey
	policyID := aa.aaConfig.GasPolicyId

	adminAddress := common.HexToAddress(aa.contractsConfig.AddrAdmin)
	adminPK := aa.contractsConfig.AdminPk

	var chainID int64 = int64(aa.aaConfig.ChainID)

	// 0 - check params
	if namesCount.Cmp(big.NewInt(0)) == 0 {
		return errors.New("names count is 0")
	}

	// TODO: optimize, cache it or move to settings
	// 1 - determine admin's SCW
	adminScw, err := aa.GetSmartWalletAddress(ctx, adminAddress)
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return err
	}

	// 2 - get nonce (from admin's SCW)
	nonce, err := aa.getNonceForSmartWalletAddress(ctx, adminScw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return err
	}
	log.Info("got nonce for admin", zap.String("adminScw", adminScw.String()), zap.Int64("nonce", nonce.Int64()))

	// 3 - create user operation
	// TODO: change amount
	callDataOriginal, err := GetCallDataForMint(userScwAddress, 100)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return err
	}
	log.Debug("prepared original call data", zap.String("callDataOriginal", hex.EncodeToString(callDataOriginal)))

	callDataOriginal2, err := GetCallDataForAprove(userScwAddress, registrarController, 100)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return err
	}
	log.Debug("prepared original call data 2", zap.String("callDataOriginal2", hex.EncodeToString(callDataOriginal2)))

	// create array of call data
	targets := []common.Address{erc20tokenAddr, erc20tokenAddr}
	callDataOriginals := [][]byte{callDataOriginal, callDataOriginal2}

	// 4 - wrap it into "execute" call
	callData, err := GetCallDataForBatchExecute(targets, callDataOriginals)
	if err != nil {
		log.Error("failed to get call data", zap.Error(err))
		return err
	}
	log.Info("prepared call data", zap.String("callData", hex.EncodeToString(callData)))

	id := aa.getNextAlchemyRequestID()

	rgapd, err := aa.alchemy.CreateRequestGasAndPaymasterData(callData, adminScw, uint64(nonce.Int64()), policyID, entryPointAddress, id)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return err
	}

	jsonDATAPre, err := json.Marshal(rgapd)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return err
	}

	log.Info("jsonDataPre is ready", zap.String("jsonDataPre", string(jsonDATAPre)))

	// 5 - send it
	response, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	// parse response
	responseStruct := alchemysdk.JSONRPCResponseGasAndPaymaster{}
	err = json.Unmarshal(response, &responseStruct)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return err
	}
	// TODO: handle "Error code": -32500  "AA25 invalid account nonce" error
	if responseStruct.Error.Code != 0 {
		log.Error("GasAndPaymaster call failed",
			zap.Int("Error code", responseStruct.Error.Code),
			zap.String("Error message", responseStruct.Error.Message),
		)
		return errors.New(responseStruct.Error.Message)
	}

	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 6 - now create new transaction
	appendEntryPoint := true
	jsonDATA, err := aa.alchemy.CreateRequestAndSign(callData, responseStruct, chainID, entryPointAddress, adminScw, uint64(nonce.Int64()), id+1, adminPK, appendEntryPoint)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return err
	}

	log.Info("created eth_sendUserOperation request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	log.Info("eth_sendUserOperation got response", zap.Any("response", response))

	// 7 - get op hash
	// will handle error in response
	opHash, err := aa.alchemy.DecodeSendUserOperationResponse(response)
	if err != nil {
		log.Error("failed to decode response", zap.Error(err))
		return err
	}
	log.Info("decoded response", zap.String("opHash", opHash))

	// TODO: loop
	// 8 - wait for receipt
	req, err := aa.alchemy.CreateRequestGetUserOperation(opHash, id+2)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return err
	}

	jsonDATA, err = json.Marshal(req)
	if err != nil {
		log.Error("can not marshal JSON", zap.Error(err))
		return err
	}

	log.Info("created eth_getUserOperationReceipt request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = aa.alchemy.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	log.Info("eth_getUserOperationReceipt got response", zap.Any("r", response))
	return nil
}

func (aa *anynsAA) GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (dataOut []byte, contextData []byte, err error) {
	// settings from config:
	entryPointAddress := common.HexToAddress(aa.aaConfig.EntryPoint)
	alchemyApiKey := aa.aaConfig.AlchemyApiKey
	policyID := aa.aaConfig.GasPolicyId

	var chainID int64 = int64(aa.aaConfig.ChainID)
	var id int = aa.getNextAlchemyRequestID()

	// 0 - check params
	err = anynsrpc.СheckRegisterParams(in)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, nil, err
	}

	// 1 - determine users's SCW
	scw, err := aa.GetSmartWalletAddress(ctx, common.HexToAddress(in.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return nil, nil, err
	}

	// 2 - get nonce
	nonce, err := aa.getNonceForSmartWalletAddress(ctx, scw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return nil, nil, err
	}
	log.Info("got nonce", zap.String("scw", scw.String()), zap.Int64("nonce", nonce.Int64()))

	// 3 - create user operation
	callData, err := aa.getCallDataForNameRegister(in.FullName, in.OwnerAnyAddress, in.OwnerEthAddress, in.SpaceId)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, nil, err
	}

	rgapd, err := aa.alchemy.CreateRequestGasAndPaymasterData(callData, scw, uint64(nonce.Int64()), policyID, entryPointAddress, id)
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

	// 5 - send it
	response, err := aa.alchemy.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return nil, nil, err
	}
	// parse response
	responseStruct := alchemysdk.JSONRPCResponseGasAndPaymaster{}
	err = json.Unmarshal(response, &responseStruct)
	if err != nil {
		log.Error("failed to unmarshal response", zap.Error(err))
		return nil, nil, err
	}
	// TODO: handle "Error code": -32500  "AA25 invalid account nonce" error
	if responseStruct.Error.Code != 0 {
		log.Error("GasAndPaymaster call failed",
			zap.Int("Error code", responseStruct.Error.Code),
			zap.String("Error message", responseStruct.Error.Message),
		)
		return nil, nil, errors.New(responseStruct.Error.Message)
	}

	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 6 - get data to sign
	jsonData, uo, err := aa.alchemy.CreateRequestStep1(callData, responseStruct, chainID, entryPointAddress, scw, uint64(nonce.Int64()))
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

func (aa *anynsAA) getCallDataForNameRegister(fullName string, ownerAnyAddress string, ownerEthAddress string, spaceID string) ([]byte, error) {
	registrarController := common.HexToAddress(aa.contractsConfig.AddrRegistrar)
	resolverAddress := common.HexToAddress(aa.contractsConfig.AddrResolver)
	registrantAccount := common.HexToAddress(ownerEthAddress)

	var nameFirstPart string = contracts.RemoveTLD(fullName)
	var REGISTRATION_TIME big.Int = *big.NewInt(365 * 24 * 60 * 60)
	var isReverseRecord bool = true
	var ownerControlledFuses uint16 = 0

	// 1 - get new random secret
	secret32, err := contracts.GenerateRandomSecret()
	if err != nil {
		log.Error("can not generate random secret", zap.Error(err))
		return nil, err
	}

	// 2 - create a commitment
	conn, err := aa.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	controller, err := aa.contracts.ConnectToController(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	commitment, err := aa.contracts.MakeCommitment(
		nameFirstPart,
		registrantAccount,
		secret32,
		controller,
		fullName,
		ownerAnyAddress,
		ownerEthAddress)

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
	callDataOriginal1, err := GetCallDataForCommit(commitment)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, err
	}

	callDataOriginal2, err := GetCallDataForRegister(
		nameFirstPart,
		registrantAccount,
		REGISTRATION_TIME,
		secret32,
		resolverAddress,
		callData,
		isReverseRecord,
		ownerControlledFuses)

	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, err
	}

	// create array of call data
	targets := []common.Address{registrarController, registrarController}
	callDataOriginals := [][]byte{callDataOriginal1, callDataOriginal2}

	// 4 - wrap it into "execute" call
	executeCallDataOut, err := GetCallDataForBatchExecute(targets, callDataOriginals)
	if err != nil {
		log.Error("failed to get call data", zap.Error(err))
		return nil, err
	}

	return executeCallDataOut, nil
}
