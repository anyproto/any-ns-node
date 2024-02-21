package cache

import (
	"context"
	"errors"
	"strings"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const CName = "any-ns.cache"

var log = logger.NewNamed(CName)

type NameDataItem struct {
	FullName string `bson:"name"`
	// always store in LOWER CASE!
	OwnerEthAddress    string `bson:"owner_eth_address"`
	OwnerScwEthAddress string `bson:"owner_scw_eth_address"`
	OwnerAnyAddress    string `bson:"owner_any_address"`
	SpaceId            string `bson:"space_id"`
	NameExpires        int64  `bson:"name_expires"`
}

// TODO: index it
type findNameDataByName struct {
	FullName string `bson:"name"`
}

// TODO: index it
type findNameDataByAddress struct {
	OwnerScwEthAddress string `bson:"owner_scw_eth_address"`
}

func New() app.Component {
	return &cacheService{}
}

type CacheService interface {
	// call it before you want to check in smart contracts
	// it will look up data in Mongo
	IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error)
	GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (out *nsp.NameByAddressResponse, err error)

	// call it when you need to read REAL data: smart contracts -> cache
	// will return "not found" if can not find name
	// will return no error if name is found and data was updated
	// will return error if something went wrong
	UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error)

	app.Component
}

type cacheService struct {
	confMongo config.Mongo
	itemColl  *mongo.Collection

	confContracts config.Contracts
	contracts     contracts.ContractsService
}

func (cs *cacheService) Name() (name string) {
	return CName
}

func (cs *cacheService) Init(a *app.App) (err error) {
	cs.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	cs.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	cs.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)

	// connect to mongo
	uri := cs.confMongo.Connect
	dbName := cs.confMongo.Database
	collectionName := "cache"

	// 1 - connect to DB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	cs.itemColl = client.Database(dbName).Collection(collectionName)
	if cs.itemColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	log.Info("mongo for cache connected!")

	return nil
}

// TODO: check if it is even called, this is not a app.ComponentRunnable instance
// so maybe it won't be called
func (cs *cacheService) Close(ctx context.Context) (err error) {
	if cs.itemColl != nil {
		err = cs.itemColl.Database().Client().Disconnect(ctx)
		cs.itemColl = nil
	}
	return
}

func (cs *cacheService) IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error) {
	// 1 - lookup in the cache
	item := &NameDataItem{}
	err = cs.itemColl.FindOne(ctx, findNameDataByName{FullName: in.FullName}).Decode(&item)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &nsp.NameAvailableResponse{Available: true}, nil
		}

		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}

	log.Debug("found item in cache", zap.String("FullName", in.FullName))

	// 2 - if found in the cache -> return false
	return &nsp.NameAvailableResponse{
		Available:          false,
		OwnerEthAddress:    item.OwnerEthAddress,
		OwnerScwEthAddress: item.OwnerScwEthAddress,
		OwnerAnyAddress:    item.OwnerAnyAddress,
		SpaceId:            item.SpaceId,
		NameExpires:        item.NameExpires,
	}, nil
}

func (cs *cacheService) GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (out *nsp.NameByAddressResponse, err error) {
	// 1 - lookup in the cache
	item := &NameDataItem{}

	// WARNING: convert to lower!
	inEthAddr := strings.ToLower(in.OwnerScwEthAddress)
	err = cs.itemColl.FindOne(ctx, findNameDataByAddress{OwnerScwEthAddress: inEthAddr}).Decode(&item)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &nsp.NameByAddressResponse{Found: false}, nil
		}

		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}

	// 2 - if found in the cache -> return
	return &nsp.NameByAddressResponse{
		Found: true,
		Name:  item.FullName,
	}, nil
}

// call it when data changes in smart contracts
// it will write to Mongo
func (cs *cacheService) setNameData(ctx context.Context, in *NameDataItem) (err error) {
	filter := findNameDataByName{FullName: in.FullName}
	opts := options.Replace().SetUpsert(true)

	// WARNING: always convert to lower case!
	in.OwnerScwEthAddress = strings.ToLower(in.OwnerScwEthAddress)
	in.OwnerEthAddress = strings.ToLower(in.OwnerEthAddress)

	_, err = cs.itemColl.ReplaceOne(ctx, filter, in, opts)
	if err != nil {
		log.Error("failed to update name data", zap.Error(err))
		return err
	}

	return nil
}

func (cs *cacheService) UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error) {
	log.Debug("reading data from smart contracts -> cache", zap.String("FullName", in.FullName))

	// 0 - create connection
	conn, err := cs.contracts.CreateEthConnection()
	if err != nil {
		log.Error("failed to connect to geth", zap.Error(err))
		return err
	}

	// 1 - convert to name hash
	nh, err := contracts.NameHash(in.FullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return err
	}

	// 2 - call contract's method
	log.Info("getting owner for name", zap.String("FullName", in.GetFullName()))
	addr, err := cs.contracts.GetOwnerForNamehash(ctx, conn, nh)
	if err != nil {
		if err.Error() == "not found" {
			log.Info("name is not registered yet...")
			return err
		}

		log.Error("can not get owner", zap.Error(err))
		return err
	}

	// the owner can be NameWrapper
	log.Info("received owner address", zap.String("Owner addr", addr.Hex()))
	if (addr == common.Address{}) {
		log.Info("name is not registered yet...")
		return nil
	}

	// 3 - if name is already registered, then get additional info
	log.Info("name is already registered...Getting additional info")
	ea, aa, si, exp, err := cs.contracts.GetAdditionalNameInfo(ctx, conn, addr, in.GetFullName())
	if err != nil {
		log.Error("failed to get additional info", zap.Error(err))
		return err
	}

	// 4 - update cache
	var ndi NameDataItem
	ndi.FullName = in.FullName
	ndi.OwnerAnyAddress = aa
	ndi.SpaceId = si
	ndi.NameExpires = exp.Int64()

	own, err := cs.contracts.GetOwnerOfSmartContractWallet(ctx, conn, common.HexToAddress(ea))
	if err != nil {
		log.Warn("failed to get SCW -> owner", zap.Error(err))

		ndi.OwnerScwEthAddress = ""
		ndi.OwnerEthAddress = strings.ToLower(ea)
	} else {
		ndi.OwnerScwEthAddress = strings.ToLower(ea)
		ndi.OwnerEthAddress = strings.ToLower(own.Hex())
	}

	// convert unixtime (big int) to string
	//timestamp := time.Unix(exp.Int64(), 0)
	//timeString := timestamp.Format("2001-01-02 15:04:05")

	err = cs.setNameData(ctx, &ndi)
	if err != nil {
		log.Error("failed to update name data after reading from smart contracts", zap.Error(err))
		return err
	}

	// success
	return nil
}
