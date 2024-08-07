package contracts

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	ac "github.com/anyproto/any-ns-node/anytype_crypto"
	"github.com/anyproto/any-ns-node/config"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.contracts"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsContracts{}
}

type MakeCommitmentParams struct {
	NameFirstPart         string
	RegisterPeriodMonths  uint32
	IsReverseRecordUpdate bool
	SpaceId               string
	OwnerAnyAddr          string
	FullName              string
	Controller            *ac.AnytypeRegistrarControllerPrivate
	Secret                [32]byte
	RegistrantAccount     common.Address
}

type CommitParams struct {
	Opts       *bind.TransactOpts
	Commitment [32]byte
	Controller *ac.AnytypeRegistrarControllerPrivate
}

type RegisterParams struct {
	AuthOpts             *bind.TransactOpts
	NameFirstPart        string
	RegistrantAccount    common.Address
	Secret               [32]byte
	Controller           *ac.AnytypeRegistrarControllerPrivate
	FullName             string
	OwnerAnyAddr         string
	SpaceId              string
	IsReverseRecord      bool
	RegisterPeriodMonths uint32
}

type RenewParams struct {
	TxOpts      *bind.TransactOpts
	FullName    string
	DurationSec uint64
	Controller  *ac.AnytypeRegistrarControllerPrivate
}

// TODO: refactor, split into several interfaces
// Low-level calls to contracts
type ContractsService interface {
	CreateEthConnection() (*ethclient.Client, error)

	// generic method to call any contract
	CallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error)

	// AA methods:
	IsContractDeployed(ctx context.Context, address common.Address) (bool, error)
	// will return .owner of the contract
	GetScwOwner(ctx context.Context, address common.Address) (common.Address, error)

	// ENS methods
	GetOwnerForNamehash(ctx context.Context, namehash [32]byte) (common.Address, error)
	GetAdditionalNameInfo(ctx context.Context, currentOwner common.Address, fullName string) (ownerEthAddress string, ownerAnyAddress string, spaceId string, expiration *big.Int, err error)

	Commit(ctx context.Context, params *CommitParams) (*types.Transaction, error)
	Register(ctx context.Context, params *RegisterParams) (*types.Transaction, error)
	Renew(ctx context.Context, params *RenewParams) (*types.Transaction, error)

	// Aux methods
	MakeCommitment(params *MakeCommitmentParams) ([32]byte, error)
	GetNameByAddress(address common.Address) (string, error)
	GetBalanceOf(ctx context.Context, tokenAddress common.Address, address common.Address) (*big.Int, error)

	ConnectToRegistryContract() (*ac.ENSRegistry, error)
	ConnectToNamewrapperContract() (*ac.AnytypeNameWrapper, error)
	ConnectToResolver() (*ac.AnytypeResolver, error)
	ConnectToRegistrar() (*ac.AnytypeRegistrarImplementation, error)
	ConnectToPrivateController() (*ac.AnytypeRegistrarControllerPrivate, error)

	GenerateAuthOptsForAdmin() (*bind.TransactOpts, error)
	CalculateTxParams(conn *ethclient.Client, address common.Address) (*big.Int, uint64, error)

	// Check if tx is even started to mine
	WaitForTxToStartMining(ctx context.Context, txHash common.Hash) error
	WaitMined(ctx context.Context, tx *types.Transaction) (wasMined bool, err error)
	TxByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, error)

	app.Component
}

type anynsContracts struct {
	config config.Contracts
}

func (acontracts *anynsContracts) Name() (name string) {
	return CName
}

func (acontracts *anynsContracts) Init(a *app.App) (err error) {
	acontracts.config = a.MustComponent(config.CName).(*config.Config).GetContracts()
	return nil
}

func (acontracts *anynsContracts) CallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	client, err := ethclient.Dial(acontracts.config.GethUrl)
	if err != nil {
		log.Error("failed to dial geth", zap.Error(err))
		return nil, err
	}

	res, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Error("failed to CallContract", zap.Error(err))
		return nil, err
	}

	return res, err
}

func (acontracts *anynsContracts) GetBalanceOf(ctx context.Context, tokenAddress common.Address, address common.Address) (*big.Int, error) {
	client, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return big.NewInt(0), err
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return big.NewInt(0), err
	}

	input, err := parsedABI.Pack("balanceOf", address)
	if err != nil {
		return big.NewInt(0), err
	}

	callMsg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: input,
	}

	res, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		log.Error("failed to call balanceOf", zap.Error(err))
		return big.NewInt(0), err
	}

	balance := big.NewInt(0)
	balance.SetBytes(res)
	return balance, nil
}

