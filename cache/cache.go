// Package cache mirrors the name contracts in Mongo (see "Name cache" in the README).
package cache

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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

// ErrNameNotRegistered is returned by UpdateInCache when the chain does not have the name and the
// cache has no record of it: "not registered (yet)".
// it is not a transport error: callers should treat it as "not registered (yet)",
// never as "the cache is up to date"
var ErrNameNotRegistered = errors.New("name is not registered")

// errNotOnChain: the latest block says that the name was never registered, but the cache has a
// record of it: the record stays as it is (taken), nothing is concluded from a read. right after a
// registration this is a lagging provider, the scheduled re-reads pick the name up. it is an
// ErrNameNotRegistered for the callers of UpdateInCache, a settled refresh for the background
var errNotOnChain = fmt.Errorf("%w on chain, the cached record is kept", ErrNameNotRegistered)

// errLapsed: readNameData read a lapsed name (expired and past the grace period). the
// observation carries the expiry and the block only (see refresh)
var errLapsed = errors.New("the name has lapsed")

// ErrNameDataIncomplete is returned by UpdateInCache when the registry confirmed the name
// (it is registered and not lapsed), but its owner could not be read. the confirmed part is
// cached and the record is marked refresh_needed, so lookups keep reporting the name as taken
var ErrNameDataIncomplete = errors.New("name is registered, but its owner could not be read")

// ErrNameReserved is returned by CheckReservation for a registration of a name that the cache has
// for another identity
var ErrNameReserved = errors.New("the name is reserved for another identity")

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

	// no background refresh takes the record before this (unix ms): a refresh in progress (a
	// lease) or a failed one (a backoff)
	RefreshNextAt int64 `bson:"refresh_next_at,omitempty"`
	// when the periodic scan takes the record (unix ms, see repairAt). indexed
	RepairAt int64 `bson:"repair_at,omitempty"`
	// scheduled re-reads of the record (unix ms, sorted): after a completed operation, see
	// rereadDelays. a read after one of them is done with it
	Rereads []int64 `bson:"rereads,omitempty"`

	// a tombstone written by 0.7.1 (the name lapsed, its owner fields were wiped). never written
	// any more: such a record is served as taken (with whatever owner fields it has), reverse
	// lookups skip it. -restore-tombstones brings the owner back and clears it
	Removed bool `bson:"removed,omitempty"`

	// the chain lets this name lapse (expired and past the grace period) at ObservedBlock; the
	// cache keeps it reserved for its owner: the owner fields stay, it is served as taken
	Lapsed bool `bson:"lapsed,omitempty"`

	// failed background refreshes in a row: the backoff of the next one grows with it (see
	// failureBackoff). reset by a successful write of the record
	RefreshFailures int `bson:"refresh_failures,omitempty"`

	// the registry confirmed the name, but the enrichment (owner, AnyID, space ID) failed: the
	// record holds what was confirmed, and what it could carry over from the previous record
	// (see carryOver). this is what orders two records of one block (see newer)
	Incomplete bool `bson:"incomplete,omitempty"`
	// the record is retried by the background (an incomplete read, or a failed refresh after an
	// operation): scheduling only, its data can be complete
	RefreshNeeded bool `bson:"refresh_needed,omitempty"`

	// the registry owner at ObservedBlock (lower case; the NameWrapper for a wrapped name, zero
	// for a reserved name without a registry owner). "" for records written before it was added
	RegistryOwner string `bson:"registry_owner,omitempty"`
	// the canonical spelling of the name (the normalizer's output), written with every record:
	// the alias check finds records under another spelling by it (see alias.go). indexed
	Canon string `bson:"canon,omitempty"`

	// not stored: which fields of an incomplete observation could not be read (see carryOver)
	unread unreadFields
}

// unreadFields: what an incomplete read could not read
type unreadFields int

const (
	// the owner, its wallet, AnyID and space ID (the NameWrapper and the resolver failed)
	unreadAll unreadFields = iota + 1
	// the owner and its wallet (the NameWrapper returned no owner)
	unreadOwner
	// the EOA owner of the wallet only
	unreadEOA
)

type findNameDataByName struct {
	FullName string `bson:"name"`
}

func New() app.Component {
	return &cacheService{}
}

