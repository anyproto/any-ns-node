package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const CName = "any-ns.cache"

// ErrNameNotRegistered is returned by UpdateInCache when the registry has no owner
// for the name, so nothing was written to the cache.
// it is not a transport error: callers should treat it as "not registered (yet)",
// never as "the cache is up to date"
var ErrNameNotRegistered = errors.New("name is not registered")

// ErrNameDataIncomplete is returned by UpdateInCache when the registry confirmed the name
// (it is registered and not lapsed), but its owner could not be read. the confirmed part is
// cached and the record is marked refresh_needed, so lookups keep reporting the name as taken
var ErrNameDataIncomplete = errors.New("name is registered, but its owner could not be read")

var log = logger.NewNamed(CName)

type NameDataItem struct {
	FullName string `bson:"name"`
	// always store in LOWER CASE!
	OwnerEthAddress    string `bson:"owner_eth_address"`
	OwnerScwEthAddress string `bson:"owner_scw_eth_address"`
	OwnerAnyAddress    string `bson:"owner_any_address"`
	SpaceId            string `bson:"space_id"`
	NameExpires        int64  `bson:"name_expires"`

	// the block the contracts were read at (all the reads of a refresh by its hash).
	// 0 for records written before it was added
	ObservedBlock     int64  `bson:"observed_block,omitempty"`
	ObservedBlockHash string `bson:"observed_block_hash,omitempty"`
	// that block's timestamp (unix seconds): lapse decisions use it, not the local clock
	ObservedBlockTime int64 `bson:"observed_block_time,omitempty"`
	// the local time of the read (unix ms)
	ObservedAt int64 `bson:"observed_at,omitempty"`

	// the registry confirmed the name, but its owner could not be read: the record holds only
	// what was confirmed (it is still a taken name)
	RefreshNeeded bool `bson:"refresh_needed,omitempty"`
}

// TODO: index it
type findNameDataByName struct {
	FullName string `bson:"name"`
}

// TODO: index it
type findNameDataByAddress struct {
	OwnerScwEthAddress string `bson:"owner_scw_eth_address"`
}

type findNameDataByAnyAddress struct {
	OwnerAnyAddress string `bson:"owner_any_address"`
}

func New() app.Component {
	return &cacheService{}
}

type CacheService interface {
	// call it before you want to check in smart contracts
	// it will look up data in Mongo
	IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error)
	GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (out *nsp.NameByAddressResponse, err error)
	GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (out *nsp.NameByAddressResponse, err error)

	// call it when you need to read REAL data: smart contracts -> cache
	// will return no error if name is found and data was updated
	// will return ErrNameNotRegistered if the registry has no owner for the name, or if the
	// name has lapsed (nameExpires + grace period has passed, it can be registered again).
	// that is an expected answer, not a failure: the cache does not have the name as taken,
	// so callers must never treat it as "the cache is up to date"
	// will return ErrNameDataIncomplete if the name is registered, but its owner could not be
	// read: the cache has it as taken, without the owner (see ErrNameDataIncomplete)
	// will return any other error if something went wrong
	UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error)

	app.Component
}

type cacheService struct {
	confMongo config.Mongo
	itemColl  *mongo.Collection

	confContracts config.Contracts
	contracts     contracts.ContractsService

	// the local clock. a field so that tests can move it
	now func() time.Time
}

func (cs *cacheService) Name() (name string) {
	return CName
}

func (cs *cacheService) Init(a *app.App) (err error) {
	cs.confMongo = a.MustComponent(config.CName).(*config.Config).Mongo
	cs.confContracts = a.MustComponent(config.CName).(*config.Config).GetContracts()
	cs.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	cs.now = time.Now

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
	item, err := cs.getNameData(ctx, in.FullName)
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}
	if item == nil {
		return &nsp.NameAvailableResponse{Available: true}, nil
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

// getNameData returns the record of the name, nil if there is none
func (cs *cacheService) getNameData(ctx context.Context, fullName string) (*NameDataItem, error) {
	item := &NameDataItem{}
	err := cs.itemColl.FindOne(ctx, findNameDataByName{FullName: fullName}).Decode(item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
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

func (cs *cacheService) GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (out *nsp.NameByAddressResponse, err error) {
	// 1 - lookup in the cache
	item := &NameDataItem{}

	// WARNING: DO NOT convert to lower!
	err = cs.itemColl.FindOne(ctx, findNameDataByAnyAddress{OwnerAnyAddress: in.AnyAddress}).Decode(&item)

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

	// 1 - read the name at the latest block (all reads pinned to it)
	obs, err := cs.readNameData(ctx, in.FullName)
	if obs == nil || errors.Is(err, ErrNameNotRegistered) {
		// nothing confirmed, or not registered: the cache is not changed.
		// (a cached record is never removed on the word of a single read)
		return err
	}

	// 2 - update cache: the confirmed part, even if the owner could not be read
	if serr := cs.setNameData(ctx, obs); serr != nil {
		log.Error("failed to update name data after reading from smart contracts", zap.Error(serr))
		return serr
	}
	return err
}
