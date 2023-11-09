package alchemysdk

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/anyproto/any-sync/app"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
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

func TestAAS_Keccak256(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		hexString := hex.EncodeToString(keccak256("0x"))
		assert.Equal(t, hexString, "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")

		hexString = hex.EncodeToString(keccak256(""))
		assert.Equal(t, hexString, "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	})
}

func TestAAS_PackUserOperation(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		var request UserOperation

		request.Sender = "0x045F756F248799F4413a026100Ae49e5E7F2031E"
		request.CallData = "0x1111"
		request.Nonce = "0x3"
		request.InitCode = "0x"
		request.CallGasLimit = "0x6000"
		request.VerificationGasLimit = "0xd8f8"
		request.PreVerificationGas = "0xab90"
		request.MaxFeePerGas = "0xf708ca6"
		request.MaxPriorityFeePerGas = "0x6dc"
		request.PaymasterAndData = "0x"
		request.Signature = "0x"

		out, err := packUserOperation(request)
		// byte array to string
		outStr := "0x" + hex.EncodeToString(out)

		shouldBe := "0x000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000003c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4704a8efaf9728aab687dd27244d1090ea0c6897fbf666dea4c60524cd49862342d0000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000d8f8000000000000000000000000000000000000000000000000000000000000ab90000000000000000000000000000000000000000000000000000000000f708ca600000000000000000000000000000000000000000000000000000000000006dcc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"

		assert.NoError(t, err)
		assert.Equal(t, outStr, shouldBe)
	})
}

func TestAAS_GetCallDataForExecute(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		erc20tokenAddr := common.HexToAddress("0x8AE88b2b35F15D6320D77ab8EC7E3410F78376F6")

		// just some random data
		data1 := "0x40c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064"
		// convert data1 string to []byte array
		data1Bytes, err := hex.DecodeString(data1[2:])
		assert.NoError(t, err)

		out, err := getCallDataForExecute(erc20tokenAddr, data1Bytes)
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
	})
}

func TestAAS_GetUserOperationHash(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		var request UserOperation

		request.Sender = "0x045F756F248799F4413a026100Ae49e5E7F2031E"
		request.InitCode = "0x"
		request.CallData = "0x1111"
		request.Nonce = "0x3"
		request.CallGasLimit = "0x6000"
		request.VerificationGasLimit = "0xd8f8"
		request.PreVerificationGas = "0xab90"
		request.MaxFeePerGas = "0xf708ca6"
		request.MaxPriorityFeePerGas = "0x6dc"
		request.PaymasterAndData = "0x"
		request.Signature = "0x"

		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		out, err := getUserOperationHash(request, 11155111, entryPointAddress)
		// byte array to string
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0x61b655b51d4cb1a6f6fb3a98f5f1e95b1955891c7488dc34d87fe32a4424e4a5")
	})

	mt.Run("success 2", func(mt *mtest.T) {
		var request UserOperation

		request.Sender = "0xb856DBD4fA1A79a46D426f537455e7d3E79ab7c4"
		request.InitCode = "0x"
		request.CallData = "0xb61d27f6000000000000000000000000b856dbd4fa1a79a46d426f537455e7d3e79ab7c4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000"
		request.Nonce = "0x1f"
		request.CallGasLimit = "0x2f6c"
		request.VerificationGasLimit = "0x114c2"
		request.PreVerificationGas = "0xa890"
		request.MaxFeePerGas = "0x59682f1e"
		request.MaxPriorityFeePerGas = "0x59682f00"
		request.PaymasterAndData = "0x"
		request.Signature = "0xd16f93b584fbfdc03a5ee85914a1f29aa35c44fea5144c387ee1040a3c1678252bf323b7e9c3e9b4dfd91cca841fc522f4d3160a1e803f2bf14eb5fa037aae4a1b"

		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		out, err := getUserOperationHash(request, 80001, entryPointAddress)
		// byte array to string
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0xa70d0af2ebb03a44dcd0714a8724f622e3ab876d0aa312f0ee04823285d6fb1b")
	})
}

func TestAAS_SignDataHashWithEthereumPrivateKey(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"
		privateKeyECDSA, err := crypto.HexToECDSA(myPK)
		assert.NoError(t, err)

		data := "0x27236e94abb05957b21cba540c0d5f2c72bdb8747457e1cc23fee757667c93cf"
		// covert 'data' string to []byte array
		dataBytes, err := hex.DecodeString(data[2:])
		assert.NoError(t, err)

		out, err := signDataHashWithEthereumPrivateKey(dataBytes, privateKeyECDSA)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0x210af945f4be3a6a179e10240fd8ed5cd3d9317734d36a7b9bb969a9139bb3fc69c0f095cc52b616b1120887ceb71fd811a849ef59dc847ad3ef8c56004c5be61b")
	})

	mt.Run("success 2", func(mt *mtest.T) {
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"
		privateKeyECDSA, err := crypto.HexToECDSA(myPK)
		assert.NoError(t, err)

		data := "0x0128b079ffdb48a614f1cf8ea1d2f1da15d9715797c63be745a77e1b1d8839b7"
		// covert 'data' string to []byte array
		dataBytes, err := hex.DecodeString(data[2:])
		assert.NoError(t, err)

		out, err := signDataHashWithEthereumPrivateKey(dataBytes, privateKeyECDSA)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0xb9258a347f35b42e3862cd9c66371c110b9429617fc371eef6b147798af397d8343ac9fcf7152fd97b476463b4c9d32db964333deb7e01feca9787dae71770981b")
	})
}

