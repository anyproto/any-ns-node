package anynsaarpc

import (
	"context"
	"math/big"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	contracts "github.com/anyproto/any-ns-node/contracts"
	as "github.com/anyproto/any-ns-node/pb/anyns_api"
)

const CName = "any-ns.aa-rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsAARpc{}
}

type anynsAARpc struct {
	contractsConfig config.Contracts
	contracts       contracts.ContractsService
	aa              accountabstraction.AccountAbstractionService
}

func (arpc *anynsAARpc) Init(a *app.App) (err error) {
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)

	return as.DRPCRegisterAnynsAccountAbstraction(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsAARpc) Name() (name string) {
	return CName
}

func (arpc *anynsAARpc) GetUserAccount(ctx context.Context, in *as.GetUserAccountRequest) (*as.UserAccount, error) {
	var res as.UserAccount
	res.OwnerEthAddress = in.OwnerEthAddress

	// even if SCW is not deployed yet -> it should be returned
	scwa, err := arpc.aa.GetSmartWalletAddress(ctx, common.HexToAddress(in.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address", zap.Error(err))
		return nil, err
	}

	res.OwnerSmartContracWalletAddress = scwa.Hex()

	res.NamesCountLeft, err = arpc.aa.GetNamesCountLeft(scwa)
	if err != nil {
		log.Error("failed to get names count left", zap.Error(err))
		return nil, err
	}

	res.OperationsCountLeft, err = arpc.aa.GetOperationsCountLeft(scwa)
	if err != nil {
		log.Error("failed to get operations count left", zap.Error(err))
		return nil, err
	}

	// return
	return &res, nil
}

func (arpc *anynsAARpc) AdminFundUserAccount(ctx context.Context, in *as.AdminFundUserAccountRequestSigned) (*as.OperationResponse, error) {
	// 1 - unmarshal the signed request
	var afuar as.AdminFundUserAccountRequest
	err := proto.Unmarshal(in.Payload, &afuar)
	if err != nil {
		log.Error("can not unmarshal AdminFundUserAccount", zap.Error(err))
		return nil, err
	}

	// 2 - check signature
	err = arpc.aa.VerifyAdminIdentity(in.Payload, in.Signature)
	if err != nil {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, err
	}

	// 3 - determine SCW of user wallet
	ua, err := arpc.GetUserAccount(ctx, &as.GetUserAccountRequest{
		OwnerEthAddress: afuar.OwnerEthAddress,
	})
	if err != nil {
		log.Error("failed to get user account", zap.Error(err))
		return nil, err
	}

	// 4 - mint tokens to that SCW
	err = arpc.aa.AdminMintAccessTokens(common.HexToAddress(ua.OwnerSmartContracWalletAddress), big.NewInt(int64(afuar.NamesCount)))
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, err
	}

	// 3 - return
	// TODO: add to queue
	var out as.OperationResponse
	out.OperationId = 0
	out.OperationState = as.OperationState_Pending

	return &out, err
}

func (arpc *anynsAARpc) AdminFundGasOperations(ctx context.Context, in *as.AdminFundGasOperationsRequestSigned) (*as.OperationResponse, error) {
	// TODO: implement

	return nil, nil
}

func (arpc *anynsAARpc) GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (*as.GetDataNameRegisterResponse, error) {
	// TODO: implement

	return nil, nil
}

func (arpc *anynsAARpc) GetDataNameUpdate(ctx context.Context, in *as.NameUpdateRequest) (*as.GetDataNameRegisterResponse, error) {
	// TODO: implement

	return nil, nil
}

func (arpc *anynsAARpc) CreateUserOperation(ctx context.Context, in *as.CreateUserOperationRequestSigned) (*as.OperationResponse, error) {
	// TODO: implement

	return nil, nil
}
