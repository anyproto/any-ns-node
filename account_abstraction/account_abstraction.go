package accountabstraction

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"strings"

	"github.com/anyproto/any-ns-node/alchemyaa"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
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
}

type AccountAbstractionService interface {
	// each EOA has an associated smart wallet address
	// even if it is not deployed yet - we can determine it
	GetSmartWalletAddress(eoa common.Address) (address common.Address, err error)
	GetNonceForSmartWalletAddress(ctx context.Context, scw common.Address) (*big.Int, error)

	VerifyAdminIdentity(payload []byte, signature []byte) (err error)

	// will mint + approve tokens to the specified smart wallet
	MintAccessTokens(scw common.Address, amount *big.Int) (err error)

	GetNamesCountLeft(scw common.Address) (count uint64, err error)
	GetOperationsCountLeft(scw common.Address) (count uint64, err error)

	app.Component
}

func (arpc *anynsAA) Init(a *app.App) (err error) {
	arpc.accountConfig = a.MustComponent(config.CName).(*config.Config).GetAccount()
	arpc.aaConfig = a.MustComponent(config.CName).(*config.Config).GetAA()
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)

	return nil
}

func (arpc *anynsAA) Name() (name string) {
	return CName
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
		log.Error("failed to get balance of", zap.Error(err))
		return 0, err
	}

	// TODO: remove hardcode
	// $20 USD per name (current testnet settings)
	// $1 USD = 10^18 wei
	oneNamePriceWei := big.NewInt(20).Exp(big.NewInt(10), big.NewInt(18), nil)

	count = balance.Div(balance, oneNamePriceWei).Uint64()

	log.Info("names count left is", zap.String("scw", scw.String()), zap.Uint64("count", count))
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

func (arpc *anynsAA) MintAccessTokens(scw common.Address, namesCount *big.Int) (err error) {
	// settings from config:
	entryPointAddress := common.HexToAddress(arpc.aaConfig.EntryPoint)
	erc20tokenAddr := common.HexToAddress(arpc.contractsConfig.AddrToken)
	registrarController := common.HexToAddress(arpc.contractsConfig.AddrRegistrar)

	alchemyApiKey := arpc.aaConfig.AlchemyApiKey
	adminPK := arpc.contractsConfig.AdminPk
	policyID := arpc.aaConfig.GasPolicyId

	var chainID int64 = int64(arpc.aaConfig.ChainID)

	// 0 - check params
	if namesCount.Cmp(big.NewInt(0)) == 0 {
		return errors.New("names count is 0")
	}

	// 1 - get nonce
	nonce, err := arpc.GetNonceForSmartWalletAddress(context.Background(), scw)
	if err != nil {
		log.Error("failed to get nonce", zap.Error(err))
		return err
	}

	// 2 - create user operation
	// TODO: change amount
	callDataOriginal, err := alchemyaa.GetCallDataForMint(scw, 100)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return err
	}
	log.Info("Prepared original call data", zap.String("callDataOriginal", hex.EncodeToString(callDataOriginal)))

	callDataOriginal2, err := alchemyaa.GetCallDataForAprove(scw, registrarController, 100)
	if err != nil {
		log.Error("failed to get original call data", zap.Error(err))
		return err
	}
	log.Info("Prepared original call data 2", zap.String("callDataOriginal2", hex.EncodeToString(callDataOriginal2)))

	// create array of call data
	targets := []common.Address{erc20tokenAddr, erc20tokenAddr}
	callDataOriginals := [][]byte{callDataOriginal, callDataOriginal2}

	// 2 - wrap it into "execute" call
	callData, err := alchemyaa.GetCallDataForBatchExecute(targets, callDataOriginals)
	if err != nil {
		log.Error("failed to get call data", zap.Error(err))
		return err
	}
	log.Info("prepared call data", zap.String("callData", hex.EncodeToString(callData)))

	// TODO: set proper ID here
	id := 1

	rgapd, err := alchemyaa.CreateRequestGasAndPaymasterData(callData, scw, uint64(nonce.Int64()), policyID, entryPointAddress, id)
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

	// 3 - send it
	response, err := alchemyaa.SendRequest(alchemyApiKey, jsonDATAPre)
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
	log.Info("alchemy_requestGasAndPaymasterAndData got response", zap.Any("responseStruct", responseStruct))

	// 4 - now create new transaction
	appendEntryPoint := true
	jsonDATA, err := alchemyaa.CreateRequest(callData, responseStruct, chainID, entryPointAddress, scw, uint64(nonce.Int64()), id+1, adminPK, appendEntryPoint)
	if err != nil {
		log.Error("failed to create request", zap.Error(err))
		return err
	}

	log.Info("created eth_sendUserOperation request", zap.String("jsonDATA", string(jsonDATA)))

	// send it
	response, err = alchemyaa.SendRequest(alchemyApiKey, jsonDATA)
	if err != nil {
		log.Error("failed to send request", zap.Error(err))
		return err
	}

	log.Info("eth_sendUserOperation got response", zap.Any("response", response))

	// TODO: wait for operation to be mined
	return nil
}
