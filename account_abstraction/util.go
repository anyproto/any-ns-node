package accountabstraction

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func getCallDataForMint(smartAccountAddress common.Address, fullTokensToMint *big.Int, tokenDecimals uint8) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	// 6 decimals
	weiPerToken := big.NewInt(1).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)
	fullTokensToMint = weiPerToken.Mul(fullTokensToMint, weiPerToken)

	inputData, err := parsedABI.Pack("mint", smartAccountAddress, fullTokensToMint)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

func getCallDataForAprove(userAddr common.Address, destAddress common.Address, fullTokensToAllow *big.Int, tokenDecimals uint8) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	// 6 decimals
	weiPerToken := big.NewInt(1).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimals)), nil)
	fullTokensToAllow = weiPerToken.Mul(fullTokensToAllow, weiPerToken)

	// as long as Admin is the owner of the contract, we can approve for any address
	inputData, err := parsedABI.Pack("approveFor", userAddr, destAddress, fullTokensToAllow)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

// in reality we call "execute" method
func getCallDataForBatchExecute(targets []common.Address, originalCallDatas [][]byte) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(executeABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	// TODO: value (Ether) is ZERO here!
	inputData, err := parsedABI.Pack("executeBatch", targets, originalCallDatas)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

func getCallDataForCommit(commitment [32]byte) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(commitABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	inputData, err := parsedABI.Pack("commit", commitment)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

func getCallDataForRegister(
	nameFirstPart string,
	registrantAccount common.Address,
	registrationTime big.Int,
	secret [32]byte,
	resolver common.Address,
	callData [][]byte,
	isReverseRecord bool,
	ownerControlledFuses uint16) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(regABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	inputData, err := parsedABI.Pack("register",
		nameFirstPart,
		registrantAccount,
		&registrationTime,
		secret,
		resolver,
		callData,
		isReverseRecord,
		ownerControlledFuses)

	if err != nil {
		return nil, err
	}

	return inputData, nil
}