type CacheService interface {
	// call it before you want to check in smart contracts
	// it will look up data in Mongo.
	// a cached name is always reported as taken, whatever its expiry: a name stays reserved
	// to the identity that registered it. only an operator removes a record (-release-name)
	IsNameAvailable(ctx context.Context, in *nsp.NameAvailableRequest) (out *nsp.NameAvailableResponse, err error)
	GetNameByAddress(ctx context.Context, in *nsp.NameByAddressRequest) (out *nsp.NameByAddressResponse, err error)
	GetNameByAnyId(ctx context.Context, in *nsp.NameByAnyIdRequest) (out *nsp.NameByAddressResponse, err error)

	// call it when you need to read REAL data: smart contracts -> cache
	// will return no error if name is found and data was updated
	// will return ErrNameNotRegistered if the chain does not have the name (never registered, or
	// lapsed) and the cache has no record of it. a cached record is never removed or wiped by a
	// read: a lapse is stored on the record (lapsed), "never registered" leaves it as it is (the
	// error then wraps ErrNameNotRegistered too). callers must never treat it as "the cache is
	// up to date"
	// will return ErrNameDataIncomplete if the name is registered, but its owner could not be
	// read: the cache has it as taken, without the owner (see ErrNameDataIncomplete)
	// will return any other error if something went wrong
	UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error)
	// UpdateInCacheAfterOperation is UpdateInCache for a name that a completed operation changed:
	// the same refresh at the latest block, and it schedules re-reads of the name (see
	// rereadDelays) in the same write, to pick up a lagging provider or a reorg
	UpdateInCacheAfterOperation(ctx context.Context, in *nsp.NameAvailableRequest) (err error)
	// RefreshAfterOperation hands a name that a completed operation changed to the background:
	// its re-reads are stored on its record (durable: a dropped request or a restart does not lose
	// them), and a refresh at the latest block runs (if it fails, the record is marked
	// refresh_needed). it never blocks and never waits for a write
	RefreshAfterOperation(fullName string)

	app.Component
}

type cacheService struct {
	confMongo config.Mongo
	itemColl  *mongo.Collection

	confContracts config.Contracts
	contracts     contracts.ContractsService

	// a one-off maintenance run (see NewMaintainer): no index changes at start, no background
	maintenance bool
	// Mongo is a replica set (or a sharded cluster): cache writes run in transactions
	txSupported bool
	// names are normalized like the registration does (config: ensip15validation)
	ensip15 bool
	// the transactions that withTx stopped waiting for (or not yet): Close lets them end (abort
	// or commit) before it disconnects, so that none is left open on the server
	inflight sync.WaitGroup
	// running transactions (a gauge, for tests)
	txActive atomic.Int64
	// the background re-read schedules (see scheduleRereadsAsync): Close waits for them
	asyncMu    sync.Mutex
	closed     bool
	async      sync.WaitGroup
	scheduling chan struct{}

	// the background refresh (see worker): the queue lookups hand names to, and the periodic
	// scan (0: off)
	queue          *refreshQueue
	repairInterval time.Duration
	stopWorker     context.CancelFunc
	workerDone     chan struct{}

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
	cs.ensip15 = conf.Ensip15Validation
	cs.scheduling = make(chan struct{}, maxScheduling)
	if !cs.maintenance {
		cs.queue = newRefreshQueue(refreshQueueSize)
	}
	switch {
	case conf.Cache.RepairIntervalSec == 0:
		cs.repairInterval = defaultRepairInterval
	case conf.Cache.RepairIntervalSec > 0:
		cs.repairInterval = time.Duration(conf.Cache.RepairIntervalSec) * time.Second
	}

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

	// 3 - one record per name (a maintenance run reports it: -dedupe-cache)
	status := "not checked"
	if !cs.maintenance {
		if status, err = cs.ensureNameIndex(initCtx, true); err != nil {
			return err
		}
		cs.ensureRepairIndex(initCtx)
		cs.ensureAliasIndex(initCtx)
		cs.ensureCanonIndex(initCtx)
		cs.warnAliases(initCtx)
	}

	log.Info("mongo for cache connected!", zap.String("unique name index", status), zap.Duration("repairInterval", cs.repairInterval))

	return nil
}

