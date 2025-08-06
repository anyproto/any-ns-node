package anynsaarpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"
	dbservice "github.com/anyproto/any-ns-node/db"
	"github.com/anyproto/any-ns-node/verification"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/any-sync/util/crypto"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	accountabstraction "github.com/anyproto/any-ns-node/account_abstraction"
	contracts "github.com/anyproto/any-ns-node/contracts"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.aa-rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsAARpc{}
}

type anynsAARpc struct {
	conf          *config.Config
	confContracts config.Contracts
	confAccount   accountservice.Config
	db            dbservice.DbService
	nodeConf      nodeconf.Service

	contracts contracts.ContractsService
	aa        accountabstraction.AccountAbstractionService
	cache     cache.CacheService
}

func (arpc *anynsAARpc) Init(a *app.App) (err error) {
	arpc.conf = a.MustComponent(config.CName).(*config.Config)
	arpc.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.db = a.MustComponent(dbservice.CName).(dbservice.DbService)
	arpc.nodeConf = a.MustComponent(nodeconf.CName).(nodeconf.Service)
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.confAccount = a.MustComponent(config.CName).(*config.Config).GetAccount()

	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)
	arpc.cache = a.MustComponent(cache.CName).(cache.CacheService)

	return nsp.DRPCRegisterAnynsAccountAbstraction(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

// TODO: check if it is even called, this is not a app.ComponentRunnable instance
// so maybe it won't be called
func (arpc *anynsAARpc) Close(ctx context.Context) (err error) {
	return nil
}

func (arpc *anynsAARpc) Name() (name string) {
	return CName
}

func (arpc *anynsAARpc) GetUserAccount(ctx context.Context, in *nsp.GetUserAccountRequest) (*nsp.UserAccount, error) {
	var res nsp.UserAccount
	res.OwnerEthAddress = in.OwnerEthAddress

	// 1 - get SCW address
	// even if SCW is not deployed yet -> it should be returned
	scwa, err := arpc.aa.GetSmartWalletAddress(ctx, common.HexToAddress(in.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address", zap.Error(err))
		return nil, errors.New("failed to get smart wallet address")
	}

	res.OwnerSmartContracWalletAddress = scwa.Hex()

	// 2 - check if SCW is deployed
	res.OwnerSmartContracWalletDeployed, err = arpc.contracts.IsContractDeployed(ctx, scwa)
	if err != nil {
		log.Error("failed to check if contract is deployed", zap.Error(err))
		return nil, errors.New("failed to get smart wallet")
	}

	// 3 - the rest
	res.NamesCountLeft, err = arpc.aa.GetNamesCountLeft(ctx, scwa)
	if err != nil {
		log.Error("failed to get names count left", zap.Error(err))
		return nil, errors.New("failed to get names count left")
	}

	res.OperationsCountLeft, err = arpc.db.GetUserOperationsCount(
		ctx,
		common.HexToAddress(in.OwnerEthAddress),
		// becuase AnyID is empty -> it will not check it
		"",
	)

	if err != nil {
		log.Error("failed to get operations count left", zap.Error(err))
		return nil, errors.New("failed to get operations count left")
	}

	// return
	return &res, nil
}

func (arpc *anynsAARpc) GetOperation(ctx context.Context, in *nsp.GetOperationStatusRequest) (*nsp.OperationResponse, error) {
	var out nsp.OperationResponse

	// 0 - get operation from Mongo first (we will need it later to update cache)
	// some operation has not been saved to Mongo (like AdminFundUserAccount)
	op, err := arpc.db.GetOperation(ctx, in.OperationId)
	operationFound := (err != mongo.ErrNoDocuments)

	// trigger error only in case Mongo returns something bad (not found is ok)
	if err != nil && operationFound {
		log.Error("failed to get operation from Mongo", zap.Error(err))
		return nil, errors.New("failed to get cache")
	}

	// 1 - get operation status from the AA service
	status, err := arpc.aa.GetOperation(ctx, in.OperationId)
	if err != nil {
		log.Error("failed to get operation info", zap.Error(err))
		return nil, errors.New("failed to get operation from cache")
	}

	out.OperationId = fmt.Sprint(in.OperationId)
	out.OperationState = status.OperationState

	// 2 - update cache (only once operation is completed)
	if operationFound && status.OperationState == nsp.OperationState_Completed {
		// 2.1 - is info already is in the cache?
		cacheRes, err := arpc.cache.IsNameAvailable(ctx, &nsp.NameAvailableRequest{
			FullName: op.FullName,
		})
		if err == nil && !cacheRes.Available {
			log.Info("name is already in cache", zap.String("FullName", op.FullName))
			return &out, nil
		}

		// 2.2 - if not -> read from smart contracts
		log.Info("operation completed, updating cache", zap.String("FullName", op.FullName))
		err = arpc.cache.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: op.FullName,
		})

		if err != nil {
			log.Error("failed to update cache", zap.Error(err))
			return nil, errors.New("failed to update in cache")
		}
	}

	return &out, nil
}

