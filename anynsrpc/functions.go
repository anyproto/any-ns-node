package anynsrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/anyproto/anyns-node/config"
	as "github.com/anyproto/anyns-node/pb/anyns_api_server"
)

func IsNameAvailable(ctx context.Context, in *as.NameAvailableRequest, config *config.Contracts) (*as.NameAvailableResponse, error) {
	conn, err := CreateEthConnection(config)
	if err != nil {
		log.Fatal("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// 1 - connect to geth
	reg, err := ConnectToRegistryContract(conn, config)
	if err != nil {
		log.Fatal("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	// 2 - convert to name hash
	nh, err := NameHash(in.FullName)
	if err != nil {
		log.Fatal("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 3 - call contract's method
	log.Info("getting owner for name", zap.String("FullName", in.GetFullName()))

	callOpts := bind.CallOpts{}
	addr, err := reg.Owner(&callOpts, nh)
	if err != nil {
		log.Fatal("failed to get owner", zap.Error(err))
		return nil, err
	}

	// 4 - covert to result
	// the owner can be NameWrapper
	log.Info("received owner address", zap.String("Owner addr", addr.Hex()))

	var res as.NameAvailableResponse
	var addrEmpty = common.Address{}

	if addr != addrEmpty {
		log.Info("name is NOT available...Getting additional info")
		// 5 - if name is not available, then get additional info
		return getAdditionalNameInfo(conn, addr, in.GetFullName(), config)
	}

	log.Info("name is not registered yet...")
	res.Available = true
	return &res, nil
}

func getAdditionalNameInfo(conn *ethclient.Client, currentOwner common.Address, fullName string, config *config.Contracts) (*as.NameAvailableResponse, error) {
	var res as.NameAvailableResponse
	res.Available = false

	// 1 - if current owner is the NW contract - then ask it again about the "real owner"
	nwAddress := config.AddrNameWrapper
	nwAddressBytes := common.HexToAddress(nwAddress)

	if currentOwner == nwAddressBytes {
		log.Info("address is owned by NameWrapper contract, ask it to retrieve real owner")

		realOwner, err := getRealOwner(conn, fullName, config)
		if err != nil {
			log.Warn("failed to get real owner of the name", zap.Error(err))
			// do not panic, try to continue
		}

		if realOwner != nil {
			res.OwnerEthAddress = *realOwner
		}
	} else {
		// if NW is not the "owner" of the contract -> then it is the real owner
		res.OwnerEthAddress = currentOwner.Hex()
	}

	// 2 - get content hash and spaceID
	ownerAnyAddress, spaceID, err := getAdditionalData(conn, fullName, config)
	if err != nil {
		log.Fatal("failed to get real additional data of the name", zap.Error(err))
		return nil, err
	}
	if ownerAnyAddress != nil {
		res.OwnerAnyAddress = *ownerAnyAddress
	}
	if spaceID != nil {
		res.SpaceId = *spaceID
	}

	return &res, nil
}

func getRealOwner(conn *ethclient.Client, fullName string, config *config.Contracts) (*string, error) {
	// 1 - connect to contract
	nw, err := ConnectToNamewrapperContract(conn, config)
	if err != nil {
		log.Fatal("failed to connect to contract", zap.Error(err))
		return nil, err
	}

	// 2 - convert to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Fatal("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 3 - call contract's method
	log.Info("getting real owner for name", zap.String("Full name", fullName))

	callOpts := bind.CallOpts{}

	// convert bytes32 -> uin256 (also 32 bytes long)
	id := new(big.Int).SetBytes(nh[:])
	addr, err := nw.OwnerOf(&callOpts, id)
	if err != nil {
		log.Fatal("failed to convert Owner", zap.Error(err))
		return nil, err
	}

	// 4 - covert to result
	// the owner can be NameWrapper
	var out string = addr.Hex()

	log.Info("received real owner address", zap.String("Owner addr", out))
	return &out, nil
}

func getAdditionalData(conn *ethclient.Client, fullName string, config *config.Contracts) (*string, *string, error) {
	// 1 - connect to contract
	ar, err := ConnectToResolver(conn, config)
	if err != nil {
		log.Fatal("failed to connect to contract", zap.Error(err))
		return nil, nil, err
	}

	// 2 - convert to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Fatal("can not convert FullName to namehash", zap.Error(err))
		return nil, nil, err
	}

	// 3 - get content hash and space ID
	callOpts := bind.CallOpts{}
	hash, err := ar.Contenthash(&callOpts, nh)
	if err != nil {
		log.Fatal("can not get contenthash", zap.Error(err))
		return nil, nil, err
	}

	space, err := ar.SpaceId(&callOpts, nh)
	if err != nil {
		log.Fatal("can not get SpaceID", zap.Error(err))
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

func NameRegister(ctx context.Context, in *as.NameRegisterRequest, config *config.Contracts) error {
	var adminAddr common.Address = common.HexToAddress(config.AddrAdmin)
	var resolverAddr common.Address = common.HexToAddress(config.AddrResolver)
	var registrantAccount common.Address = common.HexToAddress(in.OwnerEthAddress)

	conn, err := CreateEthConnection(config)
	if err != nil {
		log.Fatal("failed to connect to geth", zap.Error(err))
		return err
	}

	// 1 - connect to geth
	ac, err := ConnectToController(conn, config)
	if err != nil {
		log.Fatal("failed to connect to contract", zap.Error(err))
		return err
	}

	// 2 - get a name's first part
	// TODO: normalize string
	nameFirstPart := RemoveTLD(in.FullName)

	// 3 - calculate a commitment
	var REGISTRATION_TIME big.Int = *big.NewInt(365 * 24 * 60 * 60)

	secret, err := GenerateRandomSecret()

	if err != nil {
		log.Fatal("can not generate random secret", zap.Error(err))
		return err
	}

	callData, err := prepareCallData(in.GetFullName(), in.GetOwnerAnyAddress(), in.GetSpaceId())

	if err != nil {
		log.Fatal("can not prepare call data", zap.Error(err))
		return err
	}

	var isReverseRecord bool = false
	var ownerControlledFuses uint16 = 0

	callOpts := bind.CallOpts{}
	callOpts.From = adminAddr

	commitment, err := ac.MakeCommitment(
		&callOpts,
		nameFirstPart,
		registrantAccount,
		&REGISTRATION_TIME,
		secret,
		resolverAddr,
		callData,
		isReverseRecord,
		ownerControlledFuses)

	if err != nil {
		log.Fatal("can not calculate a commitment", zap.Error(err))
		return err
	}

	authOpts, err := GenerateAuthOptsForAdmin(conn, config)
	if err != nil {
		log.Fatal("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 2 - send a commit transaction from Admin
	tx, err := ac.Commit(authOpts, commitment)
	if err != nil {
		log.Fatal("can not Commit tx", zap.Error(err))
		return err
	}

	log.Info("commit tx sent: %s. Waiting for it to be mined...", zap.String("TX hash", tx.Hash().Hex()))

	// 3 - wait for tx to be mined
	bind.WaitMined(ctx, conn, tx)
	txRes := checkTransactionReceipt(conn, tx.Hash())
	if !txRes {
		log.Warn("commit TX failed", zap.Error(err))
		return errors.New("commit tx failed")
	}

	// update nonce again...
	authOpts, err = GenerateAuthOptsForAdmin(conn, config)
	if err != nil {
		log.Fatal("can not get auth params for admin", zap.Error(err))
		return err
	}

	// 4 - now send register tx
	tx, err = ac.Register(
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
		log.Fatal("can not Commit tx", zap.Error(err))
		return err
	}

	log.Info("register tx sent. Waiting for it to be mined",
		zap.String("TX hash", tx.Hash().Hex()))

	// 5 - wait for tx to be mined
	bind.WaitMined(ctx, conn, tx)

	// 6 - return results
	txRes = checkTransactionReceipt(conn, tx.Hash())
	if !txRes {
		// new error
		return errors.New("register tx failed")
	}

	log.Info("operation succeeded!")
	return nil
}

func Utf8ToHex(input string) string {
	encoded := hex.EncodeToString([]byte(input))
	return "0x" + encoded
}

func prepareCallData(fullName string, contentHash string, spaceID string) ([][]byte, error) {
	var out [][]byte

	// 1 -
	const jsondata = `
		[
			{
				"inputs": [
					{
						"internalType": "bytes32",
						"name": "node",
						"type": "bytes32"
					},
					{
						"internalType": "bytes",
						"name": "hash",
						"type": "bytes"
					}
				],
				"name": "setContenthash",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			},

			{
				"inputs": [
					{
						"internalType": "bytes32",
						"name": "node",
						"type": "bytes32"
					},
					{
						"internalType": "bytes",
						"name": "spaceid",
						"type": "bytes"
					}
				],
				"name": "setSpaceId",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]
	`
	contractABI, err := abi.JSON(strings.NewReader(jsondata))
	if err != nil {
		log.Fatal("error parsing ABI:", zap.Error(err))
		return nil, err
	}

	// print to debug log
	log.Info("preparing call data for name",
		zap.String("Name", fullName),
		zap.String("Content hash", contentHash),
		zap.String("Space ID", spaceID))

	// 2 - convert fullName to name hash
	nh, err := NameHash(fullName)
	if err != nil {
		log.Fatal("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 3 - if spaceID is not empty - then call setSpaceId method of resolver
	if spaceID != "" {
		data, err := contractABI.Pack("setSpaceId", nh, []byte(spaceID))
		if err != nil {
			log.Fatal("error encoding function data", zap.Error(err))
			return nil, nil
		}

		log.Info("encoded space ID",
			zap.String("Data", hex.EncodeToString(data)))

		out = append(out, data)
	}

	// 4 - if contentHash is not empty - then call encodeFunctionData method of resolver
	if contentHash != "" {
		data, err := contractABI.Pack("setContenthash", nh, []byte(contentHash))
		if err != nil {
			log.Fatal("error encoding function data", zap.Error(err))
			return nil, nil
		}

		// convert bytes to string
		log.Info("encoded data: ",
			zap.String("Data", hex.EncodeToString(data)))

		out = append(out, data)
	}

	return out, nil
}
