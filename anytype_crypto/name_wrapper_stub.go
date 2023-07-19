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

// AnytypeNameWrapperMetaData contains all meta data concerning the AnytypeNameWrapper contract.
var AnytypeNameWrapperMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"_ens\",\"type\":\"address\"},{\"internalType\":\"contractIBaseRegistrar\",\"name\":\"_registrar\",\"type\":\"address\"},{\"internalType\":\"contractIMetadataService\",\"name\":\"_metadataService\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"CannotUpgrade\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"IncompatibleParent\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"IncorrectTargetOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"IncorrectTokenType\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"labelHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedLabelhash\",\"type\":\"bytes32\"}],\"name\":\"LabelMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"}],\"name\":\"LabelTooLong\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LabelTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NameIsNotWrapped\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"OperationProhibited\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"Unauthorised\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"ControllerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"ExpiryExtended\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"}],\"name\":\"FusesSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"NameUnwrapped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"NameWrapped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"TransferBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferSingle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"URI\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"_tokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"fuseMask\",\"type\":\"uint32\"}],\"name\":\"allFusesBurned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"}],\"name\":\"balanceOfBatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"canExtendSubnames\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"canModifyName\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"controllers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"extendExpiry\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"getData\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"}],\"name\":\"isWrapped\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"isWrapped\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"metadataService\",\"outputs\":[{\"internalType\":\"contractIMetadataService\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"names\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"recoverFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"registerAndWrapETH2LD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"registrarExpiry\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registrar\",\"outputs\":[{\"internalType\":\"contractIBaseRegistrar\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeBatchTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setChildFuses\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"}],\"name\":\"setFuses\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIMetadataService\",\"name\":\"_metadataService\",\"type\":\"address\"}],\"name\":\"setMetadataService\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"setRecord\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"setResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setSubnodeOwner\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"},{\"internalType\":\"uint32\",\"name\":\"fuses\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"name\":\"setSubnodeRecord\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"ttl\",\"type\":\"uint64\"}],\"name\":\"setTTL\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractINameWrapperUpgrade\",\"name\":\"_upgradeAddress\",\"type\":\"address\"}],\"name\":\"setUpgradeContract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"parentNode\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"unwrap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"labelhash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"registrant\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"unwrapETH2LD\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"upgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"upgradeContract\",\"outputs\":[{\"internalType\":\"contractINameWrapperUpgrade\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"name\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"wrap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"wrappedOwner\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"ownerControlledFuses\",\"type\":\"uint16\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"wrapETH2LD\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"expiry\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60c06040523480156200001157600080fd5b5060405162006551380380620065518339810160408190526200003491620002f9565b8233620000418162000290565b6040516302571be360e01b81527f91d1777781884d03a6757a803996e38de2a42967fb37eeaca72729271025a9e260048201526000906001600160a01b038416906302571be390602401602060405180830381865afa158015620000a9573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190620000cf91906200034d565b604051630f41a04d60e11b81526001600160a01b03848116600483015291925090821690631e83409a906024016020604051808303816000875af11580156200011c573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019062000142919062000374565b505050506001600160a01b0383811660805282811660a052600580546001600160a01b031916918316919091179055600163fffeffff60a01b03197fa086141a224b7d4ff781ac7aad74efac04028926bdee6c7262cfb7653a85262a8190557fa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb4955604080518082019091526001815260006020808301829052908052600690527f54cdd369e4e8a8515e52ca72ec816c2101831ad1f18bf44102ed171459c9b4f89062000210908262000433565b5060408051808201909152600581526303616e7960e01b6020808301919091527fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a463600052600690527f9fab986d55bfb563cbe6b418ec514aa31486750d5ceda77474bde124347c3c319062000286908262000433565b50505050620004ff565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6001600160a01b0381168114620002f657600080fd5b50565b6000806000606084860312156200030f57600080fd5b83516200031c81620002e0565b60208501519093506200032f81620002e0565b60408501519092506200034281620002e0565b809150509250925092565b6000602082840312156200036057600080fd5b81516200036d81620002e0565b9392505050565b6000602082840312156200038757600080fd5b5051919050565b634e487b7160e01b600052604160045260246000fd5b600181811c90821680620003b957607f821691505b602082108103620003da57634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200042e57600081815260208120601f850160051c81016020861015620004095750805b601f850160051c820191505b818110156200042a5782815560010162000415565b5050505b505050565b81516001600160401b038111156200044f576200044f6200038e565b6200046781620004608454620003a4565b84620003e0565b602080601f8311600181146200049f5760008415620004865750858301515b600019600386901b1c1916600185901b1785556200042a565b600085815260208120601f198616915b82811015620004d057888601518255948401946001909101908401620004af565b5085821015620004ef5787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b60805160a051615f456200060c6000396000818161050601528181610c1501528181610cef01528181610d7901528181611c7301528181611d0901528181611db701528181611ed901528181611f4f01528181611fcf015281816122510152818161238d015281816124cc015281816126b1015281816127370152612f6c01526000818161055301528181610b9b01528181610ee4015281816110980152818161114a015281816115620152818161241201528181612551015281816127e2015281816129d901528181612ce701528181613197015281816132450152818161330e01528181613387015281816139e401528181613aff01528181613d6701526143c10152615f456000f3fe608060405234801561001057600080fd5b506004361061031f5760003560e01c80636352211e116101a7578063c93ab3fd116100ee578063e985e9c511610097578063f242432a11610071578063f242432a146107d7578063f2fde38b146107ea578063fd0cd0d9146107fd57600080fd5b8063e985e9c514610768578063eb8ae530146107a4578063ed70554d146107b757600080fd5b8063d9a50c12116100c8578063d9a50c121461071f578063da8c229e14610732578063e0dba60f1461075557600080fd5b8063c93ab3fd146106e6578063cf408823146106f9578063d8c9921a1461070c57600080fd5b8063a22cb46511610150578063b6bcad261161012a578063b6bcad26146106ad578063c475abff146106c0578063c658e086146106d357600080fd5b8063a22cb46514610674578063a401498214610687578063adf4960a1461069a57600080fd5b80638b4dfa75116101815780638b4dfa751461063d5780638cf8b41e146106505780638da5cb5b1461066357600080fd5b80636352211e146105f65780636e5d6ad214610609578063715018a61461063557600080fd5b80631f4e15041161026b5780633f15457f116102145780634e1273f4116101ee5780634e1273f4146105b057806353095467146105d05780635d3590d5146105e357600080fd5b80633f15457f1461054e578063402906fc1461057557806341415eab1461059d57600080fd5b80632b20e397116102455780632b20e397146105015780632eb2c2d61461052857806333c69ea91461053b57600080fd5b80631f4e1504146104c857806320c38e2b146104db57806324c1af44146104ee57600080fd5b80630e4cd725116102cd578063150b7a02116102a7578063150b7a02146104765780631534e177146104a25780631896f70a146104b557600080fd5b80630e4cd7251461043d5780630e89341c1461045057806314ab90381461046357600080fd5b806306fdde03116102fe57806306fdde03146103b4578063081812fc146103fd578063095ea7b31461042857600080fd5b8062fdd58e146103245780630178fe3f1461034a57806301ffc9a714610391575b600080fd5b610337610332366004614d74565b610810565b6040519081526020015b60405180910390f35b61035d610358366004614da0565b6108cf565b604080516001600160a01b03909416845263ffffffff909216602084015267ffffffffffffffff1690820152606001610341565b6103a461039f366004614dcf565b6108ff565b6040519015158152602001610341565b6103f06040518060400160405280601281526020017f416e79747970654e616d6557726170706572000000000000000000000000000081525081565b6040516103419190614e3c565b61041061040b366004614da0565b610958565b6040516001600160a01b039091168152602001610341565b61043b610436366004614d74565b61099d565b005b6103a461044b366004614e4f565b6109e3565b6103f061045e366004614da0565b610a7d565b61043b610471366004614e9c565b610aef565b610489610484366004614f11565b610c08565b6040516001600160e01b03199091168152602001610341565b61043b6104b0366004614f84565b610e1a565b61043b6104c3366004614e4f565b610e51565b600754610410906001600160a01b031681565b6103f06104e9366004614da0565b610f13565b6103376104fc36600461507c565b610fad565b6104107f000000000000000000000000000000000000000000000000000000000000000081565b61043b6105363660046151a4565b6111c1565b61043b610549366004615252565b6114eb565b6104107f000000000000000000000000000000000000000000000000000000000000000081565b6105886105833660046152aa565b6116e0565b60405163ffffffff9091168152602001610341565b6103a46105ab366004614e4f565b611782565b6105c36105be3660046152cd565b6117df565b60405161034191906153cb565b600554610410906001600160a01b031681565b61043b6105f13660046153de565b61191d565b610410610604366004614da0565b6119b7565b61061c61061736600461541f565b6119c2565b60405167ffffffffffffffff9091168152602001610341565b61043b611b17565b61043b61064b366004615454565b611b2b565b61061c61065e366004615496565b611cd5565b6000546001600160a01b0316610410565b61043b61068236600461551f565b6120a1565b61033761069536600461554d565b61218b565b6103a46106a83660046155ce565b612326565b61043b6106bb366004614f84565b61234b565b6103376106ce3660046155f1565b6125b0565b6103376106e1366004615613565b6128a7565b61043b6106f4366004615686565b612ab4565b61043b6107073660046156f2565b612c25565b61043b61071a36600461572a565b612dde565b6103a461072d3660046155f1565b612eee565b6103a4610740366004614f84565b60046020526000908152604090205460ff1681565b61043b61076336600461551f565b612ffb565b6103a4610776366004615758565b6001600160a01b03918216600090815260026020908152604080832093909416825291909152205460ff1690565b61043b6107b2366004615786565b613063565b6103376107c5366004614da0565b60016020526000908152604090205481565b61043b6107e53660046157ee565b61342e565b61043b6107f8366004614f84565b61354b565b6103a461080b366004614da0565b6135d8565b60006001600160a01b0383166108935760405162461bcd60e51b815260206004820152602b60248201527f455243313135353a2062616c616e636520717565727920666f7220746865207a60448201527f65726f206164647265737300000000000000000000000000000000000000000060648201526084015b60405180910390fd5b600061089e836119b7565b9050836001600160a01b0316816001600160a01b0316036108c35760019150506108c9565b60009150505b92915050565b60008181526001602052604090205460a081901c60c082901c6108f38383836136b0565b90959094509092509050565b60006001600160e01b031982167fd82c42d800000000000000000000000000000000000000000000000000000000148061094957506001600160e01b03198216630a85bd0160e11b145b806108c957506108c9826136e7565b600080610964836119b7565b90506001600160a01b03811661097d5750600092915050565b6000838152600360205260409020546001600160a01b03165b9392505050565b60006109a8826108cf565b50915050603f1960408216016109d45760405163a2a7201360e01b81526004810183905260240161088a565b6109de8383613769565b505050565b60008080806109f1866108cf565b925092509250846001600160a01b0316836001600160a01b03161480610a3c57506001600160a01b0380841660009081526002602090815260408083209389168352929052205460ff165b80610a6057506001600160a01b038516610a5587610958565b6001600160a01b0316145b8015610a735750610a7182826138b3565b155b9695505050505050565b6005546040516303a24d0760e21b8152600481018390526060916001600160a01b031690630e89341c90602401600060405180830381865afa158015610ac7573d6000803e3d6000fd5b505050506040513d6000823e601f3d908101601f191682016040526108c99190810190615857565b81610afa8133611782565b610b205760405163168ab55d60e31b81526004810182905233602482015260440161088a565b8260106000610b2e836108cf565b5091505063ffffffff8282161615610b5c5760405163a2a7201360e01b81526004810184905260240161088a565b6040517f14ab90380000000000000000000000000000000000000000000000000000000081526004810187905267ffffffffffffffff861660248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906314ab9038906044015b600060405180830381600087803b158015610be857600080fd5b505af1158015610bfc573d6000803e3d6000fd5b50505050505050505050565b6000336001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001614610c6c576040517f1931a53800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000808080610c7d868801886158cf565b83516020850120939750919550935091508890808214610cd3576040517fc65c3ccc000000000000000000000000000000000000000000000000000000008152600481018290526024810183905260440161088a565b604051630a3b53db60e21b8152600481018390523060248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906328ed4f6c90604401600060405180830381600087803b158015610d3b57600080fd5b505af1158015610d4f573d6000803e3d6000fd5b5050604051636b727d4360e11b8152600481018d9052600092506276a70091506001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000169063d6e4fa8690602401602060405180830381865afa158015610dc0573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610de49190615937565b610dee9190615966565b9050610e0187878761ffff1684886138e4565b50630a85bd0160e11b9c9b505050505050505050505050565b610e22613a4a565b6005805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0392909216919091179055565b81610e5c8133611782565b610e825760405163168ab55d60e31b81526004810182905233602482015260440161088a565b8260086000610e90836108cf565b5091505063ffffffff8282161615610ebe5760405163a2a7201360e01b81526004810184905260240161088a565b604051630c4b7b8560e11b8152600481018790526001600160a01b0386811660248301527f00000000000000000000000000000000000000000000000000000000000000001690631896f70a90604401610bce565b60066020526000908152604090208054610f2c9061598e565b80601f0160208091040260200160405190810160405280929190818152602001828054610f589061598e565b8015610fa55780601f10610f7a57610100808354040283529160200191610fa5565b820191906000526020600020905b815481529060010190602001808311610f8857829003601f168201915b505050505081565b600087610fba8133611782565b610fe05760405163168ab55d60e31b81526004810182905233602482015260440161088a565b875160208901206110188a82604080516020808201949094528082019290925280518083038201815260609092019052805191012090565b92506110248a84613aa4565b61102e8386613be3565b6110398a848b613c16565b506110468a848787613ce3565b935061105183613d29565b611107576040516305ef2c7f60e41b8152600481018b9052602481018290523060448201526001600160a01b03888116606483015267ffffffffffffffff881660848301527f00000000000000000000000000000000000000000000000000000000000000001690635ef2c7f09060a401600060405180830381600087803b1580156110dc57600080fd5b505af11580156110f0573d6000803e3d6000fd5b505050506111028a848b8b8989613de2565b6111b4565b6040516305ef2c7f60e41b8152600481018b9052602481018290523060448201526001600160a01b03888116606483015267ffffffffffffffff881660848301527f00000000000000000000000000000000000000000000000000000000000000001690635ef2c7f09060a401600060405180830381600087803b15801561118e57600080fd5b505af11580156111a2573d6000803e3d6000fd5b505050506111b48a848b8b8989613e19565b5050979650505050505050565b81518351146112385760405162461bcd60e51b815260206004820152602860248201527f455243313135353a2069647320616e6420616d6f756e7473206c656e6774682060448201527f6d69736d61746368000000000000000000000000000000000000000000000000606482015260840161088a565b6001600160a01b03841661129c5760405162461bcd60e51b815260206004820152602560248201527f455243313135353a207472616e7366657220746f20746865207a65726f206164604482015264647265737360d81b606482015260840161088a565b6001600160a01b0385163314806112d657506001600160a01b038516600090815260026020908152604080832033845290915290205460ff165b6113485760405162461bcd60e51b815260206004820152603260248201527f455243313135353a207472616e736665722063616c6c6572206973206e6f742060448201527f6f776e6572206e6f7220617070726f7665640000000000000000000000000000606482015260840161088a565b60005b835181101561147e576000848281518110611368576113686159c8565b602002602001015190506000848381518110611386576113866159c8565b60200260200101519050600080600061139e856108cf565b9250925092506113af858383613edd565b8360011480156113d057508a6001600160a01b0316836001600160a01b0316145b61142f5760405162461bcd60e51b815260206004820152602a60248201527f455243313135353a20696e73756666696369656e742062616c616e636520666f60448201526939103a3930b739b332b960b11b606482015260840161088a565b60008581526001602052604090206001600160a01b038b1663ffffffff60a01b60a085901b16176001600160c01b031960c084901b16179055505050505080611477906159de565b905061134b565b50836001600160a01b0316856001600160a01b0316336001600160a01b03167f4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb86866040516114ce9291906159f7565b60405180910390a46114e4338686868686613fd7565b5050505050565b6040805160208082018790528183018690528251808303840181526060909201909252805191012061151d8184613be3565b6000808061152a846108cf565b919450925090506001600160a01b03831615806115d957506040516302571be360e01b81526004810185905230906001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016906302571be390602401602060405180830381865afa1580156115a9573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906115cd9190615a25565b6001600160a01b031614155b156115f757604051635374b59960e01b815260040160405180910390fd5b6000806116038a6108cf565b90935091508a9050611644576116198633611782565b61163f5760405163168ab55d60e31b81526004810187905233602482015260440161088a565b611674565b61164e8a33611782565b6116745760405163168ab55d60e31b8152600481018b905233602482015260440161088a565b61167f86898461417c565b61168a8784836141b7565b9650620100008416158015906116ae57508363ffffffff1688851763ffffffff1614155b156116cf5760405163a2a7201360e01b81526004810187905260240161088a565b96831796610bfc86868a868b614201565b6000826116ed8133611782565b6117135760405163168ab55d60e31b81526004810182905233602482015260440161088a565b8360026000611721836108cf565b5091505063ffffffff828216161561174f5760405163a2a7201360e01b81526004810184905260240161088a565b6000808061175c8a6108cf565b9250925092506117758a84848c61ffff16178485614201565b5098975050505050505050565b6000808080611790866108cf565b925092509250846001600160a01b0316836001600160a01b03161480610a6057506001600160a01b0380841660009081526002602090815260408083209389168352929052205460ff16610a60565b606081518351146118585760405162461bcd60e51b815260206004820152602960248201527f455243313135353a206163636f756e747320616e6420696473206c656e67746860448201527f206d69736d617463680000000000000000000000000000000000000000000000606482015260840161088a565b6000835167ffffffffffffffff81111561187457611874614fa1565b60405190808252806020026020018201604052801561189d578160200160208202803683370190505b50905060005b8451811015611915576118e88582815181106118c1576118c16159c8565b60200260200101518583815181106118db576118db6159c8565b6020026020010151610810565b8282815181106118fa576118fa6159c8565b602090810291909101015261190e816159de565b90506118a3565b509392505050565b611925613a4a565b6040517fa9059cbb0000000000000000000000000000000000000000000000000000000081526001600160a01b0383811660048301526024820183905284169063a9059cbb906044016020604051808303816000875af115801561198d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906119b19190615a42565b50505050565b60006108c9826142ab565b604080516020808201869052818301859052825180830384018152606090920190925280519101206000906119f681613d29565b611a1357604051635374b59960e01b815260040160405180910390fd5b6000611a1f86336109e3565b905080158015611a365750611a348233611782565b155b15611a5d5760405163168ab55d60e31b81526004810183905233602482015260440161088a565b60008080611a6a856108cf565b92509250925083158015611a815750620400008216155b15611aa25760405163a2a7201360e01b81526004810186905260240161088a565b6000611aad8a6108cf565b92505050611abc8883836141b7565b9750611aca8685858b6142c1565b60405167ffffffffffffffff8916815286907ff675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b1329060200160405180910390a2509598975050505050505050565b611b1f613a4a565b611b296000614309565b565b604080517fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a46360208083019190915281830186905282518083038401815260609092019092528051910120611b7f8133611782565b611ba55760405163168ab55d60e31b81526004810182905233602482015260440161088a565b306001600160a01b03841603611bd957604051632ca49b0d60e11b81526001600160a01b038416600482015260240161088a565b604080517fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a46360208083019190915281830187905282518083038401815260609092019092528051910120611c2e905b83614366565b6040517f42842e0e0000000000000000000000000000000000000000000000000000000081523060048201526001600160a01b038481166024830152604482018690527f000000000000000000000000000000000000000000000000000000000000000016906342842e0e90606401600060405180830381600087803b158015611cb757600080fd5b505af1158015611ccb573d6000803e3d6000fd5b5050505050505050565b6000808686604051611ce8929190615a5f565b6040519081900381206331a9108f60e11b82526004820181905291506000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690636352211e90602401602060405180830381865afa158015611d58573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611d7c9190615a25565b90506001600160a01b0381163314801590611e24575060405163e985e9c560e01b81526001600160a01b0382811660048301523360248301527f0000000000000000000000000000000000000000000000000000000000000000169063e985e9c590604401602060405180830381865afa158015611dfe573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611e229190615a42565b155b15611e9457604080517fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a4636020808301919091528183018590528251808303840181526060830193849052805191012063168ab55d60e31b909252606481019190915233608482015260a40161088a565b6040517f23b872dd0000000000000000000000000000000000000000000000000000000081526001600160a01b038281166004830152306024830152604482018490527f000000000000000000000000000000000000000000000000000000000000000016906323b872dd90606401600060405180830381600087803b158015611f1d57600080fd5b505af1158015611f31573d6000803e3d6000fd5b5050604051630a3b53db60e21b8152600481018590523060248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031692506328ed4f6c9150604401600060405180830381600087803b158015611f9d57600080fd5b505af1158015611fb1573d6000803e3d6000fd5b5050604051636b727d4360e11b8152600481018590526276a70092507f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316915063d6e4fa8690602401602060405180830381865afa15801561201f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906120439190615937565b61204d9190615966565b925061209688888080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508a9250505061ffff881686886138e4565b505095945050505050565b6001600160a01b038216330361211f5760405162461bcd60e51b815260206004820152602960248201527f455243313135353a2073657474696e6720617070726f76616c2073746174757360448201527f20666f722073656c660000000000000000000000000000000000000000000000606482015260840161088a565b3360008181526002602090815260408083206001600160a01b03871680855290835292819020805460ff191686151590811790915590519081529192917f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31910160405180910390a35050565b3360009081526004602052604081205460ff166121fb5760405162461bcd60e51b815260206004820152602860248201527f436f6e74726f6c6c61626c653a2043616c6c6572206973206e6f74206120636f604482015267373a3937b63632b960c11b606482015260840161088a565b6000878760405161220d929190615a5f565b6040519081900381207ffca247ac000000000000000000000000000000000000000000000000000000008252600482018190523060248301526044820187905291507f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063fca247ac906064016020604051808303816000875af11580156122a2573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906122c69190615937565b915061231b88888080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508a9250505061ffff86166123156276a70087615966565b886138e4565b509695505050505050565b600080612332846108cf565b50841663ffffffff908116908516149250505092915050565b612353613a4a565b6007546001600160a01b0316156124735760075460405163a22cb46560e01b81526001600160a01b039182166004820152600060248201527f00000000000000000000000000000000000000000000000000000000000000009091169063a22cb46590604401600060405180830381600087803b1580156123d357600080fd5b505af11580156123e7573d6000803e3d6000fd5b505060075460405163a22cb46560e01b81526001600160a01b039182166004820152600060248201527f0000000000000000000000000000000000000000000000000000000000000000909116925063a22cb4659150604401600060405180830381600087803b15801561245a57600080fd5b505af115801561246e573d6000803e3d6000fd5b505050505b6007805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b038316908117909155156125ad5760075460405163a22cb46560e01b81526001600160a01b039182166004820152600160248201527f00000000000000000000000000000000000000000000000000000000000000009091169063a22cb46590604401600060405180830381600087803b15801561251257600080fd5b505af1158015612526573d6000803e3d6000fd5b505060075460405163a22cb46560e01b81526001600160a01b039182166004820152600160248201527f0000000000000000000000000000000000000000000000000000000000000000909116925063a22cb4659150604401600060405180830381600087803b15801561259957600080fd5b505af11580156114e4573d6000803e3d6000fd5b50565b3360009081526004602052604081205460ff166126205760405162461bcd60e51b815260206004820152602860248201527f436f6e74726f6c6c61626c653a2043616c6c6572206973206e6f74206120636f604482015267373a3937b63632b960c11b606482015260840161088a565b604080517fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a463602080830191909152818301869052825180830384018152606090920190925280519101206000906040517fc475abff00000000000000000000000000000000000000000000000000000000815260048101869052602481018590529091506000906001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000169063c475abff906044016020604051808303816000875af11580156126fa573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061271e9190615937565b6040516331a9108f60e11b8152600481018790529091507f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690636352211e90602401602060405180830381865afa9250505080156127a2575060408051601f3d908101601f1916820190925261279f91810190615a25565b60015b6127af5791506108c99050565b6001600160a01b0381163014158061285957506040516302571be360e01b81526004810184905230906001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016906302571be390602401602060405180830381865afa158015612829573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061284d9190615a25565b6001600160a01b031614155b15612868575091506108c99050565b5060006128786276a70083615966565b60008481526001602052604090205490915060a081901c61289b858383866142c1565b50919695505050505050565b6000866128b48133611782565b6128da5760405163168ab55d60e31b81526004810182905233602482015260440161088a565b600087876040516128ec929190615a5f565b604051809103902090506129278982604080516020808201949094528082019290925280518083038201815260609092019052805191012090565b92506129338984613aa4565b61293d8386613be3565b60006129808a858b8b8080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250613c1692505050565b905061298e8a858888613ce3565b945061299984613d29565b612a61576040517f06ab5923000000000000000000000000000000000000000000000000000000008152600481018b9052602481018390523060448201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906306ab5923906064016020604051808303816000875af1158015612a2a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190612a4e9190615937565b50612a5c8482898989614458565b612aa7565b612aa78a858b8b8080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508d92508c91508b9050613e19565b5050509695505050505050565b6000612afa600086868080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929392505061449a9050565b6007549091506001600160a01b0316612b3f576040517f24c1d6d400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b612b498133611782565b612b6f5760405163168ab55d60e31b81526004810182905233602482015260440161088a565b60008080612b7c846108cf565b919450925090506000612b8e85610958565b9050612b9985614559565b600760009054906101000a90046001600160a01b03166001600160a01b0316639198c2768a8a878787878e8e6040518963ffffffff1660e01b8152600401612be8989796959493929190615a98565b600060405180830381600087803b158015612c0257600080fd5b505af1158015612c16573d6000803e3d6000fd5b50505050505050505050505050565b83612c308133611782565b612c565760405163168ab55d60e31b81526004810182905233602482015260440161088a565b84601c6000612c64836108cf565b5091505063ffffffff8282161615612c925760405163a2a7201360e01b81526004810184905260240161088a565b6040517fcf408823000000000000000000000000000000000000000000000000000000008152600481018990523060248201526001600160a01b03878116604483015267ffffffffffffffff871660648301527f0000000000000000000000000000000000000000000000000000000000000000169063cf40882390608401600060405180830381600087803b158015612d2b57600080fd5b505af1158015612d3f573d6000803e3d6000fd5b5050506001600160a01b0388169050612da6576000612d5d896108cf565b509150506201ffff1962020000821601612d9557604051632ca49b0d60e11b81526001600160a01b038916600482015260240161088a565b612da0896000614366565b50611ccb565b6000612db1896119b7565b9050612dd381898b60001c600160405180602001604052806000815250614628565b505050505050505050565b60408051602080820186905281830185905282518083038401815260609092019092528051910120612e108133611782565b612e365760405163168ab55d60e31b81526004810182905233602482015260440161088a565b7f1781448691ae9413300646aa4093cc50b13ced291d67be7a2fe90154b2e75b9d8401612e765760405163615a470360e01b815260040160405180910390fd5b6001600160a01b0382161580612e9457506001600160a01b03821630145b15612ebd57604051632ca49b0d60e11b81526001600160a01b038316600482015260240161088a565b604080516020808201879052818301869052825180830384018152606090920190925280519101206119b190611c28565b604080516020808201859052818301849052825180830384018152606090920190925280519101206000906000612f2482613d29565b90507fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a4638514612f565791506108c99050565b6040516331a9108f60e11b8152600481018590527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690636352211e90602401602060405180830381865afa925050508015612fd7575060408051601f3d908101601f19168201909252612fd491810190615a25565b60015b612fe6576000925050506108c9565b6001600160a01b0316301492506108c9915050565b613003613a4a565b6001600160a01b038216600081815260046020908152604091829020805460ff191685151590811790915591519182527f4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf8791015b60405180910390a25050565b6000806130aa600087878080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929392505061477a9050565b9150915060006130f38288888080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929392505061449a9050565b60408051602080820184905281830187905282518083038401815260609092019092528051910120909150600090600081815260066020526040902090915061313d888a83615b47565b507f1781448691ae9413300646aa4093cc50b13ced291d67be7a2fe90154b2e75b9d820161317e5760405163615a470360e01b815260040160405180910390fd5b6040516302571be360e01b8152600481018290526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906302571be390602401602060405180830381865afa1580156131e6573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061320a9190615a25565b90506001600160a01b03811633148015906132b2575060405163e985e9c560e01b81526001600160a01b0382811660048301523360248301527f0000000000000000000000000000000000000000000000000000000000000000169063e985e9c590604401602060405180830381865afa15801561328c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906132b09190615a42565b155b156132d95760405163168ab55d60e31b81526004810183905233602482015260440161088a565b6001600160a01b0386161561336b57604051630c4b7b8560e11b8152600481018390526001600160a01b0387811660248301527f00000000000000000000000000000000000000000000000000000000000000001690631896f70a90604401600060405180830381600087803b15801561335257600080fd5b505af1158015613366573d6000803e3d6000fd5b505050505b604051635b0fc9c360e01b8152600481018390523060248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690635b0fc9c390604401600060405180830381600087803b1580156133d357600080fd5b505af11580156133e7573d6000803e3d6000fd5b50505050612dd3828a8a8080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201829052508d93509150819050614458565b6001600160a01b0384166134925760405162461bcd60e51b815260206004820152602560248201527f455243313135353a207472616e7366657220746f20746865207a65726f206164604482015264647265737360d81b606482015260840161088a565b6001600160a01b0385163314806134cc57506001600160a01b038516600090815260026020908152604080832033845290915290205460ff165b61353e5760405162461bcd60e51b815260206004820152602960248201527f455243313135353a2063616c6c6572206973206e6f74206f776e6572206e6f7260448201527f20617070726f7665640000000000000000000000000000000000000000000000606482015260840161088a565b6114e48585858585614628565b613553613a4a565b6001600160a01b0381166135cf5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f6464726573730000000000000000000000000000000000000000000000000000606482015260840161088a565b6125ad81614309565b600081815260066020526040812080548291906135f49061598e565b80601f01602080910402602001604051908101604052809291908181526020018280546136209061598e565b801561366d5780601f106136425761010080835404028352916020019161366d565b820191906000526020600020905b81548152906001019060200180831161365057829003601f168201915b5050505050905080516000036136865750600092915050565b600080613693838261477a565b909250905060006136a4848361449a565b9050610a738184612eee565b600080428367ffffffffffffffff1610156136de5761ffff19620100008516016136d957600094505b600093505b50929391925050565b60006001600160e01b031982167fd9b67a2600000000000000000000000000000000000000000000000000000000148061373157506001600160e01b031982166303a24d0760e21b145b806108c957507f01ffc9a7000000000000000000000000000000000000000000000000000000006001600160e01b03198316146108c9565b6000613774826119b7565b9050806001600160a01b0316836001600160a01b0316036137fd5760405162461bcd60e51b815260206004820152602160248201527f4552433732313a20617070726f76616c20746f2063757272656e74206f776e6560448201527f7200000000000000000000000000000000000000000000000000000000000000606482015260840161088a565b336001600160a01b038216148061383757506001600160a01b038116600090815260026020908152604080832033845290915290205460ff165b6138a95760405162461bcd60e51b815260206004820152603d60248201527f4552433732313a20617070726f76652063616c6c6572206973206e6f7420746f60448201527f6b656e206f776e6572206f7220617070726f76656420666f7220616c6c000000606482015260840161088a565b6109de8383614831565b6000620200008381161480156109965750426138d26276a70084615c07565b67ffffffffffffffff16109392505050565b84516020860120600061393e7fe87ebb796e516beccff9b955bf6c33af4ec312d6e2984185d016feab4d18a46383604080516020808201949094528082019290925280518083038201815260609092019052805191012090565b90506000613981886040518060400160405280600581526020017f03616e79000000000000000000000000000000000000000000000000000000008152506148ac565b600083815260066020526040902090915061399c8282615c28565b506139af828289620300008a1789614458565b6001600160a01b03841615611ccb57604051630c4b7b8560e11b8152600481018390526001600160a01b0385811660248301527f00000000000000000000000000000000000000000000000000000000000000001690631896f70a90604401600060405180830381600087803b158015613a2857600080fd5b505af1158015613a3c573d6000803e3d6000fd5b505050505050505050505050565b6000546001600160a01b03163314611b295760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161088a565b60008080613ab1846108cf565b919450925090504267ffffffffffffffff821610808015613b7557506001600160a01b0384161580613b7557506040516302571be360e01b8152600481018690526000906001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016906302571be390602401602060405180830381865afa158015613b46573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190613b6a9190615a25565b6001600160a01b0316145b15613bb4576000613b85876108cf565b509150506020811615613bae5760405163a2a7201360e01b81526004810187905260240161088a565b50613bdb565b62010000831615613bdb5760405163a2a7201360e01b81526004810186905260240161088a565b505050505050565b63fffdffff81811763ffffffff1614613c125760405163a2a7201360e01b81526004810183905260240161088a565b5050565b60606000613cbf83600660008881526020019081526020016000208054613c3c9061598e565b80601f0160208091040260200160405190810160405280929190818152602001828054613c689061598e565b8015613cb55780601f10613c8a57610100808354040283529160200191613cb5565b820191906000526020600020905b815481529060010190602001808311613c9857829003601f168201915b50505050506148ac565b6000858152600660205260409020909150613cda8282615c28565b50949350505050565b600080613cef856108cf565b92505050600080613d028860001c6108cf565b9250925050613d1287878461417c565b613d1d8584836141b7565b98975050505050505050565b600080613d35836119b7565b6001600160a01b0316141580156108c957506040516302571be360e01b81526004810183905230906001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016906302571be390602401602060405180830381865afa158015613dae573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190613dd29190615a25565b6001600160a01b03161492915050565b60008681526006602052604081208054613e01918791613c3c9061598e565b9050613e108682868686614458565b50505050505050565b60008080613e26886108cf565b9250925092506000613e5088600660008d81526020019081526020016000208054613c3c9061598e565b60008a8152600660205260409020805491925090613e6d9061598e565b9050600003613e90576000898152600660205260409020613e8e8282615c28565b505b613e9f89858886178589614201565b6001600160a01b038716613ebd57613eb8896000614366565b610bfc565b610bfc84888b60001c600160405180602001604052806000815250614628565b6201ffff1962020000831601613efd57613efa6276a70082615c07565b90505b428167ffffffffffffffff161015613f7a5762010000821615613f755760405162461bcd60e51b815260206004820152602a60248201527f455243313135353a20696e73756666696369656e742062616c616e636520666f60448201526939103a3930b739b332b960b11b606482015260840161088a565b613f9f565b6004821615613f9f5760405163a2a7201360e01b81526004810184905260240161088a565b604082166000036109de5750506000908152600360205260409020805473ffffffffffffffffffffffffffffffffffffffff19169055565b6001600160a01b0384163b15613bdb5760405163bc197c8160e01b81526001600160a01b0385169063bc197c819061401b9089908990889088908890600401615ce8565b6020604051808303816000875af1925050508015614056575060408051601f3d908101601f1916820190925261405391810190615d3a565b60015b61410b57614062615d57565b806308c379a00361409b5750614076615d73565b80614081575061409d565b8060405162461bcd60e51b815260040161088a9190614e3c565b505b60405162461bcd60e51b815260206004820152603460248201527f455243313135353a207472616e7366657220746f206e6f6e204552433131353560448201527f526563656976657220696d706c656d656e746572000000000000000000000000606482015260840161088a565b6001600160e01b0319811663bc197c8160e01b14613e105760405162461bcd60e51b815260206004820152602860248201527f455243313135353a204552433131353552656365697665722072656a656374656044820152676420746f6b656e7360c01b606482015260840161088a565b63ffff000082161580159060018316159082906141965750805b156114e45760405163a2a7201360e01b81526004810186905260240161088a565b60008167ffffffffffffffff168467ffffffffffffffff1611156141d9578193505b8267ffffffffffffffff168467ffffffffffffffff1610156141f9578293505b509192915050565b61420d858585846142c1565b60405163ffffffff8416815285907f39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b9060200160405180910390a28167ffffffffffffffff168167ffffffffffffffff1611156114e45760405167ffffffffffffffff8216815285907ff675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132906020015b60405180910390a25050505050565b6000806142b7836108cf565b5090949350505050565b6142cb8483614955565b60008481526001602052604090206001600160a01b03841663ffffffff60a01b60a085901b16176001600160c01b031960c084901b161790556119b1565b600080546001600160a01b0383811673ffffffffffffffffffffffffffffffffffffffff19831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b614371826001612326565b156143925760405163a2a7201360e01b81526004810183905260240161088a565b61439b82614559565b604051635b0fc9c360e01b8152600481018390526001600160a01b0382811660248301527f00000000000000000000000000000000000000000000000000000000000000001690635b0fc9c390604401600060405180830381600087803b15801561440557600080fd5b505af1158015614419573d6000803e3d6000fd5b50506040516001600160a01b03841681528492507fee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c49150602001613057565b6144648584848461498e565b847f8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd3408585858560405161429c9493929190615dfd565b60008060006144a9858561477a565b90925090508161451b57600185516144c19190615e45565b841461450f5760405162461bcd60e51b815260206004820152601d60248201527f6e616d65686173683a204a756e6b20617420656e64206f66206e616d65000000604482015260640161088a565b50600091506108c99050565b614525858261449a565b6040805160208101929092528101839052606001604051602081830303815290604052805190602001209250505092915050565b60008181526001602052604090205460a081901c60c082901c61457d8383836136b0565b6000868152600360209081526040808320805473ffffffffffffffffffffffffffffffffffffffff191690556001909152902063ffffffff60a01b60a083901b166001600160c01b031960c086901b1617905592506145d99050565b60408051858152600160208201526000916001600160a01b0386169133917fc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62910160405180910390a450505050565b6000806000614636866108cf565b925092509250614647868383613edd565b8460011480156146685750876001600160a01b0316836001600160a01b0316145b6146c75760405162461bcd60e51b815260206004820152602a60248201527f455243313135353a20696e73756666696369656e742062616c616e636520666f60448201526939103a3930b739b332b960b11b606482015260840161088a565b866001600160a01b0316836001600160a01b0316036146e8575050506114e4565b60008681526001602052604090206001600160a01b03881663ffffffff60a01b60a085901b16176001600160c01b031960c084901b1617905560408051878152602081018790526001600160a01b03808a1692908b169133917fc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62910160405180910390a4611ccb338989898989614a02565b600080835183106147cd5760405162461bcd60e51b815260206004820152601e60248201527f726561644c6162656c3a20496e646578206f7574206f6620626f756e64730000604482015260640161088a565b60008484815181106147e1576147e16159c8565b016020015160f81c9050801561480d5761480685614800866001615e58565b83614afe565b9250614812565b600092505b61481c8185615e58565b614827906001615e58565b9150509250929050565b6000818152600360205260409020805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0384169081179091558190614873826119b7565b6001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560405160405180910390a45050565b60606001835110156148ea576040517f280dacb600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60ff8351111561492857826040517fe3ba295f00000000000000000000000000000000000000000000000000000000815260040161088a9190614e3c565b8251838360405160200161493e93929190615e6b565b604051602081830303815290604052905092915050565b61ffff81161580159061496d57506201000181811614155b15613c125760405163a2a7201360e01b81526004810183905260240161088a565b6149988483614955565b6000848152600160205260409020546001600160a01b038116156149f6576149bf85614559565b6040516000815285907fee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c49060200160405180910390a25b6114e485858585614b22565b6001600160a01b0384163b15613bdb5760405163f23a6e6160e01b81526001600160a01b0385169063f23a6e6190614a469089908990889088908890600401615ecc565b6020604051808303816000875af1925050508015614a81575060408051601f3d908101601f19168201909252614a7e91810190615d3a565b60015b614a8d57614062615d57565b6001600160e01b0319811663f23a6e6160e01b14613e105760405162461bcd60e51b815260206004820152602860248201527f455243313135353a204552433131353552656365697665722072656a656374656044820152676420746f6b656e7360c01b606482015260840161088a565b8251600090614b0d8385615e58565b1115614b1857600080fd5b5091016020012090565b8360008080614b30846108cf565b9194509250905063ffff0000821667ffffffffffffffff8087169083161115614b57578195505b428267ffffffffffffffff1610614b6d57958617955b6001600160a01b03841615614bc45760405162461bcd60e51b815260206004820152601f60248201527f455243313135353a206d696e74206f66206578697374696e6720746f6b656e00604482015260640161088a565b6001600160a01b038816614c405760405162461bcd60e51b815260206004820152602160248201527f455243313135353a206d696e7420746f20746865207a65726f2061646472657360448201527f7300000000000000000000000000000000000000000000000000000000000000606482015260840161088a565b306001600160a01b03891603614cbe5760405162461bcd60e51b815260206004820152603460248201527f455243313135353a206e65774f776e65722063616e6e6f74206265207468652060448201527f4e616d655772617070657220636f6e7472616374000000000000000000000000606482015260840161088a565b60008581526001602052604090206001600160a01b03891663ffffffff60a01b60a08a901b16176001600160c01b031960c089901b1617905560408051868152600160208201526001600160a01b038a169160009133917fc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62910160405180910390a4612dd33360008a88600160405180602001604052806000815250614a02565b6001600160a01b03811681146125ad57600080fd5b60008060408385031215614d8757600080fd5b8235614d9281614d5f565b946020939093013593505050565b600060208284031215614db257600080fd5b5035919050565b6001600160e01b0319811681146125ad57600080fd5b600060208284031215614de157600080fd5b813561099681614db9565b60005b83811015614e07578181015183820152602001614def565b50506000910152565b60008151808452614e28816020860160208601614dec565b601f01601f19169290920160200192915050565b6020815260006109966020830184614e10565b60008060408385031215614e6257600080fd5b823591506020830135614e7481614d5f565b809150509250929050565b803567ffffffffffffffff81168114614e9757600080fd5b919050565b60008060408385031215614eaf57600080fd5b82359150614ebf60208401614e7f565b90509250929050565b60008083601f840112614eda57600080fd5b50813567ffffffffffffffff811115614ef257600080fd5b602083019150836020828501011115614f0a57600080fd5b9250929050565b600080600080600060808688031215614f2957600080fd5b8535614f3481614d5f565b94506020860135614f4481614d5f565b935060408601359250606086013567ffffffffffffffff811115614f6757600080fd5b614f7388828901614ec8565b969995985093965092949392505050565b600060208284031215614f9657600080fd5b813561099681614d5f565b634e487b7160e01b600052604160045260246000fd5b601f8201601f1916810167ffffffffffffffff81118282101715614fdd57614fdd614fa1565b6040525050565b600067ffffffffffffffff821115614ffe57614ffe614fa1565b50601f01601f191660200190565b600082601f83011261501d57600080fd5b813561502881614fe4565b6040516150358282614fb7565b82815285602084870101111561504a57600080fd5b82602086016020830137600092810160200192909252509392505050565b803563ffffffff81168114614e9757600080fd5b600080600080600080600060e0888a03121561509757600080fd5b87359650602088013567ffffffffffffffff8111156150b557600080fd5b6150c18a828b0161500c565b96505060408801356150d281614d5f565b945060608801356150e281614d5f565b93506150f060808901614e7f565b92506150fe60a08901615068565b915061510c60c08901614e7f565b905092959891949750929550565b600067ffffffffffffffff82111561513457615134614fa1565b5060051b60200190565b600082601f83011261514f57600080fd5b8135602061515c8261511a565b6040516151698282614fb7565b83815260059390931b850182019282810191508684111561518957600080fd5b8286015b8481101561231b578035835291830191830161518d565b600080600080600060a086880312156151bc57600080fd5b85356151c781614d5f565b945060208601356151d781614d5f565b9350604086013567ffffffffffffffff808211156151f457600080fd5b61520089838a0161513e565b9450606088013591508082111561521657600080fd5b61522289838a0161513e565b9350608088013591508082111561523857600080fd5b506152458882890161500c565b9150509295509295909350565b6000806000806080858703121561526857600080fd5b843593506020850135925061527f60408601615068565b915061528d60608601614e7f565b905092959194509250565b803561ffff81168114614e9757600080fd5b600080604083850312156152bd57600080fd5b82359150614ebf60208401615298565b600080604083850312156152e057600080fd5b823567ffffffffffffffff808211156152f857600080fd5b818501915085601f83011261530c57600080fd5b813560206153198261511a565b6040516153268282614fb7565b83815260059390931b850182019282810191508984111561534657600080fd5b948201945b8386101561536d57853561535e81614d5f565b8252948201949082019061534b565b9650508601359250508082111561538357600080fd5b506148278582860161513e565b600081518084526020808501945080840160005b838110156153c0578151875295820195908201906001016153a4565b509495945050505050565b6020815260006109966020830184615390565b6000806000606084860312156153f357600080fd5b83356153fe81614d5f565b9250602084013561540e81614d5f565b929592945050506040919091013590565b60008060006060848603121561543457600080fd5b833592506020840135915061544b60408501614e7f565b90509250925092565b60008060006060848603121561546957600080fd5b83359250602084013561547b81614d5f565b9150604084013561548b81614d5f565b809150509250925092565b6000806000806000608086880312156154ae57600080fd5b853567ffffffffffffffff8111156154c557600080fd5b6154d188828901614ec8565b90965094505060208601356154e581614d5f565b92506154f360408701615298565b9150606086013561550381614d5f565b809150509295509295909350565b80151581146125ad57600080fd5b6000806040838503121561553257600080fd5b823561553d81614d5f565b91506020830135614e7481615511565b60008060008060008060a0878903121561556657600080fd5b863567ffffffffffffffff81111561557d57600080fd5b61558989828a01614ec8565b909750955050602087013561559d81614d5f565b93506040870135925060608701356155b481614d5f565b91506155c260808801615298565b90509295509295509295565b600080604083850312156155e157600080fd5b82359150614ebf60208401615068565b6000806040838503121561560457600080fd5b50508035926020909101359150565b60008060008060008060a0878903121561562c57600080fd5b86359550602087013567ffffffffffffffff81111561564a57600080fd5b61565689828a01614ec8565b909650945050604087013561566a81614d5f565b925061567860608801615068565b91506155c260808801614e7f565b6000806000806040858703121561569c57600080fd5b843567ffffffffffffffff808211156156b457600080fd5b6156c088838901614ec8565b909650945060208701359150808211156156d957600080fd5b506156e687828801614ec8565b95989497509550505050565b6000806000806080858703121561570857600080fd5b84359350602085013561571a81614d5f565b9250604085013561527f81614d5f565b60008060006060848603121561573f57600080fd5b8335925060208401359150604084013561548b81614d5f565b6000806040838503121561576b57600080fd5b823561577681614d5f565b91506020830135614e7481614d5f565b6000806000806060858703121561579c57600080fd5b843567ffffffffffffffff8111156157b357600080fd5b6157bf87828801614ec8565b90955093505060208501356157d381614d5f565b915060408501356157e381614d5f565b939692955090935050565b600080600080600060a0868803121561580657600080fd5b853561581181614d5f565b9450602086013561582181614d5f565b93506040860135925060608601359150608086013567ffffffffffffffff81111561584b57600080fd5b6152458882890161500c565b60006020828403121561586957600080fd5b815167ffffffffffffffff81111561588057600080fd5b8201601f8101841361589157600080fd5b805161589c81614fe4565b6040516158a98282614fb7565b8281528660208486010111156158be57600080fd5b610a73836020830160208701614dec565b600080600080608085870312156158e557600080fd5b843567ffffffffffffffff8111156158fc57600080fd5b6159088782880161500c565b945050602085013561591981614d5f565b925061592760408601615298565b915060608501356157e381614d5f565b60006020828403121561594957600080fd5b5051919050565b634e487b7160e01b600052601160045260246000fd5b67ffffffffffffffff81811683821601908082111561598757615987615950565b5092915050565b600181811c908216806159a257607f821691505b6020821081036159c257634e487b7160e01b600052602260045260246000fd5b50919050565b634e487b7160e01b600052603260045260246000fd5b6000600182016159f0576159f0615950565b5060010190565b604081526000615a0a6040830185615390565b8281036020840152615a1c8185615390565b95945050505050565b600060208284031215615a3757600080fd5b815161099681614d5f565b600060208284031215615a5457600080fd5b815161099681615511565b8183823760009101908152919050565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b60c081526000615aac60c083018a8c615a6f565b6001600160a01b03898116602085015263ffffffff8916604085015267ffffffffffffffff881660608501528616608084015282810360a0840152615af2818587615a6f565b9b9a5050505050505050505050565b601f8211156109de57600081815260208120601f850160051c81016020861015615b285750805b601f850160051c820191505b81811015613bdb57828155600101615b34565b67ffffffffffffffff831115615b5f57615b5f614fa1565b615b7383615b6d835461598e565b83615b01565b6000601f841160018114615ba75760008515615b8f5750838201355b600019600387901b1c1916600186901b1783556114e4565b600083815260209020601f19861690835b82811015615bd85786850135825560209485019460019092019101615bb8565b5086821015615bf55760001960f88860031b161c19848701351681555b505060018560011b0183555050505050565b67ffffffffffffffff82811682821603908082111561598757615987615950565b815167ffffffffffffffff811115615c4257615c42614fa1565b615c5681615c50845461598e565b84615b01565b602080601f831160018114615c8b5760008415615c735750858301515b600019600386901b1c1916600185901b178555613bdb565b600085815260208120601f198616915b82811015615cba57888601518255948401946001909101908401615c9b565b5085821015615cd85787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b60006001600160a01b03808816835280871660208401525060a06040830152615d1460a0830186615390565b8281036060840152615d268186615390565b90508281036080840152613d1d8185614e10565b600060208284031215615d4c57600080fd5b815161099681614db9565b600060033d1115615d705760046000803e5060005160e01c5b90565b600060443d1015615d815790565b6040516003193d81016004833e81513d67ffffffffffffffff8160248401118184111715615db157505050505090565b8285019150815181811115615dc95750505050505090565b843d8701016020828501011115615de35750505050505090565b615df260208286010187614fb7565b509095945050505050565b608081526000615e106080830187614e10565b6001600160a01b039590951660208301525063ffffffff92909216604083015267ffffffffffffffff16606090910152919050565b818103818111156108c9576108c9615950565b808201808211156108c9576108c9615950565b7fff000000000000000000000000000000000000000000000000000000000000008460f81b16815260008351615ea8816001850160208801614dec565b835190830190615ebf816001840160208801614dec565b0160010195945050505050565b60006001600160a01b03808816835280871660208401525084604083015283606083015260a06080830152615f0460a0830184614e10565b97965050505050505056fea26469706673582212207840f866f82d8dc50f748235e9d59a75bfb85c992b34674c8ae56ba11b6f1c8b64736f6c63430008110033",
}

