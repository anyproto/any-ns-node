package anynsrpc

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"

	contracts "github.com/anyproto/any-ns-node/contracts"
	as "github.com/anyproto/any-ns-node/pb/anyns_api_server"
	"github.com/anyproto/any-ns-node/queue"
)

const CName = "any-ns.rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsRpc{}
}

type anynsRpc struct {
	contractsConfig config.Contracts
	contracts       contracts.ContractsService
	queue           queue.QueueService
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.queue = a.MustComponent(queue.CName).(queue.QueueService)

	return as.DRPCRegisterAnyns(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsRpc) Name() (name string) {
	return CName
}

func (arpc *anynsRpc) GetOperationStatus(ctx context.Context, in *as.GetOperationStatusRequest) (*as.OperationResponse, error) {
	currentState, err := arpc.queue.GetRequestStatus(ctx, in.OperationId)
	if err != nil {
		return nil, err
	}

	var res as.OperationResponse
	res.OperationId = in.OperationId
	res.OperationState = currentState
	return &res, nil
}

func (arpc *anynsRpc) IsNameAvailable(ctx context.Context, in *as.NameAvailableRequest) (*as.NameAvailableResponse, error) {
	// 0 - create connection
	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// 1 - convert to name hash
	nh, err := contracts.NameHash(in.FullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 2 - call contract's method
	log.Info("getting owner for name", zap.String("FullName", in.GetFullName()))
	addr, err := arpc.contracts.GetOwnerForNamehash(conn, nh)
	if err != nil {
		log.Error("failed to get owner", zap.Error(err))
		return nil, err
	}

	// 3 - covert to result
	// the owner can be NameWrapper
	log.Info("received owner address", zap.String("Owner addr", addr.Hex()))

	var res as.NameAvailableResponse
	var addrEmpty = common.Address{}

	if addr == addrEmpty {
		log.Info("name is not registered yet...")
		res.Available = true
		return &res, nil
	}

	// 4 - if name is not available, then get additional info
	log.Info("name is NOT available...Getting additional info")
	ea, aa, si, exp, err := arpc.contracts.GetAdditionalNameInfo(conn, addr, in.GetFullName())
	if err != nil {
		log.Error("failed to get additional info", zap.Error(err))
		return nil, err
	}

	// convert unixtime (big int) to string
	//timestamp := time.Unix(exp.Int64(), 0)
	//timeString := timestamp.Format("2001-01-02 15:04:05")

	log.Info("name is already registered...")
	res.Available = false
	res.OwnerEthAddress = ea
	res.OwnerAnyAddress = aa
	res.SpaceId = si
	res.NameExpires = exp.Int64()

	return &res, nil
}

func (arpc *anynsRpc) NameRegister(ctx context.Context, in *as.NameRegisterRequest) (*as.OperationResponse, error) {
	// 1 - check all parameters
	err := arpc.checkRegisterParams(in)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, err
	}

	// 2 - create new operation
	operationId, err := arpc.queue.AddNewRequest(ctx, in)
	if err != nil {
		log.Error("can not create new operation", zap.Error(err))
		return nil, err
	}

	return &as.OperationResponse{
		OperationState: as.OperationState_Pending,
		OperationId:    operationId,
	}, err
}

func (arpc *anynsRpc) NameRegisterSigned(ctx context.Context, in *as.NameRegisterSignedRequest) (*as.OperationResponse, error) {
	var resp as.OperationResponse

	// 1 - unmarshal the signed request
	var nrr as.NameRegisterRequest
	err := proto.Unmarshal(in.Payload, &nrr)
	if err != nil {
		resp.OperationState = as.OperationState_Error
		log.Error("can not unmarshal NameRegisterRequest", zap.Error(err))
		return &resp, err
	}

	// 2 - check signature
	err = VerifyIdentity(in, nrr.OwnerAnyAddress)
	if err != nil {
		resp.OperationState = as.OperationState_Error
		log.Error("identity is different", zap.Error(err))
		return &resp, err
	}

	// 3 - check all parameters
	err = arpc.checkRegisterParams(&nrr)
	if err != nil {
		log.Error("invalid parameters", zap.Error(err))
		return nil, err
	}

	// 4 - add to queue
	operationId, err := arpc.queue.AddNewRequest(ctx, &nrr)
	resp.OperationId = operationId
	resp.OperationState = as.OperationState_Pending
	return &resp, err
}

func (arpc *anynsRpc) checkRegisterParams(in *as.NameRegisterRequest) error {
	// 1 - check name
	if !checkName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return errors.New("invalid name")
	}

	// 2 - check ETH address
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return errors.New("invalid ETH address")
	}

	// 3 - check Any address
	if !checkAnyAddress(in.OwnerAnyAddress) {
		log.Error("invalid Any address", zap.String("Any address", in.OwnerAnyAddress))
		return errors.New("invalid Any address")
	}

	// 4 - space ID (if not empty)
	if in.SpaceId != "" {
		_, err := cid.Decode(in.SpaceId)

		if err != nil {
			log.Error("invalid SpaceId", zap.String("Any SpaceId", in.SpaceId))
			return errors.New("invalid SpaceId")
		}
	}

	// everything is OK
	return nil
}

func (arpc *anynsRpc) NameRenew(ctx context.Context, in *as.NameRenewRequest) (*as.OperationResponse, error) {
	// 1 - check all parameters
	if !checkName(in.FullName) {
		log.Error("invalid name", zap.String("name", in.FullName))
		return nil, errors.New("invalid name")
	}

	// 2 - create new operation
	operationId, err := arpc.queue.AddRenewRequest(ctx, in)
	if err != nil {
		log.Error("can not create new operation", zap.Error(err))
		return nil, err
	}

	return &as.OperationResponse{
		OperationState: as.OperationState_Pending,
		OperationId:    operationId,
	}, err
}

func (arpc *anynsRpc) GetNameByAddress(ctx context.Context, in *as.NameByAddressRequest) (*as.NameByAddressResponse, error) {
	// 0 - check parameters
	if !common.IsHexAddress(in.OwnerEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerEthAddress))
		return nil, errors.New("invalid ETH address")
	}

	// 1 - create connection
	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// convert in.OwnerEthAddress to common.Address
	var addr = common.HexToAddress(in.OwnerEthAddress)

	name, err := arpc.contracts.GetNameByAddress(conn, addr)
	if err != nil {
		log.Error("failed to get name by address", zap.Error(err))
		return nil, err
	}

	// 2 - return results
	var res as.NameByAddressResponse

	if name == "" {
		res.Found = false
		return &res, nil
	}

	res.Found = true
	res.Name = name
	return &res, nil
}
