package anynsrpc

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"

	contracts "github.com/anyproto/any-ns-node/contracts"
	as "github.com/anyproto/any-ns-node/pb/anyns_api"
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
	addr, err := arpc.contracts.GetOwnerForNamehash(ctx, conn, nh)
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
	ea, aa, si, exp, err := arpc.contracts.GetAdditionalNameInfo(ctx, conn, addr, in.GetFullName())
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