// AnytypeNameWrapperABI is the input ABI used to generate the binding from.
// Deprecated: Use AnytypeNameWrapperMetaData.ABI instead.
var AnytypeNameWrapperABI = AnytypeNameWrapperMetaData.ABI

// AnytypeNameWrapperBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AnytypeNameWrapperMetaData.Bin instead.
var AnytypeNameWrapperBin = AnytypeNameWrapperMetaData.Bin

// DeployAnytypeNameWrapper deploys a new Ethereum contract, binding an instance of AnytypeNameWrapper to it.
func DeployAnytypeNameWrapper(auth *bind.TransactOpts, backend bind.ContractBackend, _ens common.Address, _registrar common.Address, _metadataService common.Address) (common.Address, *types.Transaction, *AnytypeNameWrapper, error) {
	parsed, err := AnytypeNameWrapperMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AnytypeNameWrapperBin), backend, _ens, _registrar, _metadataService)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AnytypeNameWrapper{AnytypeNameWrapperCaller: AnytypeNameWrapperCaller{contract: contract}, AnytypeNameWrapperTransactor: AnytypeNameWrapperTransactor{contract: contract}, AnytypeNameWrapperFilterer: AnytypeNameWrapperFilterer{contract: contract}}, nil
}

// AnytypeNameWrapper is an auto generated Go binding around an Ethereum contract.
type AnytypeNameWrapper struct {
	AnytypeNameWrapperCaller     // Read-only binding to the contract
	AnytypeNameWrapperTransactor // Write-only binding to the contract
	AnytypeNameWrapperFilterer   // Log filterer for contract events
}

