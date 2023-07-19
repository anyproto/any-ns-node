package anynsrpc

import (
	"crypto/rand"
	"strings"

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
