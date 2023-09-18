package alchemysdk

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"

	"fmt"
	"math/big"
)

func padHex(hex string) string {
	// add 24 zeros to the left
	return strings.Repeat("0", 24) + strings.TrimPrefix(hex, "0x")
}

func encodeAddress(value string) string {
	return padHex(strings.ToLower(string(value)))
}

func encodeAbiParameters(types []string, values []interface{}) ([]byte, error) {
	if len(types) != len(values) {
		return nil, fmt.Errorf("mismatch between types and values length")
	}

	var encodedData []byte

	for i := 0; i < len(types); i++ {
		switch types[i] {
		case "address":
			encodedStr := encodeAddress(values[i].(string))
			addressBytes := hexToBytes(encodedStr)
			encodedData = append(encodedData, addressBytes...)
		case "uint256":
			intBytes := values[i].(*big.Int).Bytes()
			// Ensure uint256 is 32 bytes long
			padSize := 32 - len(intBytes)
			padBytes := make([]byte, padSize)
			encodedData = append(encodedData, padBytes...)
			encodedData = append(encodedData, intBytes...)
		case "bytes32":
			bytes32Value := hexToBytes32(values[i].(string))
			encodedData = append(encodedData, bytes32Value[:]...)
		default:
			return nil, fmt.Errorf("unsupported type: %s", types[i])
		}
	}

	return encodedData, nil
}

func Keccak256(data string) []byte {
	// remove 0x prefix if exists
	data = strings.TrimPrefix(data, "0x")

	// convert hex string to bytes
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		log.Error("failed to decode hex string", zap.Error(err))
		return []byte{}
	}

	// calculate hash
	hash := crypto.Keccak256Hash(dataBytes).Bytes()
	return hash
}

func hexToBytes32(hexStr string) [32]byte {
	var bytes32Value [32]byte
	copy(bytes32Value[:], hexToBytes(hexStr))
	return bytes32Value
}

func hexToBytes(hexStr string) []byte {
	data, _ := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	return data
}

func hexToBigInt(hex string) *big.Int {
	// remove 0x prefix if exists
	hex = strings.TrimPrefix(hex, "0x")

	value := new(big.Int)
	value.SetString(hex, 16)
	return value
}

func GetUserOperationHash(request UserOperation, chainID int64, entryPointAddress common.Address) ([]byte, error) {
	uoBytes, err := PackUserOperation(request)
	if err != nil {
		return nil, err
	}

	// uoBytes to hex string
	uoHex := hex.EncodeToString(uoBytes)
	uoHexKeccak := "0x" + hex.EncodeToString(Keccak256(uoHex))

	eap, err := encodeAbiParameters(
		[]string{
			"bytes32",
			"address",
			"uint256",
		},
		[]interface{}{
			uoHexKeccak,
			entryPointAddress.String(),
			big.NewInt(chainID),
		},
	)

	if err != nil {
		return nil, err
	}

	return Keccak256(hex.EncodeToString(eap)), nil
}

func PackUserOperation(request UserOperation) ([]byte, error) {
	// byte arrays
	hashedInitCode := "0x" + hex.EncodeToString(Keccak256(request.InitCode))
	hashedCallData := "0x" + hex.EncodeToString(Keccak256(request.CallData))
	hashedPaymasterAndData := "0x" + hex.EncodeToString(Keccak256(request.PaymasterAndData))

	return encodeAbiParameters(
		[]string{
			"address",
			"uint256",
			"bytes32",
			"bytes32",
			"uint256",
			"uint256",
			"uint256",
			"uint256",
			"uint256",
			"bytes32",
		},
		[]interface{}{
			request.Sender,
			hexToBigInt(request.Nonce),
			hashedInitCode,
			hashedCallData,
			hexToBigInt(request.CallGasLimit),
			hexToBigInt(request.VerificationGasLimit),
			hexToBigInt(request.PreVerificationGas),
			hexToBigInt(request.MaxFeePerGas),
			hexToBigInt(request.MaxPriorityFeePerGas),
			hashedPaymasterAndData,
		},
	)
}

func GetCallDataForExecute(dest common.Address, originalCallData []byte) ([]byte, error) {
	const executeABI = `
	[
    {
        "inputs": [
            {
                "internalType": "address",
                "name": "dest",
                "type": "address"
            },
            {
                "internalType": "uint256",
                "name": "value",
                "type": "uint256"
            },
            {
                "internalType": "bytes",
                "name": "func",
                "type": "bytes"
            }
        ],
        "name": "execute",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
	]
	`

	parsedABI, err := abi.JSON(strings.NewReader(executeABI))
	if err != nil {
		return nil, err
	}

	// TODO: value (Ether) is ZERO here!
	inputData, err := parsedABI.Pack("execute", dest, big.NewInt(0), originalCallData)
	if err != nil {
		return nil, err
	}

	return inputData, nil
}

func SignDataWithEthereumPrivateKey(data []byte, privateKeyHex string) ([]byte, error) {
	// Generate a new ECDSA private key.
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return SignDataHashWithEthereumPrivateKey(data, privateKeyECDSA)
}

func SignDataHashWithEthereumPrivateKey(dataToSign []byte, privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	if len(dataToSign) != 32 {
		return nil, errors.New("dataToSign must be 32 bytes long")
	}

	// Prepend the "Ethereum Signed Message" prefix.
	prefix := "\x19Ethereum Signed Message:\n32" + string(dataToSign)

	// Hash the message using Keccak-256.
	hash := crypto.Keccak256([]byte(prefix))

	signature, err := crypto.Sign(hash, privateKeyECDSA)
	if err != nil {
		return nil, err
	}
	//return CompactSignature(signature)

	// TODO: fix +27 please!
	// Encode the signature in Ethereum's compact signature format.
	out := append(signature[:64], signature[64]+27)
	return out, nil
}

func AppendEntryPointAddress(jsonData []byte, entryPointAddress common.Address) (error, []byte) {
	// Define a struct to represent the JSON data
	var data map[string]interface{}

	// Parse the JSON data into the struct
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		log.Error("failed to unmarshal JSON", zap.Error(err))
		return err, nil
	}

	// Add the custom value to the 'params' array
	paramsArray, ok := data["params"].([]interface{})
	if !ok {
		log.Error("'params' field is not an array")
		return errors.New("'params' field is not an array"), nil
	}
	paramsArray = append(paramsArray, entryPointAddress.String())
	data["params"] = paramsArray

	// Encode the modified data back to JSON
	outputJSON, err := json.Marshal(data)
	if err != nil {
		log.Error("error encoding JSON:", zap.Error(err))
		return err, nil
	}

	return nil, outputJSON
}
