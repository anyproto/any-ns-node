// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package anytype_crypto

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AnytypeRegistrarControllerPrivateMetaData contains all meta data concerning the AnytypeRegistrarControllerPrivate contract.
var AnytypeRegistrarControllerPrivateMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractAnytypeRegistrarImplementation\",\"name\":\"_base\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_minCommitmentAge\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_maxCommitmentAge\",\"type\":\"uint256\"},{\"internalType\":\"contractReverseRegistrar\",\"name\":\"_reverseRegistrar\",\"type\":\"address\"},{\"internalType\":\"contractINameWrapper\",\"name\":\"_nameWrapper\",\"type\":\"address\"},{\"internalType\":\"contractENS\",\"name\":\"_ens\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"CommitmentTooNew\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"CommitmentTooOld\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"DurationTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MaxCommitmentAgeTooHigh\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MaxCommitmentAgeTooLow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"NameNotAvailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ResolverRequiredWhenDataSupplied\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"UnexpiredCommitmentExists\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"label\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"label\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRenewed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MIN_REGISTRATION_DURATION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"available\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"commit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"commitments\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"},{\"internalType\":\"bool\",\"name\":\"reverseRecord\",\"type\":\"bool\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"makeCommitment\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxCommitmentAge\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minCommitmentAge\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nameWrapper\",\"outputs\":[{\"internalType\":\"contractINameWrapper\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"recoverFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"secret\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"bytes[]\",\"name\":\"data\",\"type\":\"bytes[]\"},{\"internalType\":\"bool\",\"name\":\"reverseRecord\",\"type\":\"bool\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reverseRegistrar\",\"outputs\":[{\"internalType\":\"contractReverseRegistrar\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceID\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"valid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
	Bin: "0x6101206040523480156200001257600080fd5b5060405162001b8838038062001b8883398101604081905262000035916200021e565b80336200004281620001b5565b6040516302571be360e01b81527f91d1777781884d03a6757a803996e38de2a42967fb37eeaca72729271025a9e260048201526000906001600160a01b038416906302571be390602401602060405180830381865afa158015620000aa573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190620000d091906200029b565b604051630f41a04d60e11b81526001600160a01b03848116600483015291925090821690631e83409a906024016020604051808303816000875af11580156200011d573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190620001439190620002c2565b5050505084841162000168576040516307cb550760e31b815260040160405180910390fd5b428411156200018a57604051630b4319e560e21b815260040160405180910390fd5b506001600160a01b0394851660805260a09390935260c091909152821660e0521661010052620002dc565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b03811681146200021b57600080fd5b50565b60008060008060008060c087890312156200023857600080fd5b8651620002458162000205565b8096505060208701519450604087015193506060870151620002678162000205565b60808801519093506200027a8162000205565b60a08801519092506200028d8162000205565b809150509295509295509295565b600060208284031215620002ae57600080fd5b8151620002bb8162000205565b9392505050565b600060208284031215620002d557600080fd5b5051919050565b60805160a05160c05160e0516101005161183d6200034b60003960008181610252015281816105a5015261077701526000818161019e0152610d4401526000818161029f015281816109140152610b6b0152600081816102070152610af40152600061087b015261183d6000f3fe608060405234801561001057600080fd5b506004361061011b5760003560e01c80638d839ffe116100b2578063acf1a84111610081578063ce1e09c011610066578063ce1e09c01461029a578063f14fcbc8146102c1578063f2fde38b146102d457600080fd5b8063acf1a84114610274578063aeb8ce9b1461028757600080fd5b80638d839ffe146102025780638da5cb5b146102295780639791c0971461023a578063a8e5fbc01461024d57600080fd5b806374694a2b116100ee57806374694a2b146101865780638086985314610199578063839df945146101d85780638a95b09f146101f857600080fd5b806301ffc9a7146101205780635d3590d51461014857806365a69dcf1461015d578063715018a61461017e575b600080fd5b61013361012e366004610f85565b6102e7565b60405190151581526020015b60405180910390f35b61015b610156366004610fe3565b610380565b005b61017061016b366004611150565b61041a565b60405190815260200161013f565b61015b6104bf565b61015b610194366004611253565b6104d3565b6101c07f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b03909116815260200161013f565b6101706101e636600461131d565b60016020526000908152604090205481565b6101706224ea0081565b6101707f000000000000000000000000000000000000000000000000000000000000000081565b6000546001600160a01b03166101c0565b610133610248366004611336565b610705565b6101c07f000000000000000000000000000000000000000000000000000000000000000081565b61015b610282366004611373565b61071a565b610133610295366004611336565b610832565b6101707f000000000000000000000000000000000000000000000000000000000000000081565b61015b6102cf36600461131d565b6108f5565b61015b6102e23660046113bf565b61098b565b60007fffffffff0000000000000000000000000000000000000000000000000000000082167f01ffc9a700000000000000000000000000000000000000000000000000000000148061037a57507fffffffff0000000000000000000000000000000000000000000000000000000082167fe2c97af600000000000000000000000000000000000000000000000000000000145b92915050565b610388610a1b565b6040517fa9059cbb0000000000000000000000000000000000000000000000000000000081526001600160a01b0383811660048301526024820183905284169063a9059cbb906044016020604051808303816000875af11580156103f0573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061041491906113da565b50505050565b6000610424610a1b565b895160208b0120841580159061044157506001600160a01b038716155b15610478576040517fd3f605c400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b808a8a8a8a8a8a8a8a604051602001610499999897969594939291906114b2565b604051602081830303815290604052805190602001209150509998505050505050505050565b6104c7610a1b565b6104d16000610a75565b565b6104db610a1b565b6105728a8a8080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050508861056d8d8d8080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508f92508e91508d90508c8c8c8c8c61041a565b610add565b6040517fa40149820000000000000000000000000000000000000000000000000000000081526000906001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000169063a4014982906105e4908e908e908e908e908d908a90600401611514565b6020604051808303816000875af1158015610603573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610627919061155e565b9050831561065257610652868c8c604051610643929190611577565b60405180910390208787610c5f565b821561069b5761069b8b8b8080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508a9250339150610d429050565b886001600160a01b03168b8b6040516106b5929190611577565b60405180910390207f0667086d08417333ce63f40d5bc2ef6fd330e25aaaf317b7c489541f8fe600fa8d8d856040516106f093929190611587565b60405180910390a35050505050505050505050565b6000600361071283610df6565b101592915050565b610722610a1b565b60008383604051610734929190611577565b6040519081900381207fc475abff0000000000000000000000000000000000000000000000000000000082526004820181905260248201849052915081906000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063c475abff906044016020604051808303816000875af11580156107c8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107ec919061155e565b9050827f93bc1a84707231b1d9552157299797c64a1a8c5bc79f05153716630c9c4936fc87878460405161082293929190611587565b60405180910390a2505050505050565b8051602082012060009061084583610705565b80156108ee57506040517f96e494e8000000000000000000000000000000000000000000000000000000008152600481018290527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906396e494e890602401602060405180830381865afa1580156108ca573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108ee91906113da565b9392505050565b6108fd610a1b565b6000818152600160205260409020544290610939907f0000000000000000000000000000000000000000000000000000000000000000906115c1565b10610978576040517f0a059d71000000000000000000000000000000000000000000000000000000008152600481018290526024015b60405180910390fd5b6000908152600160205260409020429055565b610993610a1b565b6001600160a01b038116610a0f5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f6464726573730000000000000000000000000000000000000000000000000000606482015260840161096f565b610a1881610a75565b50565b6000546001600160a01b031633146104d15760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161096f565b600080546001600160a01b038381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000818152600160205260409020544290610b19907f0000000000000000000000000000000000000000000000000000000000000000906115c1565b1115610b54576040517f5320bcf90000000000000000000000000000000000000000000000000000000081526004810182905260240161096f565b6000818152600160205260409020544290610b90907f0000000000000000000000000000000000000000000000000000000000000000906115c1565b11610bca576040517fcb7690d70000000000000000000000000000000000000000000000000000000081526004810182905260240161096f565b610bd383610832565b610c0b57826040517f477707e800000000000000000000000000000000000000000000000000000000815260040161096f9190611624565b6000818152600160205260408120556224ea00821015610c5a576040517f9a71997b0000000000000000000000000000000000000000000000000000000081526004810183905260240161096f565b505050565b604080517fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a463602080830191909152818301869052825180830384018152606083019384905280519101207fe32954eb0000000000000000000000000000000000000000000000000000000090925285906001600160a01b0382169063e32954eb90610cf290859088908890606401611637565b6000604051808303816000875af1158015610d11573d6000803e3d6000fd5b505050506040513d6000823e601f3d908101601f19168201604052610d39919081019061165a565b50505050505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316637a806d6b33838587604051602001610d859190611759565b6040516020818303038152906040526040518563ffffffff1660e01b8152600401610db3949392919061179a565b6020604051808303816000875af1158015610dd2573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610414919061155e565b8051600090819081905b80821015610f7c576000858381518110610e1c57610e1c6117d8565b01602001516001600160f81b03191690507f8000000000000000000000000000000000000000000000000000000000000000811015610e6757610e606001846115c1565b9250610f69565b7fe0000000000000000000000000000000000000000000000000000000000000006001600160f81b031982161015610ea457610e606002846115c1565b7ff0000000000000000000000000000000000000000000000000000000000000006001600160f81b031982161015610ee157610e606003846115c1565b7ff8000000000000000000000000000000000000000000000000000000000000006001600160f81b031982161015610f1e57610e606004846115c1565b7ffc000000000000000000000000000000000000000000000000000000000000006001600160f81b031982161015610f5b57610e606005846115c1565b610f666006846115c1565b92505b5082610f74816117ee565b935050610e00565b50909392505050565b600060208284031215610f9757600080fd5b81357fffffffff00000000000000000000000000000000000000000000000000000000811681146108ee57600080fd5b80356001600160a01b0381168114610fde57600080fd5b919050565b600080600060608486031215610ff857600080fd5b61100184610fc7565b925061100f60208501610fc7565b9150604084013590509250925092565b634e487b7160e01b600052604160045260246000fd5b604051601f8201601f1916810167ffffffffffffffff8111828210171561105e5761105e61101f565b604052919050565b600067ffffffffffffffff8211156110805761108061101f565b50601f01601f191660200190565b600082601f83011261109f57600080fd5b81356110b26110ad82611066565b611035565b8181528460208386010111156110c757600080fd5b816020850160208301376000918101602001919091529392505050565b60008083601f8401126110f657600080fd5b50813567ffffffffffffffff81111561110e57600080fd5b6020830191508360208260051b850101111561112957600080fd5b9250929050565b8015158114610a1857600080fd5b803561ffff81168114610fde57600080fd5b60008060008060008060008060006101008a8c03121561116f57600080fd5b893567ffffffffffffffff8082111561118757600080fd5b6111938d838e0161108e565b9a506111a160208d01610fc7565b995060408c0135985060608c013597506111bd60808d01610fc7565b965060a08c01359150808211156111d357600080fd5b506111e08c828d016110e4565b90955093505060c08a01356111f481611130565b915061120260e08b0161113e565b90509295985092959850929598565b60008083601f84011261122357600080fd5b50813567ffffffffffffffff81111561123b57600080fd5b60208301915083602082850101111561112957600080fd5b6000806000806000806000806000806101008b8d03121561127357600080fd5b8a3567ffffffffffffffff8082111561128b57600080fd5b6112978e838f01611211565b909c509a508a91506112ab60208e01610fc7565b995060408d0135985060608d013597506112c760808e01610fc7565b965060a08d01359150808211156112dd57600080fd5b506112ea8d828e016110e4565b90955093505060c08b01356112fe81611130565b915061130c60e08c0161113e565b90509295989b9194979a5092959850565b60006020828403121561132f57600080fd5b5035919050565b60006020828403121561134857600080fd5b813567ffffffffffffffff81111561135f57600080fd5b61136b8482850161108e565b949350505050565b60008060006040848603121561138857600080fd5b833567ffffffffffffffff81111561139f57600080fd5b6113ab86828701611211565b909790965060209590950135949350505050565b6000602082840312156113d157600080fd5b6108ee82610fc7565b6000602082840312156113ec57600080fd5b81516108ee81611130565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b81835260006020808501808196508560051b810191508460005b878110156114a55782840389528135601e1988360301811261145b57600080fd5b8701858101903567ffffffffffffffff81111561147757600080fd5b80360382131561148657600080fd5b6114918682846113f7565b9a87019a955050509084019060010161143a565b5091979650505050505050565b60006101008b83526001600160a01b03808c1660208501528a60408501528960608501528089166080850152508060a08401526114f28184018789611420565b94151560c0840152505061ffff9190911660e090910152979650505050505050565b60a08152600061152860a08301888a6113f7565b90506001600160a01b03808716602084015285604084015280851660608401525061ffff83166080830152979650505050505050565b60006020828403121561157057600080fd5b5051919050565b8183823760009101908152919050565b60408152600061159b6040830185876113f7565b9050826020830152949350505050565b634e487b7160e01b600052601160045260246000fd5b8082018082111561037a5761037a6115ab565b60005b838110156115ef5781810151838201526020016115d7565b50506000910152565b600081518084526116108160208601602086016115d4565b601f01601f19169290920160200192915050565b6020815260006108ee60208301846115f8565b838152604060208201526000611651604083018486611420565b95945050505050565b6000602080838503121561166d57600080fd5b825167ffffffffffffffff8082111561168557600080fd5b818501915085601f83011261169957600080fd5b8151818111156116ab576116ab61101f565b8060051b6116ba858201611035565b91825283810185019185810190898411156116d457600080fd5b86860192505b8383101561174c578251858111156116f25760008081fd5b8601603f81018b136117045760008081fd5b8781015160406117166110ad83611066565b8281528d8284860101111561172b5760008081fd5b61173a838c83018487016115d4565b855250505091860191908601906116da565b9998505050505050505050565b6000825161176b8184602087016115d4565b7f2e616e7900000000000000000000000000000000000000000000000000000000920191825250600401919050565b60006001600160a01b0380871683528086166020840152808516604084015250608060608301526117ce60808301846115f8565b9695505050505050565b634e487b7160e01b600052603260045260246000fd5b600060018201611800576118006115ab565b506001019056fea2646970667358221220f827500f26a4d4f318ae03b2bca0b85fb9e27d69f64a60040afa09a5ba351bad64736f6c63430008110033",
}

