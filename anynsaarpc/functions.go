package anynsaarpc

import (
	"errors"
	"strings"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"
)

func checkRegisterParams(in *nsp.NameRegisterRequest) error {
	// 1 - check name
	if !checkName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !checkAnyAddress(in.OwnerAnyAddress) {
		log.Error("invalid Any address", zap.String("Any address", in.OwnerAnyAddress))
		return errors.New("invalid Any address")
	}

	// everything is OK
	return nil
}

func checkRegisterForSpaceParams(in *nsp.NameRegisterForSpaceRequest) error {
	// 1 - check name
	if !checkName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !checkAnyAddress(in.OwnerAnyAddress) {
		log.Error("invalid Any address", zap.String("Any address", in.OwnerAnyAddress))
		return errors.New("invalid Any address")
	}

	// 4 - space ID (if not empty)
	if in.SpaceId != "" {
		_, err := cid.Decode(in.SpaceId)

		if err != nil {
			log.Error("invalid SpaceId", zap.String("Any SpaceId", in.SpaceId))
			return errors.New("invalid SpaceId")
		}
	}

	// everything is OK
	return nil
}

func checkName(name string) bool {
	// get name parts
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}

	// if extension is not 'any', then return false
	if parts[len(parts)-1] != "any" {
		return false
	}

	// if first part is less than 3 chars, then return false
	if len(parts[0]) < 3 {
		return false
	}

	// if it has slashes, then return false
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}

	return true
}

func isValidAnyAddress(address string) bool {
	// correct address format is 12D3KooWPANzVZgHqAL57CchRH4q8NGjoWDpUShVovBE3bhhXczy
	// it should start with 1
	if !strings.HasPrefix(address, "1") {
		return false
	}

	// the len should be 52
	if len(address) != 52 {
		return false
	}

	return true
}

func checkAnyAddress(addr string) bool {
	// in.OwnerAnyAddress should be a ed25519 public key hash
	return isValidAnyAddress(addr)
}