// AnytypeNameWrapperCaller is an auto generated read-only Go binding around an Ethereum contract.
type AnytypeNameWrapperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeNameWrapperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AnytypeNameWrapperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeNameWrapperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AnytypeNameWrapperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeNameWrapperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AnytypeNameWrapperSession struct {
	Contract     *AnytypeNameWrapper // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// AnytypeNameWrapperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AnytypeNameWrapperCallerSession struct {
	Contract *AnytypeNameWrapperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// AnytypeNameWrapperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AnytypeNameWrapperTransactorSession struct {
	Contract     *AnytypeNameWrapperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// AnytypeNameWrapperRaw is an auto generated low-level Go binding around an Ethereum contract.
type AnytypeNameWrapperRaw struct {
	Contract *AnytypeNameWrapper // Generic contract binding to access the raw methods on
}

// AnytypeNameWrapperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AnytypeNameWrapperCallerRaw struct {
	Contract *AnytypeNameWrapperCaller // Generic read-only contract binding to access the raw methods on
}

// AnytypeNameWrapperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AnytypeNameWrapperTransactorRaw struct {
	Contract *AnytypeNameWrapperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAnytypeNameWrapper creates a new instance of AnytypeNameWrapper, bound to a specific deployed contract.
func NewAnytypeNameWrapper(address common.Address, backend bind.ContractBackend) (*AnytypeNameWrapper, error) {
	contract, err := bindAnytypeNameWrapper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapper{AnytypeNameWrapperCaller: AnytypeNameWrapperCaller{contract: contract}, AnytypeNameWrapperTransactor: AnytypeNameWrapperTransactor{contract: contract}, AnytypeNameWrapperFilterer: AnytypeNameWrapperFilterer{contract: contract}}, nil
}