// AnytypeRegistrarControllerPrivateABI is the input ABI used to generate the binding from.
// Deprecated: Use AnytypeRegistrarControllerPrivateMetaData.ABI instead.
var AnytypeRegistrarControllerPrivateABI = AnytypeRegistrarControllerPrivateMetaData.ABI

// AnytypeRegistrarControllerPrivateBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AnytypeRegistrarControllerPrivateMetaData.Bin instead.
var AnytypeRegistrarControllerPrivateBin = AnytypeRegistrarControllerPrivateMetaData.Bin

// DeployAnytypeRegistrarControllerPrivate deploys a new Ethereum contract, binding an instance of AnytypeRegistrarControllerPrivate to it.
func DeployAnytypeRegistrarControllerPrivate(auth *bind.TransactOpts, backend bind.ContractBackend, _base common.Address, _minCommitmentAge *big.Int, _maxCommitmentAge *big.Int, _reverseRegistrar common.Address, _nameWrapper common.Address, _ens common.Address) (common.Address, *types.Transaction, *AnytypeRegistrarControllerPrivate, error) {
	parsed, err := AnytypeRegistrarControllerPrivateMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AnytypeRegistrarControllerPrivateBin), backend, _base, _minCommitmentAge, _maxCommitmentAge, _reverseRegistrar, _nameWrapper, _ens)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AnytypeRegistrarControllerPrivate{AnytypeRegistrarControllerPrivateCaller: AnytypeRegistrarControllerPrivateCaller{contract: contract}, AnytypeRegistrarControllerPrivateTransactor: AnytypeRegistrarControllerPrivateTransactor{contract: contract}, AnytypeRegistrarControllerPrivateFilterer: AnytypeRegistrarControllerPrivateFilterer{contract: contract}}, nil
}

