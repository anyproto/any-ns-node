package anynsrpc

import (
	"errors"
	"strings"

	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
	"github.com/anyproto/any-sync/util/crypto"
)

func VerifyIdentity(in *as.NameRegisterSignedRequest, ownerAnyAddress string) error {
	// convert ownerAnyAddress to array of bytes
	arr := []byte(ownerAnyAddress)

	ownerAnyIdentity, err := crypto.UnmarshalEd25519PublicKeyProto(arr)
	if err != nil {
		return err
	}

	res, err := ownerAnyIdentity.Verify(in.Payload, in.Signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// identity is OK
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
