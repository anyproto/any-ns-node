package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anyproto/any-ns-node/config"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/app/logger"
	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const CName = "any-ns.cache"

// ErrNameNotRegistered is returned by UpdateInCache when the registry has no owner for the
// name (or it has lapsed): the cache does not have it as a taken name.
// it is not a transport error: callers should treat it as "not registered (yet)",
// never as "the cache is up to date"
var ErrNameNotRegistered = errors.New("name is not registered")

// errNotFinal: the latest block says that the name is not registered, but a finalized block
// does not confirm it (yet), so the cached record was left as it is. it is an
// ErrNameNotRegistered for the callers of UpdateInCache: not registered at the latest block
var errNotFinal = fmt.Errorf("%w at the latest block, not confirmed by a finalized one", ErrNameNotRegistered)

// ErrNameDataIncomplete is returned by UpdateInCache when the registry confirmed the name
// (it is registered and not lapsed), but its owner could not be read. the confirmed part is
// cached and the record is marked refresh_needed, so lookups keep reporting the name as taken
var ErrNameDataIncomplete = errors.New("name is registered, but its owner could not be read")

var log = logger.NewNamed(CName)

type NameDataItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	FullName string             `bson:"name"`
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
	// the latest read (unix ms) of this record's fork (block and hash), if a later one than
	// ObservedAt: a read that lost to this record (e.g. an incomplete one) still counts for
	// which fork was read last (see newer)
	ForkReadAt int64 `bson:"fork_read_at,omitempty"`

	// a tombstone: the name is not registered (or has lapsed) at ObservedBlock, a finalized
	// block. it keeps the block, so that an older read can not bring the name back. lookups
	// serve it as "not in the cache"
	Removed bool `bson:"removed,omitempty"`

	// the registry confirmed the name, but its owner could not be read: the record holds only
	// what was confirmed (it is still a taken name)
	RefreshNeeded bool `bson:"refresh_needed,omitempty"`
}

type findNameDataByName struct {
	FullName string `bson:"name"`
}

func New() app.Component {
	return &cacheService{}
}

type CacheService interface {
	// call it before you want to check in smart contracts
	// it will look up data in Mongo.
	// a cached name is always reported as taken: only the contracts can prove that a name is
	// available again, then the record becomes a tombstone (see UpdateInCache)
	IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error)
	GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (out *nsp.NameByAddressResponse, err error)
	GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (out *nsp.NameByAddressResponse, err error)

	// call it when you need to read REAL data: smart contracts -> cache
	// will return no error if name is found and data was updated
	// will return ErrNameNotRegistered if the registry has no owner for the name, or if the
	// name has lapsed (nameExpires + grace period has passed, it can be registered again).
	// that is an expected answer, not a failure: the cache does not have the name as taken
	// (a cached record becomes a tombstone only when a finalized block confirms it, until then
	// it stays as it is), so callers must never treat it as "the cache is up to date"
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

	// Mongo is a replica set (or a sharded cluster): cache writes run in transactions
	txSupported bool

	// the local clock. a field so that tests can move it
	now func() time.Time
}

func (cs *cacheService) Name() (name string) {
	return CName
}

func (cs *cacheService) Init(a *app.App) (err error) {
	conf := a.MustComponent(config.CName).(*config.Config)
	cs.confMongo = conf.Mongo
	cs.confContracts = conf.GetContracts()
	cs.contracts = a.MustComponent(contracts.CName).(contracts.ContractsService)
	cs.now = time.Now

	// connect to mongo
	uri := cs.confMongo.Connect
	dbName := cs.confMongo.Database
	collectionName := "cache"

	// 1 - connect to DB
	opts := options.Client().ApplyURI(uri)
	if opts.SocketTimeout == nil {
		opts.SetSocketTimeout(mongoSocketTimeout)
	}
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return err
	}

	cs.itemColl = client.Database(dbName).Collection(collectionName,
		options.Collection().SetWriteConcern(majorityWrite(storeTimeout)))
	if cs.itemColl == nil {
		return errors.New("failed to connect to MongoDB")
	}

	// 2 - every write of a name is a transaction: a replica set, unless explicitly allowed
	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cs.txSupported, err = supportsTransactions(initCtx, client)
	if err != nil {
		return fmt.Errorf("check mongo topology: %w", err)
	}
	if !cs.txSupported {
		if !conf.Cache.AllowUnsafeStandalone {
			return errors.New("the name cache needs a Mongo replica set (transactions): " +
				"cache.allowUnsafeStandalone allows a standalone Mongo, for local development only")
		}
		log.Warn("mongo is not a replica set: name cache writes are NOT transactional, " +
			"concurrent refreshes can store an older chain state (cache.allowUnsafeStandalone)")
	}

	// 3 - one record per name
	status, err := cs.ensureNameIndex(initCtx, true)
	if err != nil {
		return err
	}

	log.Info("mongo for cache connected!", zap.String("unique name index", status))

	return nil
}

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
	// a tombstone: a finalized block confirmed that the name is free
	if item == nil || item.Removed {
		return &nsp.NameAvailableResponse{Available: true}, nil
	}

	log.Debug("found item in cache", zap.String("FullName", in.FullName))

	// 2 - if found in the cache -> return false. an expired nameExpires is only a lower bound (a
	// renewal elsewhere does not touch the cache), it never makes a name available on its own
	return &nsp.NameAvailableResponse{
		Available:          false,
		OwnerEthAddress:    item.OwnerEthAddress,
		OwnerScwEthAddress: item.OwnerScwEthAddress,
		OwnerAnyAddress:    item.OwnerAnyAddress,
		SpaceId:            item.SpaceId,
		NameExpires:        item.NameExpires,
	}, nil
}