// AnytypeRegistrarControllerPrivate is an auto generated Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivate struct {
	AnytypeRegistrarControllerPrivateCaller     // Read-only binding to the contract
	AnytypeRegistrarControllerPrivateTransactor // Write-only binding to the contract
	AnytypeRegistrarControllerPrivateFilterer   // Log filterer for contract events
}

// AnytypeRegistrarControllerPrivateCaller is an auto generated read-only Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivateCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarControllerPrivateTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivateTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarControllerPrivateFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AnytypeRegistrarControllerPrivateFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarControllerPrivateSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AnytypeRegistrarControllerPrivateSession struct {
	Contract     *AnytypeRegistrarControllerPrivate // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                      // Call options to use throughout this session
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// AnytypeRegistrarControllerPrivateCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AnytypeRegistrarControllerPrivateCallerSession struct {
	Contract *AnytypeRegistrarControllerPrivateCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                            // Call options to use throughout this session
}

// AnytypeRegistrarControllerPrivateTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AnytypeRegistrarControllerPrivateTransactorSession struct {
	Contract     *AnytypeRegistrarControllerPrivateTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                            // Transaction auth options to use throughout this session
}

// AnytypeRegistrarControllerPrivateRaw is an auto generated low-level Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivateRaw struct {
	Contract *AnytypeRegistrarControllerPrivate // Generic contract binding to access the raw methods on
}

// AnytypeRegistrarControllerPrivateCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivateCallerRaw struct {
	Contract *AnytypeRegistrarControllerPrivateCaller // Generic read-only contract binding to access the raw methods on
}

// AnytypeRegistrarControllerPrivateTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AnytypeRegistrarControllerPrivateTransactorRaw struct {
	Contract *AnytypeRegistrarControllerPrivateTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAnytypeRegistrarControllerPrivate creates a new instance of AnytypeRegistrarControllerPrivate, bound to a specific deployed contract.
func NewAnytypeRegistrarControllerPrivate(address common.Address, backend bind.ContractBackend) (*AnytypeRegistrarControllerPrivate, error) {
	contract, err := bindAnytypeRegistrarControllerPrivate(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivate{AnytypeRegistrarControllerPrivateCaller: AnytypeRegistrarControllerPrivateCaller{contract: contract}, AnytypeRegistrarControllerPrivateTransactor: AnytypeRegistrarControllerPrivateTransactor{contract: contract}, AnytypeRegistrarControllerPrivateFilterer: AnytypeRegistrarControllerPrivateFilterer{contract: contract}}, nil
}

// NewAnytypeRegistrarControllerPrivateCaller creates a new read-only instance of AnytypeRegistrarControllerPrivate, bound to a specific deployed contract.
func NewAnytypeRegistrarControllerPrivateCaller(address common.Address, caller bind.ContractCaller) (*AnytypeRegistrarControllerPrivateCaller, error) {
	contract, err := bindAnytypeRegistrarControllerPrivate(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateCaller{contract: contract}, nil
}

// NewAnytypeRegistrarControllerPrivateTransactor creates a new write-only instance of AnytypeRegistrarControllerPrivate, bound to a specific deployed contract.
func NewAnytypeRegistrarControllerPrivateTransactor(address common.Address, transactor bind.ContractTransactor) (*AnytypeRegistrarControllerPrivateTransactor, error) {
	contract, err := bindAnytypeRegistrarControllerPrivate(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateTransactor{contract: contract}, nil
}

// NewAnytypeRegistrarControllerPrivateFilterer creates a new log filterer instance of AnytypeRegistrarControllerPrivate, bound to a specific deployed contract.
func NewAnytypeRegistrarControllerPrivateFilterer(address common.Address, filterer bind.ContractFilterer) (*AnytypeRegistrarControllerPrivateFilterer, error) {
	contract, err := bindAnytypeRegistrarControllerPrivate(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateFilterer{contract: contract}, nil
}

// bindAnytypeRegistrarControllerPrivate binds a generic wrapper to an already deployed contract.
func bindAnytypeRegistrarControllerPrivate(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AnytypeRegistrarControllerPrivateABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeRegistrarControllerPrivate.Contract.AnytypeRegistrarControllerPrivateCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.AnytypeRegistrarControllerPrivateTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.AnytypeRegistrarControllerPrivateTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeRegistrarControllerPrivate.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.contract.Transact(opts, method, params...)
}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) MINREGISTRATIONDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "MIN_REGISTRATION_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) MINREGISTRATIONDURATION() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MINREGISTRATIONDURATION(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// MINREGISTRATIONDURATION is a free data retrieval call binding the contract method 0x8a95b09f.
//
// Solidity: function MIN_REGISTRATION_DURATION() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) MINREGISTRATIONDURATION() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MINREGISTRATIONDURATION(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) Available(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "available", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Available(name string) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Available(&_AnytypeRegistrarControllerPrivate.CallOpts, name)
}

// Available is a free data retrieval call binding the contract method 0xaeb8ce9b.
//
// Solidity: function available(string name) view returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) Available(name string) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Available(&_AnytypeRegistrarControllerPrivate.CallOpts, name)
}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) Commitments(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "commitments", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Commitments(arg0 [32]byte) (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Commitments(&_AnytypeRegistrarControllerPrivate.CallOpts, arg0)
}

// Commitments is a free data retrieval call binding the contract method 0x839df945.
//
// Solidity: function commitments(bytes32 ) view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) Commitments(arg0 [32]byte) (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Commitments(&_AnytypeRegistrarControllerPrivate.CallOpts, arg0)
}

