package contracts

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/idna"

	"github.com/wealdtech/go-ens/v3"

	"github.com/adraffy/ENSNormalize.go/ensip15"
)

var (
	// p       = idna.New(idna.MapForLookup(), idna.ValidateLabels(false), idna.CheckHyphens(false), idna.StrictDomainName(false), idna.Transitional(false))
	pStrict = idna.New(idna.MapForLookup(), idna.ValidateLabels(false), idna.CheckHyphens(false), idna.StrictDomainName(true), idna.Transitional(false))
)

const MAX_NAME_LENGTH = 100

func normalize(input string) (string, error) {
	// output, err := p.ToUnicode(input)
	// if name has no .any suffix -> error
	if len(input) < 4 || input[len(input)-4:] != ".any" {
		return "", errors.New("name must have .any suffix")
	}
	// remove .any suffix
	input = input[:len(input)-4]

	// somehow "github.com/wealdtech/go-ens/v3" used non-strict version of idna
	// let's use pStrict instead of p
	output, err := pStrict.ToUnicode(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert to standard unicode")
	}
	if strings.Contains(input, ".") {
		return "", errors.New("name cannot contain a period")
	}

	// now check the punycode length of the name
	punycode, err := idna.ToASCII(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert to punycode")
	}
	len := uint32(utf8.RuneCountInString(punycode))
	if len > MAX_NAME_LENGTH {
		return "", errors.New("name too long")
	}

	// add .any suffix
	output += ".any"

	return output, nil
}

func normalize15(input string) (string, error) {
	// output, err := p.ToUnicode(input)
	// if name has no .any suffix -> error
	if len(input) < 4 || input[len(input)-4:] != ".any" {
		return "", errors.New("name must have .any suffix")
	}
	// remove .any suffix
	input = input[:len(input)-4]

	// TODO: not a standard approach
	ens := ensip15.New() // or ensip15.Shared()
	output, err := ens.Normalize(input)

	if err != nil {
		return "", err
	}

	// add .any suffix
	output += ".any"

	return output, nil
}

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

func NormalizeAnyName(name string, useEnsip15 bool) (string, error) {
	if useEnsip15 {
		return normalize15(name)
	} else {
		//return ens.Normalize(name)
		return normalize(name)
	}
}

// NameHash generates a hash from a name that can be used to
// look up the name in ENS
func NameHash(name string) (hash [32]byte, err error) {
	// redirect to go-ens library

	// 1. ENSIP1 standard: ens-go v3.6.0 (current) is using it
	// 2. ENSIP15 standard: that is an another standard for ENS namehashes
	// that was accepted in June 2023.
	//
	// Current AnyNS (as of June 2024) implementation supports ENSIP1, ENSIP15
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
