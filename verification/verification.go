package verification

import (
	"errors"
	"strings"

	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"

	"go.uber.org/zap"
)

var log = logger.NewNamed("verification")

func CheckRegisterParams(in *nsp.NameRegisterRequest) error {
	// 1 - check name
	if !CheckName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !CheckAnyAddress(in.OwnerAnyAddress) {
		log.Error("invalid Any address", zap.String("Any address", in.OwnerAnyAddress))
		return errors.New("invalid Any address")
	}

	// everything is OK
	return nil
}

func CheckRegisterForSpaceParams(in *nsp.NameRegisterForSpaceRequest) error {
	// 1 - check name
	if !CheckName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !CheckAnyAddress(in.OwnerAnyAddress) {
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

func CheckName(name string) bool {
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

func IsValidAnyAddress(address string) bool {
	// correct address format is A5jC4SXWYEhdFswASPoMYAqWjZb9szm5EGXvS9CMyCE9JCD4
	// it should start with 1
	if !strings.HasPrefix(address, "A") {
		return false
	}

	// the len should be 52
	if len(address) != 48 {
		return false
	}

	return true
}

func CheckAnyAddress(addr string) bool {
	return IsValidAnyAddress(addr)
}

func VerifyAnyIdentity(ownerIdStr string, payload []byte, signature []byte) (err error) {
	// read in the PeerID format
	//ownerAnyIdentity, err := crypto.DecodePeerId(ownerIdStr)

	// read in the Account format (A5jC4SX...)
	ownerAnyIdentity, err := crypto.DecodeAccountAddress(ownerIdStr)

	if err != nil {
		log.Error("failed to unmarshal public key", zap.Error(err))
		return errors.New("failed to unmarshal public key")
	}

	// 2 - verify signature
	res, err := ownerAnyIdentity.Verify(payload, signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// success
	return nil
}

func VerifyAdminIdentity(adminSignKey string, payload []byte, signature []byte) (err error) {
	// 1 - load public key of admin
	// (should be account.signingKey in config)
	decodedSignKey, err := crypto.DecodeKeyFromString(
		adminSignKey,
		crypto.UnmarshalEd25519PrivateKey,
		nil)
	if err != nil {
		log.Error("failed to unmarshal public key", zap.Error(err))
		return err
	}

	ownerAnyIdentityStr := decodedSignKey.GetPublic().PeerId()
	ownerAnyIdentity, err := crypto.DecodePeerId(ownerAnyIdentityStr)

	if err != nil {
		log.Error("failed to unmarshal public key", zap.Error(err))
		return err
	}

	// 2 - verify signature
	res, err := ownerAnyIdentity.Verify(payload, signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// success
	return nil
}
