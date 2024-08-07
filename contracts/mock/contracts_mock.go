// Code generated by MockGen. DO NOT EDIT.
// Source: contracts/contracts.go
//
// Generated by this command:
//
//	mockgen -source=contracts/contracts.go
//

// Package mock_contracts is a generated GoMock package.
package mock_contracts

import (
	context "context"
	big "math/big"
	reflect "reflect"

	anytype_crypto "github.com/anyproto/any-ns-node/anytype_crypto"
	contracts "github.com/anyproto/any-ns-node/contracts"
	app "github.com/anyproto/any-sync/app"
	ethereum "github.com/ethereum/go-ethereum"
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
	gomock "go.uber.org/mock/gomock"
)

// MockContractsService is a mock of ContractsService interface.
type MockContractsService struct {
	ctrl     *gomock.Controller
	recorder *MockContractsServiceMockRecorder
}

// MockContractsServiceMockRecorder is the mock recorder for MockContractsService.
type MockContractsServiceMockRecorder struct {
	mock *MockContractsService
}

// NewMockContractsService creates a new mock instance.
func NewMockContractsService(ctrl *gomock.Controller) *MockContractsService {
	mock := &MockContractsService{ctrl: ctrl}
	mock.recorder = &MockContractsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContractsService) EXPECT() *MockContractsServiceMockRecorder {
	return m.recorder
}

// CalculateTxParams mocks base method.
func (m *MockContractsService) CalculateTxParams(conn *ethclient.Client, address common.Address) (*big.Int, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateTxParams", conn, address)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CalculateTxParams indicates an expected call of CalculateTxParams.
func (mr *MockContractsServiceMockRecorder) CalculateTxParams(conn, address any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateTxParams", reflect.TypeOf((*MockContractsService)(nil).CalculateTxParams), conn, address)
}

// CallContract mocks base method.
func (m *MockContractsService) CallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CallContract", ctx, msg)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CallContract indicates an expected call of CallContract.
func (mr *MockContractsServiceMockRecorder) CallContract(ctx, msg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CallContract", reflect.TypeOf((*MockContractsService)(nil).CallContract), ctx, msg)
}

// Commit mocks base method.
func (m *MockContractsService) Commit(ctx context.Context, params *contracts.CommitParams) (*types.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", ctx, params)
	ret0, _ := ret[0].(*types.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockContractsServiceMockRecorder) Commit(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockContractsService)(nil).Commit), ctx, params)
}

// ConnectToNamewrapperContract mocks base method.
func (m *MockContractsService) ConnectToNamewrapperContract() (*anytype_crypto.AnytypeNameWrapper, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectToNamewrapperContract")
	ret0, _ := ret[0].(*anytype_crypto.AnytypeNameWrapper)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectToNamewrapperContract indicates an expected call of ConnectToNamewrapperContract.
func (mr *MockContractsServiceMockRecorder) ConnectToNamewrapperContract() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectToNamewrapperContract", reflect.TypeOf((*MockContractsService)(nil).ConnectToNamewrapperContract))
}

