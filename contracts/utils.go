package contracts

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
)

func nameHashPart(prevHash [32]byte, name string) (hash [32]byte, err error) {
	sha := sha3.NewLegacyKeccak256()
	if _, err = sha.Write(prevHash[:]); err != nil {
		return
	}

	nameSha := sha3.NewLegacyKeccak256()
	if _, err = nameSha.Write([]byte(name)); err != nil {
		return
	}
	nameHash := nameSha.Sum(nil)
	if _, err = sha.Write(nameHash); err != nil {
		return
	}
	sha.Sum(hash[:0])
	return
}

// NameHash generates a hash from a name that can be used to
// look up the name in ENS
func NameHash(name string) (hash [32]byte, err error) {
	if name == "" {
		return
	}

	parts := strings.Split(name, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if hash, err = nameHashPart(hash, parts[i]); err != nil {
			return
		}
	}
	return
}

func RemoveTLD(str string) string {
	suffix := ".any"

	if strings.HasSuffix(str, suffix) {
		return str[:len(str)-len(suffix)]
	}
	return str
}

func GenerateRandomSecret() ([32]byte, error) {
	var byteArray [32]byte

	_, err := rand.Read(byteArray[:])
	if err != nil {
		return byteArray, err
	}
	return byteArray, nil
}

func PrepareCallData(fullName string, contentHash string, spaceID string) ([][]byte, error) {
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
		log.Error("error parsing ABI:", zap.Error(err))
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
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 3 - if spaceID is not empty - then call setSpaceId method of resolver
	if spaceID != "" {
		data, err := contractABI.Pack("setSpaceId", nh, []byte(spaceID))
		if err != nil {
			log.Error("error encoding function data", zap.Error(err))
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
			log.Error("error encoding function data", zap.Error(err))
			return nil, nil
		}

		// convert bytes to string
		log.Info("encoded data: ",
			zap.String("Data", hex.EncodeToString(data)))

		out = append(out, data)
	}

	return out, nil
}
