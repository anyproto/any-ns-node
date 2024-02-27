package contracts

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"go.uber.org/zap"

	"github.com/wealdtech/go-ens/v3"
)

func PeriodMonthsToTimestamp(registerPeriodMonths uint32) big.Int {
	// default value!!!
	if registerPeriodMonths == 0 {
		registerPeriodMonths = 12
	}

	// using time.Time convert number of months to seconds, but not starting from now
	currentTime := time.Now().UTC()
	newTime := currentTime.AddDate(0, int(registerPeriodMonths), 0)
	duration := newTime.Sub(currentTime)
	totalSeconds := int64(duration.Seconds())

	var regTime big.Int = *big.NewInt(totalSeconds)
	return regTime
}

func Normalize(name string) (string, error) {
	return ens.Normalize(name)
}

// NameHash generates a hash from a name that can be used to
// look up the name in ENS
func NameHash(name string) (hash [32]byte, err error) {
	// redirect to go-ens library

	// 1. ENSIP1 standard: ens-go v3.6.0 is using it
	// 2. ENSIP15 standard: that is an another standard for ENS namehashes
	// that was accepted in June 2023.
	//
	// Current AnyNS (as of February 2024) implementation does not support it
	//
	// https://eips.ethereum.org/EIPS/eip-137 (ENSIP1) grammar:
	//
	// <domain> ::= <label> | <domain> "." <label>
	// <label> ::= any valid string label per [UTS46](https://unicode.org/reports/tr46/)
	//
	// 	btw, this will also normailze name first
	return ens.NameHash(name)
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

func PrepareCallData_SetContentHashSpaceID(fullName string, contentHash string, spaceID string) ([][]byte, error) {
	var out [][]byte

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
