package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anyproto/any-sync/app"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Maintainer holds the cache maintenance tools (the -refresh-cache, -dedupe-cache and
// -purge-tombstones flags of the node). it is not a part of CacheService
type Maintainer interface {
	// RefreshAll re-reads every cached name (tombstones too) from the contracts, with the same
	// decisions as the node (finality for a removal, the order of the records).
	// apply == false is a dry run: it decides the same, but writes nothing.
	// interval is the delay between two names, every name costs several contract calls
	RefreshAll(ctx context.Context, apply bool, interval time.Duration) (RefreshStats, error)

	// VerifyNameIndex checks the unique index on {name: 1}: any unique one is accepted whatever
	// its name (prod has name_1). a missing one is created with apply (it fails if the cache has
	// duplicates of a name: they are never resolved automatically). a non-unique one is an error
	// (ErrNameIndexNotUnique), it is never dropped
	VerifyNameIndex(ctx context.Context, apply bool) (NameIndexStats, error)

	// PurgeTombstones prepares the cache for a rollback to a version older than GO-7567: such a
	// node reads ANY record of a name as a taken name, a tombstone too. it deletes the
	// tombstones. it refuses to run while the cache has incomplete records without an owner
	// (ErrIncompleteRecords, PurgeStats.Incomplete names them): an old node would serve them
	// without the owner and never repair them. apply == false is a dry run: it only counts
	PurgeTombstones(ctx context.Context, apply bool) (PurgeStats, error)
}

// NewMaintainer is the cache for a one-off maintenance run (a Maintainer): it does not touch the
// indexes at start (VerifyNameIndex reports them) and runs no background refresh
func NewMaintainer() app.Component {
	return &cacheService{maintenance: true}
}

var _ Maintainer = (*cacheService)(nil)

// RefreshStats counts what RefreshAll did (or would do, in a dry run)
type RefreshStats struct {
	Total     int
	Unchanged int
	Updated   int
	// not registered (or lapsed), confirmed at a finalized block: a tombstone
	Removed int
	// not registered at the latest block, but the finalized block does not confirm it (yet), or
	// the cache has a newer registration: left as it is
	NotFinal int
	// nothing decided (e.g. a contract read failed, the finalized block could not be read), or
	// the owner could not be read (the confirmed part is stored, marked): run it again
	Failed int
}

// NameIndexStats is what VerifyNameIndex found (or did)
type NameIndexStats struct {
	// "exists" (a unique index on exactly {name: 1}, whatever its name), "created", or
	// "missing" (a dry run: it would be created)
	NameIndex string
	// names with more than one record (only counted when the index is missing): the index can
	// not be created until they are resolved by hand
	DuplicateNames int
}

// PurgeStats counts what PurgeTombstones did (or would do, in a dry run)
type PurgeStats struct {
	Tombstones int64
	// the names whose record is incomplete without an owner: nothing is purged while there are any
	Incomplete []string
}

// ErrIncompleteRecords is returned by PurgeTombstones while the cache has incomplete records
// without an owner: a node older than GO-7567 would serve them without the owner forever.
// refresh them first (-refresh-cache -refresh-apply, or the background repair)
var ErrIncompleteRecords = errors.New("the cache has incomplete records (no owner): refresh them first (-refresh-cache -refresh-apply), then purge")

// ErrDuplicateNames is returned by VerifyNameIndex when the unique index is missing and the cache
// has more than one record of a name: they must be resolved by hand before it can be created
var ErrDuplicateNames = errors.New("the cache has duplicates of a name: resolve them by hand, then create the unique index (-dedupe-cache -refresh-apply)")

func (cs *cacheService) RefreshAll(ctx context.Context, apply bool, interval time.Duration) (stats RefreshStats, err error) {
	// 1 - collect the names first: records are replaced while we go
	values, err := cs.itemColl.Distinct(ctx, "name", bson.D{})
	if err != nil {
		return stats, err
	}

	log.Info("refreshing cached names", zap.Int("count", len(values)), zap.Bool("apply", apply), zap.Duration("interval", interval))

	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 2 - re-read them one by one
	for i, v := range values {
		if i > 0 {
			select {
			case <-ctx.Done():
				return stats, ctx.Err()
			case <-ticker.C:
			}
		}

		name, _ := v.(string)
		stats.Total++
		if err := cs.refreshOne(ctx, name, apply, &stats); err != nil {
			stats.Failed++
			log.Warn("refresh: failed", zap.String("FullName", name), zap.Error(err))
		}
	}

	log.Info("refresh done", zap.Bool("apply", apply), zap.Any("stats", stats))
	return stats, nil
}