// NewAnytypeNameWrapperCaller creates a new read-only instance of AnytypeNameWrapper, bound to a specific deployed contract.
func NewAnytypeNameWrapperCaller(address common.Address, caller bind.ContractCaller) (*AnytypeNameWrapperCaller, error) {
	contract, err := bindAnytypeNameWrapper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperCaller{contract: contract}, nil
}

// NewAnytypeNameWrapperTransactor creates a new write-only instance of AnytypeNameWrapper, bound to a specific deployed contract.
func NewAnytypeNameWrapperTransactor(address common.Address, transactor bind.ContractTransactor) (*AnytypeNameWrapperTransactor, error) {
	contract, err := bindAnytypeNameWrapper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperTransactor{contract: contract}, nil
}

// NewAnytypeNameWrapperFilterer creates a new log filterer instance of AnytypeNameWrapper, bound to a specific deployed contract.
func NewAnytypeNameWrapperFilterer(address common.Address, filterer bind.ContractFilterer) (*AnytypeNameWrapperFilterer, error) {
	contract, err := bindAnytypeNameWrapper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperFilterer{contract: contract}, nil
}

// bindAnytypeNameWrapper binds a generic wrapper to an already deployed contract.
func bindAnytypeNameWrapper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AnytypeNameWrapperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeNameWrapper *AnytypeNameWrapperRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeNameWrapper.Contract.AnytypeNameWrapperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeNameWrapper *AnytypeNameWrapperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.AnytypeNameWrapperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeNameWrapper *AnytypeNameWrapperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.AnytypeNameWrapperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeNameWrapper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.contract.Transact(opts, method, params...)
}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Tokens(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "_tokens", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Tokens(arg0 *big.Int) (*big.Int, error) {
	return _AnytypeNameWrapper.Contract.Tokens(&_AnytypeNameWrapper.CallOpts, arg0)
}

// Tokens is a free data retrieval call binding the contract method 0xed70554d.
//
// Solidity: function _tokens(uint256 ) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Tokens(arg0 *big.Int) (*big.Int, error) {
	return _AnytypeNameWrapper.Contract.Tokens(&_AnytypeNameWrapper.CallOpts, arg0)
}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) AllFusesBurned(opts *bind.CallOpts, node [32]byte, fuseMask uint32) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "allFusesBurned", node, fuseMask)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) AllFusesBurned(node [32]byte, fuseMask uint32) (bool, error) {
	return _AnytypeNameWrapper.Contract.AllFusesBurned(&_AnytypeNameWrapper.CallOpts, node, fuseMask)
}