func (acontracts *anynsContracts) IsContractDeployed(ctx context.Context, address common.Address) (bool, error) {
	client, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return false, err
	}

	bs, err := client.CodeAt(ctx, address, nil)
	if err != nil {
		log.Error("failed to get code", zap.Error(err))
		return false, err
	}

	// check if bs is not empty
	if len(bs) == 0 {
		return false, nil
	}

	return true, nil
}

func (acontracts *anynsContracts) GetOwnerForNamehash(ctx context.Context, nh [32]byte) (common.Address, error) {
	reg, err := acontracts.ConnectToRegistryContract()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return common.Address{}, err
	}

	callOpts := bind.CallOpts{}
	own, err := reg.Owner(&callOpts, nh)

	return own, err
}

func (acontracts *anynsContracts) GetScwOwner(ctx context.Context, scwAddress common.Address) (common.Address, error) {
	client, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return common.Address{}, err
	}

	// 1 - check if address is a smart contract
	isDeployed, err := acontracts.IsContractDeployed(ctx, scwAddress)
	if err != nil {
		log.Error("failed to check if contract is deployed", zap.Error(err))
		return common.Address{}, err
	}

	if !isDeployed {
		log.Info("address is not a smart contract")
		return common.Address{}, errors.New("address is not a smart contract")
	}

	scw, err := acontracts.ConnectToSCW(client, scwAddress)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return common.Address{}, err
	}

	// 2.2 - call contract's method
	callOpts := bind.CallOpts{}
	owner, err := scw.Owner(&callOpts)
	if err != nil {
		log.Error("failed to get Owner", zap.Error(err))
		return common.Address{}, err
	}

	return owner, nil
}

func (acontracts *anynsContracts) GetAdditionalNameInfo(ctx context.Context, currentOwner common.Address, fullName string) (ownerEthAddress string, ownerAnyAddress string, spaceId string, expiration *big.Int, err error) {
	var res nsp.NameAvailableResponse
	res.Available = false

	// 1 - if current owner is the NW contract - then ask it again about the "real owner"
	nwAddress := acontracts.config.AddrNameWrapper
	nwAddressBytes := common.HexToAddress(nwAddress)

	if currentOwner == nwAddressBytes {
		log.Info("address is owned by NameWrapper contract, ask it to retrieve real owner")

		realOwner, err := acontracts.getRealOwner(fullName)
		if err != nil {
			log.Warn("failed to get real owner of the name", zap.Error(err))
			// do not panic, try to continue
		}

		if realOwner != nil {
			ownerEthAddress = *realOwner
		}
	} else {
		// if NW is not the "owner" of the contract -> then it is the real owner
		ownerEthAddress = currentOwner.Hex()
	}

	// 2 - get content hash and spaceID
	owner, spaceID, err := acontracts.getAdditionalData(fullName)
	if err != nil {
		log.Error("failed to get real additional data of the name", zap.Error(err))
		return "", "", "", nil, err
	}
	if owner != nil {
		ownerAnyAddress = *owner
	}
	if spaceID != nil {
		spaceId = *spaceID
	}

	// 3 - get expiration date
	expiration, err = acontracts.getExpirationDate(fullName)
	if err != nil {
		log.Error("failed to get expiration of the name", zap.Error(err))
		return "", "", "", nil, err
	}

	return ownerEthAddress, ownerAnyAddress, spaceId, expiration, nil
}

func (acontracts *anynsContracts) getRealOwner(fullName string) (*string, error) {
	// 1 - connect to contract
	nw, err := acontracts.ConnectToNamewrapperContract()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	// 2 - convert to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 3 - call contract's method
	log.Info("getting real owner for name", zap.String("Full name", fullName))

	callOpts := bind.CallOpts{}

	// convert bytes32 -> uin256 (also 32 bytes long)
	id := new(big.Int).SetBytes(nh[:])
	addr, err := nw.OwnerOf(&callOpts, id)
	if err != nil {
		log.Error("failed to convert Owner", zap.Error(err))
		return nil, err
	}

	// 4 - covert to result
	// the owner can be NameWrapper
	var out string = addr.Hex()

	log.Info("received real owner address", zap.String("Owner addr", out))
	return &out, nil
}

