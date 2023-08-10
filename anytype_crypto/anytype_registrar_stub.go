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

// AnytypeRegistrarImplementationMetaData contains all meta data concerning the AnytypeRegistrarImplementation contract.
var AnytypeRegistrarImplementationMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"_ens\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_baseNode\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"ControllerAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"ControllerRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expires\",\"type\":\"uint256\"}],\"name\":\"NameRenewed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"GRACE_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"addController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"available\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"baseNode\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"controllers\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"nameExpires\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"reclaim\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"register\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"registerOnly\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"removeController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"renew\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"setResolver\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceID\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b506040516200232138038062002321833981016040819052620000349162000109565b60408051602080820183526000808352835191820190935282815290916200005d8382620001ea565b5060016200006c8282620001ea565b5050506200008962000083620000b360201b60201c565b620000b7565b600880546001600160a01b0319166001600160a01b039390931692909217909155600955620002b6565b3390565b600680546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b600080604083850312156200011d57600080fd5b82516001600160a01b03811681146200013557600080fd5b6020939093015192949293505050565b634e487b7160e01b600052604160045260246000fd5b600181811c908216806200017057607f821691505b6020821081036200019157634e487b7160e01b600052602260045260246000fd5b50919050565b601f821115620001e557600081815260208120601f850160051c81016020861015620001c05750805b601f850160051c820191505b81811015620001e157828155600101620001cc565b5050505b505050565b81516001600160401b0381111562000206576200020662000145565b6200021e816200021784546200015b565b8462000197565b602080601f8311600181146200025657600084156200023d5750858301515b600019600386901b1c1916600185901b178555620001e1565b600085815260208120601f198616915b82811015620002875788860151825594840194600190910190840162000266565b5085821015620002a65787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b61205b80620002c66000396000f3fe608060405234801561001057600080fd5b50600436106101cf5760003560e01c806395d89b4111610104578063c87b56dd116100a2578063e985e9c511610071578063e985e9c5146103e0578063f2fde38b1461041c578063f6a74ed71461042f578063fca247ac1461044257600080fd5b8063c87b56dd14610381578063d6e4fa8614610394578063da8c229e146103b4578063ddf7fcb0146103d757600080fd5b8063a7fc7a07116100de578063a7fc7a071461033e578063b88d4fde14610351578063c1a287e214610364578063c475abff1461036e57600080fd5b806395d89b411461031057806396e494e814610318578063a22cb4651461032b57600080fd5b80633f15457f116101715780636352211e1161014b5780636352211e146102d157806370a08231146102e4578063715018a6146102f75780638da5cb5b146102ff57600080fd5b80633f15457f1461029857806342842e0e146102ab5780634e543b26146102be57600080fd5b8063095ea7b3116101ad578063095ea7b31461023c5780630e297b451461025157806323b872dd1461027257806328ed4f6c1461028557600080fd5b806301ffc9a7146101d457806306fdde03146101fc578063081812fc14610211575b600080fd5b6101e76101e2366004611be9565b610455565b60405190151581526020015b60405180910390f35b6102046104f2565b6040516101f39190611c56565b61022461021f366004611c69565b610584565b6040516001600160a01b0390911681526020016101f3565b61024f61024a366004611c97565b6105ab565b005b61026461025f366004611cc3565b6106e1565b6040519081526020016101f3565b61024f610280366004611cfb565b6106f8565b61024f610293366004611d2b565b61077f565b600854610224906001600160a01b031681565b61024f6102b9366004611cfb565b610898565b61024f6102cc366004611d5b565b6108b3565b6102246102df366004611c69565b610941565b6102646102f2366004611d5b565b610964565b61024f6109fe565b6006546001600160a01b0316610224565b610204610a12565b6101e7610326366004611c69565b610a21565b61024f610339366004611d78565b610a47565b61024f61034c366004611d5b565b610a56565b61024f61035f366004611dc1565b610aaa565b6102646276a70081565b61026461037c366004611ea1565b610b38565b61020461038f366004611c69565b610cc9565b6102646103a2366004611c69565b60009081526007602052604090205490565b6101e76103c2366004611d5b565b600a6020526000908152604090205460ff1681565b61026460095481565b6101e76103ee366004611ec3565b6001600160a01b03918216600090815260056020908152604080832093909416825291909152205460ff1690565b61024f61042a366004611d5b565b610d3d565b61024f61043d366004611d5b565b610dcd565b610264610450366004611cc3565b610e1e565b60006001600160e01b031982167f01ffc9a70000000000000000000000000000000000000000000000000000000014806104b857506001600160e01b031982167f80ac58cd00000000000000000000000000000000000000000000000000000000145b806104ec57506001600160e01b031982167f28ed4f6c00000000000000000000000000000000000000000000000000000000145b92915050565b60606000805461050190611ef1565b80601f016020809104026020016040519081016040528092919081815260200182805461052d90611ef1565b801561057a5780601f1061054f5761010080835404028352916020019161057a565b820191906000526020600020905b81548152906001019060200180831161055d57829003601f168201915b5050505050905090565b600061058f82610e2d565b506000908152600460205260409020546001600160a01b031690565b60006105b682610e91565b9050806001600160a01b0316836001600160a01b0316036106445760405162461bcd60e51b815260206004820152602160248201527f4552433732313a20617070726f76616c20746f2063757272656e74206f776e6560448201527f720000000000000000000000000000000000000000000000000000000000000060648201526084015b60405180910390fd5b336001600160a01b0382161480610660575061066081336103ee565b6106d25760405162461bcd60e51b815260206004820152603d60248201527f4552433732313a20617070726f76652063616c6c6572206973206e6f7420746f60448201527f6b656e206f776e6572206f7220617070726f76656420666f7220616c6c000000606482015260840161063b565b6106dc8383610ef6565b505050565b60006106f08484846000610f71565b949350505050565b6107023382611181565b6107745760405162461bcd60e51b815260206004820152602d60248201527f4552433732313a2063616c6c6572206973206e6f7420746f6b656e206f776e6560448201527f72206f7220617070726f76656400000000000000000000000000000000000000606482015260840161063b565b6106dc8383836111fc565b6008546009546040516302571be360e01b8152600481019190915230916001600160a01b0316906302571be390602401602060405180830381865afa1580156107cc573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107f09190611f2b565b6001600160a01b03161461080357600080fd5b61080d3383611181565b61081657600080fd5b6008546009546040516306ab592360e01b81526004810191909152602481018490526001600160a01b038381166044830152909116906306ab5923906064016020604051808303816000875af1158015610874573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106dc9190611f48565b6106dc83838360405180602001604052806000815250610aaa565b6108bb61140f565b6008546009546040517f1896f70a00000000000000000000000000000000000000000000000000000000815260048101919091526001600160a01b03838116602483015290911690631896f70a90604401600060405180830381600087803b15801561092657600080fd5b505af115801561093a573d6000803e3d6000fd5b5050505050565b600081815260076020526040812054421061095b57600080fd5b6104ec82610e91565b60006001600160a01b0382166109e25760405162461bcd60e51b815260206004820152602960248201527f4552433732313a2061646472657373207a65726f206973206e6f74206120766160448201527f6c6964206f776e65720000000000000000000000000000000000000000000000606482015260840161063b565b506001600160a01b031660009081526003602052604090205490565b610a0661140f565b610a106000611469565b565b60606001805461050190611ef1565b6000818152600760205260408120544290610a40906276a70090611f77565b1092915050565b610a523383836114c8565b5050565b610a5e61140f565b6001600160a01b0381166000818152600a6020526040808220805460ff19166001179055517f0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d74749190a250565b610ab43383611181565b610b265760405162461bcd60e51b815260206004820152602d60248201527f4552433732313a2063616c6c6572206973206e6f7420746f6b656e206f776e6560448201527f72206f7220617070726f76656400000000000000000000000000000000000000606482015260840161063b565b610b3284848484611596565b50505050565b6008546009546040516302571be360e01b8152600481019190915260009130916001600160a01b03909116906302571be390602401602060405180830381865afa158015610b8a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610bae9190611f2b565b6001600160a01b031614610bc157600080fd5b336000908152600a602052604090205460ff16610bdd57600080fd5b6000838152600760205260409020544290610bfc906276a70090611f77565b1015610c0757600080fd5b610c146276a70083611f77565b6000848152600760205260409020546276a70090610c33908590611f77565b610c3d9190611f77565b11610c4757600080fd5b60008381526007602052604081208054849290610c65908490611f77565b90915550506000838152600760205260409081902054905184917f9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd691610cad91815260200190565b60405180910390a2505060009081526007602052604090205490565b6060610cd482610e2d565b6000610ceb60408051602081019091526000815290565b90506000815111610d0b5760405180602001604052806000815250610d36565b80610d158461161f565b604051602001610d26929190611f8a565b6040516020818303038152906040525b9392505050565b610d4561140f565b6001600160a01b038116610dc15760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f6464726573730000000000000000000000000000000000000000000000000000606482015260840161063b565b610dca81611469565b50565b610dd561140f565b6001600160a01b0381166000818152600a6020526040808220805460ff19169055517f33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e811139190a250565b60006106f08484846001610f71565b6000818152600260205260409020546001600160a01b0316610dca5760405162461bcd60e51b815260206004820152601860248201527f4552433732313a20696e76616c696420746f6b656e2049440000000000000000604482015260640161063b565b6000818152600260205260408120546001600160a01b0316806104ec5760405162461bcd60e51b815260206004820152601860248201527f4552433732313a20696e76616c696420746f6b656e2049440000000000000000604482015260640161063b565b6000818152600460205260409020805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0384169081179091558190610f3882610e91565b6001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560405160405180910390a45050565b6008546009546040516302571be360e01b8152600481019190915260009130916001600160a01b03909116906302571be390602401602060405180830381865afa158015610fc3573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610fe79190611f2b565b6001600160a01b031614610ffa57600080fd5b336000908152600a602052604090205460ff1661101657600080fd5b61101f85610a21565b61102857600080fd5b6110356276a70042611f77565b6276a7006110438542611f77565b61104d9190611f77565b1161105757600080fd5b6110618342611f77565b6000868152600760209081526040808320939093556002905220546001600160a01b03161561109357611093856116bf565b61109d848661176f565b8115611127576008546009546040516306ab592360e01b81526004810191909152602481018790526001600160a01b038681166044830152909116906306ab5923906064016020604051808303816000875af1158015611101573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906111259190611f48565b505b6001600160a01b038416857fb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d961115d8642611f77565b60405190815260200160405180910390a36111788342611f77565b95945050505050565b60008061118d83610941565b9050806001600160a01b0316846001600160a01b031614806111c85750836001600160a01b03166111bd84610584565b6001600160a01b0316145b806106f057506001600160a01b0380821660009081526005602090815260408083209388168352929052205460ff166106f0565b826001600160a01b031661120f82610e91565b6001600160a01b0316146112735760405162461bcd60e51b815260206004820152602560248201527f4552433732313a207472616e736665722066726f6d20696e636f72726563742060448201526437bbb732b960d91b606482015260840161063b565b6001600160a01b0382166112ee5760405162461bcd60e51b8152602060048201526024808201527f4552433732313a207472616e7366657220746f20746865207a65726f2061646460448201527f7265737300000000000000000000000000000000000000000000000000000000606482015260840161063b565b6112fb8383836001611915565b826001600160a01b031661130e82610e91565b6001600160a01b0316146113725760405162461bcd60e51b815260206004820152602560248201527f4552433732313a207472616e736665722066726f6d20696e636f72726563742060448201526437bbb732b960d91b606482015260840161063b565b6000818152600460209081526040808320805473ffffffffffffffffffffffffffffffffffffffff199081169091556001600160a01b0387811680865260038552838620805460001901905590871680865283862080546001019055868652600290945282852080549092168417909155905184937fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef91a4505050565b6006546001600160a01b03163314610a105760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161063b565b600680546001600160a01b0383811673ffffffffffffffffffffffffffffffffffffffff19831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b816001600160a01b0316836001600160a01b0316036115295760405162461bcd60e51b815260206004820152601960248201527f4552433732313a20617070726f766520746f2063616c6c657200000000000000604482015260640161063b565b6001600160a01b03838116600081815260056020908152604080832094871680845294825291829020805460ff191686151590811790915591519182527f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31910160405180910390a3505050565b6115a18484846111fc565b6115ad8484848461199d565b610b325760405162461bcd60e51b815260206004820152603260248201527f4552433732313a207472616e7366657220746f206e6f6e20455243373231526560448201527f63656976657220696d706c656d656e7465720000000000000000000000000000606482015260840161063b565b6060600061162c83611af1565b600101905060008167ffffffffffffffff81111561164c5761164c611dab565b6040519080825280601f01601f191660200182016040528015611676576020820181803683370190505b5090508181016020015b600019017f3031323334353637383961626364656600000000000000000000000000000000600a86061a8153600a850494508461168057509392505050565b60006116ca82610e91565b90506116da816000846001611915565b6116e382610e91565b6000838152600460209081526040808320805473ffffffffffffffffffffffffffffffffffffffff199081169091556001600160a01b0385168085526003845282852080546000190190558785526002909352818420805490911690555192935084927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef908390a45050565b6001600160a01b0382166117c55760405162461bcd60e51b815260206004820181905260248201527f4552433732313a206d696e7420746f20746865207a65726f2061646472657373604482015260640161063b565b6000818152600260205260409020546001600160a01b03161561182a5760405162461bcd60e51b815260206004820152601c60248201527f4552433732313a20746f6b656e20616c7265616479206d696e74656400000000604482015260640161063b565b611838600083836001611915565b6000818152600260205260409020546001600160a01b03161561189d5760405162461bcd60e51b815260206004820152601c60248201527f4552433732313a20746f6b656e20616c7265616479206d696e74656400000000604482015260640161063b565b6001600160a01b0382166000818152600360209081526040808320805460010190558483526002909152808220805473ffffffffffffffffffffffffffffffffffffffff19168417905551839291907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef908290a45050565b6001811115610b32576001600160a01b0384161561195b576001600160a01b03841660009081526003602052604081208054839290611955908490611fb9565b90915550505b6001600160a01b03831615610b32576001600160a01b03831660009081526003602052604081208054839290611992908490611f77565b909155505050505050565b60006001600160a01b0384163b15611ae957604051630a85bd0160e11b81526001600160a01b0385169063150b7a02906119e1903390899088908890600401611fcc565b6020604051808303816000875af1925050508015611a1c575060408051601f3d908101601f19168201909252611a1991810190612008565b60015b611acf573d808015611a4a576040519150601f19603f3d011682016040523d82523d6000602084013e611a4f565b606091505b508051600003611ac75760405162461bcd60e51b815260206004820152603260248201527f4552433732313a207472616e7366657220746f206e6f6e20455243373231526560448201527f63656976657220696d706c656d656e7465720000000000000000000000000000606482015260840161063b565b805181602001fd5b6001600160e01b031916630a85bd0160e11b1490506106f0565b5060016106f0565b6000807a184f03e93ff9f4daa797ed6e38ed64bf6a1f0100000000000000008310611b3a577a184f03e93ff9f4daa797ed6e38ed64bf6a1f010000000000000000830492506040015b6d04ee2d6d415b85acef81000000008310611b66576d04ee2d6d415b85acef8100000000830492506020015b662386f26fc100008310611b8457662386f26fc10000830492506010015b6305f5e1008310611b9c576305f5e100830492506008015b6127108310611bb057612710830492506004015b60648310611bc2576064830492506002015b600a83106104ec5760010192915050565b6001600160e01b031981168114610dca57600080fd5b600060208284031215611bfb57600080fd5b8135610d3681611bd3565b60005b83811015611c21578181015183820152602001611c09565b50506000910152565b60008151808452611c42816020860160208601611c06565b601f01601f19169290920160200192915050565b602081526000610d366020830184611c2a565b600060208284031215611c7b57600080fd5b5035919050565b6001600160a01b0381168114610dca57600080fd5b60008060408385031215611caa57600080fd5b8235611cb581611c82565b946020939093013593505050565b600080600060608486031215611cd857600080fd5b833592506020840135611cea81611c82565b929592945050506040919091013590565b600080600060608486031215611d1057600080fd5b8335611d1b81611c82565b92506020840135611cea81611c82565b60008060408385031215611d3e57600080fd5b823591506020830135611d5081611c82565b809150509250929050565b600060208284031215611d6d57600080fd5b8135610d3681611c82565b60008060408385031215611d8b57600080fd5b8235611d9681611c82565b915060208301358015158114611d5057600080fd5b634e487b7160e01b600052604160045260246000fd5b60008060008060808587031215611dd757600080fd5b8435611de281611c82565b93506020850135611df281611c82565b925060408501359150606085013567ffffffffffffffff80821115611e1657600080fd5b818701915087601f830112611e2a57600080fd5b813581811115611e3c57611e3c611dab565b604051601f8201601f19908116603f01168101908382118183101715611e6457611e64611dab565b816040528281528a6020848701011115611e7d57600080fd5b82602086016020830137600060208483010152809550505050505092959194509250565b60008060408385031215611eb457600080fd5b50508035926020909101359150565b60008060408385031215611ed657600080fd5b8235611ee181611c82565b91506020830135611d5081611c82565b600181811c90821680611f0557607f821691505b602082108103611f2557634e487b7160e01b600052602260045260246000fd5b50919050565b600060208284031215611f3d57600080fd5b8151610d3681611c82565b600060208284031215611f5a57600080fd5b5051919050565b634e487b7160e01b600052601160045260246000fd5b808201808211156104ec576104ec611f61565b60008351611f9c818460208801611c06565b835190830190611fb0818360208801611c06565b01949350505050565b818103818111156104ec576104ec611f61565b60006001600160a01b03808716835280861660208401525083604083015260806060830152611ffe6080830184611c2a565b9695505050505050565b60006020828403121561201a57600080fd5b8151610d3681611bd356fea26469706673582212204d61b743acc45a3ef9ec751b6f2b482085732ce5b2391bdc1ae31da5ccceafd764736f6c63430008110033",
}