// the background refresh runs with the app
var _ app.ComponentRunnable = (*cacheService)(nil)

func (cs *cacheService) Run(_ context.Context) error {
	if cs.maintenance {
		return nil
	}
	cs.startWorker()
	return nil
}

func (cs *cacheService) Close(ctx context.Context) (err error) {
	if cs.stopWorker != nil {
		cs.stopWorker()
		<-cs.workerDone
		cs.stopWorker = nil
	}
	cs.asyncMu.Lock()
	cs.closed = true
	cs.asyncMu.Unlock()
	if cs.itemColl != nil {
		// bounded: every transaction ends within its storeTimeout, its session within another one
		ended := make(chan struct{})
		go func() {
			cs.async.Wait()
			cs.inflight.Wait()
			close(ended)
		}()
		select {
		case <-ended:
		case <-time.After(2 * storeTimeout):
			log.Warn("closing the name cache with a transaction still running")
		}
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
	// no canonical record, or a 0.7.1 tombstone (no owner): a record under another spelling keeps
	// the name taken too, and a live one answers with its owner
	if item == nil || item.Removed {
		if hookBeforeAliasCheck != nil {
			hookBeforeAliasCheck()
		}
		canonical, cerr := cs.canonical(in.FullName)
		if cerr != nil {
			canonical = in.FullName
		}
		alias, err := cs.liveAlias(ctx, canonical)
		if err != nil {
			log.Error("failed to get item from DB", zap.Error(err))
			return nil, err
		}
		if alias != nil && (item == nil || !alias.Removed) {
			if needsRefresh(alias, cs.now()) {
				cs.requestRefresh(refreshRequest{name: alias.FullName})
			}
			return nameTaken(alias), nil
		}
	}
	if item == nil {
		return &nsp.NameAvailableResponse{Available: true}, nil
	}
	// the record is served as it is. only an incomplete one (or one whose refresh after an
	// operation failed) is handed to the background refresh, never refreshed here: an expired
	// or lapsed name needs no chain read, it stays reserved for its owner
	if needsRefresh(item, cs.now()) {
		cs.requestRefresh(refreshRequest{name: item.FullName})
	}

	log.Debug("found item in cache", zap.String("FullName", in.FullName))

	// 2 - if found in the cache -> return false, whatever the expiry: clients can tell an
	// expired name by NameExpires. a tombstone of 0.7.1 too (taken, without an owner)
	return nameTaken(item), nil
}

func nameTaken(item *NameDataItem) *nsp.NameAvailableResponse {
	return &nsp.NameAvailableResponse{
		Available:          false,
		OwnerEthAddress:    item.OwnerEthAddress,
		OwnerScwEthAddress: item.OwnerScwEthAddress,
		OwnerAnyAddress:    item.OwnerAnyAddress,
		SpaceId:            item.SpaceId,
		NameExpires:        item.NameExpires,
	}
}

// getNameData returns the record of the name (a removed one too), nil if there is none.
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

// reverseLookupAliases bounds the aliases a reverse lookup keeps while it looks for a canonical
// record (one per name otherwise)
const reverseLookupAliases = 64

// reverseLookup finds a name by an owner field. 0.7.1 tombstones never match (their owner fields
// are empty anyway). a record under a
// non-canonical spelling (see alias.go) counts only if no canonical record matches and its name
// has no canonical record; the name is answered in its canonical spelling
func (cs *cacheService) reverseLookup(ctx context.Context, filter bson.M) (*nsp.NameByAddressResponse, error) {
	filter["removed"] = bson.M{"$ne": true}

	cur, err := cs.itemColl.Find(ctx, filter)
	if err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()

	found := func(item *NameDataItem, name string) (*nsp.NameByAddressResponse, error) {
		if needsRefresh(item, cs.now()) {
			cs.requestRefresh(refreshRequest{name: item.FullName})
		}
		return &nsp.NameByAddressResponse{Found: true, Name: name}, nil
	}
	// 1 - the first canonical record, whatever number of aliases match before it
	var aliases []NameDataItem
	for cur.Next(ctx) {
		var item NameDataItem
		if err = cur.Decode(&item); err != nil {
			return nil, err
		}
		canonical, err := cs.canonical(item.FullName)
		if err != nil || canonical == item.FullName {
			return found(&item, item.FullName)
		}
		if len(aliases) < reverseLookupAliases {
			aliases = append(aliases, item)
		}
	}
	if err = cur.Err(); err != nil {
		log.Error("failed to get item from DB", zap.Error(err))
		return nil, err
	}
	// 2 - an alias whose name has no canonical record
	for i := range aliases {
		canonical, _ := cs.canonical(aliases[i].FullName)
		c, err := cs.getNameData(ctx, canonical)
		if err != nil {
			log.Error("failed to get item from DB", zap.Error(err))
			return nil, err
		}
		if c == nil {
			return found(&aliases[i], canonical)
		}
	}
	return &nsp.NameByAddressResponse{Found: false}, nil
}

func (cs *cacheService) UpdateInCache(ctx context.Context, in *nsp.NameAvailableRequest) (err error) {
	log.Debug("reading data from smart contracts -> cache", zap.String("FullName", in.FullName))

	_, err = cs.refresh(ctx, in.FullName, refreshOpts{})
	return err
}

func (cs *cacheService) UpdateInCacheAfterOperation(ctx context.Context, in *nsp.NameAvailableRequest) (err error) {
	log.Debug("reading data from smart contracts -> cache, after an operation", zap.String("FullName", in.FullName))

	_, err = cs.refresh(ctx, in.FullName, refreshOpts{rereads: cs.rereadTimes()})
	return err
}

func (cs *cacheService) RefreshAfterOperation(fullName string) {
	cs.scheduleRereadsAsync(fullName)
	cs.requestRefresh(refreshRequest{name: fullName, afterOp: true})
}

// CheckReservation checks that the identity ownerAnyID may register the name: a name stays reserved to the
// identity that registered it, even after it lapsed on chain: the cache is the record of who holds
// a name (every registration goes through an operation of ours), so any record of the name (a
// lapsed one, a 0.7.1 tombstone, one under another spelling) reserves it. no record, or the same
// owner (a re-registration after a lapse, the payment node's renew-past-grace fallback): nil.
// a record without an owner (a tombstone, an incomplete read): ErrNameReserved. a cache that can
// not be read: an error too (fail closed, the caller retries). moving a name to another identity
// is a support action (-release-name)
func CheckReservation(ctx context.Context, cs CacheService, fullName, ownerAnyID string, ensip15 bool) error {
	name, err := contracts.NormalizeAnyName(fullName, ensip15)
	if err != nil {
		return fmt.Errorf("normalize the name %q: %w", fullName, err)
	}
	cached, err := cs.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: name})
	if err != nil {
		log.Error("failed to check the name in the cache", zap.String("FullName", name), zap.Error(err))
		return fmt.Errorf("check the name in the cache: %w", err)
	}
	if cached.Available || (ownerAnyID != "" && cached.OwnerAnyAddress == ownerAnyID) {
		return nil
	}
	log.Warn("a registration of a name reserved for another identity", zap.String("FullName", name),
		zap.String("owner", cached.OwnerAnyAddress), zap.String("requested for", ownerAnyID))
	return fmt.Errorf("%w: %s", ErrNameReserved, name)
}