func (acontracts *anynsContracts) getAdditionalData(fullName string) (*string, *string, error) {
	// 1 - connect to contract
	ar, err := acontracts.ConnectToResolver()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return nil, nil, err
	}

	// 2 - convert to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, nil, err
	}

	// 3 - get content hash and space ID
	callOpts := bind.CallOpts{}
	hash, err := ar.Contenthash(&callOpts, nh)
	if err != nil {
		log.Error("can not get contenthash", zap.Error(err))
		return nil, nil, err
	}

	space, err := ar.SpaceId(&callOpts, nh)
	if err != nil {
		log.Error("can not get SpaceID", zap.Error(err))
		return nil, nil, err
	}

	// convert hex values to string
	hexString := hex.EncodeToString(hash)
	contentHashDecoded, _ := hex.DecodeString(hexString)
	ownerAnyAddressOut := string(contentHashDecoded)

	hexString = hex.EncodeToString(space)
	spaceIDDecoded, _ := hex.DecodeString(hexString)
	spaceIDOut := string(spaceIDDecoded)

	return &ownerAnyAddressOut, &spaceIDOut, nil
}

func (acontracts *anynsContracts) getExpirationDate(fullName string) (*big.Int, error) {
	// 1 - connect to contract
	ar, err := acontracts.ConnectToRegistrar()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	// 2 - convert to name hash
	parts := strings.Split(fullName, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid full name")
	}

	label := parts[0]
	labelHash := crypto.Keccak256([]byte(label))
	// convert nh (32 bytes array) to big.Int
	nhAsTokenID := new(big.Int).SetBytes(labelHash[:])

	// 3 - get content hash and space ID
	callOpts := bind.CallOpts{}
	out, err := ar.NameExpires(&callOpts, nhAsTokenID)
	if err != nil {
		log.Error("can not get nameexpires", zap.Error(err))
		return nil, err
	}
	return out, nil
}

func (acontracts *anynsContracts) CreateEthConnection() (*ethclient.Client, error) {
	connStr := acontracts.config.GethUrl
	conn, err := ethclient.Dial(connStr)
	return conn, err
}

func (acontracts *anynsContracts) ConnectToRegistryContract() (*ac.ENSRegistry, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - create new contract instance
	contractRegAddr := acontracts.config.AddrRegistry

	reg, err := ac.NewENSRegistry(common.HexToAddress(contractRegAddr), conn)
	if err != nil || reg == nil {
		log.Error("failed to instantiate ENSRegistry contract", zap.Error(err))
		return nil, err
	}

	return reg, err
}

func (acontracts *anynsContracts) ConnectToNamewrapperContract() (*ac.AnytypeNameWrapper, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrNameWrapper

	nw, err := ac.NewAnytypeNameWrapper(common.HexToAddress(contractAddr), conn)
	if err != nil || nw == nil {
		log.Error("failed to instantiate AnytypeNameWrapper contract", zap.Error(err))
		return nil, err
	}

	return nw, err
}

func (acontracts *anynsContracts) ConnectToResolver() (*ac.AnytypeResolver, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrResolver

	ar, err := ac.NewAnytypeResolver(common.HexToAddress(contractAddr), conn)
	if err != nil || ar == nil {
		log.Error("failed to instantiate AnytypeResolver contract", zap.Error(err))
		return nil, err
	}

	return ar, err
}

func (acontracts *anynsContracts) ConnectToRegistrar() (*ac.AnytypeRegistrarImplementation, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrRegistrarImplementation

	ar, err := ac.NewAnytypeRegistrarImplementation(common.HexToAddress(contractAddr), conn)
	if err != nil || ar == nil {
		log.Error("failed to instantiate AnytypeRegistrar contract", zap.Error(err))
		return nil, err
	}

	return ar, err
}

func (acontracts *anynsContracts) ConnectToPrivateController() (*ac.AnytypeRegistrarControllerPrivate, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrRegistrarPrivateController

	ac, err := ac.NewAnytypeRegistrarControllerPrivate(common.HexToAddress(contractAddr), conn)
	if err != nil || ac == nil {
		log.Error("failed to instantiate AnytypeRegistrarControllerPrivate contract", zap.Error(err))
		return nil, err
	}

	return ac, err
}

