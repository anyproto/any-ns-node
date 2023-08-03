package contracts

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	"go.uber.org/zap"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	ac "github.com/anyproto/any-ns-node/anytype_crypto"
	"github.com/anyproto/any-ns-node/config"
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
)

const CName = "any-ns.contracts"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsContracts{}
}

// Low-level calls to contracts
type ContractsService interface {
	CreateEthConnection() (*ethclient.Client, error)

	GetOwnerForNamehash(client *ethclient.Client, namehash [32]byte) (common.Address, error)
	GetAdditionalNameInfo(conn *ethclient.Client, currentOwner common.Address, fullName string) (ownerEthAddress string, ownerAnyAddress string, spaceId string, err error)

	ConnectToRegistryContract(conn *ethclient.Client) (*ac.ENSRegistry, error)
	ConnectToNamewrapperContract(conn *ethclient.Client) (*ac.AnytypeNameWrapper, error)
	ConnectToResolver(conn *ethclient.Client) (*ac.AnytypeResolver, error)
	ConnectToController(conn *ethclient.Client) (*ac.AnytypeRegistrarControllerPrivate, error)

	MakeCommitment(nameFirstPart string, registrantAccount common.Address, secret [32]byte, controller *ac.AnytypeRegistrarControllerPrivate, fullName string, ownerAnyAddr string, spaceId string) ([32]byte, error)
	Commit(opts *bind.TransactOpts, commitment [32]byte, controller *ac.AnytypeRegistrarControllerPrivate) (*types.Transaction, error)
	Register(authOpts *bind.TransactOpts, nameFirstPart string, registrantAccount common.Address, secret [32]byte, controller *ac.AnytypeRegistrarControllerPrivate, fullName string, ownerAnyAddr string, spaceId string) (*types.Transaction, error)

	// aux
	GenerateAuthOptsForAdmin(conn *ethclient.Client) (*bind.TransactOpts, error)
	// Wait for tx and get result
	WaitMined(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (wasMined bool, err error)
	TxByHash(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*types.Transaction, error)

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

func (acontracts *anynsContracts) GetOwnerForNamehash(conn *ethclient.Client, nh [32]byte) (common.Address, error) {
	reg, err := acontracts.ConnectToRegistryContract(conn)
	if err != nil {
		log.Error("failed to connect to contract", zap.Error(err))
		return common.Address{}, err
	}

	callOpts := bind.CallOpts{}
	own, err := reg.Owner(&callOpts, nh)

	return own, err
}

func (acontracts *anynsContracts) GetAdditionalNameInfo(conn *ethclient.Client, currentOwner common.Address, fullName string) (ownerEthAddress string, ownerAnyAddress string, spaceId string, err error) {
	var res as.NameAvailableResponse
	res.Available = false

	// 1 - if current owner is the NW contract - then ask it again about the "real owner"
	nwAddress := acontracts.config.AddrNameWrapper
	nwAddressBytes := common.HexToAddress(nwAddress)

	if currentOwner == nwAddressBytes {
		log.Info("address is owned by NameWrapper contract, ask it to retrieve real owner")

		realOwner, err := acontracts.getRealOwner(conn, fullName)
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
	owner, spaceID, err := acontracts.getAdditionalData(conn, fullName)
	if err != nil {
		log.Error("failed to get real additional data of the name", zap.Error(err))
		return "", "", "", err
	}
	if owner != nil {
		ownerAnyAddress = *owner
	}
	if spaceID != nil {
		spaceId = *spaceID
	}

	return ownerEthAddress, ownerAnyAddress, spaceId, nil
}

func (acontracts *anynsContracts) getRealOwner(conn *ethclient.Client, fullName string) (*string, error) {
	// 1 - connect to contract
	nw, err := acontracts.ConnectToNamewrapperContract(conn)
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

func (acontracts *anynsContracts) getAdditionalData(conn *ethclient.Client, fullName string) (*string, *string, error) {
	// 1 - connect to contract
	ar, err := acontracts.ConnectToResolver(conn)
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

func (acontracts *anynsContracts) CreateEthConnection() (*ethclient.Client, error) {
	connStr := acontracts.config.GethUrl
	conn, err := ethclient.Dial(connStr)
	return conn, err
}

func (acontracts *anynsContracts) ConnectToRegistryContract(conn *ethclient.Client) (*ac.ENSRegistry, error) {
	// 1 - create new contract instance
	contractRegAddr := acontracts.config.AddrRegistry

	reg, err := ac.NewENSRegistry(common.HexToAddress(contractRegAddr), conn)
	if err != nil || reg == nil {
		log.Error("failed to instantiate ENSRegistry contract", zap.Error(err))
		return nil, err
	}

	return reg, err
}

func (acontracts *anynsContracts) ConnectToNamewrapperContract(conn *ethclient.Client) (*ac.AnytypeNameWrapper, error) {
	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrNameWrapper

	nw, err := ac.NewAnytypeNameWrapper(common.HexToAddress(contractAddr), conn)
	if err != nil || nw == nil {
		log.Error("failed to instantiate AnytypeNameWrapper contract", zap.Error(err))
		return nil, err
	}

	return nw, err
}

func (acontracts *anynsContracts) ConnectToResolver(conn *ethclient.Client) (*ac.AnytypeResolver, error) {
	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrResolver

	ar, err := ac.NewAnytypeResolver(common.HexToAddress(contractAddr), conn)
	if err != nil || ar == nil {
		log.Error("failed to instantiate AnytypeResolver contract", zap.Error(err))
		return nil, err
	}

	return ar, err
}

func (acontracts *anynsContracts) ConnectToController(conn *ethclient.Client) (*ac.AnytypeRegistrarControllerPrivate, error) {
	// 1 - create new contract instance
	contractAddr := acontracts.config.AddrPrivateController

	ac, err := ac.NewAnytypeRegistrarControllerPrivate(common.HexToAddress(contractAddr), conn)
	if err != nil || ac == nil {
		log.Error("failed to instantiate AnytypeRegistrarControllerPrivate contract", zap.Error(err))
		return nil, err
	}

	return ac, err
}

func (acontracts *anynsContracts) GenerateAuthOptsForAdmin(conn *ethclient.Client) (*bind.TransactOpts, error) {
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
	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Error("can not get nonce", zap.Error(err))
		return nil, err
	}

	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("can not get gas price", zap.Error(err))
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

func (acontracts *anynsContracts) Commit(opts *bind.TransactOpts, commitment [32]byte, controller *ac.AnytypeRegistrarControllerPrivate) (*types.Transaction, error) {
	tx, err := controller.Commit(opts, commitment)
	if err != nil {
		log.Error("failed to commit", zap.Error(err))
		return nil, err
	}

	log.Info("commit tx sent", zap.String("TX hash", tx.Hash().Hex()))
	return tx, nil
}

func (acontracts *anynsContracts) WaitMined(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (wasMined bool, err error) {
	// receipt is not used
	_, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Error("failed to wait for tx", zap.Error(err))
		return false, err
	}

	wasMined = acontracts.checkTransactionReceipt(client, tx.Hash())
	return wasMined, nil
}

func (acontracts *anynsContracts) TxByHash(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*types.Transaction, error) {
	tx, _, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		log.Error("failed to get tx", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (acontracts *anynsContracts) MakeCommitment(nameFirstPart string, registrantAccount common.Address, secret [32]byte, controller *ac.AnytypeRegistrarControllerPrivate, fullName string, ownerAnyAddr string, spaceId string) ([32]byte, error) {
	var adminAddr common.Address = common.HexToAddress(acontracts.config.AddrAdmin)
	var resolverAddr common.Address = common.HexToAddress(acontracts.config.AddrResolver)
	var REGISTRATION_TIME big.Int = *big.NewInt(365 * 24 * 60 * 60)

	callData, err := PrepareCallData(fullName, ownerAnyAddr, spaceId)
	if err != nil {
		log.Error("can not prepare call data", zap.Error(err))
		return [32]byte{}, err
	}

	var isReverseRecord bool = false
	var ownerControlledFuses uint16 = 0
	callOpts := bind.CallOpts{}
	callOpts.From = adminAddr

	return controller.MakeCommitment(
		&callOpts,
		nameFirstPart,
		registrantAccount,
		&REGISTRATION_TIME,
		secret,
		resolverAddr,
		callData,
		isReverseRecord,
		ownerControlledFuses)
}

func (acontracts *anynsContracts) Register(authOpts *bind.TransactOpts, nameFirstPart string, registrantAccount common.Address, secret [32]byte, controller *ac.AnytypeRegistrarControllerPrivate, fullName string, ownerAnyAddr string, spaceId string) (*types.Transaction, error) {
	var resolverAddr common.Address = common.HexToAddress(acontracts.config.AddrResolver)
	var REGISTRATION_TIME big.Int = *big.NewInt(365 * 24 * 60 * 60)

	callData, err := PrepareCallData(fullName, ownerAnyAddr, spaceId)
	if err != nil {
		log.Error("can not prepare call data", zap.Error(err))
		return nil, err
	}

	var isReverseRecord bool = false
	var ownerControlledFuses uint16 = 0

	tx, err := controller.Register(
		authOpts,
		nameFirstPart,
		registrantAccount,
		&REGISTRATION_TIME,
		secret,
		resolverAddr,
		callData,
		isReverseRecord,
		ownerControlledFuses)

	if err != nil {
		log.Error("failed to register", zap.Error(err))
		return nil, err
	}

	log.Info("register tx sent", zap.String("TX hash", tx.Hash().Hex()))
	return tx, nil
}