// AnytypeRegistrarImplementationABI is the input ABI used to generate the binding from.
// Deprecated: Use AnytypeRegistrarImplementationMetaData.ABI instead.
var AnytypeRegistrarImplementationABI = AnytypeRegistrarImplementationMetaData.ABI

// AnytypeRegistrarImplementationBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AnytypeRegistrarImplementationMetaData.Bin instead.
var AnytypeRegistrarImplementationBin = AnytypeRegistrarImplementationMetaData.Bin

// DeployAnytypeRegistrarImplementation deploys a new Ethereum contract, binding an instance of AnytypeRegistrarImplementation to it.
func DeployAnytypeRegistrarImplementation(auth *bind.TransactOpts, backend bind.ContractBackend, _ens common.Address, _baseNode [32]byte) (common.Address, *types.Transaction, *AnytypeRegistrarImplementation, error) {
	parsed, err := AnytypeRegistrarImplementationMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AnytypeRegistrarImplementationBin), backend, _ens, _baseNode)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AnytypeRegistrarImplementation{AnytypeRegistrarImplementationCaller: AnytypeRegistrarImplementationCaller{contract: contract}, AnytypeRegistrarImplementationTransactor: AnytypeRegistrarImplementationTransactor{contract: contract}, AnytypeRegistrarImplementationFilterer: AnytypeRegistrarImplementationFilterer{contract: contract}}, nil
}

