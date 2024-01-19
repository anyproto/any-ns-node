package anynsrpc

import (
	"context"
	"errors"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/config"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	contracts "github.com/anyproto/any-ns-node/contracts"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsRpc{}
}

type anynsRpc struct {
	cache           cache.CacheService
	contractsConfig config.Contracts
	contracts       contracts.ContractsService

	readFromCache bool
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.cache = a.MustComponent(cache.CName).(cache.CacheService)
	arpc.contractsConfig = a.MustComponent(config.CName).(*config.Config).GetContracts()
	arpc.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	arpc.readFromCache = a.MustComponent(config.CName).(*config.Config).ReadFromCache

	return nsp.DRPCRegisterAnyns(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsRpc) Name() (name string) {
	return CName
}

func (arpc *anynsRpc) IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (*nsp.NameAvailableResponse, error) {
	// 1 - if ReadFromCache is false -> always first read from smart contracts
	// if not, then always just read quickly from cache
	if !arpc.readFromCache {
		log.Debug("EXCPLICIT: read data from smart contracts -> cache", zap.String("FullName", in.FullName))
		err := arpc.cache.UpdateInCache(ctx, &nsp.NameAvailableRequest{
			FullName: in.FullName,
		})

		if err != nil {
			log.Error("failed to update in cache", zap.Error(err))
			return nil, err
		}
	}

	// 2 - check in cache (Mongo)
	return arpc.cache.IsNameAvailable(ctx, in)
}

func (arpc *anynsRpc) GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
	// 1 - if ReadFromCache is false -> always first read from smart contracts
	// if not, then always just read quickly from cache
	if !arpc.readFromCache {
		log.Debug("EXCPLICIT: reverse resolve using no cache", zap.String("FullName", in.OwnerScwEthAddress))
		return arpc.getNameByAddressDirectly(ctx, in)
	}

	// check in cache (Mongo)
	return arpc.cache.GetNameByAddress(ctx, in)
}

func (arpc *anynsRpc) getNameByAddressDirectly(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
	// 0 - check parameters
	if !common.IsHexAddress(in.OwnerScwEthAddress) {
		log.Error("invalid ETH address", zap.String("ETH address", in.OwnerScwEthAddress))
		return nil, errors.New("invalid ETH address")
	}

	// 1 - create connection
	conn, err := arpc.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return nil, err
	}

	// convert in.OwnerScwEthAddress to common.Address
	var addr = common.HexToAddress(in.OwnerScwEthAddress)

	name, err := arpc.contracts.GetNameByAddress(conn, addr)
	if err != nil {
		log.Error("failed to get name by address", zap.Error(err))
		return nil, err
	}

	// 2 - return results
	var res nsp.NameByAddressResponse

	if name == "" {
		res.Found = false
		return &res, nil
	}

	res.Found = true
	res.Name = name
	return &res, nil
}