// MakeCommitment is a free data retrieval call binding the contract method 0x65a69dcf.
//
// Solidity: function makeCommitment(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) view returns(bytes32)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) MakeCommitment(opts *bind.CallOpts, name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) ([32]byte, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "makeCommitment", name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MakeCommitment is a free data retrieval call binding the contract method 0x65a69dcf.
//
// Solidity: function makeCommitment(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) view returns(bytes32)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) MakeCommitment(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) ([32]byte, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MakeCommitment(&_AnytypeRegistrarControllerPrivate.CallOpts, name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)
}

// MakeCommitment is a free data retrieval call binding the contract method 0x65a69dcf.
//
// Solidity: function makeCommitment(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) view returns(bytes32)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) MakeCommitment(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) ([32]byte, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MakeCommitment(&_AnytypeRegistrarControllerPrivate.CallOpts, name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)
}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) MaxCommitmentAge(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "maxCommitmentAge")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) MaxCommitmentAge() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MaxCommitmentAge(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// MaxCommitmentAge is a free data retrieval call binding the contract method 0xce1e09c0.
//
// Solidity: function maxCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) MaxCommitmentAge() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MaxCommitmentAge(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) MinCommitmentAge(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "minCommitmentAge")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) MinCommitmentAge() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MinCommitmentAge(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// MinCommitmentAge is a free data retrieval call binding the contract method 0x8d839ffe.
//
// Solidity: function minCommitmentAge() view returns(uint256)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) MinCommitmentAge() (*big.Int, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.MinCommitmentAge(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// NameWrapper is a free data retrieval call binding the contract method 0xa8e5fbc0.
//
// Solidity: function nameWrapper() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) NameWrapper(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "nameWrapper")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NameWrapper is a free data retrieval call binding the contract method 0xa8e5fbc0.
//
// Solidity: function nameWrapper() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) NameWrapper() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.NameWrapper(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// NameWrapper is a free data retrieval call binding the contract method 0xa8e5fbc0.
//
// Solidity: function nameWrapper() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) NameWrapper() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.NameWrapper(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Owner() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Owner(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) Owner() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Owner(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// ReverseRegistrar is a free data retrieval call binding the contract method 0x80869853.
//
// Solidity: function reverseRegistrar() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) ReverseRegistrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "reverseRegistrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ReverseRegistrar is a free data retrieval call binding the contract method 0x80869853.
//
// Solidity: function reverseRegistrar() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) ReverseRegistrar() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.ReverseRegistrar(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// ReverseRegistrar is a free data retrieval call binding the contract method 0x80869853.
//
// Solidity: function reverseRegistrar() view returns(address)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) ReverseRegistrar() (common.Address, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.ReverseRegistrar(&_AnytypeRegistrarControllerPrivate.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) SupportsInterface(opts *bind.CallOpts, interfaceID [4]byte) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "supportsInterface", interfaceID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.SupportsInterface(&_AnytypeRegistrarControllerPrivate.CallOpts, interfaceID)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.SupportsInterface(&_AnytypeRegistrarControllerPrivate.CallOpts, interfaceID)
}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCaller) Valid(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarControllerPrivate.contract.Call(opts, &out, "valid", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Valid(name string) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Valid(&_AnytypeRegistrarControllerPrivate.CallOpts, name)
}

// Valid is a free data retrieval call binding the contract method 0x9791c097.
//
// Solidity: function valid(string name) pure returns(bool)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateCallerSession) Valid(name string) (bool, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Valid(&_AnytypeRegistrarControllerPrivate.CallOpts, name)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) Commit(opts *bind.TransactOpts, commitment [32]byte) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "commit", commitment)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Commit(commitment [32]byte) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Commit(&_AnytypeRegistrarControllerPrivate.TransactOpts, commitment)
}