// AnytypeRegistrarImplementation is an auto generated Go binding around an Ethereum contract.
type AnytypeRegistrarImplementation struct {
	AnytypeRegistrarImplementationCaller     // Read-only binding to the contract
	AnytypeRegistrarImplementationTransactor // Write-only binding to the contract
	AnytypeRegistrarImplementationFilterer   // Log filterer for contract events
}

// AnytypeRegistrarImplementationCaller is an auto generated read-only Go binding around an Ethereum contract.
type AnytypeRegistrarImplementationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarImplementationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AnytypeRegistrarImplementationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarImplementationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AnytypeRegistrarImplementationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnytypeRegistrarImplementationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AnytypeRegistrarImplementationSession struct {
	Contract     *AnytypeRegistrarImplementation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                   // Call options to use throughout this session
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// AnytypeRegistrarImplementationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AnytypeRegistrarImplementationCallerSession struct {
	Contract *AnytypeRegistrarImplementationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                         // Call options to use throughout this session
}

// AnytypeRegistrarImplementationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AnytypeRegistrarImplementationTransactorSession struct {
	Contract     *AnytypeRegistrarImplementationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                         // Transaction auth options to use throughout this session
}

// AnytypeRegistrarImplementationRaw is an auto generated low-level Go binding around an Ethereum contract.
type AnytypeRegistrarImplementationRaw struct {
	Contract *AnytypeRegistrarImplementation // Generic contract binding to access the raw methods on
}

// AnytypeRegistrarImplementationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AnytypeRegistrarImplementationCallerRaw struct {
	Contract *AnytypeRegistrarImplementationCaller // Generic read-only contract binding to access the raw methods on
}

// AnytypeRegistrarImplementationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AnytypeRegistrarImplementationTransactorRaw struct {
	Contract *AnytypeRegistrarImplementationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAnytypeRegistrarImplementation creates a new instance of AnytypeRegistrarImplementation, bound to a specific deployed contract.
func NewAnytypeRegistrarImplementation(address common.Address, backend bind.ContractBackend) (*AnytypeRegistrarImplementation, error) {
	contract, err := bindAnytypeRegistrarImplementation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementation{AnytypeRegistrarImplementationCaller: AnytypeRegistrarImplementationCaller{contract: contract}, AnytypeRegistrarImplementationTransactor: AnytypeRegistrarImplementationTransactor{contract: contract}, AnytypeRegistrarImplementationFilterer: AnytypeRegistrarImplementationFilterer{contract: contract}}, nil
}

// NewAnytypeRegistrarImplementationCaller creates a new read-only instance of AnytypeRegistrarImplementation, bound to a specific deployed contract.
func NewAnytypeRegistrarImplementationCaller(address common.Address, caller bind.ContractCaller) (*AnytypeRegistrarImplementationCaller, error) {
	contract, err := bindAnytypeRegistrarImplementation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationCaller{contract: contract}, nil
}