func (arpc *anynsAARpc) isAdmin(peerId string) bool {
	// 1 - check if peer is a payment node!
	if slices.Contains(arpc.nodeConf.NodeTypes(peerId), nodeconf.NodeTypePaymentProcessingNode) {
		return true
	}

	// 2 - admin
	err := verification.VerifyAdminIdentity(arpc.confAccount.PeerKey, peerId)
	return (err == nil)
}

// WARNING: There is no way here to check that EthAddress of user matches AnyID
// we trust user! If he passed wrong AnyID - it is his problem
// and then Admin just passes those values to current method
func (arpc *anynsAARpc) AdminFundUserAccount(ctx context.Context, in *nsp.AdminFundUserAccountRequestSigned) (*nsp.OperationResponse, error) {
	peerId, err := peer.CtxPeerId(ctx)
	if err != nil {
		return nil, err
	}

	// 1 - unmarshal the signed request
	var afuar nsp.AdminFundUserAccountRequest
	err = afuar.UnmarshalVT(in.Payload)
	if err != nil {
		log.Error("can not unmarshal AdminFundUserAccount", zap.Error(err))
		return nil, errors.New("can not unmarshal AdminFundUserAccount")
	}

	// 2 - check signature
	isAllow := arpc.isAdmin(peerId)
	if !isAllow {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, errors.New("not an Admin!!!")
	}

	// 3 - determine SCW of user wallet
	scwa, err := arpc.aa.GetSmartWalletAddress(ctx, common.HexToAddress(afuar.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address", zap.Error(err))
		return nil, errors.New("failed to get smart wallet address")
	}

	// 4 - mint tokens to that SCW
	opID, err := arpc.aa.AdminMintAccessTokens(ctx, scwa, big.NewInt(int64(afuar.NamesCount)))
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, errors.New("failed to mint access tokens")
	}

	/*
		// 5 - save operation to mongo
		err = arpc.db.SaveOperation(ctx, opID, cuor)
		if err != nil {
			log.Error("failed to save operation to Mongo", zap.Error(err))
			return nil, err
		}
	*/

	// 6 - return
	var out nsp.OperationResponse
	out.OperationId = fmt.Sprint(opID)
	out.OperationState = nsp.OperationState_Pending
	return &out, err
}

func (arpc *anynsAARpc) AdminFundGasOperations(ctx context.Context, in *nsp.AdminFundGasOperationsRequestSigned) (*nsp.OperationResponse, error) {
	peerId, err := peer.CtxPeerId(ctx)
	if err != nil {
		return nil, err
	}

	// 1 - unmarshal the signed request
	var afgor nsp.AdminFundGasOperationsRequest
	err = afgor.UnmarshalVT(in.Payload)
	if err != nil {
		log.Error("can not unmarshal AdminFundGasOperationsRequest", zap.Error(err))
		return nil, errors.New("can not unmarshal AdminFundGasOperationsRequest")
	}

	// 2 - check signature
	isAllow := arpc.isAdmin(peerId)
	if !isAllow {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, errors.New("not an Admin!!!")
	}

	// validate all params
	// TODO: check format
	if afgor.OwnerEthAddress == "" {
		log.Error("wrong OwnerEthAddress", zap.String("OwnerEthAddress", afgor.OwnerEthAddress))
		return nil, errors.New("wrong OwnerEthAddress")
	}
	if afgor.OwnerAnyID == "" {
		log.Error("wrong OwnerAnyID", zap.String("OwnerAnyID", afgor.OwnerAnyID))
		return nil, errors.New("wrong OwnerAnyID")
	}
	if afgor.OperationsCount == 0 {
		log.Error("wrong OperationsCount", zap.Uint64("OperationsCount", afgor.OperationsCount))
		return nil, errors.New("wrong OperationsCount")
	}

	var out nsp.OperationResponse
	err = arpc.db.AddUserToTheWhitelist(ctx, common.HexToAddress(afgor.OwnerEthAddress), afgor.OwnerAnyID, afgor.OperationsCount)
	if err != nil {
		log.Error("failed to add user to the whitelist", zap.Error(err))
		return nil, errors.New("failed to add user to the whitelist")
	}

	return &out, err
}

