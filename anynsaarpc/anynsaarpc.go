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
	"github.com/anyproto/any-sync/util/crypto"
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

	itemColl *mongo.Collection

	contracts contracts.ContractsService
	aa        accountabstraction.AccountAbstractionService
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

func (arpc *anynsAARpc) Init(a *app.App) (err error) {
	arpc.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.aa = a.MustComponent(accountabstraction.CName).(accountabstraction.AccountAbstractionService)

	// connect to mongo
	uri := arpc.confMongo.Connect
	dbName := arpc.confMongo.Database
	collectionName := "aa-users"

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

	return nsp.DRPCRegisterAnynsAccountAbstraction(a.MustComponent(server.CName).(server.DRPCServer), arpc)
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

func (arpc *anynsAARpc) GetUserAccount(ctx context.Context, in *nsp.GetUserAccountRequest) (*nsp.UserAccount, error) {
	var res nsp.UserAccount
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

	res.OperationsCountLeft, err = arpc.mongoGetUserOperationsCount(
		ctx,
		common.HexToAddress(in.OwnerEthAddress),
		// becuase AnyID is empty -> it will not check it
		"",
	)

	if err != nil {
		log.Error("failed to get operations count left", zap.Error(err))
		return nil, err
	}

	// return
	return &res, nil
}

func (arpc *anynsAARpc) GetOperation(ctx context.Context, in *nsp.GetOperationStatusRequest) (*nsp.OperationResponse, error) {
	var out nsp.OperationResponse

	status, err := arpc.aa.GetOperation(ctx, in.OperationId)
	if err != nil {
		log.Error("failed to get operation info", zap.Error(err))
		return nil, err
	}

	out.OperationId = fmt.Sprint(in.OperationId)
	out.OperationState = status.OperationState

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
		return nil, err
	}

	// 2 - check signature
	err = arpc.aa.AdminVerifyIdentity(in.Payload, in.Signature)
	if err != nil {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, err
	}

	// 3 - determine SCW of user wallet
	scwa, err := arpc.aa.GetSmartWalletAddress(ctx, common.HexToAddress(afuar.OwnerEthAddress))
	if err != nil {
		log.Error("failed to get smart wallet address", zap.Error(err))
		return nil, err
	}

	// 4 - mint tokens to that SCW
	opID, err := arpc.aa.AdminMintAccessTokens(ctx, scwa, big.NewInt(int64(afuar.NamesCount)))
	if err != nil {
		log.Error("failed to mint tokens", zap.Error(err))
		return nil, err
	}

	// 3 - return
	// TODO: add to queue?
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
		return nil, err
	}

	// 2 - check signature
	err = arpc.aa.AdminVerifyIdentity(in.Payload, in.Signature)
	if err != nil {
		log.Error("not an Admin!!!", zap.Error(err))
		return nil, err
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
		return nil, err
	}

	return &out, err
}

func (arpc *anynsAARpc) GetDataNameRegister(ctx context.Context, in *nsp.NameRegisterRequest) (*nsp.GetDataNameRegisterResponse, error) {
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

	var out nsp.GetDataNameRegisterResponse
	// user should sign it
	out.Data = dataOut
	// user should pass it back to us
	out.Context = contextData

	return &out, nil
}

func (arpc *anynsAARpc) VerifyAnyIdentity(ownerIdStr string, payload []byte, signature []byte) (err error) {
	// to read in the PeerID format
	ownerAnyIdentity, err := crypto.DecodePeerId(ownerIdStr)

	// to read ID in the marshaled format
	//arr := []byte(ownerIdStr)
	//ownerAnyIdentity, err := crypto.UnmarshalEd25519PublicKeyProto(arr)

	if err != nil {
		log.Error("failed to unmarshal public key", zap.Error(err))
		return err
	}

	// 2 - verify signature
	res, err := ownerAnyIdentity.Verify(payload, signature)
	if err != nil || !res {
		return errors.New("signature is different")
	}

	// success
	return nil
}

// once user got data by using method like GetDataNameRegister, and signed it, now he can create a new operation
func (arpc *anynsAARpc) CreateUserOperation(ctx context.Context, in *nsp.CreateUserOperationRequestSigned) (*nsp.OperationResponse, error) {
	// 1 - unmarshal the signed request
	var cuor nsp.CreateUserOperationRequest
	err := proto.Unmarshal(in.Payload, &cuor)
	if err != nil {
		log.Error("can not unmarshal CreateUserOperationRequest", zap.Error(err))
		return nil, err
	}

	// 2 - check users's signature
	err = arpc.VerifyAnyIdentity(cuor.OwnerAnyID, in.Payload, in.Signature)
	if err != nil {
		log.Error("wrong Anytype signature", zap.Error(err))
		return nil, err
	}

	// 3 - check if user has enough operations left
	// will fail if AnyID was different
	ops, err := arpc.mongoGetUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress), cuor.OwnerAnyID)
	if err != nil {
		log.Error("failed to get operations count", zap.Error(err))
		return nil, err
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
		return nil, err
	}

	// 5 - decrease operations count for that user
	err = arpc.mongoDecreaseUserOperationsCount(ctx, common.HexToAddress(cuor.OwnerEthAddress))
	if err != nil {
		log.Error("failed to decrease operations count", zap.Error(err))
		return nil, err
	}

	// 6 - return result
	var out nsp.OperationResponse
	out.OperationId = opID
	out.OperationState = nsp.OperationState_Pending

	return nil, nil
}