// NewAnytypeRegistrarImplementationTransactor creates a new write-only instance of AnytypeRegistrarImplementation, bound to a specific deployed contract.
func NewAnytypeRegistrarImplementationTransactor(address common.Address, transactor bind.ContractTransactor) (*AnytypeRegistrarImplementationTransactor, error) {
	contract, err := bindAnytypeRegistrarImplementation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationTransactor{contract: contract}, nil
}

// NewAnytypeRegistrarImplementationFilterer creates a new log filterer instance of AnytypeRegistrarImplementation, bound to a specific deployed contract.
func NewAnytypeRegistrarImplementationFilterer(address common.Address, filterer bind.ContractFilterer) (*AnytypeRegistrarImplementationFilterer, error) {
	contract, err := bindAnytypeRegistrarImplementation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationFilterer{contract: contract}, nil
}

// bindAnytypeRegistrarImplementation binds a generic wrapper to an already deployed contract.
func bindAnytypeRegistrarImplementation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AnytypeRegistrarImplementationABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeRegistrarImplementation.Contract.AnytypeRegistrarImplementationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.AnytypeRegistrarImplementationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.AnytypeRegistrarImplementationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnytypeRegistrarImplementation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.contract.Transact(opts, method, params...)
}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) GRACEPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "GRACE_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) GRACEPERIOD() (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.GRACEPERIOD(&_AnytypeRegistrarImplementation.CallOpts)
}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) GRACEPERIOD() (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.GRACEPERIOD(&_AnytypeRegistrarImplementation.CallOpts)
}

// Available is a free data retrieval call binding the contract method 0x96e494e8.
//
// Solidity: function available(uint256 id) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Available(opts *bind.CallOpts, id *big.Int) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "available", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Available is a free data retrieval call binding the contract method 0x96e494e8.
//
// Solidity: function available(uint256 id) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Available(id *big.Int) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.Available(&_AnytypeRegistrarImplementation.CallOpts, id)
}

// Available is a free data retrieval call binding the contract method 0x96e494e8.
//
// Solidity: function available(uint256 id) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Available(id *big.Int) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.Available(&_AnytypeRegistrarImplementation.CallOpts, id)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.BalanceOf(&_AnytypeRegistrarImplementation.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.BalanceOf(&_AnytypeRegistrarImplementation.CallOpts, owner)
}

// BaseNode is a free data retrieval call binding the contract method 0xddf7fcb0.
//
// Solidity: function baseNode() view returns(bytes32)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) BaseNode(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "baseNode")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BaseNode is a free data retrieval call binding the contract method 0xddf7fcb0.
//
// Solidity: function baseNode() view returns(bytes32)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) BaseNode() ([32]byte, error) {
	return _AnytypeRegistrarImplementation.Contract.BaseNode(&_AnytypeRegistrarImplementation.CallOpts)
}

// BaseNode is a free data retrieval call binding the contract method 0xddf7fcb0.
//
// Solidity: function baseNode() view returns(bytes32)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) BaseNode() ([32]byte, error) {
	return _AnytypeRegistrarImplementation.Contract.BaseNode(&_AnytypeRegistrarImplementation.CallOpts)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Controllers(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "controllers", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Controllers(arg0 common.Address) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.Controllers(&_AnytypeRegistrarImplementation.CallOpts, arg0)
}

// Controllers is a free data retrieval call binding the contract method 0xda8c229e.
//
// Solidity: function controllers(address ) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Controllers(arg0 common.Address) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.Controllers(&_AnytypeRegistrarImplementation.CallOpts, arg0)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Ens() (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.Ens(&_AnytypeRegistrarImplementation.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Ens() (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.Ens(&_AnytypeRegistrarImplementation.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.GetApproved(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.GetApproved(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.IsApprovedForAll(&_AnytypeRegistrarImplementation.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.IsApprovedForAll(&_AnytypeRegistrarImplementation.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Name() (string, error) {
	return _AnytypeRegistrarImplementation.Contract.Name(&_AnytypeRegistrarImplementation.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Name() (string, error) {
	return _AnytypeRegistrarImplementation.Contract.Name(&_AnytypeRegistrarImplementation.CallOpts)
}

// NameExpires is a free data retrieval call binding the contract method 0xd6e4fa86.
//
// Solidity: function nameExpires(uint256 id) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) NameExpires(opts *bind.CallOpts, id *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "nameExpires", id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NameExpires is a free data retrieval call binding the contract method 0xd6e4fa86.
//
// Solidity: function nameExpires(uint256 id) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) NameExpires(id *big.Int) (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.NameExpires(&_AnytypeRegistrarImplementation.CallOpts, id)
}