// Commit is a paid mutator transaction binding the contract method 0xf14fcbc8.
//
// Solidity: function commit(bytes32 commitment) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) Commit(commitment [32]byte) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Commit(&_AnytypeRegistrarControllerPrivate.TransactOpts, commitment)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) RecoverFunds(opts *bind.TransactOpts, _token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "recoverFunds", _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.RecoverFunds(&_AnytypeRegistrarControllerPrivate.TransactOpts, _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.RecoverFunds(&_AnytypeRegistrarControllerPrivate.TransactOpts, _token, _to, _amount)
}

// Register is a paid mutator transaction binding the contract method 0x74694a2b.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) Register(opts *bind.TransactOpts, name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "register", name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)
}

// Register is a paid mutator transaction binding the contract method 0x74694a2b.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Register(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Register(&_AnytypeRegistrarControllerPrivate.TransactOpts, name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)
}

// Register is a paid mutator transaction binding the contract method 0x74694a2b.
//
// Solidity: function register(string name, address owner, uint256 duration, bytes32 secret, address resolver, bytes[] data, bool reverseRecord, uint16 ownerControlledFuses) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) Register(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address, data [][]byte, reverseRecord bool, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Register(&_AnytypeRegistrarControllerPrivate.TransactOpts, name, owner, duration, secret, resolver, data, reverseRecord, ownerControlledFuses)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) Renew(opts *bind.TransactOpts, name string, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "renew", name, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) Renew(name string, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Renew(&_AnytypeRegistrarControllerPrivate.TransactOpts, name, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xacf1a841.
//
// Solidity: function renew(string name, uint256 duration) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) Renew(name string, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.Renew(&_AnytypeRegistrarControllerPrivate.TransactOpts, name, duration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.RenounceOwnership(&_AnytypeRegistrarControllerPrivate.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.RenounceOwnership(&_AnytypeRegistrarControllerPrivate.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.TransferOwnership(&_AnytypeRegistrarControllerPrivate.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarControllerPrivate.Contract.TransferOwnership(&_AnytypeRegistrarControllerPrivate.TransactOpts, newOwner)
}