func (arpc *anynsAARpc) GetDataNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (*nsp.GetDataNameRegisterResponse, error) {
	// 1 - check params
	useEnsip15 := arpc.conf.Ensip15Validation

	err := verification.CheckRegisterParams(in, useEnsip15)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, errors.New("invalid parameters")
	}

	// 2 - get data to sign
	dataOut, contextData, err := arpc.aa.GetDataNameRegister(ctx, in)
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, errors.New("failed to mint tokens")
	}

	var out nsp.GetDataNameRegisterResponse
	// user should sign it
	out.Data = dataOut
	// user should pass it back to us
	out.Context = contextData

	return &out, nil
}

func (arpc *anynsAARpc) GetDataNameRegisterForSpace(ctx context.Context, in *nsp.NameRegisterForSpaceRequest) (*nsp.GetDataNameRegisterResponse, error) {
	// 1 - check params
	useEnsip15 := arpc.conf.Ensip15Validation

	err := verification.CheckRegisterForSpaceParams(in, useEnsip15)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, errors.New("invalid parameters")
	}

	// 2 - get data to sign
	dataOut, contextData, err := arpc.aa.GetDataNameRegisterForSpace(ctx, in)
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, errors.New("failed to mint tokens")
	}

	var out nsp.GetDataNameRegisterResponse
	// user should sign it
	out.Data = dataOut
	// user should pass it back to us
	out.Context = contextData

	return &out, nil
}

// once user got data by using method like GetDataNameRegister, and signed it, now he can create a new operation
func (arpc *anynsAARpc) CreateUserOperation(ctx context.Context, in *nsp.CreateUserOperationRequestSigned) (*nsp.OperationResponse, error) {
	userAnyID, err := peer.CtxIdentity(ctx)
	if err != nil {
		return nil, err
	}

	// 1 - unmarshal the signed request
	var cuor nsp.CreateUserOperationRequest
	err = cuor.UnmarshalVT(in.Payload)
	if err != nil {
		log.Error("can not unmarshal CreateUserOperationRequest", zap.Error(err))
		return nil, errors.New("can not unmarshal CreateUserOperationRequest")
	}

	// 2 - check users's signature
	err = verification.VerifyAnyIdentity(crypto.EncodeBytesToString(userAnyID), in.Payload, in.Signature)
	if err != nil {
		log.Error("wrong Anytype signature", zap.Error(err))
		return nil, errors.New("wrong Anytype signature")
	}

	// 3 - check if user has enough operations left
	// will fail if AnyID was different
	ops, err := arpc.db.GetUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress), cuor.OwnerAnyID)
	if err != nil {
		log.Error("failed to get operations count", zap.Error(err))
		return nil, errors.New("failed to get operations count")
	}

	if ops < 1 {
		log.Error("not enough operations left", zap.Uint64("ops", ops))
		return nil, errors.New("not enough operations left")
	}

	// 4 - now send it!
	// TODO: add to queue???
	opID, err := arpc.aa.SendUserOperation(ctx, cuor.Context, cuor.SignedData)
	if err != nil {
		log.Error("failed to send user operation", zap.Error(err))
		return nil, errors.New("failed to send user operation")
	}

	// 5 - decrease operations count for that user
	err = arpc.db.DecreaseUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress))
	if err != nil {
		log.Error("failed to decrease operations count", zap.Error(err))
		return nil, errors.New("failed to decrease operations count")
	}

	// 6 - save operation to mongo (can be used later)
	err = arpc.db.SaveOperation(ctx, opID, cuor)
	if err != nil {
		log.Error("failed to save operation to Mongo", zap.Error(err))
		return nil, errors.New("failed to save operation")
	}

	// 7 - return result
	var out nsp.OperationResponse
	out.OperationId = opID
	out.OperationState = nsp.OperationState_Pending

	return nil, nil
}