// NameExpires is a free data retrieval call binding the contract method 0xd6e4fa86.
//
// Solidity: function nameExpires(uint256 id) view returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) NameExpires(id *big.Int) (*big.Int, error) {
	return _AnytypeRegistrarImplementation.Contract.NameExpires(&_AnytypeRegistrarImplementation.CallOpts, id)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Owner() (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.Owner(&_AnytypeRegistrarImplementation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Owner() (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.Owner(&_AnytypeRegistrarImplementation.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.OwnerOf(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AnytypeRegistrarImplementation.Contract.OwnerOf(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) SupportsInterface(opts *bind.CallOpts, interfaceID [4]byte) (bool, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "supportsInterface", interfaceID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.SupportsInterface(&_AnytypeRegistrarImplementation.CallOpts, interfaceID)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) view returns(bool)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _AnytypeRegistrarImplementation.Contract.SupportsInterface(&_AnytypeRegistrarImplementation.CallOpts, interfaceID)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Symbol() (string, error) {
	return _AnytypeRegistrarImplementation.Contract.Symbol(&_AnytypeRegistrarImplementation.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) Symbol() (string, error) {
	return _AnytypeRegistrarImplementation.Contract.Symbol(&_AnytypeRegistrarImplementation.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _AnytypeRegistrarImplementation.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AnytypeRegistrarImplementation.Contract.TokenURI(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AnytypeRegistrarImplementation.Contract.TokenURI(&_AnytypeRegistrarImplementation.CallOpts, tokenId)
}

// AddController is a paid mutator transaction binding the contract method 0xa7fc7a07.
//
// Solidity: function addController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) AddController(opts *bind.TransactOpts, controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "addController", controller)
}

// AddController is a paid mutator transaction binding the contract method 0xa7fc7a07.
//
// Solidity: function addController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) AddController(controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.AddController(&_AnytypeRegistrarImplementation.TransactOpts, controller)
}

// AddController is a paid mutator transaction binding the contract method 0xa7fc7a07.
//
// Solidity: function addController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) AddController(controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.AddController(&_AnytypeRegistrarImplementation.TransactOpts, controller)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Approve(&_AnytypeRegistrarImplementation.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Approve(&_AnytypeRegistrarImplementation.TransactOpts, to, tokenId)
}

// Reclaim is a paid mutator transaction binding the contract method 0x28ed4f6c.
//
// Solidity: function reclaim(uint256 id, address owner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) Reclaim(opts *bind.TransactOpts, id *big.Int, owner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "reclaim", id, owner)
}

// Reclaim is a paid mutator transaction binding the contract method 0x28ed4f6c.
//
// Solidity: function reclaim(uint256 id, address owner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Reclaim(id *big.Int, owner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Reclaim(&_AnytypeRegistrarImplementation.TransactOpts, id, owner)
}

// Reclaim is a paid mutator transaction binding the contract method 0x28ed4f6c.
//
// Solidity: function reclaim(uint256 id, address owner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) Reclaim(id *big.Int, owner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Reclaim(&_AnytypeRegistrarImplementation.TransactOpts, id, owner)
}

// Register is a paid mutator transaction binding the contract method 0xfca247ac.
//
// Solidity: function register(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) Register(opts *bind.TransactOpts, id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "register", id, owner, duration)
}

// Register is a paid mutator transaction binding the contract method 0xfca247ac.
//
// Solidity: function register(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Register(id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Register(&_AnytypeRegistrarImplementation.TransactOpts, id, owner, duration)
}

// Register is a paid mutator transaction binding the contract method 0xfca247ac.
//
// Solidity: function register(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) Register(id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Register(&_AnytypeRegistrarImplementation.TransactOpts, id, owner, duration)
}

// RegisterOnly is a paid mutator transaction binding the contract method 0x0e297b45.
//
// Solidity: function registerOnly(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) RegisterOnly(opts *bind.TransactOpts, id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "registerOnly", id, owner, duration)
}

// RegisterOnly is a paid mutator transaction binding the contract method 0x0e297b45.
//
// Solidity: function registerOnly(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) RegisterOnly(id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RegisterOnly(&_AnytypeRegistrarImplementation.TransactOpts, id, owner, duration)
}

// RegisterOnly is a paid mutator transaction binding the contract method 0x0e297b45.
//
// Solidity: function registerOnly(uint256 id, address owner, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) RegisterOnly(id *big.Int, owner common.Address, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RegisterOnly(&_AnytypeRegistrarImplementation.TransactOpts, id, owner, duration)
}

// RemoveController is a paid mutator transaction binding the contract method 0xf6a74ed7.
//
// Solidity: function removeController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) RemoveController(opts *bind.TransactOpts, controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "removeController", controller)
}

// RemoveController is a paid mutator transaction binding the contract method 0xf6a74ed7.
//
// Solidity: function removeController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) RemoveController(controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RemoveController(&_AnytypeRegistrarImplementation.TransactOpts, controller)
}

// RemoveController is a paid mutator transaction binding the contract method 0xf6a74ed7.
//
// Solidity: function removeController(address controller) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) RemoveController(controller common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RemoveController(&_AnytypeRegistrarImplementation.TransactOpts, controller)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 id, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) Renew(opts *bind.TransactOpts, id *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "renew", id, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 id, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) Renew(id *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Renew(&_AnytypeRegistrarImplementation.TransactOpts, id, duration)
}

// Renew is a paid mutator transaction binding the contract method 0xc475abff.
//
// Solidity: function renew(uint256 id, uint256 duration) returns(uint256)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) Renew(id *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.Renew(&_AnytypeRegistrarImplementation.TransactOpts, id, duration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RenounceOwnership(&_AnytypeRegistrarImplementation.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.RenounceOwnership(&_AnytypeRegistrarImplementation.TransactOpts)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SafeTransferFrom(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SafeTransferFrom(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SafeTransferFrom0(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SafeTransferFrom0(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SetApprovalForAll(&_AnytypeRegistrarImplementation.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SetApprovalForAll(&_AnytypeRegistrarImplementation.TransactOpts, operator, approved)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address resolver) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) SetResolver(opts *bind.TransactOpts, resolver common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "setResolver", resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address resolver) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) SetResolver(resolver common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SetResolver(&_AnytypeRegistrarImplementation.TransactOpts, resolver)
}

