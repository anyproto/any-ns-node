package accountabstraction

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func GetCallDataForMint(smartAccountAddress common.Address, usdToMint uint) ([]byte, error) {
	const erc20ABI = `
		[
			{
				"constant": false,
				"inputs": [
					{
						"name": "_to",
						"type": "address"
					},
					{
						"name": "_amount",
						"type": "uint256"
					}
				],
				"name": "mint",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]	
	`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	inputData, err := parsedABI.Pack("mint", smartAccountAddress, big.NewInt(int64(usdToMint)))
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

func GetCallDataForAprove(srcAddress common.Address, destAddress common.Address, usdToMint uint) ([]byte, error) {
	const erc20ABI = `
	[
		{
			"constant": false,
			"inputs": [
				{
					"name": "_spender",
					"type": "address"
				},
				{
					"name": "_value",
					"type": "uint256"
				}
			],
			"name": "approve",
			"outputs": [
				{
					"name": "success",
					"type": "bool"
				}
			],
			"payable": false,
			"stateMutability": "nonpayable",
			"type": "function"
		},

		{
      "inputs": [
        {
          "internalType": "address",
          "name": "from",
          "type": "address"
        },
        {
          "internalType": "address",
          "name": "spender",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "value",
          "type": "uint256"
        }
      ],
      "name": "approveFor",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "nonpayable",
      "type": "function"
    }
	]
	`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal("failed to parse ABI", zap.Error(err))
		return nil, err
	}

	approveAmount := big.NewInt(int64(usdToMint))
	// multiply it by 1000000 (because approve is not in USD like mint above!)
	approveAmount = approveAmount.Mul(approveAmount, big.NewInt(1000000))

	// as long as Admin is the owner of the contract, we can approve for any address
	inputData, err := parsedABI.Pack("approveFor", srcAddress, destAddress, approveAmount)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

// in reality we call "execute" method
func GetCallDataForBatchExecute(targets []common.Address, originalCallDatas [][]byte) ([]byte, error) {
	const executeABI = `
		[
			{
				"inputs": [
					{
						"internalType": "address[]",
						"name": "dest",
						"type": "address[]"
					},
					{
						"internalType": "bytes[]",
						"name": "func",
						"type": "bytes[]"
					}
				],
				"name": "executeBatch",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]
	`

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

func GetCallDataForCommit(commitment [32]byte) ([]byte, error) {
	const commitABI = `
		[
			{
				"inputs": [
					{
						"internalType": "bytes32",
						"name": "commitment",
						"type": "bytes32"
					}
				],
				"name": "commit",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]
	`

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

func GetCallDataForRegister(
	nameFirstPart string,
	registrantAccount common.Address,
	registrationTime big.Int,
	secret [32]byte,
	resolver common.Address,
	callData [][]byte,
	isReverseRecord bool,
	ownerControlledFuses uint16) ([]byte, error) {

	const regABI = `
		[
			{
				"inputs": [
					{
						"internalType": "string",
						"name": "name",
						"type": "string"
					},
					{
						"internalType": "address",
						"name": "owner",
						"type": "address"
					},
					{
						"internalType": "uint256",
						"name": "duration",
						"type": "uint256"
					},
					{
						"internalType": "bytes32",
						"name": "secret",
						"type": "bytes32"
					},
					{
						"internalType": "address",
						"name": "resolver",
						"type": "address"
					},
					{
						"internalType": "bytes[]",
						"name": "data",
						"type": "bytes[]"
					},
					{
						"internalType": "bool",
						"name": "reverseRecord",
						"type": "bool"
					},
					{
						"internalType": "uint16",
						"name": "ownerControlledFuses",
						"type": "uint16"
					}
				],
				"name": "register",
				"outputs": [],
				"stateMutability": "nonpayable",
				"type": "function"
			}	
		]
	`

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