// ConnectToPrivateController mocks base method.
func (m *MockContractsService) ConnectToPrivateController() (*anytype_crypto.AnytypeRegistrarControllerPrivate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectToPrivateController")
	ret0, _ := ret[0].(*anytype_crypto.AnytypeRegistrarControllerPrivate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectToPrivateController indicates an expected call of ConnectToPrivateController.
func (mr *MockContractsServiceMockRecorder) ConnectToPrivateController() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectToPrivateController", reflect.TypeOf((*MockContractsService)(nil).ConnectToPrivateController))
}

// ConnectToRegistrar mocks base method.
func (m *MockContractsService) ConnectToRegistrar() (*anytype_crypto.AnytypeRegistrarImplementation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectToRegistrar")
	ret0, _ := ret[0].(*anytype_crypto.AnytypeRegistrarImplementation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectToRegistrar indicates an expected call of ConnectToRegistrar.
func (mr *MockContractsServiceMockRecorder) ConnectToRegistrar() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectToRegistrar", reflect.TypeOf((*MockContractsService)(nil).ConnectToRegistrar))
}

// ConnectToRegistryContract mocks base method.
func (m *MockContractsService) ConnectToRegistryContract() (*anytype_crypto.ENSRegistry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectToRegistryContract")
	ret0, _ := ret[0].(*anytype_crypto.ENSRegistry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectToRegistryContract indicates an expected call of ConnectToRegistryContract.
func (mr *MockContractsServiceMockRecorder) ConnectToRegistryContract() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectToRegistryContract", reflect.TypeOf((*MockContractsService)(nil).ConnectToRegistryContract))
}

// ConnectToResolver mocks base method.
func (m *MockContractsService) ConnectToResolver() (*anytype_crypto.AnytypeResolver, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectToResolver")
	ret0, _ := ret[0].(*anytype_crypto.AnytypeResolver)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectToResolver indicates an expected call of ConnectToResolver.
func (mr *MockContractsServiceMockRecorder) ConnectToResolver() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectToResolver", reflect.TypeOf((*MockContractsService)(nil).ConnectToResolver))
}

// CreateEthConnection mocks base method.
func (m *MockContractsService) CreateEthConnection() (*ethclient.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEthConnection")
	ret0, _ := ret[0].(*ethclient.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEthConnection indicates an expected call of CreateEthConnection.
func (mr *MockContractsServiceMockRecorder) CreateEthConnection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEthConnection", reflect.TypeOf((*MockContractsService)(nil).CreateEthConnection))
}

// GenerateAuthOptsForAdmin mocks base method.
func (m *MockContractsService) GenerateAuthOptsForAdmin() (*bind.TransactOpts, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateAuthOptsForAdmin")
	ret0, _ := ret[0].(*bind.TransactOpts)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateAuthOptsForAdmin indicates an expected call of GenerateAuthOptsForAdmin.
func (mr *MockContractsServiceMockRecorder) GenerateAuthOptsForAdmin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateAuthOptsForAdmin", reflect.TypeOf((*MockContractsService)(nil).GenerateAuthOptsForAdmin))
}

// GetAdditionalNameInfo mocks base method.
func (m *MockContractsService) GetAdditionalNameInfo(ctx context.Context, currentOwner common.Address, fullName string) (string, string, string, *big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAdditionalNameInfo", ctx, currentOwner, fullName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(*big.Int)
	ret4, _ := ret[4].(error)
	return ret0, ret1, ret2, ret3, ret4
}

// GetAdditionalNameInfo indicates an expected call of GetAdditionalNameInfo.
func (mr *MockContractsServiceMockRecorder) GetAdditionalNameInfo(ctx, currentOwner, fullName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAdditionalNameInfo", reflect.TypeOf((*MockContractsService)(nil).GetAdditionalNameInfo), ctx, currentOwner, fullName)
}

// GetBalanceOf mocks base method.
func (m *MockContractsService) GetBalanceOf(ctx context.Context, tokenAddress, address common.Address) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalanceOf", ctx, tokenAddress, address)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalanceOf indicates an expected call of GetBalanceOf.
func (mr *MockContractsServiceMockRecorder) GetBalanceOf(ctx, tokenAddress, address any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalanceOf", reflect.TypeOf((*MockContractsService)(nil).GetBalanceOf), ctx, tokenAddress, address)
}

// GetNameByAddress mocks base method.
func (m *MockContractsService) GetNameByAddress(address common.Address) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNameByAddress", address)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNameByAddress indicates an expected call of GetNameByAddress.
func (mr *MockContractsServiceMockRecorder) GetNameByAddress(address any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNameByAddress", reflect.TypeOf((*MockContractsService)(nil).GetNameByAddress), address)
}

// GetOwnerForNamehash mocks base method.
func (m *MockContractsService) GetOwnerForNamehash(ctx context.Context, namehash [32]byte) (common.Address, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOwnerForNamehash", ctx, namehash)
	ret0, _ := ret[0].(common.Address)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOwnerForNamehash indicates an expected call of GetOwnerForNamehash.
func (mr *MockContractsServiceMockRecorder) GetOwnerForNamehash(ctx, namehash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOwnerForNamehash", reflect.TypeOf((*MockContractsService)(nil).GetOwnerForNamehash), ctx, namehash)
}

// GetScwOwner mocks base method.
func (m *MockContractsService) GetScwOwner(ctx context.Context, address common.Address) (common.Address, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScwOwner", ctx, address)
	ret0, _ := ret[0].(common.Address)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetScwOwner indicates an expected call of GetScwOwner.
func (mr *MockContractsServiceMockRecorder) GetScwOwner(ctx, address any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScwOwner", reflect.TypeOf((*MockContractsService)(nil).GetScwOwner), ctx, address)
}

// Init mocks base method.
func (m *MockContractsService) Init(a *app.App) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", a)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockContractsServiceMockRecorder) Init(a any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockContractsService)(nil).Init), a)
}

// IsContractDeployed mocks base method.
func (m *MockContractsService) IsContractDeployed(ctx context.Context, address common.Address) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsContractDeployed", ctx, address)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsContractDeployed indicates an expected call of IsContractDeployed.
func (mr *MockContractsServiceMockRecorder) IsContractDeployed(ctx, address any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsContractDeployed", reflect.TypeOf((*MockContractsService)(nil).IsContractDeployed), ctx, address)
}

// MakeCommitment mocks base method.
func (m *MockContractsService) MakeCommitment(params *contracts.MakeCommitmentParams) ([32]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeCommitment", params)
	ret0, _ := ret[0].([32]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MakeCommitment indicates an expected call of MakeCommitment.
func (mr *MockContractsServiceMockRecorder) MakeCommitment(params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeCommitment", reflect.TypeOf((*MockContractsService)(nil).MakeCommitment), params)
}

// Name mocks base method.
func (m *MockContractsService) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockContractsServiceMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockContractsService)(nil).Name))
}

// Register mocks base method.
func (m *MockContractsService) Register(ctx context.Context, params *contracts.RegisterParams) (*types.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", ctx, params)
	ret0, _ := ret[0].(*types.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockContractsServiceMockRecorder) Register(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockContractsService)(nil).Register), ctx, params)
}

// Renew mocks base method.
func (m *MockContractsService) Renew(ctx context.Context, params *contracts.RenewParams) (*types.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Renew", ctx, params)
	ret0, _ := ret[0].(*types.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Renew indicates an expected call of Renew.
func (mr *MockContractsServiceMockRecorder) Renew(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Renew", reflect.TypeOf((*MockContractsService)(nil).Renew), ctx, params)
}

// TxByHash mocks base method.
func (m *MockContractsService) TxByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TxByHash", ctx, txHash)
	ret0, _ := ret[0].(*types.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TxByHash indicates an expected call of TxByHash.
func (mr *MockContractsServiceMockRecorder) TxByHash(ctx, txHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TxByHash", reflect.TypeOf((*MockContractsService)(nil).TxByHash), ctx, txHash)
}

// WaitForTxToStartMining mocks base method.
func (m *MockContractsService) WaitForTxToStartMining(ctx context.Context, txHash common.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForTxToStartMining", ctx, txHash)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForTxToStartMining indicates an expected call of WaitForTxToStartMining.
func (mr *MockContractsServiceMockRecorder) WaitForTxToStartMining(ctx, txHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForTxToStartMining", reflect.TypeOf((*MockContractsService)(nil).WaitForTxToStartMining), ctx, txHash)
}

// WaitMined mocks base method.
func (m *MockContractsService) WaitMined(ctx context.Context, tx *types.Transaction) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitMined", ctx, tx)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitMined indicates an expected call of WaitMined.
func (mr *MockContractsServiceMockRecorder) WaitMined(ctx, tx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitMined", reflect.TypeOf((*MockContractsService)(nil).WaitMined), ctx, tx)
}
