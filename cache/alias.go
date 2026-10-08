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
//   - an alias keeps its canonical name taken (see liveAlias: checked on every lookup that
//     finds no canonical record or a removed one, nothing is remembered)
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

// canonIndexName: the index of the alias check by the canonical spelling (non-unique, sparse,
// with aliasCollation: the check is one query with that collation)
const canonIndexName = "canon_ci"

func (cs *cacheService) ensureCanonIndex(ctx context.Context) {
	_, err := cs.itemColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "canon", Value: 1}},
		Options: options.Index().SetName(canonIndexName).SetSparse(true).SetCollation(aliasCollation),
	})
	if err != nil {
		log.Warn("failed to create the canon index of the name cache", zap.Error(err))
	}
}

// aliasCandidates bounds the records the collation fallback looks at (the spellings of a name,
// and names that differ only by accents)
const aliasCandidates = 64

// aliasesOf: the records whose name normalizes to canonical (the canonical record itself only
// with withCanonical). filter narrows them (e.g. live ones).
//
// one query (aliasCollation, both branches indexed): a record matches by its canonical spelling
// (canon: every record this version wrote or -dedupe-cache migrated) or by its name. one query,
// so a migration in between can not make a record slip between two of them: a record renamed to
// the canonical name or given its canon still matches the branch it matched before. the
// collation is broader than the normalization (e.g. "tést" and "test"), the candidates are
// checked by canon or by the normalization itself. a spelling the collation does not equate
// (e.g. punycode) is found only once -dedupe-cache gave the record its canon
func (cs *cacheService) aliasesOf(ctx context.Context, canonical string, filter bson.M, withCanonical bool) ([]NameDataItem, error) {
	query := bson.M{"$or": bson.A{bson.M{"canon": canonical}, bson.M{"name": canonical}}}
	for k, v := range filter {
		query[k] = v
	}
	cur, err := cs.itemColl.Find(ctx, query, options.Find().SetCollation(aliasCollation).SetLimit(aliasCandidates))
	if err != nil {
		return nil, err
	}
	var candidates []NameDataItem
	if err = cur.All(ctx, &candidates); err != nil {
		return nil, err
	}
	var out []NameDataItem
	for _, d := range candidates {
		if d.FullName == canonical && !withCanonical {
			continue
		}
		if d.Canon == canonical {
			out = append(out, d)
			continue
		}
		if c, err := cs.canonical(d.FullName); err == nil && c == canonical {
			out = append(out, d)
		}
	}
	return out, nil
}

// liveAlias: a record whose name normalizes to canonical, nil if there is none (see aliasesOf):
// a live one if there is one, otherwise a removed one (a 0.7.1 tombstone keeps its name taken
// too). it also finds a record renamed to the canonical name in between, or one written by an
// old node a moment ago
func (cs *cacheService) liveAlias(ctx context.Context, canonical string) (*NameDataItem, error) {
	// the canonical record itself counts too: it can have been written (or renamed to its name,
	// with its canon) since the lookup missed it
	aliases, err := cs.aliasesOf(ctx, canonical, nil, true)
	if err != nil || len(aliases) == 0 {
		return nil, err
	}
	for i := range aliases {
		if !aliases[i].Removed {
			return &aliases[i], nil
		}
	}
	return &aliases[0], nil
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
	aliases, err := cs.itemColl.CountDocuments(ctx, bson.M{"canon": bson.M{"$exists": true}, "$expr": bson.M{"$ne": bson.A{"$canon", "$name"}}})
	if err != nil {
		log.Warn("failed to count the records under a non-canonical spelling", zap.Error(err))
		return
	}
	if aliases > 0 {
		log.Warn("the name cache has records under a non-canonical spelling that -dedupe-cache kept: they keep "+
			"their names taken until an operator resolves them (see the README)", zap.Int64("records", aliases))
	}
}

// settleAlias takes an alias out of the periodic scan and out of the lookups' refreshes (its
// canonical name is what is refreshed): no repair_at, no refresh_needed, and a while before a
// lookup could hand it again anyway
func (cs *cacheService) settleAlias(ctx context.Context, alias string) {
	ctx, cancel := boundedCtx(ctx)
	defer cancel()
	_, err := cs.itemColl.UpdateOne(ctx, bson.M{"name": alias}, bson.M{
		"$unset": bson.M{"repair_at": "", "refresh_needed": ""},
		"$set":   bson.M{"refresh_next_at": cs.now().Add(aliasSettleInterval).UnixMilli()},
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
