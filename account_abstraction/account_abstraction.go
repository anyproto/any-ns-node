package accountabstraction

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"strings"

	"github.com/anyproto/any-ns-node/alchemyaa"
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
	aa              alchemyaa.AlchemyAAService
}

type AccountAbstractionService interface {
	// each EOA has an associated smart wallet address
	// even if it is not deployed yet - we can determine it
	GetSmartWalletAddress(ctx context.Context, eoa common.Address) (address common.Address, err error)

	GetNonceForSmartWalletAddress(ctx context.Context, scw common.Address) (*big.Int, error)

	VerifyAdminIdentity(payload []byte, signature []byte) (err error)

	// will mint + approve tokens to the specified smart wallet
	AdminMintAccessTokens(scw common.Address, amount *big.Int) (err error)

	GetNamesCountLeft(scw common.Address) (count uint64, err error)
	GetOperationsCountLeft(scw common.Address) (count uint64, err error)

	GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (dataOut []byte, contextData []byte, err error)

	//
	GetNextAlchemyRequestID() int

	app.Component
}

func (arpc *anynsAA) Init(a *app.App) (err error) {
	arpc.accountConfig = a.MustComponent(config.CName).(*config.Config).GetAccount()
	arpc.aaConfig = a.MustComponent(config.CName).(*config.Config).GetAA()
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.aa = a.MustComponent(alchemyaa.CName).(alchemyaa.AlchemyAAService)

	return nil
}

func (arpc *anynsAA) Name() (name string) {
	return CName
}

func (arpc *anynsAA) GetNextAlchemyRequestID() int {
	// return something real
	return 1
}

func (arpc *anynsAA) GetSmartWalletAddress(ctx context.Context, eoa common.Address) (address common.Address, err error) {
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

	addr := common.HexToAddress(arpc.aaConfig.AccountFactory)
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: input,
	}

	res, err := arpc.contracts.CallContract(ctx, callMsg)
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

func (arpc *anynsAA) GetNonceForSmartWalletAddress(ctx context.Context, scw common.Address) (*big.Int, error) {

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

	addr := common.HexToAddress(arpc.aaConfig.EntryPoint)
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: input,
	}

	res, err := arpc.contracts.CallContract(ctx, callMsg)
	if err != nil {
		log.Error("failed to call getNonce", zap.Error(err))
		return nil, err
	}

	out := big.NewInt(0)
	out.SetBytes(res)
	return out, nil
}

func (arpc *anynsAA) GetNamesCountLeft(scw common.Address) (count uint64, err error) {
	tokenAddress := common.HexToAddress(arpc.contractsConfig.AddrToken)

	client, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create eth connection", zap.Error(err))
		return 0, err
	}

	balance, err := arpc.contracts.GetBalanceOf(client, tokenAddress, scw)
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

func (arpc *anynsAA) GetOperationsCountLeft(scw common.Address) (count uint64, err error) {
	return 0, nil
}

func (arpc *anynsAA) VerifyAdminIdentity(payload []byte, signature []byte) (err error) {
	// 1 - load public key of admin
	ownerAnyIdentity, err := crypto.DecodePeerId(arpc.accountConfig.PeerId)

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
func (arpc *anynsAA) AdminMintAccessTokens(userScwAddress common.Address, namesCount *big.Int) (err error) {
	// settings from config:
	entryPointAddress := common.HexToAddress(arpc.aaConfig.EntryPoint)
	erc20tokenAddr := common.HexToAddress(arpc.contractsConfig.AddrToken)
	registrarController := common.HexToAddress(arpc.contractsConfig.AddrRegistrar)
	alchemyApiKey := arpc.aaConfig.AlchemyApiKey
	policyID := arpc.aaConfig.GasPolicyId

	adminAddress := common.HexToAddress(arpc.contractsConfig.AddrAdmin)
	adminPK := arpc.contractsConfig.AdminPk

	var chainID int64 = int64(arpc.aaConfig.ChainID)

	// 0 - check params
	if namesCount.Cmp(big.NewInt(0)) == 0 {
		return errors.New("names count is 0")
	}

	// TODO: optimize, cache it or move to settings
	// 1 - determine admin's SCW
	adminScw, err := arpc.GetSmartWalletAddress(context.Background(), adminAddress)
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return err
	}

	// 2 - get nonce (from admin's SCW)
	nonce, err := arpc.GetNonceForSmartWalletAddress(context.Background(), adminScw)
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

	id := arpc.GetNextAlchemyRequestID()

	rgapd, err := arpc.aa.CreateRequestGasAndPaymasterData(callData, adminScw, uint64(nonce.Int64()), policyID, entryPointAddress, id)
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
	response, err := arpc.aa.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	// parse response
	responseStruct := alchemyaa.JSONRPCResponseGasAndPaymaster{}
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
	jsonDATA, err := arpc.aa.CreateRequestAndSign(callData, responseStruct, chainID, entryPointAddress, adminScw, uint64(nonce.Int64()), id+1, adminPK, appendEntryPoint)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return err
	}

	log.Info("created eth_sendUserOperation request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = arpc.aa.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	log.Info("eth_sendUserOperation got response", zap.Any("response", response))

	// 7 - get op hash
	// will handle error in response
	opHash, err := arpc.aa.DecodeSendUserOperationResponse(response)
	if err != nil {
		log.Error("failed to decode response", zap.Error(err))
		return err
	}
	log.Info("decoded response", zap.String("opHash", opHash))

	// TODO: loop
	// 8 - wait for receipt
	req, err := arpc.aa.CreateRequestGetUserOperation(opHash, id+2)
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
	response, err = arpc.aa.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	log.Info("eth_getUserOperationReceipt got response", zap.Any("r", response))
	return nil
}

func (arpc *anynsAA) GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (dataOut []byte, contextData []byte, err error) {
	// settings from config:
	entryPointAddress := common.HexToAddress(arpc.aaConfig.EntryPoint)
	alchemyApiKey := arpc.aaConfig.AlchemyApiKey
	policyID := arpc.aaConfig.GasPolicyId

	var chainID int64 = int64(arpc.aaConfig.ChainID)
	var id int = arpc.GetNextAlchemyRequestID()

	// 0 - check params
	err = anynsrpc.СheckRegisterParams(in)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, nil, err
	}

	// 1 - determine users's SCW
	scw, err := arpc.GetSmartWalletAddress(context.Background(), common.HexToAddress(in.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address for admin", zap.Error(err))
		return nil, nil, err
	}

	// 2 - get nonce
	nonce, err := arpc.GetNonceForSmartWalletAddress(context.Background(), scw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return nil, nil, err
	}
	log.Info("got nonce", zap.String("scw", scw.String()), zap.Int64("nonce", nonce.Int64()))

	// 3 - create user operation
	callData, err := arpc.aa.GetCallDataForNameRegister(in.FullName, in.OwnerAnyAddress, in.OwnerEthAddress, in.SpaceId)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return nil, nil, err
	}

	rgapd, err := arpc.aa.CreateRequestGasAndPaymasterData(callData, scw, uint64(nonce.Int64()), policyID, entryPointAddress, id)
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
	response, err := arpc.aa.SendRequest(alchemyApiKey, jsonDATAPre)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return nil, nil, err
	}
	// parse response
	responseStruct := alchemyaa.JSONRPCResponseGasAndPaymaster{}
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
	jsonData, uo, err := arpc.aa.CreateRequestStep1(callData, responseStruct, chainID, entryPointAddress, scw, uint64(nonce.Int64()))
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
