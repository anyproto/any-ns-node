package anynsaarpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/verification"
	"github.com/anyproto/any-sync/accountservice"
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
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.aa-rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsAARpc{}
}

type anynsAARpc struct {
	confContracts config.Contracts
	confMongo     config.Mongo
	confAccount   accountservice.Config

	usersColl *mongo.Collection
	opColl    *mongo.Collection

	contracts contracts.ContractsService
	aa        accountabstraction.AccountAbstractionService
	cache     cache.CacheService
}

// TODO: index it
type AAUser struct {
	Address         string `bson:"address"`
	AnyID           string `bson:"any_id"`
	OperationsCount uint64 `bson:"operations"`
}

type findAAUserByAddress struct {
	Address string `bson:"address"`
}

// see nameserviceprotoюCreateUserOperationRequest
type AAUserOperation struct {
	OperationID string `bson:"operation_id"`

	Data       []byte `bson:"data"`
	SignedData []byte `bson:"signed_data"`
	Context    []byte `bson:"context"`

	OwnerEthAddress string `bson:"owner_eth_address"`
	OwnerAnyID      string `bson:"owner_any_id"`
	FullName        string `bson:"full_name"`
}

type findUserOperationByID struct {
	OperationID string `bson:"operation_id"`
}

func (arpc *anynsAARpc) Init(a *app.App) (err error) {
	arpc.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.confAccount = a.MustComponent(config.CName).(*config.Config).GetAccount()

	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)
	arpc.cache = a.MustComponent(cache.CName).(cache.CacheService)

	// connect to mongo
	uri := arpc.confMongo.Connect
	dbName := arpc.confMongo.Database

	// 1 - connect to DB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	arpc.usersColl = client.Database(dbName).Collection("aa-users")
	if arpc.usersColl == nil {
		return errors.New("failed to connect to MongoDB")
	}
	arpc.opColl = client.Database(dbName).Collection("aa-operations")
	if arpc.opColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("mongo connected!")

	return nsp.DRPCRegisterAnynsAccountAbstraction(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

// TODO: check if it is even called, this is not a app.ComponentRunnable instance
// so maybe it won't be called
func (arpc *anynsAARpc) Close(ctx context.Context) (err error) {
	if arpc.usersColl != nil {
		err = arpc.usersColl.Database().Client().Disconnect(ctx)
		arpc.usersColl = nil
	}
	if arpc.opColl != nil {
		err = arpc.opColl.Database().Client().Disconnect(ctx)
		arpc.opColl = nil
	}
	return
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
	client, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to create eth connection", zap.Error(err))
		return nil, errors.New("failed to create eth connection")
	}

	res.OwnerSmartContracWalletDeployed, err = arpc.contracts.IsContractDeployed(ctx, client, scwa)
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

	res.OperationsCountLeft, err = arpc.mongoGetUserOperationsCount(
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
	op, err := arpc.mongoGetOperation(ctx, in.OperationId)
	cacheOperationFound := (err != mongo.ErrNoDocuments)

	// trigger error only in case Mongo returns something bad (not found is ok)
	if err != nil && cacheOperationFound {
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
	if cacheOperationFound && (status.OperationState == nsp.OperationState_Completed) {
		/*
			// is cache empty?
			_, err := arpc.cache.IsNameAvailable(ctx, &nsp.NameAvailableRequest{
				FullName: op.FullName,
			})

			if err == nil {
				log.Debug("name is already in cache", zap.String("FullName", op.FullName))
				return &out, nil
			}
		*/

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

// WARNING: There is no way here to check that EthAddress of user matches AnyID
// we trust user! If he passed wrong AnyID - it is his problem
// and then Admin just passes those values to current method
func (arpc *anynsAARpc) AdminFundUserAccount(ctx context.Context, in *nsp.AdminFundUserAccountRequestSigned) (*nsp.OperationResponse, error) {
	// 1 - unmarshal the signed request
	var afuar nsp.AdminFundUserAccountRequest
	err := proto.Unmarshal(in.Payload, &afuar)
	if err != nil {
		log.Error("can not unmarshal AdminFundUserAccount", zap.Error(err))
		return nil, errors.New("can not unmarshal AdminFundUserAccount")
	}

	// 2 - check signature
	err = verification.VerifyAdminIdentity(arpc.confAccount.SigningKey, in.Payload, in.Signature)
	if err != nil {
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
		err = arpc.mongoSaveOperation(ctx, opID, cuor)
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
	// 1 - unmarshal the signed request
	var afgor nsp.AdminFundGasOperationsRequest
	err := proto.Unmarshal(in.Payload, &afgor)
	if err != nil {
		log.Error("can not unmarshal AdminFundGasOperationsRequest", zap.Error(err))
		return nil, errors.New("can not unmarshal AdminFundGasOperationsRequest")
	}

	// 2 - check signature
	err = verification.VerifyAdminIdentity(arpc.confAccount.SigningKey, in.Payload, in.Signature)
	if err != nil {
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
	err = arpc.mongoAddUserToTheWhitelist(ctx, common.HexToAddress(afgor.OwnerEthAddress), afgor.OwnerAnyID, afgor.OperationsCount)
	if err != nil {
		log.Error("failed to add user to the whitelist", zap.Error(err))
		return nil, errors.New("failed to add user to the whitelist")
	}

	return &out, err
}

func (arpc *anynsAARpc) GetDataNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (*nsp.GetDataNameRegisterResponse, error) {
	// 1 - check params
	err := verification.CheckRegisterParams(in)
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
	err := verification.CheckRegisterForSpaceParams(in)
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
	// 1 - unmarshal the signed request
	var cuor nsp.CreateUserOperationRequest
	err := proto.Unmarshal(in.Payload, &cuor)
	if err != nil {
		log.Error("can not unmarshal CreateUserOperationRequest", zap.Error(err))
		return nil, errors.New("can not unmarshal CreateUserOperationRequest")
	}

	// 2 - check users's signature
	err = verification.VerifyAnyIdentity(cuor.OwnerAnyID, in.Payload, in.Signature)
	if err != nil {
		log.Error("wrong Anytype signature", zap.Error(err))
		return nil, errors.New("wrong Anytype signature")
	}

	// 3 - check if user has enough operations left
	// will fail if AnyID was different
	ops, err := arpc.mongoGetUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress), cuor.OwnerAnyID)
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
	err = arpc.mongoDecreaseUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress))
	if err != nil {
		log.Error("failed to decrease operations count", zap.Error(err))
		return nil, errors.New("failed to decrease operations count")
	}

	// 6 - save operation to mongo (can be used later)
	err = arpc.mongoSaveOperation(ctx, opID, cuor)
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
