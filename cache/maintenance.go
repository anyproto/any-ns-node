package cache

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/anyproto/any-sync/app"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Maintainer holds the cache maintenance tools (the -refresh-cache, -dedupe-cache, -release-name
// and -restore-tombstones flags of the node). it is not a part of CacheService
type Maintainer interface {
	// RefreshAll re-reads every cached name from the contracts, with the same decisions as the
	// node (a lapse is stored on the record, nothing is ever removed, the order of the records).
	// apply == false is a dry run: it decides the same, but writes nothing.
	// interval is the delay between two names, every name costs several contract calls
	RefreshAll(ctx context.Context, apply bool, interval time.Duration) (RefreshStats, error)

	// VerifyNameIndex checks the unique index on {name: 1}: any unique one is accepted whatever
	// its name (prod has name_1). a missing one is created with apply (it fails if the cache has
	// duplicates of a name: they are never resolved automatically). a non-unique one is an error
	// (ErrNameIndexNotUnique), it is never dropped
	VerifyNameIndex(ctx context.Context, apply bool) (NameIndexStats, error)

	// MigrateAliases resolves the records cached under a non-canonical spelling (see alias.go),
	// one transaction per record: no canonical record -> the alias is renamed to the canonical
	// name (its data kept, marked for a refresh); the canonical record is live, complete and at
	// least as new -> the alias is deleted; otherwise both stay (Kept). apply == false is a dry run
	MigrateAliases(ctx context.Context, apply bool) (AliasStats, error)

	// ReleaseName is the support transfer of a name: a name stays reserved to its identity until
	// an operator releases it. it finds the record of the name (normalized) and, with apply,
	// deletes exactly that record; the records under other spellings (aliases) are listed, never
	// deleted. ErrNameNotCached if there is no record. apply == false is a dry run
	ReleaseName(ctx context.Context, fullName string, apply bool) (ReleaseStats, error)

	// RestoreTombstones brings back the owners of the 0.7.1 tombstones (removed records, their
	// owner fields wiped) from a mongoexport of the cache (JSON lines, MongoDB Extended JSON).
	// one transaction per record, only while it is still removed. a tombstone that the export
	// has no owner for is listed (Missing) and left as it is (served as taken, without an owner).
	// apply == false is a dry run
	RestoreTombstones(ctx context.Context, export io.Reader, apply bool) (RestoreStats, error)
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
	// lapsed on chain (expired and past the grace period): marked lapsed, or kept so; the
	// record and its owner stay
	Lapsed int
	// the chain says that a cached name was never registered: the record is kept as it is, for
	// an operator to look at
	NotOnChain int
	// nothing decided (e.g. a contract read failed), or the owner could not be read (the
	// confirmed part is stored, marked): run it again
	Failed int
	// cached under a spelling that is not the canonical one (e.g. "Foo.any"): the canonical
	// name was refreshed, the record under this spelling is left as it is (taken), for an
	// operator to look at
	NonCanonical int
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

// ReleaseStats is what ReleaseName found (and did)
type ReleaseStats struct {
	// the normalized name
	Name string
	// the record of the name, nil if there is none
	Record *NameDataItem
	// the records of the name under other spellings: listed, never deleted
	Aliases []string
	// the record was deleted (apply)
	Deleted bool
}

// ErrNameNotCached is returned by ReleaseName when the cache has no record of the name
var ErrNameNotCached = errors.New("the name is not in the cache")

// RestoreStats is what RestoreTombstones did (or would do, in a dry run)
type RestoreStats struct {
	// the removed records of the cache
	Tombstones int
	// given their owner back from the export
	Restored int
	// the export has no owner for them: left as they are (taken, without an owner)
	Missing []string
}

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
	name, err := cs.canonical(fullName)
	if err != nil {
		return err
	}
	if name != fullName {
		stats.NonCanonical++
		log.Warn("refresh: a cached name is not in its canonical spelling, refreshing the canonical one, the record is left as it is",
			zap.String("FullName", fullName), zap.String("canonical", name))
		fullName = name
	}
	old, err := cs.getNameData(ctx, fullName)
	if err != nil {
		return err
	}

	// the same decisions with or without apply: only the writes are left out
	fresh, err := cs.refresh(ctx, fullName, refreshOpts{dry: !apply, noChangeRereads: true})

	switch {
	case errors.Is(err, errNotOnChain):
		stats.NotOnChain++
		log.Warn("refresh: the chain says that a cached name was never registered, keeping the record",
			zap.String("FullName", fullName), zap.Bool("apply", apply), zap.Any("cached", old))
		return nil
	case errors.Is(err, ErrNameNotRegistered):
		// nothing cached under the canonical spelling (only an alias has the name): nothing to do
		stats.Unchanged++
		return nil
	case err != nil:
		// ErrNameDataIncomplete too: with apply, the confirmed part was stored and marked
		return err
	}

	if fresh.Lapsed {
		stats.Lapsed++
		if old == nil || !old.Lapsed || old.NameExpires != fresh.NameExpires {
			log.Info("refresh: the name has lapsed on chain, keeping it reserved for its owner", zap.String("FullName", fullName),
				zap.Bool("apply", apply), zap.Any("cached", old), zap.Any("fresh", fresh))
		}
		return nil
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
		a.Lapsed == b.Lapsed &&
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

func (cs *cacheService) ReleaseName(ctx context.Context, fullName string, apply bool) (stats ReleaseStats, err error) {
	stats.Name, err = cs.canonical(fullName)
	if err != nil {
		return stats, err
	}
	stats.Record, err = cs.getNameData(ctx, stats.Name)
	if err != nil {
		return stats, err
	}
	aliases, err := cs.aliasesOf(ctx, stats.Name, nil, false)
	if err != nil {
		return stats, err
	}
	for _, a := range aliases {
		stats.Aliases = append(stats.Aliases, a.FullName)
	}
	sort.Strings(stats.Aliases)
	if stats.Record == nil {
		return stats, fmt.Errorf("%w: %s", ErrNameNotCached, stats.Name)
	}
	if !apply {
		return stats, nil
	}
	res, err := cs.itemColl.DeleteOne(ctx, bson.M{"_id": stats.Record.ID})
	if err != nil {
		return stats, err
	}
	stats.Deleted = res.DeletedCount == 1
	log.Warn("release: the record of the name was deleted by an operator, the name can be registered again",
		zap.String("FullName", stats.Name), zap.Any("record", stats.Record), zap.Strings("aliases", stats.Aliases))
	return stats, nil
}

// exportedRecord: what RestoreTombstones takes from a record of the export
type exportedRecord struct {
	FullName           string `bson:"name"`
	OwnerEthAddress    string `bson:"owner_eth_address"`
	OwnerScwEthAddress string `bson:"owner_scw_eth_address"`
	OwnerAnyAddress    string `bson:"owner_any_address"`
	SpaceID            string `bson:"space_id"`
	NameExpires        int64  `bson:"name_expires"`
	Removed            bool   `bson:"removed"`
}

// test hook, nil in production: runs in RestoreTombstones before the write of a record
var hookBeforeRestore func(name string)

// maxExportLine bounds one line (one record) of the export
const maxExportLine = 1 << 20

// readExport: the records of the export that have an owner, by name (exact and canonical)
func (cs *cacheService) readExport(export io.Reader) (map[string]exportedRecord, error) {
	out := map[string]exportedRecord{}
	sc := bufio.NewScanner(export)
	sc.Buffer(make([]byte, 0, 64*1024), maxExportLine)
	for line := 1; sc.Scan(); line++ {
		raw := bytes.TrimSpace(sc.Bytes())
		if len(raw) == 0 {
			continue
		}
		var r exportedRecord
		if err := bson.UnmarshalExtJSON(raw, false, &r); err != nil {
			return nil, fmt.Errorf("the export, line %d: %w", line, err)
		}
		if r.Removed || r.OwnerAnyAddress == "" {
			continue
		}
		// the exact spelling wins over another one that normalizes to the same name
		out[r.FullName] = r
		if c, err := cs.canonical(r.FullName); err == nil && c != r.FullName {
			if _, ok := out[c]; !ok {
				out[c] = r
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("the export: %w", err)
	}
	return out, nil
}

func (cs *cacheService) RestoreTombstones(ctx context.Context, export io.Reader, apply bool) (stats RestoreStats, err error) {
	byName, err := cs.readExport(export)
	if err != nil {
		return stats, err
	}
	cur, err := cs.itemColl.Find(ctx, bson.M{"removed": true}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return stats, err
	}
	var tombstones []NameDataItem
	if err = cur.All(ctx, &tombstones); err != nil {
		return stats, err
	}

	now := cs.now()
	for i := range tombstones {
		t := &tombstones[i]
		stats.Tombstones++
		r, ok := byName[t.FullName]
		if !ok {
			stats.Missing = append(stats.Missing, t.FullName)
			log.Warn("restore: the export has no owner for a tombstone, it stays (taken, without an owner)", zap.String("FullName", t.FullName))
			continue
		}

		// the record as it will be: the owner from the export, the block of the tombstone
		rec := *t
		rec.Removed, rec.Incomplete = false, false
		rec.OwnerEthAddress = strings.ToLower(r.OwnerEthAddress)
		rec.OwnerScwEthAddress = strings.ToLower(r.OwnerScwEthAddress)
		rec.OwnerAnyAddress = r.OwnerAnyAddress
		rec.SpaceId = r.SpaceID
		rec.NameExpires = r.NameExpires
		if rec.Canon == "" {
			if c, err := cs.canonical(rec.FullName); err == nil {
				rec.Canon = c
			}
		}
		set := bson.M{
			"owner_eth_address":     rec.OwnerEthAddress,
			"owner_scw_eth_address": rec.OwnerScwEthAddress,
			"owner_any_address":     rec.OwnerAnyAddress,
			"space_id":              rec.SpaceId,
			"name_expires":          rec.NameExpires,
		}
		unset := bson.M{"removed": "", "incomplete": ""}
		if rec.Canon != "" {
			set["canon"] = rec.Canon
		}
		if isLapsed(rec.NameExpires, now) {
			// the reason it was removed: kept reserved, no chain read
			rec.Lapsed = true
			set["lapsed"] = true
		} else {
			// renewed since the export, or removed by a read that was wrong: the background
			// reads it once
			rec.Lapsed, rec.RefreshNeeded = false, true
			set["refresh_needed"] = true
			unset["lapsed"] = ""
		}
		rec.RepairAt = repairAt(&rec)
		setOrUnset(set, unset, "repair_at", rec.RepairAt, rec.RepairAt == 0)

		if apply {
			if hookBeforeRestore != nil {
				hookBeforeRestore(t.FullName)
			}
			matched, err := withTx(ctx, cs, func(ctx context.Context) (int64, error) {
				res, err := cs.itemColl.UpdateOne(ctx, bson.M{"_id": t.ID, "removed": true}, updateDoc(set, unset))
				if err != nil {
					return 0, err
				}
				return res.MatchedCount, nil
			})
			if err != nil {
				return stats, fmt.Errorf("restore %s: %w", t.FullName, err)
			}
			if matched == 0 {
				// written since the scan (a registration read, say): not a tombstone any more
				log.Info("restore: the record is not removed any more, leaving it", zap.String("FullName", t.FullName))
				continue
			}
		}
		stats.Restored++
		log.Info("restore: the owner of a tombstone", zap.String("FullName", t.FullName), zap.Bool("apply", apply),
			zap.String("owner", rec.OwnerAnyAddress), zap.Bool("lapsed", rec.Lapsed))
	}
	log.Info("restore done", zap.Bool("apply", apply), zap.Int("tombstones", stats.Tombstones),
		zap.Int("restored", stats.Restored), zap.Int("missing", len(stats.Missing)))
	return stats, nil
}