// AllFusesBurned is a free data retrieval call binding the contract method 0xadf4960a.
//
// Solidity: function allFusesBurned(bytes32 node, uint32 fuseMask) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) AllFusesBurned(node [32]byte, fuseMask uint32) (bool, error) {
	return _AnytypeNameWrapper.Contract.AllFusesBurned(&_AnytypeNameWrapper.CallOpts, node, fuseMask)
}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) BalanceOf(opts *bind.CallOpts, account common.Address, id *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "balanceOf", account, id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) BalanceOf(account common.Address, id *big.Int) (*big.Int, error) {
	return _AnytypeNameWrapper.Contract.BalanceOf(&_AnytypeNameWrapper.CallOpts, account, id)
}

// BalanceOf is a free data retrieval call binding the contract method 0x00fdd58e.
//
// Solidity: function balanceOf(address account, uint256 id) view returns(uint256)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) BalanceOf(account common.Address, id *big.Int) (*big.Int, error) {
	return _AnytypeNameWrapper.Contract.BalanceOf(&_AnytypeNameWrapper.CallOpts, account, id)
}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) BalanceOfBatch(opts *bind.CallOpts, accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "balanceOfBatch", accounts, ids)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) BalanceOfBatch(accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	return _AnytypeNameWrapper.Contract.BalanceOfBatch(&_AnytypeNameWrapper.CallOpts, accounts, ids)
}