// canonical is the spelling of the name that the registration uses (and the cache key)
func (cs *cacheService) canonical(fullName string) (string, error) {
	name, err := contracts.NormalizeAnyName(fullName, cs.ensip15)
	if err != nil {
		return "", fmt.Errorf("normalize the name %q: %w", fullName, err)
	}
	return name, nil
}

// refreshOpts: how a refresh stores what it read
type refreshOpts struct {
	// decide exactly the same (the reads, the order of the records), but write nothing (the dry
	// run of the backfill)
	dry bool
	// re-reads to schedule on the record (unix ms, see rereadDelays), in the same write
	rereads []int64
}

// refresh reads the name from the contracts at the latest block and stores what the chain
// confirmed (see applyObservation). it never removes a record and never wipes its owner.
// returns the record that is in the cache afterwards (nil: none); it can be a newer one written
// by somebody else. the error follows that record:
//   - nil: a complete registration, or a lapsed name kept for its owner (Lapsed)
//   - ErrNameDataIncomplete: a registration without the owner
//   - ErrNameNotRegistered: the chain does not have the name (or it lapsed) and nothing is
//     cached. errNotOnChain: the chain says never registered, the cached record stays
//   - any other error: nothing was decided, nothing was written
func (cs *cacheService) refresh(ctx context.Context, fullName string, o refreshOpts) (*NameDataItem, error) {
	// every contract read and the cache key use the canonical spelling: a raw "Foo.any" would
	// hash its label as "Foo" for nameExpires (0: never registered) while the registry
	// normalizes it to "foo.any"
	fullName, err := cs.canonical(fullName)
	if err != nil {
		return nil, err
	}

	obs, readErr := cs.readNameData(ctx, fullName)
	switch {
	case errors.Is(readErr, ErrNameNotRegistered):
		return cs.notOnChain(ctx, obs, o)
	case errors.Is(readErr, errLapsed):
		// obs.Lapsed: stored on the cached record, if there is one (see applyObservationTx)
		readErr = nil
	case obs == nil:
		// nothing to store
		return nil, readErr
	}

	var stored *NameDataItem
	if o.dry {
		stored, err = cs.applyObservationTx(ctx, obs, o)
	} else {
		stored, err = cs.applyObservation(ctx, obs, o)
	}
	if err != nil {
		log.Error("failed to store name data", zap.String("FullName", fullName), zap.Error(err))
		return nil, fmt.Errorf("store the name: %w", err)
	}

	switch {
	case stored == nil:
		// a lapsed name that the cache does not have: never ours, or the cache lost it (an
		// operator matter, -refresh-cache reports it). nothing is written
		return nil, ErrNameNotRegistered
	case stored.Incomplete:
		if errors.Is(readErr, ErrNameDataIncomplete) {
			return stored, readErr
		}
		return stored, ErrNameDataIncomplete
	}
	return stored, nil
}

