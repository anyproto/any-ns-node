package anynsrpc

import (
	"context"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	"github.com/anyproto/any-sync/net/rpc/server"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
)

const CName = "any-ns.rpc"

var log = logger.NewNamed(CName)

func New() app.Component {
	return &anynsRpc{}
}

type anynsRpc struct {
	cache cache.CacheService
}

func (arpc *anynsRpc) Init(a *app.App) (err error) {
	arpc.cache = a.MustComponent(cache.CName).(cache.CacheService)

	return nsp.DRPCRegisterAnyns(a.MustComponent(server.CName).(server.DRPCServer), arpc)
}

func (arpc *anynsRpc) Name() (name string) {
	return CName
}

func (arpc *anynsRpc) IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (*nsp.NameAvailableResponse, error) {
	// check in cache (Mongo)
	return arpc.cache.IsNameAvailable(ctx, in)
}

func (arpc *anynsRpc) GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
	// check in cache (Mongo)
	return arpc.cache.GetNameByAddress(ctx, in)
}

/*
func (arpc *anynsRpc) GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (*nsp.NameByAddressResponse, error) {
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
	var res nsp.NameByAddressResponse

	if name == "" {
		res.Found = false
		return &res, nil
	}

	res.Found = true
	res.Name = name
	return &res, nil
}*/