// BalanceOfBatch is a free data retrieval call binding the contract method 0x4e1273f4.
//
// Solidity: function balanceOfBatch(address[] accounts, uint256[] ids) view returns(uint256[])
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) BalanceOfBatch(accounts []common.Address, ids []*big.Int) ([]*big.Int, error) {
	return _AnytypeNameWrapper.Contract.BalanceOfBatch(&_AnytypeNameWrapper.CallOpts, accounts, ids)
}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) CanExtendSubnames(opts *bind.CallOpts, node [32]byte, addr common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "canExtendSubnames", node, addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) CanExtendSubnames(node [32]byte, addr common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.CanExtendSubnames(&_AnytypeNameWrapper.CallOpts, node, addr)
}

// CanExtendSubnames is a free data retrieval call binding the contract method 0x0e4cd725.
//
// Solidity: function canExtendSubnames(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) CanExtendSubnames(node [32]byte, addr common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.CanExtendSubnames(&_AnytypeNameWrapper.CallOpts, node, addr)
}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) CanModifyName(opts *bind.CallOpts, node [32]byte, addr common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "canModifyName", node, addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) CanModifyName(node [32]byte, addr common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.CanModifyName(&_AnytypeNameWrapper.CallOpts, node, addr)
}

// CanModifyName is a free data retrieval call binding the contract method 0x41415eab.
//
// Solidity: function canModifyName(bytes32 node, address addr) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) CanModifyName(node [32]byte, addr common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.CanModifyName(&_AnytypeNameWrapper.CallOpts, node, addr)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Controllers(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "controllers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Controllers(arg0 common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.Controllers(&_AnytypeNameWrapper.CallOpts, arg0)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Controllers(arg0 common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.Controllers(&_AnytypeNameWrapper.CallOpts, arg0)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Ens() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Ens(&_AnytypeNameWrapper.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Ens() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Ens(&_AnytypeNameWrapper.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) GetApproved(opts *bind.CallOpts, id *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "getApproved", id)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) GetApproved(id *big.Int) (common.Address, error) {
	return _AnytypeNameWrapper.Contract.GetApproved(&_AnytypeNameWrapper.CallOpts, id)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 id) view returns(address operator)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) GetApproved(id *big.Int) (common.Address, error) {
	return _AnytypeNameWrapper.Contract.GetApproved(&_AnytypeNameWrapper.CallOpts, id)
}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) GetData(opts *bind.CallOpts, id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "getData", id)

	outstruct := new(struct {
		Owner  common.Address
		Fuses  uint32
		Expiry uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Fuses = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.Expiry = *abi.ConvertType(out[2], new(uint64)).(*uint64)

	return *outstruct, err

}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) GetData(id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	return _AnytypeNameWrapper.Contract.GetData(&_AnytypeNameWrapper.CallOpts, id)
}

// GetData is a free data retrieval call binding the contract method 0x0178fe3f.
//
// Solidity: function getData(uint256 id) view returns(address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) GetData(id *big.Int) (struct {
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
}, error) {
	return _AnytypeNameWrapper.Contract.GetData(&_AnytypeNameWrapper.CallOpts, id)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) IsApprovedForAll(opts *bind.CallOpts, account common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "isApprovedForAll", account, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) IsApprovedForAll(account common.Address, operator common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsApprovedForAll(&_AnytypeNameWrapper.CallOpts, account, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address account, address operator) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) IsApprovedForAll(account common.Address, operator common.Address) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsApprovedForAll(&_AnytypeNameWrapper.CallOpts, account, operator)
}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) IsWrapped(opts *bind.CallOpts, parentNode [32]byte, labelhash [32]byte) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "isWrapped", parentNode, labelhash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) IsWrapped(parentNode [32]byte, labelhash [32]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsWrapped(&_AnytypeNameWrapper.CallOpts, parentNode, labelhash)
}

// IsWrapped is a free data retrieval call binding the contract method 0xd9a50c12.
//
// Solidity: function isWrapped(bytes32 parentNode, bytes32 labelhash) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) IsWrapped(parentNode [32]byte, labelhash [32]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsWrapped(&_AnytypeNameWrapper.CallOpts, parentNode, labelhash)
}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) IsWrapped0(opts *bind.CallOpts, node [32]byte) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "isWrapped0", node)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) IsWrapped0(node [32]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsWrapped0(&_AnytypeNameWrapper.CallOpts, node)
}

// IsWrapped0 is a free data retrieval call binding the contract method 0xfd0cd0d9.
//
// Solidity: function isWrapped(bytes32 node) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) IsWrapped0(node [32]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.IsWrapped0(&_AnytypeNameWrapper.CallOpts, node)
}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) MetadataService(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "metadataService")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) MetadataService() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.MetadataService(&_AnytypeNameWrapper.CallOpts)
}

// MetadataService is a free data retrieval call binding the contract method 0x53095467.
//
// Solidity: function metadataService() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) MetadataService() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.MetadataService(&_AnytypeNameWrapper.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Name() (string, error) {
	return _AnytypeNameWrapper.Contract.Name(&_AnytypeNameWrapper.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Name() (string, error) {
	return _AnytypeNameWrapper.Contract.Name(&_AnytypeNameWrapper.CallOpts)
}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Names(opts *bind.CallOpts, arg0 [32]byte) ([]byte, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "names", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Names(arg0 [32]byte) ([]byte, error) {
	return _AnytypeNameWrapper.Contract.Names(&_AnytypeNameWrapper.CallOpts, arg0)
}

// Names is a free data retrieval call binding the contract method 0x20c38e2b.
//
// Solidity: function names(bytes32 ) view returns(bytes)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Names(arg0 [32]byte) ([]byte, error) {
	return _AnytypeNameWrapper.Contract.Names(&_AnytypeNameWrapper.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Owner() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Owner(&_AnytypeNameWrapper.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Owner() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Owner(&_AnytypeNameWrapper.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) OwnerOf(opts *bind.CallOpts, id *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "ownerOf", id)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) OwnerOf(id *big.Int) (common.Address, error) {
	return _AnytypeNameWrapper.Contract.OwnerOf(&_AnytypeNameWrapper.CallOpts, id)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 id) view returns(address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) OwnerOf(id *big.Int) (common.Address, error) {
	return _AnytypeNameWrapper.Contract.OwnerOf(&_AnytypeNameWrapper.CallOpts, id)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Registrar(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "registrar")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Registrar() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Registrar(&_AnytypeNameWrapper.CallOpts)
}

// Registrar is a free data retrieval call binding the contract method 0x2b20e397.
//
// Solidity: function registrar() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Registrar() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.Registrar(&_AnytypeNameWrapper.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.SupportsInterface(&_AnytypeNameWrapper.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AnytypeNameWrapper.Contract.SupportsInterface(&_AnytypeNameWrapper.CallOpts, interfaceId)
}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) UpgradeContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "upgradeContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) UpgradeContract() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.UpgradeContract(&_AnytypeNameWrapper.CallOpts)
}

// UpgradeContract is a free data retrieval call binding the contract method 0x1f4e1504.
//
// Solidity: function upgradeContract() view returns(address)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) UpgradeContract() (common.Address, error) {
	return _AnytypeNameWrapper.Contract.UpgradeContract(&_AnytypeNameWrapper.CallOpts)
}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperCaller) Uri(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _AnytypeNameWrapper.contract.Call(opts, &out, "uri", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Uri(tokenId *big.Int) (string, error) {
	return _AnytypeNameWrapper.Contract.Uri(&_AnytypeNameWrapper.CallOpts, tokenId)
}

// Uri is a free data retrieval call binding the contract method 0x0e89341c.
//
// Solidity: function uri(uint256 tokenId) view returns(string)
func (_AnytypeNameWrapper *AnytypeNameWrapperCallerSession) Uri(tokenId *big.Int) (string, error) {
	return _AnytypeNameWrapper.Contract.Uri(&_AnytypeNameWrapper.CallOpts, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Approve(&_AnytypeNameWrapper.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Approve(&_AnytypeNameWrapper.TransactOpts, to, tokenId)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) ExtendExpiry(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "extendExpiry", parentNode, labelhash, expiry)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) ExtendExpiry(parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.ExtendExpiry(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, expiry)
}

// ExtendExpiry is a paid mutator transaction binding the contract method 0x6e5d6ad2.
//
// Solidity: function extendExpiry(bytes32 parentNode, bytes32 labelhash, uint64 expiry) returns(uint64)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) ExtendExpiry(parentNode [32]byte, labelhash [32]byte, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.ExtendExpiry(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, expiry)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) OnERC721Received(opts *bind.TransactOpts, to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "onERC721Received", to, arg1, tokenId, data)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) OnERC721Received(to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.OnERC721Received(&_AnytypeNameWrapper.TransactOpts, to, arg1, tokenId, data)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address to, address , uint256 tokenId, bytes data) returns(bytes4)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) OnERC721Received(to common.Address, arg1 common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.OnERC721Received(&_AnytypeNameWrapper.TransactOpts, to, arg1, tokenId, data)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) RecoverFunds(opts *bind.TransactOpts, _token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "recoverFunds", _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RecoverFunds(&_AnytypeNameWrapper.TransactOpts, _token, _to, _amount)
}

// RecoverFunds is a paid mutator transaction binding the contract method 0x5d3590d5.
//
// Solidity: function recoverFunds(address _token, address _to, uint256 _amount) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) RecoverFunds(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RecoverFunds(&_AnytypeNameWrapper.TransactOpts, _token, _to, _amount)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) RegisterAndWrapETH2LD(opts *bind.TransactOpts, label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "registerAndWrapETH2LD", label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) RegisterAndWrapETH2LD(label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RegisterAndWrapETH2LD(&_AnytypeNameWrapper.TransactOpts, label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// RegisterAndWrapETH2LD is a paid mutator transaction binding the contract method 0xa4014982.
//
// Solidity: function registerAndWrapETH2LD(string label, address wrappedOwner, uint256 duration, address resolver, uint16 ownerControlledFuses) returns(uint256 registrarExpiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) RegisterAndWrapETH2LD(label string, wrappedOwner common.Address, duration *big.Int, resolver common.Address, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RegisterAndWrapETH2LD(&_AnytypeNameWrapper.TransactOpts, label, wrappedOwner, duration, resolver, ownerControlledFuses)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) Renew(opts *bind.TransactOpts, tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "renew", tokenId, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Renew(tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Renew(&_AnytypeNameWrapper.TransactOpts, tokenId, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 tokenId, uint256 duration) returns(uint256 expires)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) Renew(tokenId *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Renew(&_AnytypeNameWrapper.TransactOpts, tokenId, duration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RenounceOwnership(&_AnytypeNameWrapper.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.RenounceOwnership(&_AnytypeNameWrapper.TransactOpts)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SafeBatchTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "safeBatchTransferFrom", from, to, ids, amounts, data)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SafeBatchTransferFrom(from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SafeBatchTransferFrom(&_AnytypeNameWrapper.TransactOpts, from, to, ids, amounts, data)
}

// SafeBatchTransferFrom is a paid mutator transaction binding the contract method 0x2eb2c2d6.
//
// Solidity: function safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SafeBatchTransferFrom(from common.Address, to common.Address, ids []*big.Int, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SafeBatchTransferFrom(&_AnytypeNameWrapper.TransactOpts, from, to, ids, amounts, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "safeTransferFrom", from, to, id, amount, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SafeTransferFrom(from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SafeTransferFrom(&_AnytypeNameWrapper.TransactOpts, from, to, id, amount, data)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0xf242432a.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SafeTransferFrom(from common.Address, to common.Address, id *big.Int, amount *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SafeTransferFrom(&_AnytypeNameWrapper.TransactOpts, from, to, id, amount, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetApprovalForAll(&_AnytypeNameWrapper.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetApprovalForAll(&_AnytypeNameWrapper.TransactOpts, operator, approved)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetChildFuses(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setChildFuses", parentNode, labelhash, fuses, expiry)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetChildFuses(parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetChildFuses(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, fuses, expiry)
}

