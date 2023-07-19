package anynsrpc

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	ac "github.com/anyproto/anyns-node/anytype_crypto"
	"github.com/anyproto/anyns-node/config"
)

func CreateEthConnection(config *config.Contracts) (*ethclient.Client, error) {
	connStr := config.GethUrl

	conn, err := ethclient.Dial(connStr)
	return conn, err
}

func ConnectToRegistryContract(conn *ethclient.Client, config *config.Contracts) (*ac.ENSRegistry, error) {
	// 1 - create new contract instance
	contractRegAddr := config.AddrRegistry

	reg, err := ac.NewENSRegistry(common.HexToAddress(contractRegAddr), conn)
	if err != nil || reg == nil {
		log.Fatal("failed to instantiate ENSRegistry contract", zap.Error(err))
		return nil, err
	}

	return reg, err
}

func ConnectToNamewrapperContract(conn *ethclient.Client, config *config.Contracts) (*ac.AnytypeNameWrapper, error) {
	// 1 - create new contract instance
	contractAddr := config.AddrNameWrapper

	nw, err := ac.NewAnytypeNameWrapper(common.HexToAddress(contractAddr), conn)
	if err != nil || nw == nil {
		log.Fatal("failed to instantiate AnytypeNameWrapper contract", zap.Error(err))
		return nil, err
	}

	return nw, err
}

func ConnectToResolver(conn *ethclient.Client, config *config.Contracts) (*ac.AnytypeResolver, error) {
	// 1 - create new contract instance
	contractAddr := config.AddrResolver

	ar, err := ac.NewAnytypeResolver(common.HexToAddress(contractAddr), conn)
	if err != nil || ar == nil {
		log.Fatal("failed to instantiate AnytypeResolver contract", zap.Error(err))
		return nil, err
	}

	return ar, err
}

func ConnectToController(conn *ethclient.Client, config *config.Contracts) (*ac.AnytypeRegistrarControllerPrivate, error) {
	// 1 - create new contract instance
	contractAddr := config.AddrPrivateController

	ac, err := ac.NewAnytypeRegistrarControllerPrivate(common.HexToAddress(contractAddr), conn)
	if err != nil || ac == nil {
		log.Fatal("failed to instantiate AnytypeRegistrarControllerPrivate contract", zap.Error(err))
		return nil, err
	}

	return ac, err
}

func GenerateAuthOptsForAdmin(conn *ethclient.Client, config *config.Contracts) (*bind.TransactOpts, error) {
	// 1 - load private key
	// TODO: move PK to secure place
	privateKey, err := crypto.HexToECDSA(config.AdminPk)

	if err != nil {
		log.Fatal("can not get admin PK", zap.Error(err))
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 2 - get gas costs, etc
	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("can not get nonce", zap.Error(err))
		return nil, err
	}

	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("can not get gas price", zap.Error(err))
		return nil, err
	}

	auth := bind.NewKeyedTransactor(privateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(500000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func checkTransactionReceipt(conn *ethclient.Client, txHash common.Hash) bool {
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