// SetResolver is a paid mutator transaction binding the contract method 0x4e543b26.
//
// Solidity: function setResolver(address resolver) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) SetResolver(resolver common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.SetResolver(&_AnytypeRegistrarImplementation.TransactOpts, resolver)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.TransferFrom(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.TransferFrom(&_AnytypeRegistrarImplementation.TransactOpts, from, to, tokenId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.TransferOwnership(&_AnytypeRegistrarImplementation.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AnytypeRegistrarImplementation.Contract.TransferOwnership(&_AnytypeRegistrarImplementation.TransactOpts, newOwner)
}

// AnytypeRegistrarImplementationApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationApprovalIterator struct {
	Event *AnytypeRegistrarImplementationApproval // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationApproval)
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
		it.Event = new(AnytypeRegistrarImplementationApproval)
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
func (it *AnytypeRegistrarImplementationApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationApproval represents a Approval event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*AnytypeRegistrarImplementationApprovalIterator, error) {

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

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationApprovalIterator{contract: _AnytypeRegistrarImplementation.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationApproval)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseApproval(log types.Log) (*AnytypeRegistrarImplementationApproval, error) {
	event := new(AnytypeRegistrarImplementationApproval)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationApprovalForAllIterator struct {
	Event *AnytypeRegistrarImplementationApprovalForAll // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationApprovalForAll)
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
		it.Event = new(AnytypeRegistrarImplementationApprovalForAll)
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
func (it *AnytypeRegistrarImplementationApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationApprovalForAll represents a ApprovalForAll event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*AnytypeRegistrarImplementationApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationApprovalForAllIterator{contract: _AnytypeRegistrarImplementation.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationApprovalForAll)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseApprovalForAll(log types.Log) (*AnytypeRegistrarImplementationApprovalForAll, error) {
	event := new(AnytypeRegistrarImplementationApprovalForAll)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationControllerAddedIterator is returned from FilterControllerAdded and is used to iterate over the raw logs and unpacked data for ControllerAdded events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationControllerAddedIterator struct {
	Event *AnytypeRegistrarImplementationControllerAdded // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationControllerAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationControllerAdded)
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
		it.Event = new(AnytypeRegistrarImplementationControllerAdded)
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
func (it *AnytypeRegistrarImplementationControllerAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationControllerAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationControllerAdded represents a ControllerAdded event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationControllerAdded struct {
	Controller common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerAdded is a free log retrieval operation binding the contract event 0x0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d7474.
//
// Solidity: event ControllerAdded(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterControllerAdded(opts *bind.FilterOpts, controller []common.Address) (*AnytypeRegistrarImplementationControllerAddedIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "ControllerAdded", controllerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationControllerAddedIterator{contract: _AnytypeRegistrarImplementation.contract, event: "ControllerAdded", logs: logs, sub: sub}, nil
}

// WatchControllerAdded is a free log subscription operation binding the contract event 0x0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d7474.
//
// Solidity: event ControllerAdded(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchControllerAdded(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationControllerAdded, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "ControllerAdded", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationControllerAdded)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ControllerAdded", log); err != nil {
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

// ParseControllerAdded is a log parse operation binding the contract event 0x0a8bb31534c0ed46f380cb867bd5c803a189ced9a764e30b3a4991a9901d7474.
//
// Solidity: event ControllerAdded(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseControllerAdded(log types.Log) (*AnytypeRegistrarImplementationControllerAdded, error) {
	event := new(AnytypeRegistrarImplementationControllerAdded)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ControllerAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationControllerRemovedIterator is returned from FilterControllerRemoved and is used to iterate over the raw logs and unpacked data for ControllerRemoved events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationControllerRemovedIterator struct {
	Event *AnytypeRegistrarImplementationControllerRemoved // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationControllerRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationControllerRemoved)
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
		it.Event = new(AnytypeRegistrarImplementationControllerRemoved)
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
func (it *AnytypeRegistrarImplementationControllerRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationControllerRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationControllerRemoved represents a ControllerRemoved event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationControllerRemoved struct {
	Controller common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerRemoved is a free log retrieval operation binding the contract event 0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113.
//
// Solidity: event ControllerRemoved(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterControllerRemoved(opts *bind.FilterOpts, controller []common.Address) (*AnytypeRegistrarImplementationControllerRemovedIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "ControllerRemoved", controllerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationControllerRemovedIterator{contract: _AnytypeRegistrarImplementation.contract, event: "ControllerRemoved", logs: logs, sub: sub}, nil
}

// WatchControllerRemoved is a free log subscription operation binding the contract event 0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113.
//
// Solidity: event ControllerRemoved(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchControllerRemoved(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationControllerRemoved, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "ControllerRemoved", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationControllerRemoved)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ControllerRemoved", log); err != nil {
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

// ParseControllerRemoved is a log parse operation binding the contract event 0x33d83959be2573f5453b12eb9d43b3499bc57d96bd2f067ba44803c859e81113.
//
// Solidity: event ControllerRemoved(address indexed controller)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseControllerRemoved(log types.Log) (*AnytypeRegistrarImplementationControllerRemoved, error) {
	event := new(AnytypeRegistrarImplementationControllerRemoved)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "ControllerRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationNameRegisteredIterator is returned from FilterNameRegistered and is used to iterate over the raw logs and unpacked data for NameRegistered events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationNameRegisteredIterator struct {
	Event *AnytypeRegistrarImplementationNameRegistered // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationNameRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationNameRegistered)
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
		it.Event = new(AnytypeRegistrarImplementationNameRegistered)
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
func (it *AnytypeRegistrarImplementationNameRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationNameRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationNameRegistered represents a NameRegistered event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationNameRegistered struct {
	Id      *big.Int
	Owner   common.Address
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRegistered is a free log retrieval operation binding the contract event 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9.
//
// Solidity: event NameRegistered(uint256 indexed id, address indexed owner, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterNameRegistered(opts *bind.FilterOpts, id []*big.Int, owner []common.Address) (*AnytypeRegistrarImplementationNameRegisteredIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "NameRegistered", idRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationNameRegisteredIterator{contract: _AnytypeRegistrarImplementation.contract, event: "NameRegistered", logs: logs, sub: sub}, nil
}

// WatchNameRegistered is a free log subscription operation binding the contract event 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9.
//
// Solidity: event NameRegistered(uint256 indexed id, address indexed owner, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchNameRegistered(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationNameRegistered, id []*big.Int, owner []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "NameRegistered", idRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationNameRegistered)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "NameRegistered", log); err != nil {
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

// ParseNameRegistered is a log parse operation binding the contract event 0xb3d987963d01b2f68493b4bdb130988f157ea43070d4ad840fee0466ed9370d9.
//
// Solidity: event NameRegistered(uint256 indexed id, address indexed owner, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseNameRegistered(log types.Log) (*AnytypeRegistrarImplementationNameRegistered, error) {
	event := new(AnytypeRegistrarImplementationNameRegistered)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "NameRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationNameRenewedIterator is returned from FilterNameRenewed and is used to iterate over the raw logs and unpacked data for NameRenewed events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationNameRenewedIterator struct {
	Event *AnytypeRegistrarImplementationNameRenewed // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationNameRenewedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationNameRenewed)
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
		it.Event = new(AnytypeRegistrarImplementationNameRenewed)
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
func (it *AnytypeRegistrarImplementationNameRenewedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationNameRenewedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationNameRenewed represents a NameRenewed event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationNameRenewed struct {
	Id      *big.Int
	Expires *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNameRenewed is a free log retrieval operation binding the contract event 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6.
//
// Solidity: event NameRenewed(uint256 indexed id, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterNameRenewed(opts *bind.FilterOpts, id []*big.Int) (*AnytypeRegistrarImplementationNameRenewedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "NameRenewed", idRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationNameRenewedIterator{contract: _AnytypeRegistrarImplementation.contract, event: "NameRenewed", logs: logs, sub: sub}, nil
}

// WatchNameRenewed is a free log subscription operation binding the contract event 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6.
//
// Solidity: event NameRenewed(uint256 indexed id, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchNameRenewed(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationNameRenewed, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "NameRenewed", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationNameRenewed)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "NameRenewed", log); err != nil {
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

// ParseNameRenewed is a log parse operation binding the contract event 0x9b87a00e30f1ac65d898f070f8a3488fe60517182d0a2098e1b4b93a54aa9bd6.
//
// Solidity: event NameRenewed(uint256 indexed id, uint256 expires)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseNameRenewed(log types.Log) (*AnytypeRegistrarImplementationNameRenewed, error) {
	event := new(AnytypeRegistrarImplementationNameRenewed)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "NameRenewed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationOwnershipTransferredIterator struct {
	Event *AnytypeRegistrarImplementationOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationOwnershipTransferred)
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
		it.Event = new(AnytypeRegistrarImplementationOwnershipTransferred)
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
func (it *AnytypeRegistrarImplementationOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationOwnershipTransferred represents a OwnershipTransferred event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AnytypeRegistrarImplementationOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationOwnershipTransferredIterator{contract: _AnytypeRegistrarImplementation.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationOwnershipTransferred)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseOwnershipTransferred(log types.Log) (*AnytypeRegistrarImplementationOwnershipTransferred, error) {
	event := new(AnytypeRegistrarImplementationOwnershipTransferred)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnytypeRegistrarImplementationTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationTransferIterator struct {
	Event *AnytypeRegistrarImplementationTransfer // Event containing the contract specifics and raw log

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
func (it *AnytypeRegistrarImplementationTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnytypeRegistrarImplementationTransfer)
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
		it.Event = new(AnytypeRegistrarImplementationTransfer)
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
func (it *AnytypeRegistrarImplementationTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnytypeRegistrarImplementationTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnytypeRegistrarImplementationTransfer represents a Transfer event raised by the AnytypeRegistrarImplementation contract.
type AnytypeRegistrarImplementationTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*AnytypeRegistrarImplementationTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AnytypeRegistrarImplementationTransferIterator{contract: _AnytypeRegistrarImplementation.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *AnytypeRegistrarImplementationTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AnytypeRegistrarImplementation.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnytypeRegistrarImplementationTransfer)
				if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AnytypeRegistrarImplementation *AnytypeRegistrarImplementationFilterer) ParseTransfer(log types.Log) (*AnytypeRegistrarImplementationTransfer, error) {
	event := new(AnytypeRegistrarImplementationTransfer)
	if err := _AnytypeRegistrarImplementation.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
