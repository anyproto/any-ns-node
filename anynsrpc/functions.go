package anynsrpc

import (
	"errors"

	"github.com/anyproto/any-sync/util/crypto"
	as "github.com/anyproto/anyns-node/pb/anyns_api_server"
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
