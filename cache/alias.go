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

// the collation of the alias check: primary strength, it ignores case, width, compatibility
// forms and accents, so it is broader than the normalization (strength 2 is not: it tells "ſ"
// from "s", which the normalization maps to one). the candidates are then checked by the
// normalization itself
var aliasCollation = &options.Collation{Locale: "en", Strength: 1}

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

// canonIndexName: the index of the alias check by the canonical spelling (non-unique, sparse)
const canonIndexName = "canon"

func (cs *cacheService) ensureCanonIndex(ctx context.Context) {
	_, err := cs.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "canon", Value: 1}},
		Options: options.Index().SetName(canonIndexName).SetSparse(true),
	})
	if err != nil {
		log.Warn("failed to create the canon index of the name cache", zap.Error(err))
	}
}

// aliasCandidates bounds the records the collation fallback looks at (the spellings of a name,
// and names that differ only by accents)
const aliasCandidates = 64

// aliasesOf: the records whose name normalizes to canonical (the canonical record itself only
// with withCanonical). filter narrows them (e.g. live ones). authoritative at the time of the call:
//   - every record this version wrote (and every record -dedupe-cache migrated) has its
//     canonical spelling in canon: one indexed query
//   - a record without canon (written by an older node, not migrated yet): the collation
//     fallback, checked by the normalization itself. it can miss a spelling the collation does
//     not equate (e.g. punycode): -dedupe-cache -refresh-apply gives such records their canon
func (cs *cacheService) aliasesOf(ctx context.Context, canonical string, filter bson.M, withCanonical bool) ([]NameDataItem, error) {
	byCanon := bson.M{"canon": canonical}
	for k, v := range filter {
		byCanon[k] = v
	}
	cur, err := cs.itemColl.Find(ctx, byCanon)
	if err != nil {
		return nil, err
	}
	var out []NameDataItem
	if err = cur.All(ctx, &out); err != nil {
		return nil, err
	}

	legacy := bson.M{"name": canonical, "canon": bson.M{"$exists": false}}
	for k, v := range filter {
		legacy[k] = v
	}
	cur, err = cs.itemColl.Find(ctx, legacy, options.Find().SetCollation(aliasCollation).SetLimit(aliasCandidates))
	if err != nil {
		return nil, err
	}
	var candidates []NameDataItem
	if err = cur.All(ctx, &candidates); err != nil {
		return nil, err
	}
	for i := range candidates {
		// the collation is broader than the normalization (e.g. "tést" and "test"): only a record
		// whose name normalizes to the requested one counts
		if c, err := cs.canonical(candidates[i].FullName); err == nil && c == canonical {
			out = append(out, candidates[i])
		}
	}

	aliases := out[:0]
	for _, d := range out {
		if withCanonical || d.FullName != canonical {
			aliases = append(aliases, d)
		}
	}
	return aliases, nil
}

// liveAlias: a live (not a tombstone) record whose name normalizes to canonical, nil if there is
// none (see aliasesOf). it also finds a record renamed to the canonical name in between, or one
// written by an old node a moment ago
func (cs *cacheService) liveAlias(ctx context.Context, canonical string) (*NameDataItem, error) {
	// the canonical record itself counts too: it can have been written (or renamed to its name,
	// with its canon) since the lookup missed it
	aliases, err := cs.aliasesOf(ctx, canonical, bson.M{"removed": bson.M{"$ne": true}}, true)
	if err != nil || len(aliases) == 0 {
		return nil, err
	}
	return &aliases[0], nil
}

// retireLegacyAliases deletes the live records of the name under another spelling that have no
// block of their own (legacy): called with a removal confirmed at a finalized block, in its
// transaction. an alias read at a block (newer data than "unknown") stays
func (cs *cacheService) retireLegacyAliases(ctx context.Context, canonical string) error {
	aliases, err := cs.aliasesOf(ctx, canonical, bson.M{"removed": bson.M{"$ne": true}}, false)
	if err != nil {
		return err
	}
	for _, a := range aliases {
		if a.ObservedBlock != 0 {
			continue
		}
		log.Info("a removal retires a legacy record of the name under another spelling",
			zap.String("FullName", a.FullName), zap.String("canonical", canonical))
		if _, err := cs.itemColl.DeleteOne(ctx, bson.M{"_id": a.ID}); err != nil {
			return err
		}
	}
	return nil
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

// warnAliases logs how many records have no canonical spelling stored (written by an older
// node, not migrated yet): a hint for the operator, nothing depends on it
func (cs *cacheService) warnAliases(ctx context.Context) {
	n, err := cs.itemColl.CountDocuments(ctx, bson.M{"canon": bson.M{"$exists": false}})
	if err != nil {
		log.Warn("failed to count the records without a canonical spelling", zap.Error(err))
		return
	}
	if n > 0 {
		log.Warn("the name cache has records without a canonical spelling (canon): records under another "+
			"spelling keep their names taken, run -dedupe-cache -refresh-apply to migrate them", zap.Int64("records", n))
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
	// records given their canonical spelling (canon), aliases and canonical records alike
	Canon int
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

	// every record that has no canon gets it (the alias check finds a spelling by it that the
	// collation fallback can not)
	cur, err := cs.itemColl.Find(ctx, bson.M{"canon": bson.M{"$exists": false}}, options.Find().SetProjection(bson.M{"name": 1}))
	if err != nil {
		return stats, err
	}
	var rows []NameDataItem
	if err = cur.All(ctx, &rows); err != nil {
		return stats, err
	}
	for _, r := range rows {
		canonical, err := cs.canonical(r.FullName)
		if err != nil {
			log.Warn("dedupe: a cached name can not be normalized, no canon", zap.String("FullName", r.FullName), zap.Error(err))
			continue
		}
		stats.Canon++
		if apply {
			if _, err := cs.itemColl.UpdateOne(ctx, bson.M{"_id": r.ID, "canon": bson.M{"$exists": false}},
				bson.M{"$set": bson.M{"canon": canonical}}); err != nil {
				return stats, err
			}
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
				"name": canonical, "canon": canonical, "refresh_needed": true, "repair_at": max(int64(1), a.RefreshNextAt),
			}})
		}
		return "renamed", err
	case c.Removed && a.ObservedBlock == 0:
		// the canonical name was removed at a finalized block, the alias has no block of its own
		// (legacy): the chain is authoritative for the name
		if apply {
			_, err = cs.itemColl.DeleteOne(ctx, bson.M{"_id": a.ID})
		}
		return "deleted", err
	case c.ObservedBlock > 0 && !c.Removed && !c.Incomplete && c.ObservedBlock >= a.ObservedBlock:
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
