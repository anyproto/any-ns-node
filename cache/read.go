package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/contracts"
)

// errInconsistentRegistry: the registry has an owner for the name, but the registrar has no
// expiry for it. a failure (retried), never "not registered"
var errInconsistentRegistry = errors.New("the registry has an owner, but the registrar has no expiry")

// gracePeriodSec is the registrar's GRACE_PERIOD (a constant in the contract: 7776000 sec).
// a name can be renewed until nameExpires + gracePeriod, and only after that it is
// available again: the registrar's available() is "nameExpires + GRACE_PERIOD < block.timestamp"
const gracePeriodSec = 7776000

// vars and not consts only so that tests can shrink them:
//   - confirmTimeout bounds the reads that decide if the name is registered (the block, the
//     registry owner, nameExpires)
//   - enrichTimeout bounds the reads of the owner, AnyID and space ID that follow. they get
//     their own budget: a slow owner read must not turn a confirmed registration into a failure
var (
	confirmTimeout = 10 * time.Second
	enrichTimeout  = 10 * time.Second
)

// isLapsed is the registrar's available(): the name can be registered again.
// whole seconds, strict comparison, like the contract does with block.timestamp
func isLapsed(nameExpires int64, at time.Time) bool {
	return at.Unix() > nameExpires+gracePeriodSec
}

// observation: an empty record of the name, read at the block
func (cs *cacheService) observation(fullName string, block *contracts.Block) *NameDataItem {
	return &NameDataItem{
		FullName:          fullName,
		ObservedBlock:     block.Number,
		ObservedBlockHash: block.Hash.Hex(),
		ObservedBlockTime: int64(block.Time),
		ObservedAt:        cs.now().UnixMilli(),
	}
}

// registration: what the registrar says about the name at a block
type registration int

const (
	// never registered (no registry owner, no expiry)
	regNotRegistered registration = iota
	// registered (in the grace period too)
	regRegistered
	// expired and past the grace period: the registrar would let anyone register it again
	regLapsed
)

// readRegistration: is the name registered (not available at the registrar) at the block?
// returns the registry owner (zero: a reserved name, see below) and nameExpires if it is
// registered or lapsed
func (cs *cacheService) readRegistration(ctx context.Context, fullName string, nh [32]byte, block *contracts.Block) (owner common.Address, expires int64, state registration, err error) {
	addr, err := cs.contracts.GetOwnerForNamehash(ctx, nh, block.Hash)
	if err != nil {
		if err.Error() != "not found" {
			log.Error("can not get owner", zap.Error(err))
			return common.Address{}, 0, regNotRegistered, err
		}
		// the registry has nothing for that name: the same as a zero owner
		addr = common.Address{}
	}

	// the registrar decides, even without a registry owner: its registrant can reclaim the name
	// to address(0) and it stays reserved (not available) until it lapses
	exp, err := cs.contracts.GetNameExpires(ctx, fullName, block.Hash)
	if err != nil {
		log.Error("failed to get expiration of the name", zap.Error(err))
		return common.Address{}, 0, regNotRegistered, err
	}
	noExpiry := exp == nil || exp.Sign() <= 0

	// the registry has an owner, the registrar has no expiry: the two reads disagree (e.g. a
	// label hashed in another spelling). never a lapse: nothing is concluded from it
	if (addr != common.Address{}) && noExpiry {
		log.Error("the registry has an owner, but the registrar has no expiry", zap.String("FullName", fullName), zap.Int64("block", block.Number))
		return common.Address{}, 0, regNotRegistered, fmt.Errorf("%w: %s at block %d", errInconsistentRegistry, fullName, block.Number)
	}
	if noExpiry {
		// never registered (right after a registration this can also mean that the
		// provider has simply not caught up yet: not necessarily a free name)
		log.Warn("registry has no owner for the name", zap.String("FullName", fullName), zap.Int64("block", block.Number))
		return common.Address{}, 0, regNotRegistered, nil
	}

	// a lapsed name stays in the registry (owned by the NameWrapper, or reclaimed to zero), the
	// registrar would let anyone register it again. the cache keeps it reserved for its owner
	if isLapsed(exp.Int64(), time.Unix(int64(block.Time), 0)) {
		log.Info("name has lapsed (expired and past the grace period)",
			zap.String("FullName", fullName), zap.Int64("NameExpires", exp.Int64()), zap.Int64("block", block.Number))
		return addr, exp.Int64(), regLapsed, nil
	}
	if (addr == common.Address{}) {
		log.Info("the name has no registry owner, but the registrar still reserves it",
			zap.String("FullName", fullName), zap.Int64("NameExpires", exp.Int64()), zap.Int64("block", block.Number))
	}
	return addr, exp.Int64(), regRegistered, nil
}