// SetChildFuses is a paid mutator transaction binding the contract method 0x33c69ea9.
//
// Solidity: function setChildFuses(bytes32 parentNode, bytes32 labelhash, uint32 fuses, uint64 expiry) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetChildFuses(parentNode [32]byte, labelhash [32]byte, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetChildFuses(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, fuses, expiry)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetController(opts *bind.TransactOpts, controller common.Address, active bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setController", controller, active)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetController(controller common.Address, active bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetController(&_AnytypeNameWrapper.TransactOpts, controller, active)
}

// SetController is a paid mutator transaction binding the contract method 0xe0dba60f.
//
// Solidity: function setController(address controller, bool active) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetController(controller common.Address, active bool) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetController(&_AnytypeNameWrapper.TransactOpts, controller, active)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetFuses(opts *bind.TransactOpts, node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setFuses", node, ownerControlledFuses)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetFuses(node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetFuses(&_AnytypeNameWrapper.TransactOpts, node, ownerControlledFuses)
}

// SetFuses is a paid mutator transaction binding the contract method 0x402906fc.
//
// Solidity: function setFuses(bytes32 node, uint16 ownerControlledFuses) returns(uint32)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetFuses(node [32]byte, ownerControlledFuses uint16) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetFuses(&_AnytypeNameWrapper.TransactOpts, node, ownerControlledFuses)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetMetadataService(opts *bind.TransactOpts, _metadataService common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setMetadataService", _metadataService)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetMetadataService(_metadataService common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetMetadataService(&_AnytypeNameWrapper.TransactOpts, _metadataService)
}

// SetMetadataService is a paid mutator transaction binding the contract method 0x1534e177.
//
// Solidity: function setMetadataService(address _metadataService) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetMetadataService(_metadataService common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetMetadataService(&_AnytypeNameWrapper.TransactOpts, _metadataService)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetRecord(opts *bind.TransactOpts, node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setRecord", node, owner, resolver, ttl)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetRecord(node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetRecord(&_AnytypeNameWrapper.TransactOpts, node, owner, resolver, ttl)
}

// SetRecord is a paid mutator transaction binding the contract method 0xcf408823.
//
// Solidity: function setRecord(bytes32 node, address owner, address resolver, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetRecord(node [32]byte, owner common.Address, resolver common.Address, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetRecord(&_AnytypeNameWrapper.TransactOpts, node, owner, resolver, ttl)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetResolver(opts *bind.TransactOpts, node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setResolver", node, resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetResolver(node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetResolver(&_AnytypeNameWrapper.TransactOpts, node, resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x1896f70a.
//
// Solidity: function setResolver(bytes32 node, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetResolver(node [32]byte, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetResolver(&_AnytypeNameWrapper.TransactOpts, node, resolver)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetSubnodeOwner(opts *bind.TransactOpts, parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setSubnodeOwner", parentNode, label, owner, fuses, expiry)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetSubnodeOwner(parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetSubnodeOwner(&_AnytypeNameWrapper.TransactOpts, parentNode, label, owner, fuses, expiry)
}

// SetSubnodeOwner is a paid mutator transaction binding the contract method 0xc658e086.
//
// Solidity: function setSubnodeOwner(bytes32 parentNode, string label, address owner, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetSubnodeOwner(parentNode [32]byte, label string, owner common.Address, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetSubnodeOwner(&_AnytypeNameWrapper.TransactOpts, parentNode, label, owner, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetSubnodeRecord(opts *bind.TransactOpts, parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setSubnodeRecord", parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetSubnodeRecord(parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetSubnodeRecord(&_AnytypeNameWrapper.TransactOpts, parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetSubnodeRecord is a paid mutator transaction binding the contract method 0x24c1af44.
//
// Solidity: function setSubnodeRecord(bytes32 parentNode, string label, address owner, address resolver, uint64 ttl, uint32 fuses, uint64 expiry) returns(bytes32 node)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetSubnodeRecord(parentNode [32]byte, label string, owner common.Address, resolver common.Address, ttl uint64, fuses uint32, expiry uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetSubnodeRecord(&_AnytypeNameWrapper.TransactOpts, parentNode, label, owner, resolver, ttl, fuses, expiry)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetTTL(opts *bind.TransactOpts, node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setTTL", node, ttl)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetTTL(node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetTTL(&_AnytypeNameWrapper.TransactOpts, node, ttl)
}

// SetTTL is a paid mutator transaction binding the contract method 0x14ab9038.
//
// Solidity: function setTTL(bytes32 node, uint64 ttl) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetTTL(node [32]byte, ttl uint64) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetTTL(&_AnytypeNameWrapper.TransactOpts, node, ttl)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) SetUpgradeContract(opts *bind.TransactOpts, _upgradeAddress common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "setUpgradeContract", _upgradeAddress)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) SetUpgradeContract(_upgradeAddress common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetUpgradeContract(&_AnytypeNameWrapper.TransactOpts, _upgradeAddress)
}

// SetUpgradeContract is a paid mutator transaction binding the contract method 0xb6bcad26.
//
// Solidity: function setUpgradeContract(address _upgradeAddress) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) SetUpgradeContract(_upgradeAddress common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.SetUpgradeContract(&_AnytypeNameWrapper.TransactOpts, _upgradeAddress)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.TransferOwnership(&_AnytypeNameWrapper.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.TransferOwnership(&_AnytypeNameWrapper.TransactOpts, newOwner)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) Unwrap(opts *bind.TransactOpts, parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "unwrap", parentNode, labelhash, controller)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Unwrap(parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Unwrap(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, controller)
}

// Unwrap is a paid mutator transaction binding the contract method 0xd8c9921a.
//
// Solidity: function unwrap(bytes32 parentNode, bytes32 labelhash, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) Unwrap(parentNode [32]byte, labelhash [32]byte, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Unwrap(&_AnytypeNameWrapper.TransactOpts, parentNode, labelhash, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) UnwrapETH2LD(opts *bind.TransactOpts, labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "unwrapETH2LD", labelhash, registrant, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) UnwrapETH2LD(labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.UnwrapETH2LD(&_AnytypeNameWrapper.TransactOpts, labelhash, registrant, controller)
}

// UnwrapETH2LD is a paid mutator transaction binding the contract method 0x8b4dfa75.
//
// Solidity: function unwrapETH2LD(bytes32 labelhash, address registrant, address controller) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) UnwrapETH2LD(labelhash [32]byte, registrant common.Address, controller common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.UnwrapETH2LD(&_AnytypeNameWrapper.TransactOpts, labelhash, registrant, controller)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) Upgrade(opts *bind.TransactOpts, name []byte, extraData []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "upgrade", name, extraData)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Upgrade(name []byte, extraData []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Upgrade(&_AnytypeNameWrapper.TransactOpts, name, extraData)
}

// Upgrade is a paid mutator transaction binding the contract method 0xc93ab3fd.
//
// Solidity: function upgrade(bytes name, bytes extraData) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) Upgrade(name []byte, extraData []byte) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Upgrade(&_AnytypeNameWrapper.TransactOpts, name, extraData)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) Wrap(opts *bind.TransactOpts, name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "wrap", name, wrappedOwner, resolver)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) Wrap(name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Wrap(&_AnytypeNameWrapper.TransactOpts, name, wrappedOwner, resolver)
}