func (cs *cacheService) refreshOne(ctx context.Context, fullName string, apply bool, stats *RefreshStats) error {
	old, err := cs.getNameData(ctx, fullName)
	if err != nil {
		return err
	}

	// the same decisions with or without apply: only the writes are left out
	fresh, err := cs.refresh(ctx, fullName, refreshOpts{dry: !apply, confirm: true})

	switch {
	case errors.Is(err, errNotFinal):
		stats.NotFinal++
		log.Info("refresh: name is not registered at the latest block, not final, keeping it",
			zap.String("FullName", fullName), zap.Bool("apply", apply), zap.Error(err))
		return nil
	case errors.Is(err, ErrNameNotRegistered):
		if old == nil || old.Removed {
			stats.Unchanged++
			return nil
		}
		stats.Removed++
		log.Info("refresh: name is not registered any more (confirmed at a finalized block), a tombstone",
			zap.String("FullName", fullName), zap.Bool("apply", apply), zap.Any("cached", old))
		return nil
	case err != nil:
		// ErrNameDataIncomplete too: with apply, the confirmed part was stored and marked
		return err
	}

	if old != nil && sameNameData(old, fresh) {
		stats.Unchanged++
		return nil
	}

	stats.Updated++
	log.Info("refresh: cached name is stale, updating", zap.String("FullName", fullName),
		zap.Bool("apply", apply), zap.Any("cached", old), zap.Any("fresh", fresh))
	return nil
}

// the name data only, no bookkeeping
func sameNameData(a, b *NameDataItem) bool {
	return a.FullName == b.FullName &&
		a.Removed == b.Removed &&
		a.OwnerEthAddress == b.OwnerEthAddress &&
		a.OwnerScwEthAddress == b.OwnerScwEthAddress &&
		a.OwnerAnyAddress == b.OwnerAnyAddress &&
		a.SpaceId == b.SpaceId &&
		a.NameExpires == b.NameExpires
}

func (cs *cacheService) VerifyNameIndex(ctx context.Context, apply bool) (stats NameIndexStats, err error) {
	status, err := cs.ensureNameIndex(ctx, false)
	if err != nil || status == "exists" {
		stats.NameIndex = status
		log.Info("dedupe: the unique index on the name", zap.String("status", status), zap.Error(err))
		return stats, err
	}

	// missing: can it be created?
	cur, err := cs.itemColl.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$name"}, {Key: "n", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		{{Key: "$match", Value: bson.D{{Key: "n", Value: bson.D{{Key: "$gt", Value: 1}}}}}},
		{{Key: "$count", Value: "names"}},
	})
	if err != nil {
		return stats, err
	}
	var counted []struct {
		Names int `bson:"names"`
	}
	if err = cur.All(ctx, &counted); err != nil {
		return stats, err
	}
	if len(counted) > 0 {
		stats.DuplicateNames = counted[0].Names
	}
	stats.NameIndex = "missing"
	if stats.DuplicateNames > 0 {
		return stats, fmt.Errorf("%w: %d names", ErrDuplicateNames, stats.DuplicateNames)
	}
	if !apply {
		return stats, nil
	}
	stats.NameIndex, err = cs.ensureNameIndex(ctx, true)
	return stats, err
}

func (cs *cacheService) PurgeTombstones(ctx context.Context, apply bool) (stats PurgeStats, err error) {
	// 1 - no incomplete record without an owner
	incomplete, err := cs.itemColl.Distinct(ctx, "name", bson.M{
		"refresh_needed":    true,
		"removed":           bson.M{"$ne": true},
		"owner_eth_address": bson.M{"$in": bson.A{"", nil}},
	})
	if err != nil {
		return stats, err
	}
	for _, v := range incomplete {
		name, _ := v.(string)
		stats.Incomplete = append(stats.Incomplete, name)
	}
	if len(stats.Incomplete) > 0 {
		log.Warn("purge: incomplete records", zap.Strings("names", stats.Incomplete))
		return stats, fmt.Errorf("%w: %d names", ErrIncompleteRecords, len(stats.Incomplete))
	}

	// 2 - the tombstones. a tombstone that is replaced in between is not one any more: the
	// filter is checked per record, atomically
	filter := bson.M{"removed": true}
	if !apply {
		stats.Tombstones, err = cs.itemColl.CountDocuments(ctx, filter)
		return stats, err
	}
	res, err := cs.itemColl.DeleteMany(ctx, filter)
	if err != nil {
		return stats, err
	}
	stats.Tombstones = res.DeletedCount
	log.Info("purge done", zap.Bool("apply", apply), zap.Any("stats", stats))
	return stats, nil
}
