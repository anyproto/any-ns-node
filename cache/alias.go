package cache

import (
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// aliases: records cached under a spelling that is not the canonical one of their name (e.g.
// "Foo.any": a node older than GO-7567 cached an operation's name as the client sent it). every
// read and write of this version uses the canonical spelling, so such a record is never
// refreshed or removed. until -dedupe-cache -refresh-apply migrates them:
//   - a live alias keeps its canonical name taken (see liveAlias: checked on every lookup that
//     finds no live canonical record, nothing is remembered)
//   - reverse lookups prefer the canonical record, an alias only counts when there is none
//   - the background refresh of an alias refreshes the canonical name and settles the alias

// the collation of the alias check: case- and width-insensitive (strength 2 ignores tertiary
// differences), so every spelling that the normalization maps to the canonical one matches (the
// candidates are then checked by the normalization itself)
var aliasCollation = &options.Collation{Locale: "en", Strength: 2}

// aliasIndexName: the index of the alias check (non-unique, with aliasCollation). a failure to
// create it only makes the check a scan of the cache (a few thousand records)
const aliasIndexName = "name_ci"

func (cs *cacheService) ensureAliasIndex(ctx context.Context) {
	_, err := cs.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName(aliasIndexName).SetCollation(aliasCollation),
	})
	if err != nil {
		log.Warn("failed to create the alias index of the name cache", zap.Error(err))
	}
}

// aliasCandidates bounds the records the alias check looks at (one per spelling of a name)
const aliasCandidates = 16

// liveAlias: a live (not a tombstone) record whose name normalizes to canonical, nil if there is
// none. authoritative: one query of the collection at the time of the lookup (it also finds a
// record renamed to the canonical name in between, or written by an old node a moment ago)
func (cs *cacheService) liveAlias(ctx context.Context, canonical string) (*NameDataItem, error) {
	cur, err := cs.itemColl.Find(ctx, bson.M{"name": canonical, "removed": bson.M{"$ne": true}},
		options.Find().SetCollation(aliasCollation).SetLimit(aliasCandidates))
	if err != nil {
		return nil, err
	}
	var items []NameDataItem
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	for i := range items {
		// the collation is broader than the normalization (e.g. "ss" and "ß"): only a record
		// whose name normalizes to the requested one counts
		if c, err := cs.canonical(items[i].FullName); err == nil && c == canonical {
			return &items[i], nil
		}
	}
	return nil, nil
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

// warnAliases logs how many names have aliases (a hint for the operator, nothing depends on it)
func (cs *cacheService) warnAliases(ctx context.Context) {
	m, err := cs.scanAliases(ctx)
	if err != nil {
		log.Warn("failed to scan the name cache for non-canonical names", zap.Error(err))
		return
	}
	if n := len(m); n > 0 {
		log.Warn("the name cache has records under a non-canonical spelling: they keep their names taken, "+
			"run -dedupe-cache -refresh-apply to migrate them", zap.Int("names", n))
	}
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

// AliasStats is what MigrateAliases did (or would do, in a dry run)
type AliasStats struct {
	// records under a non-canonical spelling
	Aliases int
	// renamed to the canonical name (there was no canonical record), kept as they are, marked
	// for a refresh
	Renamed int
	// deleted: the canonical record has its block, is live, complete and at least as new
	Deleted int
	// left as they are (the canonical record has no block, is a tombstone, incomplete or older),
	// the canonical record marked for a refresh: for an operator, or a later run
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
	case c.ObservedBlock > 0 && !c.Removed && !c.RefreshNeeded && c.ObservedBlock >= a.ObservedBlock:
		// the canonical record was read from the chain (it has its block), is live, complete and
		// at least as new
		if apply {
			_, err = cs.itemColl.DeleteOne(ctx, bson.M{"_id": a.ID})
		}
		return "deleted", err
	}
	// unknown or conflicting freshness (e.g. two legacy records without a block): both stay, the
	// canonical one is refreshed as soon as possible, a later run decides
	if apply {
		_, err = cs.itemColl.UpdateOne(ctx, bson.M{"_id": c.ID}, bson.M{"$set": bson.M{
			"refresh_needed": true, "repair_at": max(int64(1), c.RefreshNextAt),
		}})
	}
	return "kept", err
}
