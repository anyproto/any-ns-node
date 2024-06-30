package contracts

import (
	"encoding/hex"
	"testing"
	"unicode/utf8"

	"github.com/zeebo/assert"
	"golang.org/x/net/idna"
)

func checksum(out [32]byte) int {
	sum := 0
	for _, v := range out {
		sum += int(v)
	}
	return sum
}

// This is the old/current standard for ENS namehashes
//
// test NameHash function
func TestNameHash1_ENSIP1(t *testing.T) {
	// 1
	out, err := NameHash("")
	assert.NoError(t, err)
	// convert [32]byte out to string of 0x plus hex
	hexOut := "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", hexOut)

	// 2
	out, err = NameHash("eth")
	assert.NoError(t, err)
	// convert [32]byte out to string of 0x plus hex
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x93cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae", hexOut)

	// 3
	out, err = NameHash("foo.eth")
	assert.NoError(t, err)
	// convert [32]byte out to string of 0x plus hex
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f", hexOut)

	// 3.2
	out, err = NameHash("FOO.eth")
	assert.NoError(t, err)
	// convert [32]byte out to string of 0x plus hex
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f", hexOut)

	// 4 - should normailze to foo.eth (with normal 'o' letter)
	out, err = NameHash("fоо.eth") // with cyrillic 'o'
	assert.NoError(t, err)
	// convert [32]byte out to string of 0x plus hex
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x4ba2c679a3fd1e83c41104c61c8b149647e61d171805ef29338d789509c47be3", hexOut)

	// 5
	out, err = NameHash("🦚.eth")
	assert.NoError(t, err)
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x396cc1de067d8acd061f5f965f6af2e9c17422f04e37601e526ac86210e2b235", hexOut)

	// 6
	out, err = NameHash("🦚.eth")
	assert.NoError(t, err)
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x396cc1de067d8acd061f5f965f6af2e9c17422f04e37601e526ac86210e2b235", hexOut)

	// 6 - in current ENSIP15 implementation this is invalid!!!
	// check here https://app.ens.domains/
	out, err = NameHash("Ꮟіtⲥ০i̇ռ")
	assert.NoError(t, err)
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x9c655d3b0a6c9865e20311ad9d3c7394073729c60b44815591381f1479974fe4", hexOut)

	// 7 - in current ENSIP15 implementation this is invalid!!!
	// check here https://app.ens.domains/
	out, err = NameHash("❶❷❸❹❺❻❼❽❾❿")
	assert.NoError(t, err)
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0x67c15275f31d2da214a5d84deec25f4af9848b2fd0631f891a0093f404f225df", hexOut)

	// 8 - spaces
	out, err = NameHash("hello world")
	assert.NoError(t, err)
	hexOut = "0x" + hex.EncodeToString(out[:])
	assert.Equal(t, "0xad4f933a04969d30ef4d7caa6ff10c8af110b25045454179e01999cc69cc34c8", hexOut)
}

func TestNormalizeEnsip1(t *testing.T) {
	useEnsip15 := false

	// 1
	_, err := NormalizeAnyName("", useEnsip15)
	assert.Error(t, err)

	// 2
	out, err := NormalizeAnyName("Foo.any", useEnsip15)
	assert.NoError(t, err)
	assert.Equal(t, "foo.any", out)

	// 3
	out, err = NormalizeAnyName("❶❷❸❹❺❻❼❽❾❿.any", useEnsip15)
	assert.NoError(t, err)
	assert.Equal(t, "❶❷❸❹❺❻❼❽❾❿.any", out)

	// 4
	out, err = NormalizeAnyName("fоо.any", useEnsip15)
	assert.NoError(t, err)
	assert.Equal(t, "fоо.any", out)

	// 5
	_, err = NormalizeAnyName("hello world.any", useEnsip15)
	assert.Error(t, err)

	// 6 - too long
	s := "😚😉😉☺😊😉😚😚😙☺😗☺😙😙☺☺😚☺☺😚👱‍♂🙍🙍‍♀👩‍🦲👱‍♂🙎‍♂🙍‍♀🧑‍🦲🙎‍♂🙎‍♂💁‍♂🧏‍♀🧏‍♀🧏‍♂🙋🧏‍♂🤷‍♂🧏‍♂🧏‍♂🤷‍♂"
	count := utf8.RuneCountInString(s)
	assert.Equal(t, 76, count)

	// punycode it please
	punycode, err := idna.ToASCII(s)
	assert.NoError(t, err)
	len := uint32(utf8.RuneCountInString(punycode))
	assert.Equal(t, 124, len)

	_, err = NormalizeAnyName(s+".any", useEnsip15)
	assert.Error(t, err)
}

func TestNormalizeEnsip15(t *testing.T) {
	useEnsip15 := true

	// 1
	_, err := NormalizeAnyName("", useEnsip15)
	assert.Error(t, err)

	// 2
	out, err := NormalizeAnyName("Foo.any", useEnsip15)
	assert.NoError(t, err)
	assert.Equal(t, "foo.any", out)

	// 3 - HERE!
	_, err = NormalizeAnyName("❶❷❸❹❺❻❼❽❾❿.any", useEnsip15)
	assert.Error(t, err)

	// 4
	_, err = NormalizeAnyName("fоо.any", useEnsip15)
	assert.Error(t, err)

	// 5
	out, err = NormalizeAnyName("hello world.any", useEnsip15)
	assert.Error(t, err)
}

/*
// This is the new standard for ENS namehashes
// that was accepted in June 2023
//
// current AnyNS (as of February 2024) implementation does not support it
// if you uncommend this test - it will fail
//
// you can check name validation here - https://app.ens.domains/
//
// see https://github.com/adraffy/ens-normalize.js for more information
func TestNameHash2_ENSIP15_FromFile(t *testing.T) {
	type Item struct {
		Name    string `json:"name"`
		Error   bool   `json:"error"`
		Comment string `json:"comment"`
	}

	// Load the JSON file
	file, err := ioutil.ReadFile("namehash_tests.json")
	assert.NoError(t, err)

	// Unmarshal the JSON data into a slice of Item structs
	var items []Item
	err = json.Unmarshal(file, &items)
	assert.NoError(t, err)

	// print items count
	log.Info("Items count", zap.Int("Count", len(items)))

	// Iterate over the slice of Items
	index := 0
	for _, item := range items {
		_, err := NameHash(item.Name)

		// Check if an error was expected and if it matches the .error field
		if (err != nil) && !item.Error {
			log.Error("Expected NO error but got one",
				zap.Int("Index", index),
				zap.String("Name", item.Name))

			assert.Equal(t, item.Error, (err != nil))
		}
		if (err == nil) && item.Error {
			log.Error("Expected error but got nil",
				zap.Int("Index", index),
				zap.String("Name", item.Name))

			assert.Equal(t, item.Error, (err != nil))
		}

		index++
	}
}
*/