// readNameData reads everything we cache about the name from the contracts at the latest block:
// the block header is fetched once, and every read is pinned to it by its hash. it writes nothing.
//   - nil and an error: nothing could be confirmed
//   - an observation and ErrNameNotRegistered: the name was never registered at the latest block
//     (no owner in the registry, no expiry). a single read
//   - a lapsed observation (Lapsed: the expiry, the registry owner and the block, nothing else)
//     and errLapsed: after a lapse the NameWrapper no longer reports the owner, so the owner,
//     AnyID and space ID are not read (they would fail forever): the cache keeps its own
//   - an incomplete observation (RefreshNeeded) and ErrNameDataIncomplete: registered, the owner
//     could not be read
//   - an observation and nil
func (cs *cacheService) readNameData(ctx context.Context, fullName string) (*NameDataItem, error) {
	// 1 - convert to name hash
	nh, err := contracts.NameHash(fullName)
	if err != nil {
		log.Error("can not convert FullName to namehash", zap.Error(err))
		return nil, err
	}

	// 2 - confirm: is the name registered at the latest block?
	confirmCtx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	head, err := cs.contracts.LatestBlock(confirmCtx)
	if err != nil {
		log.Error("can not get the latest block", zap.Error(err))
		return nil, err
	}
	obs := cs.observation(fullName, head)

	log.Info("getting owner for name", zap.String("FullName", fullName), zap.Int64("block", head.Number))
	addr, exp, state, err := cs.readRegistration(confirmCtx, fullName, nh, head)
	if err != nil {
		return nil, err
	}
	if state == regNotRegistered {
		return obs, ErrNameNotRegistered
	}
	obs.NameExpires = exp
	obs.RegistryOwner = strings.ToLower(addr.Hex())
	if state == regLapsed {
		obs.Lapsed = true
		return obs, errLapsed
	}
	if (addr == common.Address{}) {
		// reserved without a registry owner (reclaimed to address(0)): taken, complete as it is.
		// its owner can not be read from the chain: the cached one stays (see carryOver)
		obs.unread = unreadAll
		return obs, nil
	}

	// 3 - the name is confirmed: everything below only enriches it, with its own time budget.
	// a failure here is ErrNameDataIncomplete, never a plain error
	enrichCtx, cancelEnrich := context.WithTimeout(ctx, enrichTimeout)
	defer cancelEnrich()

	incomplete := func(unread unreadFields, cause error) (*NameDataItem, error) {
		log.Warn("name is registered, but its owner could not be read",
			zap.String("FullName", fullName), zap.Error(cause))
		obs.Incomplete, obs.RefreshNeeded, obs.unread = true, true, unread
		// re-read in the background, after a backoff
		obs.RefreshNextAt = cs.now().Add(refreshFailureBackoff).UnixMilli()
		return obs, fmt.Errorf("%w: %w", ErrNameDataIncomplete, cause)
	}

	// the owner, AnyID and space ID: read together, all or nothing
	ea, aa, si, err := cs.contracts.GetAdditionalNameInfo(enrichCtx, addr, fullName, head.Hash)
	if err != nil {
		return incomplete(unreadAll, err)
	}
	obs.OwnerAnyAddress = aa
	obs.SpaceId = si

	// never cache a guess: an owner we could not read must not end up in the cache
	if !common.IsHexAddress(ea) || (common.HexToAddress(ea) == common.Address{}) {
		return incomplete(unreadOwner, fmt.Errorf("the name wrapper returned owner %q", ea))
	}

	own, err := cs.contracts.GetScwOwner(enrichCtx, common.HexToAddress(ea), head.Hash)
	switch {
	case err == nil && (own != common.Address{}):
		// the name is owned by a smart contract wallet
		obs.OwnerScwEthAddress = strings.ToLower(ea)
		obs.OwnerEthAddress = strings.ToLower(own.Hex())
	case errors.Is(err, contracts.ErrNotAContract):
		// no code at the address: the name is owned by an EOA directly.
		// (a smart contract wallet that is not deployed yet looks the same, and its owner
		// can not be read from the chain)
		obs.OwnerScwEthAddress = ""
		obs.OwnerEthAddress = strings.ToLower(ea)
	case err != nil:
		// a failed lookup: do not store the wallet as if it was the owner
		obs.OwnerScwEthAddress = strings.ToLower(ea)
		return incomplete(unreadEOA, fmt.Errorf("wallet owner: %w", err))
	default:
		obs.OwnerScwEthAddress = strings.ToLower(ea)
		return incomplete(unreadEOA, fmt.Errorf("wallet %s has no owner", ea))
	}

	return obs, nil
}
