package anynsaarpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/anyproto/any-ns-node/anynsrpc"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	confContracts config.Contracts
	confMongo     config.Mongo

	itemColl *mongo.Collection

	contracts contracts.ContractsService
	aa        accountabstraction.AccountAbstractionService
}

func (arpc *anynsAARpc) Init(a *app.App) (err error) {
	arpc.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)

	// connect to mongo
	uri := arpc.confMongo.Connect
	dbName := arpc.confMongo.Database
	collectionName := "aa"

	// 1 - connect to DB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	arpc.itemColl = client.Database(dbName).Collection(collectionName)
	if arpc.itemColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("mongo connected!")

	return as.DRPCRegisterAnynsAccountAbstraction(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

// TODO: check if it is even called, this is not a app.ComponentRunnable instance
// so maybe it won't be called
func (arpc *anynsAARpc) Close(ctx context.Context) (err error) {
	if arpc.itemColl != nil {
		err = arpc.itemColl.Database().Client().Disconnect(ctx)
		arpc.itemColl = nil
	}
	return
}

func (arpc *anynsAARpc) Name() (name string) {
	return CName
}

func (arpc *anynsAARpc) GetUserAccount(ctx context.Context, in *as.GetUserAccountRequest) (*as.UserAccount, error) {
	var res as.UserAccount
	res.OwnerEthAddress = in.OwnerEthAddress

	// 1 - get SCW address
	// even if SCW is not deployed yet -> it should be returned
	scwa, err := arpc.aa.GetSmartWalletAddress(ctx, common.HexToAddress(in.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address", zap.Error(err))
		return nil, err
	}

	res.OwnerSmartContracWalletAddress = scwa.Hex()

	// 2 - check if SCW is deployed
	client, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create eth connection", zap.Error(err))
		return nil, err
	}

	res.OwnerSmartContracWalletDeployed, err = arpc.contracts.IsContractDeployed(ctx, client, scwa)
	if err != nil {
		log.Error("failed to check if contract is deployed", zap.Error(err))
		return nil, err
	}

	// 3 - the rest
	res.NamesCountLeft, err = arpc.aa.GetNamesCountLeft(ctx, scwa)
	if err != nil {
		log.Error("failed to get names count left", zap.Error(err))
		return nil, err
	}

	res.OperationsCountLeft, err = arpc.aa.GetOperationsCountLeft(ctx, scwa)
	if err != nil {
		log.Error("failed to get operations count left", zap.Error(err))
		return nil, err
	}

	// return
	return &res, nil
}

func (arpc *anynsAARpc) GetOperation(ctx context.Context, in *as.GetOperationStatusRequest) (*as.OperationResponse, error) {
	var out as.OperationResponse

	status, err := arpc.aa.GetOperation(ctx, in.OperationId)
	if err != nil {
		log.Error("failed to get operation info", zap.Error(err))
		return nil, err
	}

	out.OperationId = fmt.Sprint(in.OperationId)
	out.OperationState = status.OperationState

	return &out, nil
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
	err = arpc.aa.AdminVerifyIdentity(in.Payload, in.Signature)
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
	opID, err := arpc.aa.AdminMintAccessTokens(ctx, common.HexToAddress(ua.OwnerSmartContracWalletAddress), big.NewInt(int64(afuar.NamesCount)))
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, err
	}

	// 3 - return
	// TODO: add to queue
	var out as.OperationResponse
	out.OperationId = fmt.Sprint(opID)
	out.OperationState = as.OperationState_Pending

	return &out, err
}

func (arpc *anynsAARpc) AdminFundGasOperations(ctx context.Context, in *as.AdminFundGasOperationsRequestSigned) (*as.OperationResponse, error) {
	// 1 - unmarshal the signed request
	var afgor as.AdminFundGasOperationsRequest
	err := proto.Unmarshal(in.Payload, &afgor)
	if err != nil {
		log.Error("can not unmarshal AdminFundGasOperationsRequest", zap.Error(err))
		return nil, err
	}

	// 2 - check signature
	err = arpc.aa.AdminVerifyIdentity(in.Payload, in.Signature)
	if err != nil {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, err
	}

	// TODO: allow +N operations
	log.Fatal("TODO: not implemented")

	return nil, nil
}

func (arpc *anynsAARpc) GetDataNameRegister(ctx context.Context, in *as.NameRegisterRequest) (*as.GetDataNameRegisterResponse, error) {
	// 1 - check params
	err := anynsrpc.СheckRegisterParams(in)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, err
	}

	// 2 - get data to sign
	dataOut, contextData, err := arpc.aa.GetDataNameRegister(ctx, in)
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, err
	}

	var out as.GetDataNameRegisterResponse
	// user should sign it
	out.Data = dataOut
	// user should pass it back to us
	out.Context = contextData

	return &out, nil
}

func (arpc *anynsAARpc) GetDataNameUpdate(ctx context.Context, in *as.NameUpdateRequest) (*as.GetDataNameRegisterResponse, error) {
	// TODO: implement

	return nil, nil
}

// once user got data by using method like GetDataNameRegister, and signed it, now he can create a new operation
func (arpc *anynsAARpc) CreateUserOperation(ctx context.Context, in *as.CreateUserOperationRequestSigned) (*as.OperationResponse, error) {
	// 1 - unmarshal the signed request
	var cuor as.CreateUserOperationRequest
	err := proto.Unmarshal(in.Payload, &cuor)
	if err != nil {
		log.Error("can not unmarshal CreateUserOperationRequest", zap.Error(err))
		return nil, err
	}

	// TODO:
	// 2 - check users's signature
	/*
		err = arpc.aa.VerifyIdentity(in.Payload, in.Signature, cuor)
		if err != nil {
			log.Error("not an Admin!!!", zap.Error(err))
			return nil, err
		}
	*/

	// TODO:
	// 3 - check if user has enough "GetOperationsCountLeft"

	// 4 - now send it!
	opID, err := arpc.aa.SendUserOperation(ctx, cuor.Context, cuor.SignedData)
	if err != nil {
		log.Error("failed to send user operation", zap.Error(err))
		return nil, err
	}

	// TODO: add to queue

	// 5 - return result
	var out as.OperationResponse
	out.OperationId = opID
	out.OperationState = as.OperationState_Pending

	return nil, nil
}
