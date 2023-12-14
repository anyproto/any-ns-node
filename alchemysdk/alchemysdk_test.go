package alchemysdk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	asdk "github.com/anyproto/alchemy-aa-sdk/alchemysdk"
	"github.com/anyproto/any-sync/app"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/mock/gomock"
)

var ctx = context.Background()

type fixture struct {
	a    *app.App
	ctrl *gomock.Controller

	*alchemysdk
}

func newFixture(t *testing.T) *fixture {
	fx := &fixture{
		a:          new(app.App),
		ctrl:       gomock.NewController(t),
		alchemysdk: New().(*alchemysdk),
	}

	require.NoError(t, fx.a.Start(ctx))
	return fx
}

func (fx *fixture) finish(t *testing.T) {
	assert.NoError(t, fx.a.Close(ctx))
	fx.ctrl.Finish()
}

func TestAA_CreateRequestGasAndPaymasterData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		/*
		   json := `
		     {
		       "jsonrpc": "2.0",
		       "id": 13,
		       "method": "alchemy_requestGasAndPaymasterAndData",
		       "params": [
		         {
		           "policyId": "22032aca-2101-40d5-8550-14a6a11366ba",
		           "entryPoint": "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789",
		           "userOperation": {
		             "initCode": "0x",
		             "sender": "0x045F756F248799F4413a026100Ae49e5E7F2031E",
		             "nonce": "0x4",
		             "callData": "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000",
		             "signature": "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c",
		             "paymasterAndData": "0x",
		             "maxFeePerGas": "0x0",
		             "maxPriorityFeePerGas": "0x0",
		             "callGasLimit": "0x0",
		             "preVerificationGas": "0x0",
		             "verificationGasLimit": "0x0"
		           },
		           "dummySignature": "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c"
		         }
		       ]
		     }
		   `
		*/

		// convert string to byte array
		callData, _ := hex.DecodeString("b61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		sender := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		scw := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")

		nonce := uint64(4)
		id := 13

		policyID := "22032aca-2101-40d5-8550-14a6a11366ba"
		entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		factoryAddress := common.Address{}

		out, err := fx.CreateRequestGasAndPaymasterData(callData, sender, scw, nonce, policyID, entryPoint, factoryAddress, id)
		assert.NoError(t, err)

		assert.Equal(t, out.ID, 13)
		assert.Equal(t, out.JSONRPC, "2.0")
		assert.Equal(t, out.Method, "alchemy_requestGasAndPaymasterAndData")
		assert.Equal(t, out.Params[0].PolicyID, "22032aca-2101-40d5-8550-14a6a11366ba")
		assert.Equal(t, out.Params[0].EntryPoint, "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		assert.Equal(t, out.Params[0].UserOperation.InitCode, "0x")
		assert.Equal(t, out.Params[0].UserOperation.Sender, "0x045F756F248799F4413a026100Ae49e5E7F2031E")
		assert.Equal(t, out.Params[0].UserOperation.Nonce, "0x4")
		assert.Equal(t, out.Params[0].UserOperation.CallData, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].UserOperation.Signature, "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c")
		assert.Equal(t, out.Params[0].UserOperation.PaymasterAndData, "0x")
		assert.Equal(t, out.Params[0].UserOperation.MaxFeePerGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.MaxPriorityFeePerGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.CallGasLimit, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.PreVerificationGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.VerificationGasLimit, "0x0")

		assert.Equal(t, out.Params[0].DummySignature, "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c")
	})

	t.Run("success with non-null InitCode", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// convert string to byte array
		callData, _ := hex.DecodeString("b61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		sender := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		scw := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(4)
		id := 13

		policyID := "22032aca-2101-40d5-8550-14a6a11366ba"
		entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		factoryAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2111")

		out, err := fx.CreateRequestGasAndPaymasterData(callData, sender, scw, nonce, policyID, entryPoint, factoryAddress, id)
		assert.NoError(t, err)

		assert.Equal(t, out.ID, 13)
		assert.Equal(t, out.JSONRPC, "2.0")
		assert.Equal(t, out.Method, "alchemy_requestGasAndPaymasterAndData")
		assert.Equal(t, out.Params[0].PolicyID, "22032aca-2101-40d5-8550-14a6a11366ba")
		assert.Equal(t, out.Params[0].EntryPoint, "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		// Should not be 0x !
		assert.Equal(t, out.Params[0].UserOperation.InitCode, "0x5ff137d4b0fdcd49dca30c7cf57e578a026d21115fbfb9cf00000000000000000000000061d1eee7fbf652482dea98a1df591c626ba09a600000000000000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].UserOperation.Sender, "0x045F756F248799F4413a026100Ae49e5E7F2031E")
		assert.Equal(t, out.Params[0].UserOperation.Nonce, "0x4")
		assert.Equal(t, out.Params[0].UserOperation.CallData, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].UserOperation.Signature, "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c")
		assert.Equal(t, out.Params[0].UserOperation.PaymasterAndData, "0x")
		assert.Equal(t, out.Params[0].UserOperation.MaxFeePerGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.MaxPriorityFeePerGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.CallGasLimit, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.PreVerificationGas, "0x0")
		assert.Equal(t, out.Params[0].UserOperation.VerificationGasLimit, "0x0")

		assert.Equal(t, out.Params[0].DummySignature, "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c")
	})
}