func (acontracts *anynsContracts) ConnectToSCW(conn *ethclient.Client, address common.Address) (*ac.SCW, error) {
	// 1 - create new contract instance
	scw, err := ac.NewSCW(address, conn)

	if err != nil || scw == nil {
		log.Error("failed to instantiate SCW contract", zap.Error(err))
		return nil, err
	}

	return scw, err
}

func (acontracts *anynsContracts) GenerateAuthOptsForAdmin() (*bind.TransactOpts, error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	// 1 - load private key
	// TODO: move PK to secure place
	privateKey, err := crypto.HexToECDSA(acontracts.config.AdminPk)

	if err != nil {
		log.Error("can not get admin PK", zap.Error(err))
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 2 - get gas costs, etc
	gasPrice, nonce, err := acontracts.CalculateTxParams(conn, fromAddress)
	if err != nil {
		log.Error("can not calculate tx params", zap.Error(err))
		return nil, err
	}

	// increase gas price - multiply gasPrice BigInt twice
	gasPrice.Mul(gasPrice, big.NewInt(2))

	// TODO: change to
	//bind.NewKeyedTransactorWithChainID()
	auth := bind.NewKeyedTransactor(privateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(500000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (acontracts *anynsContracts) CalculateTxParams(conn *ethclient.Client, address common.Address) (*big.Int, uint64, error) {
	nonce, err := conn.PendingNonceAt(context.Background(), address)
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return nil, 0, err
	}

	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("can not get gas price", zap.Error(err))
		return nil, 0, err
	}

	return gasPrice, nonce, nil
}

func (acontracts *anynsContracts) checkTransactionReceipt(conn *ethclient.Client, txHash common.Hash) bool {
	tx, err := conn.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return false
	}

	// success
	if tx.Status == 1 {
		return true
	}

	return false
}

func (acontracts *anynsContracts) WaitForTxToStartMining(ctx context.Context, txHash common.Hash) (err error) {
	// if transaction is not returned by node immediately after it is sent... it is either:
	// 1. "nonce is too high" error
	// 2. normal behaviour
	//
	// so we will wait N times each X seconds long
	// if tx is not mined after N*X seconds - we will assume that it is "nonce is too high" error
	var i uint = 0
	for ; i < acontracts.config.WaitMiningRetryCount; i++ {
		tx, err := acontracts.TxByHash(ctx, txHash)
		if (err == nil) && (tx != nil) {
			// tx mined!
			// TODO: sometimes it gives us false positives here :-(((
			log.Debug("NOT a HIGH NONCE!!!", zap.Any("tx", tx))
			return nil
		}

		if err.Error() == "not found" {
			// wait and try again
			log.Warn("tx is still not found. waiting...", zap.Any("tx hash", txHash), zap.Any("try", i))

			time.Sleep(5 * time.Second)
			continue
		}
		// for any other error -> return it
		return err
	}

	log.Warn("Probably we have HIGH NONCE...")
	return ErrNonceTooHigh
}

func (acontracts *anynsContracts) WaitMined(ctx context.Context, tx *types.Transaction) (wasMined bool, err error) {
	conn, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return false, err
	}

	// receipt is not used
	_, err = bind.WaitMined(ctx, conn, tx)
	if err != nil {
		log.Error("failed to wait for tx", zap.Error(err))
		return false, err
	}

	// please note that transaction receipts are not available for pending transactions
	wasMined = acontracts.checkTransactionReceipt(conn, tx.Hash())
	return wasMined, nil
}