// notOnChain: the latest block says that the name was never registered (no registry owner, no
// expiry). nothing cached: ErrNameNotRegistered. a cached record stays as it is (a lagging
// provider right after a registration, or a record the chain never had: for an operator), only
// its due re-reads are done by this read (the later ones stay), and a marked one backs off:
// errNotOnChain
func (cs *cacheService) notOnChain(ctx context.Context, obs *NameDataItem, o refreshOpts) (*NameDataItem, error) {
	settle := func(ctx context.Context) (*NameDataItem, error) {
		stored, err := cs.getNameData(ctx, obs.FullName)
		if err != nil || stored == nil {
			return stored, err
		}
		set, unset := bson.M{}, bson.M{}
		// a marked record stays marked (it is incomplete, or restored for one read): it is
		// retried after a backoff that grows with every such read, not at every lease
		if stored.RefreshNeeded {
			stored.RefreshFailures++
			stored.RefreshNextAt = max(stored.RefreshNextAt, cs.now().Add(failureBackoff(stored.RefreshFailures)).UnixMilli())
			set["refresh_failures"], set["refresh_next_at"] = stored.RefreshFailures, stored.RefreshNextAt
		}
		rereads := mergeRereads(stored.Rereads, o.rereads, obs.ObservedAt)
		if len(set) == 0 && slices.Equal(rereads, stored.Rereads) {
			return stored, nil
		}
		stored.Rereads = rereads
		stored.RepairAt = repairAt(stored)
		if o.dry {
			return stored, nil
		}
		setOrUnset(set, unset, "rereads", rereads, len(rereads) == 0)
		setOrUnset(set, unset, "repair_at", stored.RepairAt, stored.RepairAt == 0)
		_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": stored.ID}, updateDoc(set, unset))
		return stored, err
	}
	var stored *NameDataItem
	var err error
	if o.dry {
		stored, err = settle(ctx)
	} else {
		stored, err = withTx(ctx, cs, settle)
	}
	if err != nil {
		return nil, fmt.Errorf("store the name: %w", err)
	}
	if stored == nil {
		return nil, ErrNameNotRegistered
	}
	log.Warn("the chain does not have a cached name (never registered at the latest block), keeping the record",
		zap.String("FullName", obs.FullName), zap.Int64("block", obs.ObservedBlock), zap.Int64("cached block", stored.ObservedBlock))
	return stored, errNotOnChain
}
