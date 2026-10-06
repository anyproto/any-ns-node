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

// readRegistration: is the name registered (owned and not lapsed) at the block?
// returns the registry owner and nameExpires if it is
func (cs *cacheService) readRegistration(ctx context.Context, fullName string, nh [32]byte, block *contracts.Block) (owner common.Address, expires int64, registered bool, err error) {
	addr, err := cs.contracts.GetOwnerForNamehash(ctx, nh, block.Hash)
	if err != nil {
		if err.Error() == "not found" {
			// the registry has nothing for that name. same outcome as a zero owner
			log.Info("registry does not know the name", zap.String("FullName", fullName))
			return common.Address{}, 0, false, nil
		}
		log.Error("can not get owner", zap.Error(err))
		return common.Address{}, 0, false, err
	}
	if (addr == common.Address{}) {
		// right after a registration this can also mean that the registry provider
		// has simply not caught up yet, so it is not necessarily a free name
		log.Warn("registry has no owner for the name", zap.String("FullName", fullName), zap.Int64("block", block.Number))
		return common.Address{}, 0, false, nil
	}

	exp, err := cs.contracts.GetNameExpires(ctx, fullName, block.Hash)
	if err != nil {
		log.Error("failed to get expiration of the name", zap.Error(err))
		return common.Address{}, 0, false, err
	}

	// a lapsed name stays in the registry (owned by the NameWrapper), but the registrar
	// lets anyone register it again: it is free, whatever the previous owner was
	if isLapsed(exp.Int64(), time.Unix(int64(block.Time), 0)) {
		log.Info("name has lapsed (expired and past the grace period)",
			zap.String("FullName", fullName), zap.Int64("NameExpires", exp.Int64()), zap.Int64("block", block.Number))
		return common.Address{}, 0, false, nil
	}
	return addr, exp.Int64(), true, nil
}

// readNameData reads everything we cache about the name from the contracts at the latest block:
// the block header is fetched once, and every read is pinned to it by its hash. it writes nothing.
//   - nil and an error: nothing could be confirmed
//   - an observation and ErrNameNotRegistered: the name is free at the latest block (no owner in
//     the registry, or lapsed). a single read, not final
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
	addr, exp, registered, err := cs.readRegistration(confirmCtx, fullName, nh, head)
	if err != nil {
		return nil, err
	}
	if !registered {
		return obs, ErrNameNotRegistered
	}
	obs.NameExpires = exp

	// 3 - the name is confirmed: everything below only enriches it, with its own time budget.
	// a failure here is ErrNameDataIncomplete, never a plain error
	enrichCtx, cancelEnrich := context.WithTimeout(ctx, enrichTimeout)
	defer cancelEnrich()

	incomplete := func(cause error) (*NameDataItem, error) {
		log.Warn("name is registered, but its owner could not be read",
			zap.String("FullName", fullName), zap.Error(cause))
		obs.RefreshNeeded = true
		return obs, fmt.Errorf("%w: %w", ErrNameDataIncomplete, cause)
	}

	// the owner, AnyID and space ID: read together, all or nothing
	ea, aa, si, err := cs.contracts.GetAdditionalNameInfo(enrichCtx, addr, fullName, head.Hash)
	if err != nil {
		return incomplete(err)
	}
	obs.OwnerAnyAddress = aa
	obs.SpaceId = si

	// never cache a guess: an owner we could not read must not end up in the cache
	if !common.IsHexAddress(ea) || (common.HexToAddress(ea) == common.Address{}) {
		return incomplete(fmt.Errorf("the name wrapper returned owner %q", ea))
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
		return incomplete(fmt.Errorf("wallet owner: %w", err))
	default:
		obs.OwnerScwEthAddress = strings.ToLower(ea)
		return incomplete(fmt.Errorf("wallet %s has no owner", ea))
	}

	return obs, nil
}