func (acontracts *anynsContracts) TxByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	client, err := acontracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create connection", zap.Error(err))
		return nil, err
	}

	tx, _, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		// this can happen!
		log.Warn("failed to get tx", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (acontracts *anynsContracts) MakeCommitment(params *MakeCommitmentParams) ([32]byte, error) {
	var adminAddr common.Address = common.HexToAddress(acontracts.config.AddrAdmin)
	var resolverAddr common.Address = common.HexToAddress(acontracts.config.AddrResolver)

	var regTime = PeriodMonthsToTimestamp(params.RegisterPeriodMonths)

	callData, err := PrepareCallData_SetContentHashSpaceID(params.FullName, params.OwnerAnyAddr, params.SpaceId)
	if err != nil {
		log.Error("can not prepare call data", zap.Error(err))
		return [32]byte{}, err
	}

	var ownerControlledFuses uint16 = 0
	callOpts := bind.CallOpts{}
	callOpts.From = adminAddr

	return params.Controller.MakeCommitment(
		&callOpts,
		params.NameFirstPart,
		params.RegistrantAccount,
		&regTime,
		params.Secret,
		resolverAddr,
		callData,
		params.IsReverseRecordUpdate,
		ownerControlledFuses)
}

func (acontracts *anynsContracts) Commit(ctx context.Context, params *CommitParams) (*types.Transaction, error) {
	tx, err := params.Controller.Commit(params.Opts, params.Commitment)
	if err != nil {
		// TODO - handle the "replacement transaction underpriced" error
		log.Error("failed to commit", zap.Error(err), zap.Any("tx", tx))

		if err.Error() == "nonce too low" {
			return tx, ErrNonceTooLow
		}

		return tx, err
	}

	// wait until TX is "seen" by the network (N times)
	// can return ErrNonceTooHigh or just error
	err = acontracts.WaitForTxToStartMining(ctx, tx.Hash())
	if err != nil {
		log.Error("can not Commit tx, can not start", zap.Error(err), zap.Any("tx", tx))
		return tx, err
	}

	log.Info("commit tx sent", zap.String("TX hash", tx.Hash().Hex()))
	return tx, nil
}

func (acontracts *anynsContracts) Register(ctx context.Context, params *RegisterParams) (*types.Transaction, error) {
	var resolverAddr common.Address = common.HexToAddress(acontracts.config.AddrResolver)
	var regTime = PeriodMonthsToTimestamp(params.RegisterPeriodMonths)

	callData, err := PrepareCallData_SetContentHashSpaceID(params.FullName, params.OwnerAnyAddr, params.SpaceId)
	if err != nil {
		log.Error("can not prepare call data", zap.Error(err))
		return nil, err
	}

	var ownerControlledFuses uint16 = 0

	tx, err := params.Controller.Register(
		params.AuthOpts,
		params.NameFirstPart,
		params.RegistrantAccount,
		&regTime,
		params.Secret,
		resolverAddr,
		callData,
		params.IsReverseRecord,
		ownerControlledFuses)

	if err != nil {
		log.Error("failed to register", zap.Error(err), zap.Any("tx", tx))

		if err.Error() == "nonce too low" {
			return tx, ErrNonceTooLow
		}

		return tx, err
	}

	// wait until TX is "seen" by the network (N times)
	// can return ErrNonceTooHigh or just error
	err = acontracts.WaitForTxToStartMining(ctx, tx.Hash())
	if err != nil {
		log.Error("can not Register tx, can not start", zap.Error(err), zap.Any("tx", tx))
		return tx, err
	}

	log.Info("register tx sent", zap.String("TX hash", tx.Hash().Hex()))
	return tx, nil
}

func (acontracts *anynsContracts) Renew(ctx context.Context, params *RenewParams) (*types.Transaction, error) {
	tx, err := params.Controller.Renew(
		params.TxOpts,
		params.FullName,
		big.NewInt(int64(params.DurationSec)),
	)

	if err != nil {
		log.Error("failed to renew", zap.Error(err), zap.Any("tx", tx))

		if err.Error() == "nonce too low" {
			return tx, ErrNonceTooLow
		}
		return tx, err
	}

	// wait until TX is "seen" by the network (N times)
	// can return ErrNonceTooHigh or just error
	err = acontracts.WaitForTxToStartMining(ctx, tx.Hash())
	if err != nil {
		log.Error("can not Register tx, can not start", zap.Error(err), zap.Any("tx", tx))
		return tx, err
	}

	log.Info("renew tx sent", zap.String("TX hash", tx.Hash().Hex()))
	return tx, nil
}

func (acontracts *anynsContracts) GetNameByAddress(address common.Address) (string, error) {
	// 1 - connect to contract
	ar, err := acontracts.ConnectToResolver()
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return "", err
	}

	// 2 - convert address to .addr.reverse
	// remove 0x
	fullName := strings.ToLower(address.Hex()[2:] + ".addr.reverse")

	// convert to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return "", err
	}

	// convert namehash from bytes32 to string
	nhStr := hex.EncodeToString(nh[:])

	log.Info("getting name for address",
		zap.String("Address", address.Hex()),
		zap.String("FullName", fullName),
		zap.String("NameHash", nhStr))

	// 3 - call contract's method
	callOpts := bind.CallOpts{}
	name, err := ar.Name(&callOpts, nh)
	if err != nil {
		log.Error("can not get SpaceID", zap.Error(err))
		return "", err
	}
	log.Info("got name", zap.String("Name", name))

	return name, nil
}