// getNameData returns the record of the name (a tombstone too), nil if there is none.
// the unique index on the name keeps it at one record per name
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
	// WARNING: convert to lower!
	inEthAddr := strings.ToLower(in.OwnerScwEthAddress)
	return cs.reverseLookup(ctx, bson.M{"owner_scw_eth_address": inEthAddr})
}

func (cs *cacheService) GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (out *nsp.NameByAddressResponse, err error) {
	// WARNING: DO NOT convert to lower!
	return cs.reverseLookup(ctx, bson.M{"owner_any_address": in.AnyAddress})
}

// reverseLookup finds a name by an owner field. tombstones never match
func (cs *cacheService) reverseLookup(ctx context.Context, filter bson.M) (*nsp.NameByAddressResponse, error) {
	filter["removed"] = bson.M{"$ne": true}

	item := &NameDataItem{}
	err := cs.itemColl.FindOne(ctx, filter).Decode(item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &nsp.NameByAddressResponse{Found: false}, nil
	}
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}
	return &nsp.NameByAddressResponse{Found: true, Name: item.FullName}, nil
}

func (cs *cacheService) UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error) {
	log.Debug("reading data from smart contracts -> cache", zap.String("FullName", in.FullName))

	_, err = cs.refresh(ctx, in.FullName, refreshOpts{})
	return err
}

// refreshOpts: how a refresh stores what it read
type refreshOpts struct {
	// decide exactly the same (the reads, the finality confirmation, the order of the records),
	// but write nothing (the dry run of the backfill)
	dry bool
}

// refresh reads the name from the contracts at the latest block and stores what the chain
// confirmed (see applyObservation). returns the record that is in the cache afterwards (nil:
// none); it can be a newer one written by somebody else. the error follows that record:
//   - nil: a complete registration
//   - ErrNameDataIncomplete: a registration without the owner
//   - ErrNameNotRegistered: none, or a tombstone. errNotFinal: the latest block says that the
//     name is not registered, but the cached registration stays (not confirmed by a finalized
//     block, or the cache has a newer one)
//   - any other error: nothing was decided, nothing was written
func (cs *cacheService) refresh(ctx context.Context, fullName string, o refreshOpts) (*NameDataItem, error) {
	obs, readErr := cs.readNameData(ctx, fullName)
	if errors.Is(readErr, ErrNameNotRegistered) {
		// a destructive change: only a finalized block can decide it
		obs, readErr = cs.confirmNotRegistered(ctx, fullName)
	}
	if obs == nil {
		// nothing to store
		return nil, readErr
	}

	var (
		stored *NameDataItem
		err    error
	)
	if o.dry {
		stored, err = cs.applyObservationTx(ctx, obs, true)
	} else {
		stored, err = cs.applyObservation(ctx, obs)
	}
	if err != nil {
		log.Error("failed to store name data", zap.String("FullName", fullName), zap.Error(err))
		return nil, fmt.Errorf("store the name: %w", err)
	}

	switch {
	case obs.Removed && stored != nil && !stored.Removed:
		// the confirmed removal is older than the cached registration: it stays
		return stored, fmt.Errorf("%w: the cache has a registration at a later block %d", errNotFinal, stored.ObservedBlock)
	case stored == nil || stored.Removed:
		return stored, ErrNameNotRegistered
	case stored.RefreshNeeded:
		if errors.Is(readErr, ErrNameDataIncomplete) {
			return stored, readErr
		}
		return stored, ErrNameDataIncomplete
	}
	return stored, nil
}

// confirmNotRegistered: the latest block says that the name is not registered. that makes a
// cached name available, so it must be final: a read of a block that is reorged away later (or
// of another fork, or of a lagging backend behind a load balancer) would make a registered name
// available. it reads the registry again at the finalized block:
//   - a tombstone at that block and ErrNameNotRegistered: confirmed
//   - nil and ErrNameNotRegistered: the cache has no registration of the name, nothing to remove
//   - nil and errNotFinal: the finalized block still has the name registered: the cache must
//     stay as it is (for now)
//   - nil and another error: the confirmation failed (e.g. the finalized block could not be read)
func (cs *cacheService) confirmNotRegistered(ctx context.Context, fullName string) (*NameDataItem, error) {
	stored, err := cs.getNameData(ctx, fullName)
	if err != nil {
		return nil, err
	}
	if stored == nil || stored.Removed {
		return nil, ErrNameNotRegistered
	}

	nh, err := contracts.NameHash(fullName)
	if err != nil {
		return nil, err
	}

	confirmCtx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	fin, err := cs.contracts.FinalizedBlock(confirmCtx)
	if err != nil {
		// a failure, not errNotFinal: nothing says that the name is still registered there
		log.Warn("can not get the finalized block", zap.String("FullName", fullName), zap.Error(err))
		return nil, fmt.Errorf("the finalized block: %w", err)
	}
	_, _, registered, err := cs.readRegistration(confirmCtx, fullName, nh, fin)
	if err != nil {
		return nil, fmt.Errorf("the registry at the finalized block: %w", err)
	}
	if registered {
		log.Info("the name is not registered at the latest block, but it is at the finalized one",
			zap.String("FullName", fullName), zap.Int64("finalized", fin.Number))
		return nil, errNotFinal
	}

	obs := cs.observation(fullName, fin)
	obs.Removed = true
	return obs, ErrNameNotRegistered
}
