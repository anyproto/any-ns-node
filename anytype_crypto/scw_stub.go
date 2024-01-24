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

// SCWMetaData contains all meta data concerning the SCW contract.
var SCWMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SCWABI is the input ABI used to generate the binding from.
// Deprecated: Use SCWMetaData.ABI instead.
var SCWABI = SCWMetaData.ABI

// SCW is an auto generated Go binding around an Ethereum contract.
type SCW struct {
	SCWCaller     // Read-only binding to the contract
	SCWTransactor // Write-only binding to the contract
	SCWFilterer   // Log filterer for contract events
}

// SCWCaller is an auto generated read-only Go binding around an Ethereum contract.
type SCWCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCWTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SCWTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCWFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SCWFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SCWSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SCWSession struct {
	Contract     *SCW              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SCWCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SCWCallerSession struct {
	Contract *SCWCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SCWTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SCWTransactorSession struct {
	Contract     *SCWTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SCWRaw is an auto generated low-level Go binding around an Ethereum contract.
type SCWRaw struct {
	Contract *SCW // Generic contract binding to access the raw methods on
}

// SCWCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SCWCallerRaw struct {
	Contract *SCWCaller // Generic read-only contract binding to access the raw methods on
}

// SCWTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SCWTransactorRaw struct {
	Contract *SCWTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSCW creates a new instance of SCW, bound to a specific deployed contract.
func NewSCW(address common.Address, backend bind.ContractBackend) (*SCW, error) {
	contract, err := bindSCW(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SCW{SCWCaller: SCWCaller{contract: contract}, SCWTransactor: SCWTransactor{contract: contract}, SCWFilterer: SCWFilterer{contract: contract}}, nil
}

// NewSCWCaller creates a new read-only instance of SCW, bound to a specific deployed contract.
func NewSCWCaller(address common.Address, caller bind.ContractCaller) (*SCWCaller, error) {
	contract, err := bindSCW(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SCWCaller{contract: contract}, nil
}

// NewSCWTransactor creates a new write-only instance of SCW, bound to a specific deployed contract.
func NewSCWTransactor(address common.Address, transactor bind.ContractTransactor) (*SCWTransactor, error) {
	contract, err := bindSCW(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SCWTransactor{contract: contract}, nil
}

// NewSCWFilterer creates a new log filterer instance of SCW, bound to a specific deployed contract.
func NewSCWFilterer(address common.Address, filterer bind.ContractFilterer) (*SCWFilterer, error) {
	contract, err := bindSCW(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SCWFilterer{contract: contract}, nil
}

// bindSCW binds a generic wrapper to an already deployed contract.
func bindSCW(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SCWABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SCW *SCWRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SCW.Contract.SCWCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SCW *SCWRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SCW.Contract.SCWTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SCW *SCWRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SCW.Contract.SCWTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SCW *SCWCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SCW.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SCW *SCWTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SCW.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SCW *SCWTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SCW.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SCW *SCWCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SCW.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SCW *SCWSession) Owner() (common.Address, error) {
	return _SCW.Contract.Owner(&_SCW.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SCW *SCWCallerSession) Owner() (common.Address, error) {
	return _SCW.Contract.Owner(&_SCW.CallOpts)
}