// Wrap is a paid mutator transaction binding the contract method 0xeb8ae530.
//
// Solidity: function wrap(bytes name, address wrappedOwner, address resolver) returns()
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) Wrap(name []byte, wrappedOwner common.Address, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.Wrap(&_AnytypeNameWrapper.TransactOpts, name, wrappedOwner, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactor) WrapETH2LD(opts *bind.TransactOpts, label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.contract.Transact(opts, "wrapETH2LD", label, wrappedOwner, ownerControlledFuses, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperSession) WrapETH2LD(label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.WrapETH2LD(&_AnytypeNameWrapper.TransactOpts, label, wrappedOwner, ownerControlledFuses, resolver)
}

// WrapETH2LD is a paid mutator transaction binding the contract method 0x8cf8b41e.
//
// Solidity: function wrapETH2LD(string label, address wrappedOwner, uint16 ownerControlledFuses, address resolver) returns(uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperTransactorSession) WrapETH2LD(label string, wrappedOwner common.Address, ownerControlledFuses uint16, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeNameWrapper.Contract.WrapETH2LD(&_AnytypeNameWrapper.TransactOpts, label, wrappedOwner, ownerControlledFuses, resolver)
}

// AnytypeNameWrapperApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperApprovalIterator struct {
	Event *AnytypeNameWrapperApproval // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperApproval)
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
		it.Event = new(AnytypeNameWrapperApproval)
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
func (it *AnytypeNameWrapperApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperApproval represents a Approval event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*AnytypeNameWrapperApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperApprovalIterator{contract: _AnytypeNameWrapper.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperApproval)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseApproval(log types.Log) (*AnytypeNameWrapperApproval, error) {
	event := new(AnytypeNameWrapperApproval)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperApprovalForAllIterator struct {
	Event *AnytypeNameWrapperApprovalForAll // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperApprovalForAll)
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
		it.Event = new(AnytypeNameWrapperApprovalForAll)
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
func (it *AnytypeNameWrapperApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperApprovalForAll represents a ApprovalForAll event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperApprovalForAll struct {
	Account  common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterApprovalForAll(opts *bind.FilterOpts, account []common.Address, operator []common.Address) (*AnytypeNameWrapperApprovalForAllIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "ApprovalForAll", accountRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperApprovalForAllIterator{contract: _AnytypeNameWrapper.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperApprovalForAll, account []common.Address, operator []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "ApprovalForAll", accountRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperApprovalForAll)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed account, address indexed operator, bool approved)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseApprovalForAll(log types.Log) (*AnytypeNameWrapperApprovalForAll, error) {
	event := new(AnytypeNameWrapperApprovalForAll)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperControllerChangedIterator is returned from FilterControllerChanged and is used to iterate over the raw logs and unpacked data for ControllerChanged events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperControllerChangedIterator struct {
	Event *AnytypeNameWrapperControllerChanged // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperControllerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperControllerChanged)
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
		it.Event = new(AnytypeNameWrapperControllerChanged)
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
func (it *AnytypeNameWrapperControllerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperControllerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperControllerChanged represents a ControllerChanged event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperControllerChanged struct {
	Controller common.Address
	Active     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerChanged is a free log retrieval operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterControllerChanged(opts *bind.FilterOpts, controller []common.Address) (*AnytypeNameWrapperControllerChangedIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperControllerChangedIterator{contract: _AnytypeNameWrapper.contract, event: "ControllerChanged", logs: logs, sub: sub}, nil
}

// WatchControllerChanged is a free log subscription operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchControllerChanged(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperControllerChanged, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "ControllerChanged", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperControllerChanged)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
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

// ParseControllerChanged is a log parse operation binding the contract event 0x4c97694570a07277810af7e5669ffd5f6a2d6b74b6e9a274b8b870fd5114cf87.
//
// Solidity: event ControllerChanged(address indexed controller, bool active)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseControllerChanged(log types.Log) (*AnytypeNameWrapperControllerChanged, error) {
	event := new(AnytypeNameWrapperControllerChanged)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ControllerChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperExpiryExtendedIterator is returned from FilterExpiryExtended and is used to iterate over the raw logs and unpacked data for ExpiryExtended events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperExpiryExtendedIterator struct {
	Event *AnytypeNameWrapperExpiryExtended // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperExpiryExtendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperExpiryExtended)
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
		it.Event = new(AnytypeNameWrapperExpiryExtended)
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
func (it *AnytypeNameWrapperExpiryExtendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperExpiryExtendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperExpiryExtended represents a ExpiryExtended event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperExpiryExtended struct {
	Node   [32]byte
	Expiry uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterExpiryExtended is a free log retrieval operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterExpiryExtended(opts *bind.FilterOpts, node [][32]byte) (*AnytypeNameWrapperExpiryExtendedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "ExpiryExtended", nodeRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperExpiryExtendedIterator{contract: _AnytypeNameWrapper.contract, event: "ExpiryExtended", logs: logs, sub: sub}, nil
}

// WatchExpiryExtended is a free log subscription operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchExpiryExtended(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperExpiryExtended, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "ExpiryExtended", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperExpiryExtended)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ExpiryExtended", log); err != nil {
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

// ParseExpiryExtended is a log parse operation binding the contract event 0xf675815a0817338f93a7da433f6bd5f5542f1029b11b455191ac96c7f6a9b132.
//
// Solidity: event ExpiryExtended(bytes32 indexed node, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseExpiryExtended(log types.Log) (*AnytypeNameWrapperExpiryExtended, error) {
	event := new(AnytypeNameWrapperExpiryExtended)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "ExpiryExtended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperFusesSetIterator is returned from FilterFusesSet and is used to iterate over the raw logs and unpacked data for FusesSet events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperFusesSetIterator struct {
	Event *AnytypeNameWrapperFusesSet // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperFusesSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperFusesSet)
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
		it.Event = new(AnytypeNameWrapperFusesSet)
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
func (it *AnytypeNameWrapperFusesSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperFusesSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperFusesSet represents a FusesSet event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperFusesSet struct {
	Node  [32]byte
	Fuses uint32
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFusesSet is a free log retrieval operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterFusesSet(opts *bind.FilterOpts, node [][32]byte) (*AnytypeNameWrapperFusesSetIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "FusesSet", nodeRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperFusesSetIterator{contract: _AnytypeNameWrapper.contract, event: "FusesSet", logs: logs, sub: sub}, nil
}

// WatchFusesSet is a free log subscription operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchFusesSet(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperFusesSet, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "FusesSet", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperFusesSet)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "FusesSet", log); err != nil {
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

// ParseFusesSet is a log parse operation binding the contract event 0x39873f00c80f4f94b7bd1594aebcf650f003545b74824d57ddf4939e3ff3a34b.
//
// Solidity: event FusesSet(bytes32 indexed node, uint32 fuses)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseFusesSet(log types.Log) (*AnytypeNameWrapperFusesSet, error) {
	event := new(AnytypeNameWrapperFusesSet)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "FusesSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperNameUnwrappedIterator is returned from FilterNameUnwrapped and is used to iterate over the raw logs and unpacked data for NameUnwrapped events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperNameUnwrappedIterator struct {
	Event *AnytypeNameWrapperNameUnwrapped // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperNameUnwrappedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperNameUnwrapped)
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
		it.Event = new(AnytypeNameWrapperNameUnwrapped)
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
func (it *AnytypeNameWrapperNameUnwrappedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperNameUnwrappedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperNameUnwrapped represents a NameUnwrapped event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperNameUnwrapped struct {
	Node  [32]byte
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterNameUnwrapped is a free log retrieval operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterNameUnwrapped(opts *bind.FilterOpts, node [][32]byte) (*AnytypeNameWrapperNameUnwrappedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "NameUnwrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperNameUnwrappedIterator{contract: _AnytypeNameWrapper.contract, event: "NameUnwrapped", logs: logs, sub: sub}, nil
}

// WatchNameUnwrapped is a free log subscription operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchNameUnwrapped(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperNameUnwrapped, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "NameUnwrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperNameUnwrapped)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "NameUnwrapped", log); err != nil {
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

// ParseNameUnwrapped is a log parse operation binding the contract event 0xee2ba1195c65bcf218a83d874335c6bf9d9067b4c672f3c3bf16cf40de7586c4.
//
// Solidity: event NameUnwrapped(bytes32 indexed node, address owner)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseNameUnwrapped(log types.Log) (*AnytypeNameWrapperNameUnwrapped, error) {
	event := new(AnytypeNameWrapperNameUnwrapped)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "NameUnwrapped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperNameWrappedIterator is returned from FilterNameWrapped and is used to iterate over the raw logs and unpacked data for NameWrapped events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperNameWrappedIterator struct {
	Event *AnytypeNameWrapperNameWrapped // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperNameWrappedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperNameWrapped)
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
		it.Event = new(AnytypeNameWrapperNameWrapped)
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
func (it *AnytypeNameWrapperNameWrappedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperNameWrappedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperNameWrapped represents a NameWrapped event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperNameWrapped struct {
	Node   [32]byte
	Name   []byte
	Owner  common.Address
	Fuses  uint32
	Expiry uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNameWrapped is a free log retrieval operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterNameWrapped(opts *bind.FilterOpts, node [][32]byte) (*AnytypeNameWrapperNameWrappedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "NameWrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperNameWrappedIterator{contract: _AnytypeNameWrapper.contract, event: "NameWrapped", logs: logs, sub: sub}, nil
}

// WatchNameWrapped is a free log subscription operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchNameWrapped(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperNameWrapped, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "NameWrapped", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperNameWrapped)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "NameWrapped", log); err != nil {
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

// ParseNameWrapped is a log parse operation binding the contract event 0x8ce7013e8abebc55c3890a68f5a27c67c3f7efa64e584de5fb22363c606fd340.
//
// Solidity: event NameWrapped(bytes32 indexed node, bytes name, address owner, uint32 fuses, uint64 expiry)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseNameWrapped(log types.Log) (*AnytypeNameWrapperNameWrapped, error) {
	event := new(AnytypeNameWrapperNameWrapped)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "NameWrapped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperOwnershipTransferredIterator struct {
	Event *AnytypeNameWrapperOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperOwnershipTransferred)
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
		it.Event = new(AnytypeNameWrapperOwnershipTransferred)
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
func (it *AnytypeNameWrapperOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperOwnershipTransferred represents a OwnershipTransferred event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AnytypeNameWrapperOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperOwnershipTransferredIterator{contract: _AnytypeNameWrapper.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperOwnershipTransferred)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseOwnershipTransferred(log types.Log) (*AnytypeNameWrapperOwnershipTransferred, error) {
	event := new(AnytypeNameWrapperOwnershipTransferred)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperTransferBatchIterator is returned from FilterTransferBatch and is used to iterate over the raw logs and unpacked data for TransferBatch events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperTransferBatchIterator struct {
	Event *AnytypeNameWrapperTransferBatch // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperTransferBatchIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperTransferBatch)
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
		it.Event = new(AnytypeNameWrapperTransferBatch)
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
func (it *AnytypeNameWrapperTransferBatchIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperTransferBatchIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperTransferBatch represents a TransferBatch event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperTransferBatch struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Ids      []*big.Int
	Values   []*big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTransferBatch is a free log retrieval operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterTransferBatch(opts *bind.FilterOpts, operator []common.Address, from []common.Address, to []common.Address) (*AnytypeNameWrapperTransferBatchIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "TransferBatch", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperTransferBatchIterator{contract: _AnytypeNameWrapper.contract, event: "TransferBatch", logs: logs, sub: sub}, nil
}

// WatchTransferBatch is a free log subscription operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchTransferBatch(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperTransferBatch, operator []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "TransferBatch", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperTransferBatch)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "TransferBatch", log); err != nil {
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

// ParseTransferBatch is a log parse operation binding the contract event 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb.
//
// Solidity: event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseTransferBatch(log types.Log) (*AnytypeNameWrapperTransferBatch, error) {
	event := new(AnytypeNameWrapperTransferBatch)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "TransferBatch", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperTransferSingleIterator is returned from FilterTransferSingle and is used to iterate over the raw logs and unpacked data for TransferSingle events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperTransferSingleIterator struct {
	Event *AnytypeNameWrapperTransferSingle // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperTransferSingleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperTransferSingle)
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
		it.Event = new(AnytypeNameWrapperTransferSingle)
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
func (it *AnytypeNameWrapperTransferSingleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperTransferSingleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperTransferSingle represents a TransferSingle event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperTransferSingle struct {
	Operator common.Address
	From     common.Address
	To       common.Address
	Id       *big.Int
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTransferSingle is a free log retrieval operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterTransferSingle(opts *bind.FilterOpts, operator []common.Address, from []common.Address, to []common.Address) (*AnytypeNameWrapperTransferSingleIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "TransferSingle", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperTransferSingleIterator{contract: _AnytypeNameWrapper.contract, event: "TransferSingle", logs: logs, sub: sub}, nil
}

// WatchTransferSingle is a free log subscription operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchTransferSingle(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperTransferSingle, operator []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "TransferSingle", operatorRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperTransferSingle)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "TransferSingle", log); err != nil {
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

// ParseTransferSingle is a log parse operation binding the contract event 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62.
//
// Solidity: event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseTransferSingle(log types.Log) (*AnytypeNameWrapperTransferSingle, error) {
	event := new(AnytypeNameWrapperTransferSingle)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "TransferSingle", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeNameWrapperURIIterator is returned from FilterURI and is used to iterate over the raw logs and unpacked data for URI events raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperURIIterator struct {
	Event *AnytypeNameWrapperURI // Event containing the contract specifics and raw log

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
func (it *AnytypeNameWrapperURIIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeNameWrapperURI)
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
		it.Event = new(AnytypeNameWrapperURI)
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
func (it *AnytypeNameWrapperURIIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeNameWrapperURIIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeNameWrapperURI represents a URI event raised by the AnytypeNameWrapper contract.
type AnytypeNameWrapperURI struct {
	Value string
	Id    *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterURI is a free log retrieval operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) FilterURI(opts *bind.FilterOpts, id []*big.Int) (*AnytypeNameWrapperURIIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.FilterLogs(opts, "URI", idRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeNameWrapperURIIterator{contract: _AnytypeNameWrapper.contract, event: "URI", logs: logs, sub: sub}, nil
}

// WatchURI is a free log subscription operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) WatchURI(opts *bind.WatchOpts, sink chan<- *AnytypeNameWrapperURI, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _AnytypeNameWrapper.contract.WatchLogs(opts, "URI", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeNameWrapperURI)
				if err := _AnytypeNameWrapper.contract.UnpackLog(event, "URI", log); err != nil {
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

// ParseURI is a log parse operation binding the contract event 0x6bb7ff708619ba0610cba295a58592e0451dee2622938c8755667688daf3529b.
//
// Solidity: event URI(string value, uint256 indexed id)
func (_AnytypeNameWrapper *AnytypeNameWrapperFilterer) ParseURI(log types.Log) (*AnytypeNameWrapperURI, error) {
	event := new(AnytypeNameWrapperURI)
	if err := _AnytypeNameWrapper.contract.UnpackLog(event, "URI", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
