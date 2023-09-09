package alchemyaa

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeebo/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestAAS_Keccak256(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		hexString := hex.EncodeToString(Keccak256("0x"))
		assert.Equal(t, hexString, "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")

		hexString = hex.EncodeToString(Keccak256(""))
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

		out, err := PackUserOperation(request)
		// byte array to string
		outStr := "0x" + hex.EncodeToString(out)

		shouldBe := "0x000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000003c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4704a8efaf9728aab687dd27244d1090ea0c6897fbf666dea4c60524cd49862342d0000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000d8f8000000000000000000000000000000000000000000000000000000000000ab90000000000000000000000000000000000000000000000000000000000f708ca600000000000000000000000000000000000000000000000000000000000006dcc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"

		assert.NoError(t, err)
		assert.Equal(t, outStr, shouldBe)
	})
}

func TestAAS_GetCallDataForMint(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		smartAccountAddress := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		var usdToMint uint = 100

		out, err := GetCallDataForMint(smartAccountAddress, usdToMint)
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0x40c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064")
	})
}

func TestAAS_GetCallDataForAprove(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		from := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		registrarController := common.HexToAddress("0xB6bF17cBe45CbC7609e4f8fA56154c9DeF8590CA")
		var usdToMint uint = 100

		out, err := GetCallDataForAprove(from, registrarController, usdToMint)
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0x2b991746000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000b6bf17cbe45cbc7609e4f8fa56154c9def8590ca0000000000000000000000000000000000000000000000000000000005f5e100")
	})
}

func TestAAS_GetCallDataForExecute(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		smartAccountAddress := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		var usdToMint uint = 100
		erc20tokenAddr := common.HexToAddress("0x8AE88b2b35F15D6320D77ab8EC7E3410F78376F6")

		data1, err := GetCallDataForMint(smartAccountAddress, usdToMint)

		out, err := GetCallDataForExecute(erc20tokenAddr, data1)
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0xb61d27f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000")
	})
}

func TestAAS_GetCallDataForBatchExecute(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		smartAccountAddress := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		registrarController := common.HexToAddress("0xB6bF17cBe45CbC7609e4f8fA56154c9DeF8590CA")

		var usdToMint uint = 100
		erc20tokenAddr := common.HexToAddress("0x8AE88b2b35F15D6320D77ab8EC7E3410F78376F6")
		callDataOriginal1, _ := GetCallDataForMint(smartAccountAddress, usdToMint)
		callDataOriginal2, _ := GetCallDataForAprove(smartAccountAddress, registrarController, 100)

		// convert []byte to hex string
		callDataOriginal1Str := "0x" + hex.EncodeToString(callDataOriginal1)
		callDataOriginal2Str := "0x" + hex.EncodeToString(callDataOriginal2)

		assert.Equal(t, callDataOriginal1Str, "0x40c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e0000000000000000000000000000000000000000000000000000000000000064")
		assert.Equal(t, callDataOriginal2Str, "0x2b991746000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000b6bf17cbe45cbc7609e4f8fa56154c9def8590ca0000000000000000000000000000000000000000000000000000000005f5e100")

		// put address and address2 into array
		// both are the same
		addresses := []common.Address{erc20tokenAddr, erc20tokenAddr}
		// put data1 and callDataOriginal2 into array
		datas := [][]byte{callDataOriginal1, callDataOriginal2}

		//////
		out, err := GetCallDataForBatchExecute(addresses, datas)
		outStr := "0x" + hex.EncodeToString(out)

		assert.NoError(t, err)
		assert.Equal(t, outStr, "0x18dfb3c7000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000020000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000008ae88b2b35f15d6320d77ab8ec7e3410f78376f60000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000004440c10f19000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e00000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000642b991746000000000000000000000000045f756f248799f4413a026100ae49e5e7f2031e000000000000000000000000b6bf17cbe45cbc7609e4f8fa56154c9def8590ca0000000000000000000000000000000000000000000000000000000005f5e10000000000000000000000000000000000000000000000000000000000")
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

		out, err := GetUserOperationHash(request, 11155111, entryPointAddress)
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

		out, err := GetUserOperationHash(request, 80001, entryPointAddress)
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

		out, err := SignDataHashWithEthereumPrivateKey(dataBytes, privateKeyECDSA)
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

		out, err := SignDataHashWithEthereumPrivateKey(dataBytes, privateKeyECDSA)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0xb9258a347f35b42e3862cd9c66371c110b9429617fc371eef6b147798af397d8343ac9fcf7152fd97b476463b4c9d32db964333deb7e01feca9787dae71770981b")
	})
}

func TestAA_CreateRequestGasAndPaymasterData(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
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
		sender := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(4)
		id := 13

		policyID := "22032aca-2101-40d5-8550-14a6a11366ba"
		entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		out, err := CreateRequestGasAndPaymasterData(callData, sender, nonce, policyID, entryPoint, id)
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
		dataToSign, err := GetUserOperationHash(request, chainID, entryPointAddress)
		assert.NoError(t, err)

		out, err := SignDataWithEthereumPrivateKey(dataToSign, myPK)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+hex.EncodeToString(out), "0x5327e90dcb6136769a302eb6ef846b01382a97dc4b1a1422bc1911436c9a0f291ec24a0ce2f3d082a03c31d0c4b966d75dafacd1e5fd926e55ecc251cd3fc47f1c")
	})
}

func TestAA_CreateRequest(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
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
		sender := common.HexToAddress("0x045F756F248799F4413a026100Ae49e5E7F2031E")
		nonce := uint64(8)
		id := 32

		// do not append it only for test, otherwise JSON Unmarshal won't work
		myPK := "ac4bab11ad6b7ec2c84e5e293710828234ab63b62d377a23681228be588fab57"
		appendEntryPoint := false

		// TODO: params!
		var chainID int64 = 11155111
		entryPointAddress := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
		outBytes, err := CreateRequest(callData, rgap, chainID, entryPointAddress, sender, nonce, id, myPK, appendEntryPoint)
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
}

func TestAA_DecodeSendUserOperationResponse(t *testing.T) {
	var mt = mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("fail if wrong input", func(mt *mtest.T) {
		// convert string to byte array
		h := []byte("0x1")
		_, err := DecodeSendUserOperationResponse(h)
		assert.Error(t, err)
	})

	mt.Run("success", func(mt *mtest.T) {
		respStr := `{"jsonrpc":"2.0","id":2,"result":"0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792"}`

		hash, err := DecodeSendUserOperationResponse([]byte(respStr))
		assert.NoError(t, err)
		assert.Equal(t, hash, "0xa417d6e564c27e7803097f7c712490896d093e27c6f9f44b0192252d82522792")
	})
}
