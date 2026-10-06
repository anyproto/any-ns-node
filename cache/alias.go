package cache

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// aliases: records cached under a spelling that is not the canonical one of their name (e.g.
// "Foo.any": a node older than GO-7567 cached an operation's name as the client sent it). every
// read and write of this version uses the canonical spelling, so such a record is never
// refreshed or removed. until -dedupe-cache -refresh-apply migrates them:
//   - a live alias keeps its canonical name taken (lookups never answer available because of it)
//   - reverse lookups prefer the canonical record, an alias only counts when there is none
//   - the background refresh of an alias refreshes the canonical name and settles the alias
//
// the node knows them from a scan of the names (at start, then every aliasRescanInterval)

// aliasRescanInterval: how often the worker scans the names for aliases again (only an old node
// in the mixed window of an upgrade can write new ones)
const aliasRescanInterval = 10 * time.Minute

type aliasIndex struct {
	mu          sync.RWMutex
	byCanonical map[string][]string
	scannedAt   time.Time
}

func (a *aliasIndex) of(canonical string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.byCanonical[canonical]
}

// scanAliases: the non-canonical names of the cache, by canonical name
func (cs *cacheService) scanAliases(ctx context.Context) (map[string][]string, error) {
	values, err := cs.itemColl.Distinct(ctx, "name", bson.D{})
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, v := range values {
		name, _ := v.(string)
		canonical, err := cs.canonical(name)
		if err != nil || canonical == name {
			continue
		}
		out[canonical] = append(out[canonical], name)
	}
	for _, names := range out {
		sort.Strings(names)
	}
	return out, nil
}

// loadAliases scans the names for aliases; it warns while there are any
func (cs *cacheService) loadAliases(ctx context.Context) {
	m, err := cs.scanAliases(ctx)
	if err != nil {
		log.Warn("failed to scan the name cache for non-canonical names", zap.Error(err))
		return
	}
	cs.aliases.mu.Lock()
	cs.aliases.byCanonical = m
	cs.aliases.scannedAt = cs.now()
	cs.aliases.mu.Unlock()
	if n := len(m); n > 0 {
		log.Warn("the name cache has records under a non-canonical spelling: they keep their names taken, "+
			"run -dedupe-cache -refresh-apply to migrate them", zap.Int("names", n))
	}
}

// liveAlias: a live (not a tombstone) record cached under another spelling of the canonical name,
// nil if there is none. one indexed query, only if there are aliases of the name
func (cs *cacheService) liveAlias(ctx context.Context, canonical string) (*NameDataItem, error) {
	names := cs.aliases.of(canonical)
	if len(names) == 0 {
		return nil, nil
	}
	item := &NameDataItem{}
	err := cs.itemColl.FindOne(ctx, bson.M{"name": bson.M{"$in": names}, "removed": bson.M{"$ne": true}}).Decode(item)
	if err != nil {
		return nil, ignoreNoDocuments(err)
	}
	return item, nil
}

// settleAlias takes an alias out of the periodic scan (its canonical name is what is refreshed)
// and keeps lookups from handing it to the background again for a while
func (cs *cacheService) settleAlias(ctx context.Context, alias string) {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()
	_, err := cs.itemColl.UpdateOne(ctx, bson.M{"name": alias}, bson.M{
		"$unset": bson.M{"repair_at": ""},
		"$set":   bson.M{"refresh_next_at": cs.now().Add(expiredRefreshInterval).UnixMilli()},
	})
	if err != nil {
		log.Warn("failed to settle a non-canonical record", zap.String("FullName", alias), zap.Error(err))
	}
}

func ignoreNoDocuments(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	return err
}

// AliasStats is what MigrateAliases did (or would do, in a dry run)
type AliasStats struct {
	// records under a non-canonical spelling
	Aliases int
	// renamed to the canonical name (there was no canonical record), kept as they are, marked
	// for a refresh
	Renamed int
	// deleted: the canonical record is live, complete and at least as new
	Deleted int
	// left as they are (the canonical record is a tombstone, incomplete or older): for an operator
	Kept []string
}

func (cs *cacheService) MigrateAliases(ctx context.Context, apply bool) (stats AliasStats, err error) {
	byCanonical, err := cs.scanAliases(ctx)
	if err != nil {
		return stats, err
	}
	canonicals := make([]string, 0, len(byCanonical))
	for c := range byCanonical {
		canonicals = append(canonicals, c)
	}
	sort.Strings(canonicals)

	for _, canonical := range canonicals {
		for _, alias := range byCanonical[canonical] {
			stats.Aliases++
			decide := func(ctx context.Context) (string, error) {
				return cs.migrateAlias(ctx, alias, canonical, apply)
			}
			var outcome string
			if apply {
				outcome, err = withTx(ctx, cs, decide)
			} else {
				outcome, err = decide(ctx)
			}
			if err != nil {
				return stats, err
			}
			switch outcome {
			case "renamed":
				stats.Renamed++
			case "deleted":
				stats.Deleted++
			case "kept":
				stats.Kept = append(stats.Kept, alias)
			}
			log.Info("dedupe: a non-canonical record", zap.String("FullName", alias), zap.String("canonical", canonical),
				zap.String("outcome", outcome), zap.Bool("apply", apply))
		}
	}
	return stats, nil
}

// migrateAlias decides on one alias (in a transaction with apply): "renamed", "deleted", "kept"
// or "" (it is gone)
func (cs *cacheService) migrateAlias(ctx context.Context, alias, canonical string, apply bool) (string, error) {
	a, err := cs.getNameData(ctx, alias)
	if err != nil || a == nil {
		return "", err
	}
	c, err := cs.getNameData(ctx, canonical)
	if err != nil {
		return "", err
	}
	switch {
	case c == nil:
		// the record becomes the canonical one, as it is, and is refreshed as soon as possible
		if apply {
			_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": a.ID}, bson.M{"$set": bson.M{
				"name": canonical, "refresh_needed": true, "repair_at": max(int64(1), a.RefreshNextAt),
			}})
		}
		return "renamed", err
	case !c.Removed && !c.RefreshNeeded && c.ObservedBlock >= a.ObservedBlock:
		if apply {
			_, err = cs.itemColl.DeleteOne(ctx, bson.M{"_id": a.ID})
		}
		return "deleted", err
	}
	return "kept", nil
}