// AnytypeRegistrarControllerPrivateNameRegisteredIterator is returned from FilterNameRegistered and is used to iterate over the raw logs and unpacked data for NameRegistered events raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateNameRegisteredIterator struct {
	Event *AnytypeRegistrarControllerPrivateNameRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnytypeRegistrarControllerPrivateNameRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarControllerPrivateNameRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnytypeRegistrarControllerPrivateNameRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnytypeRegistrarControllerPrivateNameRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarControllerPrivateNameRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarControllerPrivateNameRegistered represents a NameRegistered event raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateNameRegistered struct {
	Name    string
	Label   [32]byte
	Owner   common.Address
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRegistered is a free log retrieval operation binding the contract event 0x0667086d08417333ce63f40d5bc2ef6fd330e25aaaf317b7c489541f8fe600fa.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) FilterNameRegistered(opts *bind.FilterOpts, label [][32]byte, owner []common.Address) (*AnytypeRegistrarControllerPrivateNameRegisteredIterator, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.FilterLogs(opts, "NameRegistered", labelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateNameRegisteredIterator{contract: _AnytypeRegistrarControllerPrivate.contract, event: "NameRegistered", logs: logs, sub: sub}, nil
}

// WatchNameRegistered is a free log subscription operation binding the contract event 0x0667086d08417333ce63f40d5bc2ef6fd330e25aaaf317b7c489541f8fe600fa.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) WatchNameRegistered(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarControllerPrivateNameRegistered, label [][32]byte, owner []common.Address) (event.Subscription, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.WatchLogs(opts, "NameRegistered", labelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarControllerPrivateNameRegistered)
				if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "NameRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNameRegistered is a log parse operation binding the contract event 0x0667086d08417333ce63f40d5bc2ef6fd330e25aaaf317b7c489541f8fe600fa.
//
// Solidity: event NameRegistered(string name, bytes32 indexed label, address indexed owner, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) ParseNameRegistered(log types.Log) (*AnytypeRegistrarControllerPrivateNameRegistered, error) {
	event := new(AnytypeRegistrarControllerPrivateNameRegistered)
	if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "NameRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarControllerPrivateNameRenewedIterator is returned from FilterNameRenewed and is used to iterate over the raw logs and unpacked data for NameRenewed events raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateNameRenewedIterator struct {
	Event *AnytypeRegistrarControllerPrivateNameRenewed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnytypeRegistrarControllerPrivateNameRenewedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarControllerPrivateNameRenewed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnytypeRegistrarControllerPrivateNameRenewed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnytypeRegistrarControllerPrivateNameRenewedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarControllerPrivateNameRenewedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarControllerPrivateNameRenewed represents a NameRenewed event raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateNameRenewed struct {
	Name    string
	Label   [32]byte
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRenewed is a free log retrieval operation binding the contract event 0x93bc1a84707231b1d9552157299797c64a1a8c5bc79f05153716630c9c4936fc.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) FilterNameRenewed(opts *bind.FilterOpts, label [][32]byte) (*AnytypeRegistrarControllerPrivateNameRenewedIterator, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.FilterLogs(opts, "NameRenewed", labelRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateNameRenewedIterator{contract: _AnytypeRegistrarControllerPrivate.contract, event: "NameRenewed", logs: logs, sub: sub}, nil
}

// WatchNameRenewed is a free log subscription operation binding the contract event 0x93bc1a84707231b1d9552157299797c64a1a8c5bc79f05153716630c9c4936fc.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) WatchNameRenewed(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarControllerPrivateNameRenewed, label [][32]byte) (event.Subscription, error) {

	var labelRule []interface{}
	for _, labelItem := range label {
		labelRule = append(labelRule, labelItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.WatchLogs(opts, "NameRenewed", labelRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarControllerPrivateNameRenewed)
				if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "NameRenewed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNameRenewed is a log parse operation binding the contract event 0x93bc1a84707231b1d9552157299797c64a1a8c5bc79f05153716630c9c4936fc.
//
// Solidity: event NameRenewed(string name, bytes32 indexed label, uint256 expires)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) ParseNameRenewed(log types.Log) (*AnytypeRegistrarControllerPrivateNameRenewed, error) {
	event := new(AnytypeRegistrarControllerPrivateNameRenewed)
	if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "NameRenewed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarControllerPrivateOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateOwnershipTransferredIterator struct {
	Event *AnytypeRegistrarControllerPrivateOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnytypeRegistrarControllerPrivateOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarControllerPrivateOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnytypeRegistrarControllerPrivateOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnytypeRegistrarControllerPrivateOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarControllerPrivateOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarControllerPrivateOwnershipTransferred represents a OwnershipTransferred event raised by the AnytypeRegistrarControllerPrivate contract.
type AnytypeRegistrarControllerPrivateOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AnytypeRegistrarControllerPrivateOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarControllerPrivateOwnershipTransferredIterator{contract: _AnytypeRegistrarControllerPrivate.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarControllerPrivateOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeRegistrarControllerPrivate.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarControllerPrivateOwnershipTransferred)
				if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeRegistrarControllerPrivate *AnytypeRegistrarControllerPrivateFilterer) ParseOwnershipTransferred(log types.Log) (*AnytypeRegistrarControllerPrivateOwnershipTransferred, error) {
	event := new(AnytypeRegistrarControllerPrivateOwnershipTransferred)
	if err := _AnytypeRegistrarControllerPrivate.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