func TestAA_CreateRequestGasAndPaymasterData(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
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

	mt.Run("success with non-null InitCode", func(mt *mtest.T) {
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

func TestAAS_SignDataWithEthereumPrivateKey(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"

		// prepare data
		var request UserOperation
		var chainID int64 = 11155111

		// this is Admin's SCW
		request.Sender = "0x045F756F248799F4413a026100Ae49e5E7F2031E"
		request.InitCode = "0x"
		request.CallData = "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000"
		request.Nonce = "0x7"
		request.CallGasLimit = "0x5000"
		request.VerificationGasLimit = "0xd7e2"
		request.PreVerificationGas = "0xab84"
		request.MaxFeePerGas = "0x1bcaa03d6"
		request.MaxPriorityFeePerGas = "0x60b"
		request.PaymasterAndData = "0xc03aac639bb21233e0139381970328db8bceeb67000064f9a155000064f9b1bd0000000000000000000000000000000000000000c252780de3a555372ac2a971eaf2ee453cc51ddd82e8f360420e6847ce1f78442afffdf79f154e74e1952c1fb351106060a9a4c0fe216dcc36a2f024491fee381c"
		request.Signature = "0x"

		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

		// sign it
		dataToSign, err := getUserOperationHash(request, chainID, entryPointAddress)
		assert.NoError(t, err)

		out, err := signDataWithEthereumPrivateKey(dataToSign, myPK)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0x5327e90dcb6136769a302eb6ef846b01382a97dc4b1a1422bc1911436c9a0f291ec24a0ce2f3d082a03c31d0c4b966d75dafacd1e5fd926e55ecc251cd3fc47f1c")
	})
}

func TestAA_CreateRequest(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
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

		var rgap JSONRPCResponseGasAndPaymaster

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
		var out JSONRPCRequest = JSONRPCRequest{}
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

	mt.Run("success - include InitCode", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		var rgap JSONRPCResponseGasAndPaymaster

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
		var out JSONRPCRequest = JSONRPCRequest{}
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

	mt.Run("success - ADMIN for ADMIN", func(mt *mtest.T) {
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
		         "nonce": "0x15",
		         "callData": "0x18dfb3c7000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000020000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b3000000000000000000000000c8b944dda833fb33134b96199e52f999dfbd66890000000000000000000000000000000000000000000000000000000005f5e10000000000000000000000000000000000000000000000000000000000",

		         "signature": "0xf2d8779117dce7444bc499f61cdc080ef415b68c0b8467f8fed5bd66ad494e9e1b76b03fb6654a072fe5ac0efbb184f0da03c8bda2da34b1355b500ef37dcf941c",
		         "paymasterAndData": "0xc03aac639bb21233e0139381970328db8bceeb670000652d4dfa0000652d5e62000000000000000000000000000000000000000008ffdc2f37b611e7a11839283f4bbe14abc0d7a3ff6f418e40f36cf494b9cacf1bb5c34020b38ce526c8940d6a8d64fa2ef2198667093cf09b088d98459bc89c1b",
		       },
		       "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"
		     ]
		   }
		   `
		*/

		var rgap JSONRPCResponseGasAndPaymaster

		rgap.Result.MaxPriorityFeePerGas = "0x6422c4b"
		rgap.Result.MaxFeePerGas = "0xa23be3181"
		rgap.Result.PaymasterAndData = "0xc03aac639bb21233e0139381970328db8bceeb670000652d4dfa0000652d5e62000000000000000000000000000000000000000008ffdc2f37b611e7a11839283f4bbe14abc0d7a3ff6f418e40f36cf494b9cacf1bb5c34020b38ce526c8940d6a8d64fa2ef2198667093cf09b088d98459bc89c1b"
		rgap.Result.VerificationGasLimit = "0xdb25"
		rgap.Result.CallGasLimit = "0x7000"
		rgap.Result.PreVerificationGas = "0xc3fc"

		// Sepolia
		var chainID int64 = 11155111

		// 1 - TEST getUserOperationHash
		// prepare data
		var request UserOperation

		// this is Admin's SCW
		request.Sender = "0x045F756F248799F4413a026100Ae49e5E7F2031E"
		request.InitCode = "0x"
		request.CallData = "0x18dfb3c7000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000020000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b3000000000000000000000000c8b944dda833fb33134b96199e52f999dfbd66890000000000000000000000000000000000000000000000000000000005f5e10000000000000000000000000000000000000000000000000000000000"
		request.Nonce = "0x15"
		request.CallGasLimit = "0x7000"
		request.VerificationGasLimit = "0xdb25"
		request.PreVerificationGas = "0xc3fc"
		request.MaxFeePerGas = "0xa23be3181"
		request.MaxPriorityFeePerGas = "0x6422c4b"
		request.PaymasterAndData = "0xc03aac639bb21233e0139381970328db8bceeb670000652d4dfa0000652d5e62000000000000000000000000000000000000000008ffdc2f37b611e7a11839283f4bbe14abc0d7a3ff6f418e40f36cf494b9cacf1bb5c34020b38ce526c8940d6a8d64fa2ef2198667093cf09b088d98459bc89c1b"
		request.Signature = "0x"

		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		dataToSign, err := getUserOperationHash(request, chainID, entryPointAddress)
		assert.NoError(t, err)
		// covert dataToSign to hex string
		dataToSignHexStr := "0x" + hex.EncodeToString(dataToSign)
		assert.Equal(t, dataToSignHexStr, "0xf629143c5622adb70b0f0aac4c56c644d1d3f22f67aaea673014ab307795e94f")

		// sign
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"

		out, err := signDataWithEthereumPrivateKey(dataToSign, myPK)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0x48e0411bf17e1beb381b65b9d460e174589cc7d6ff96b307cad5a7c37bd14ac53846077023f38b07b7811d7dd62b374ca14d204f6f59a71753f452d0dd2048db1b")

		// 2 - now pack the same request!
		callData, err := hex.DecodeString("18dfb3c7000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000020000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b3000000000000000000000000c8b944dda833fb33134b96199e52f999dfbd66890000000000000000000000000000000000000000000000000000000005f5e10000000000000000000000000000000000000000000000000000000000")
		assert.NoError(t, err)

		sender := common.HexToAddress("0x61d1eeE7FBF652482DEa98A1Df591C626bA09a60")
		senderScw := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(21)
		id := 18
		appendEntryPoint := false

		// 2 - TEST CreateRequestAndSign
		factoryAddr := common.Address{}
		outBytes, err := fx.CreateRequestAndSign(callData, rgap, chainID, entryPointAddress, sender, senderScw, nonce, id, myPK, factoryAddr, appendEntryPoint)
		assert.NoError(t, err)

		// convert byte array to JSON
		var r JSONRPCRequest = JSONRPCRequest{}
		err = json.Unmarshal(outBytes, &r)
		assert.NoError(t, err)

		assert.Equal(t, r.ID, 18)
		assert.Equal(t, r.JSONRPC, "2.0")
		assert.Equal(t, r.Method, "eth_sendUserOperation")
		assert.Equal(t, r.Params[0].InitCode, "0x")
		assert.Equal(t, r.Params[0].Sender, "0x045F756F248799F4413a026100Ae49e5E7F2031E")
		assert.Equal(t, r.Params[0].Nonce, "0x15")
		assert.Equal(t, r.Params[0].CallData, "0x18dfb3c7000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000020000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000044095ea7b3000000000000000000000000c8b944dda833fb33134b96199e52f999dfbd66890000000000000000000000000000000000000000000000000000000005f5e10000000000000000000000000000000000000000000000000000000000")
		assert.Equal(t, r.Params[0].PaymasterAndData, "0xc03aac639bb21233e0139381970328db8bceeb670000652d4dfa0000652d5e62000000000000000000000000000000000000000008ffdc2f37b611e7a11839283f4bbe14abc0d7a3ff6f418e40f36cf494b9cacf1bb5c34020b38ce526c8940d6a8d64fa2ef2198667093cf09b088d98459bc89c1b")
		assert.Equal(t, r.Params[0].Signature, "0x48e0411bf17e1beb381b65b9d460e174589cc7d6ff96b307cad5a7c37bd14ac53846077023f38b07b7811d7dd62b374ca14d204f6f59a71753f452d0dd2048db1b")
	})
}

func TestAA_DecodeSendUserOperationResponse(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if wrong input", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// convert string to byte array
		h := []byte("0x1")
		_, err := fx.DecodeResponseSendRequest(h)
		assert.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		respStr := `{"jsonrpc":"2.0","id":2,"result":"0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792"}`

		hash, err := fx.DecodeResponseSendRequest([]byte(respStr))
		assert.NoError(t, err)
		assert.Equal(t, hash, "0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792")
	})
}

func TestAA_GetAccountInitCode(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		addr := common.HexToAddress("0xB5CA7eBFc4BA773683810a59954820CfC42f4AcD")
		factoryAddr := common.HexToAddress("0x9406Cc6185a346906296840746125a0E44976454")

		code, err := getAccountInitCode(addr, factoryAddr)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(code), "0x9406cc6185a346906296840746125a0e449764545fbfb9cf000000000000000000000000b5ca7ebfc4ba773683810a59954820cfc42f4acd0000000000000000000000000000000000000000000000000000000000000000")
	})
}