func TestAA_CreateRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		/*
		   {
		     "jsonrpc": "2.0",
		     "id": 32,
		     "method": "eth_sendUserOperation",
		     "params": [
		       {
		         "initCode": "0x",
		         "sender": "0x045F756F248799F4413a026100Ae49e5E7F2031E",
		         "nonce": "0x8",
		         "callData": "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000",

		         "signature": "0x985c208f2ce62b0ce1b8e4a2099d86d186ed48edc96b3361e0fbc50361581c565d289b1e477260edb32c89110dd615a721da8ed586470c0e6a253ad3c50f7f3d1b",
		         "paymasterAndData": "0xc03aac639bb21233e0139381970328db8bceeb67000064f9ca6c000064f9dad40000000000000000000000000000000000000000796f4ebcef9ae51a6d5131b1344228c971982353cc698f67e309ffb320ef04787ccec730f240788c78e6e1d096e3376a49782f51d38ba28b9eaeed1bca833be01c",
		       },
		       "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"
		     ]
		   }
		   `
		*/

		var rgap asdk.JSONRPCResponseGasAndPaymaster

		rgap.Result.MaxPriorityFeePerGas = "0x60b"
		rgap.Result.MaxFeePerGas = "0xf732015ce"
		rgap.Result.PaymasterAndData = "0xc03aac639bb21233e0139381970328db8bceeb67000064f9ca6c000064f9dad40000000000000000000000000000000000000000796f4ebcef9ae51a6d5131b1344228c971982353cc698f67e309ffb320ef04787ccec730f240788c78e6e1d096e3376a49782f51d38ba28b9eaeed1bca833be01c"
		rgap.Result.VerificationGasLimit = "0xd7e2"
		rgap.Result.CallGasLimit = "0x5000"
		rgap.Result.PreVerificationGas = "0xab84"

		// now pack the request
		callData, _ := hex.DecodeString("b61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		sender := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		senderScw := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(8)
		id := 32

		// do not append it only for test, otherwise JSON Unmarshal won't work
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"
		appendEntryPoint := false

		var chainID int64 = 11155111
		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		factoryAddr := common.Address{}
		outBytes, err := fx.CreateRequestAndSign(callData, rgap, chainID, entryPointAddress, sender, senderScw, nonce, id, myPK, factoryAddr, appendEntryPoint)
		assert.NoError(t, err)

		// convert byte array to JSON
		var out asdk.JSONRPCRequest = asdk.JSONRPCRequest{}
		err = json.Unmarshal(outBytes, &out)
		assert.NoError(t, err)

		assert.Equal(t, out.ID, 32)
		assert.Equal(t, out.JSONRPC, "2.0")
		assert.Equal(t, out.Method, "eth_sendUserOperation")
		assert.Equal(t, out.Params[0].InitCode, "0x")
		assert.Equal(t, out.Params[0].Sender, "0x045F756F248799F4413a026100Ae49e5E7F2031E")
		assert.Equal(t, out.Params[0].Nonce, "0x8")
		assert.Equal(t, out.Params[0].CallData, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].PaymasterAndData, "0xc03aac639bb21233e0139381970328db8bceeb67000064f9ca6c000064f9dad40000000000000000000000000000000000000000796f4ebcef9ae51a6d5131b1344228c971982353cc698f67e309ffb320ef04787ccec730f240788c78e6e1d096e3376a49782f51d38ba28b9eaeed1bca833be01c")
		assert.Equal(t, out.Params[0].Signature, "0x571ec8a77c9ed42958db1f2f31b3883f773cb4bb6a225208fa13ea8f53dc435939c22e6fae79da717977881d5288bc7de2b840b54b27df6230906244c665e6d51b")
	})

	t.Run("success - include InitCode", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var rgap asdk.JSONRPCResponseGasAndPaymaster

		rgap.Result.MaxPriorityFeePerGas = "0x60b"
		rgap.Result.MaxFeePerGas = "0xf732015ce"
		rgap.Result.PaymasterAndData = "0xc03aac639bb21233e0139381970328db8bceeb67000064f9ca6c000064f9dad40000000000000000000000000000000000000000796f4ebcef9ae51a6d5131b1344228c971982353cc698f67e309ffb320ef04787ccec730f240788c78e6e1d096e3376a49782f51d38ba28b9eaeed1bca833be01c"
		rgap.Result.VerificationGasLimit = "0xd7e2"
		rgap.Result.CallGasLimit = "0x5000"
		rgap.Result.PreVerificationGas = "0xab84"

		// now pack the request
		callData, _ := hex.DecodeString("b61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		sender := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		senderScw := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(8)
		id := 32

		// do not append it only for test, otherwise JSON Unmarshal won't work
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"
		appendEntryPoint := false

		var chainID int64 = 11155111
		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		factoryAddr := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		outBytes, err := fx.CreateRequestAndSign(callData, rgap, chainID, entryPointAddress, sender, senderScw, nonce, id, myPK, factoryAddr, appendEntryPoint)
		assert.NoError(t, err)

		// convert byte array to JSON
		var out asdk.JSONRPCRequest = asdk.JSONRPCRequest{}
		err = json.Unmarshal(outBytes, &out)
		assert.NoError(t, err)

		assert.Equal(t, out.ID, 32)
		assert.Equal(t, out.JSONRPC, "2.0")
		assert.Equal(t, out.Method, "eth_sendUserOperation")
		// Should not be 0x !
		assert.Equal(t, out.Params[0].InitCode, "0x61d1eee7fbf652482dea98a1df591c626ba09a605fbfb9cf00000000000000000000000061d1eee7fbf652482dea98a1df591c626ba09a600000000000000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].Sender, "0x045F756F248799F4413a026100Ae49e5E7F2031E")
		assert.Equal(t, out.Params[0].Nonce, "0x8")
		assert.Equal(t, out.Params[0].CallData, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, out.Params[0].PaymasterAndData, "0xc03aac639bb21233e0139381970328db8bceeb67000064f9ca6c000064f9dad40000000000000000000000000000000000000000796f4ebcef9ae51a6d5131b1344228c971982353cc698f67e309ffb320ef04787ccec730f240788c78e6e1d096e3376a49782f51d38ba28b9eaeed1bca833be01c")
		assert.Equal(t, out.Params[0].Signature, "0xd1e8c6a31b68ea76f58428f95c59e6eaea030869ffd198acd1bf767448a726553bc94a719f5fcdccc28a95bd527c9f526ce813e38791ffcf7ba5d5cfb0b854011c")
	})
}

func TestAA_DecodeSendUserOperationResponse(t *testing.T) {
	t.Run("fail if wrong input", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// convert string to byte array
		h := []byte("0x1")
		_, err := fx.DecodeResponseSendRequest(h)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		respStr := `{"jsonrpc":"2.0","id":2,"result":"0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792"}`

		hash, err := fx.DecodeResponseSendRequest([]byte(respStr))
		assert.NoError(t, err)
		assert.Equal(t, hash, "0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792")
	})
}
