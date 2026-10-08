package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/app"
)

// cacheTask is what the cache maintenance mode of the node runs
type cacheTask struct {
	dedupe bool
	// -release-name: the name to release ("" : none)
	release string
	// -restore-tombstones -restore-from: the export to restore the tombstones from
	restore     bool
	restoreFrom string
	refresh     bool
	apply       bool
	interval    time.Duration
}

// runCacheMaintenance runs -dedupe-cache, -release-name, -restore-tombstones and/or -refresh-cache
// (in that order) as a one-off process (no RPC server, no background refresh).
// a dry run unless -refresh-apply is set. returns the process exit code: not 0 if anything failed
// (or needs an operator: a tombstone the export has no owner for, a name that is not cached)
func runCacheMaintenance(ctx context.Context, a *app.App, task cacheTask) int {
	a.Register(contracts.New()).
		Register(cache.NewMaintainer())

	if err := a.Start(ctx); err != nil {
		log.Error("can't start app", zap.Error(err))
		return 1
	}
	defer func() {
		if err := a.Close(ctx); err != nil {
			log.Error("close error", zap.Error(err))
		}
	}()

	return maintainCache(ctx, a.MustComponent(cache.CName).(cache.Maintainer), task, os.Stdout)
}

func maintainCache(ctx context.Context, m cache.Maintainer, task cacheTask, out io.Writer) int {
	mode := "DRY RUN (nothing was written, use -refresh-apply to write)"
	if task.apply {
		mode = "APPLIED"
	}

	// 1 - one record per name: the unique index
	if task.dedupe {
		stats, err := m.VerifyNameIndex(ctx, task.apply)
		_, _ = fmt.Fprintf(out, "dedupe %s: unique name index=%s names with duplicates=%d\n",
			mode, stats.NameIndex, stats.DuplicateNames)
		if err != nil {
			log.Error("dedupe failed", zap.Error(err))
			return 1
		}
		aliases, err := m.MigrateAliases(ctx, task.apply)
		_, _ = fmt.Fprintf(out, "dedupe %s: non-canonical records=%d renamed=%d deleted=%d kept=%d canon set=%d\n",
			mode, aliases.Aliases, aliases.Renamed, aliases.Deleted, len(aliases.Kept), aliases.Canon)
		for _, name := range aliases.Kept {
			_, _ = fmt.Fprintf(out, "  kept: %s\n", name)
		}
		if err != nil {
			log.Error("dedupe failed", zap.Error(err))
			return 1
		}
	}

	// 2 - a support transfer: the record of the name goes, the name can be registered again
	if task.release != "" {
		stats, err := m.ReleaseName(ctx, task.release, task.apply)
		if stats.Record != nil {
			r := stats.Record
			_, _ = fmt.Fprintf(out, "release %s: name=%s owner_any_address=%q owner_eth_address=%q owner_scw_eth_address=%q "+
				"space_id=%q name_expires=%d lapsed=%v removed=%v deleted=%v\n",
				mode, stats.Name, r.OwnerAnyAddress, r.OwnerEthAddress, r.OwnerScwEthAddress, r.SpaceId, r.NameExpires,
				r.Lapsed, r.Removed, stats.Deleted)
		}
		for _, name := range stats.Aliases {
			_, _ = fmt.Fprintf(out, "  alias (not deleted, it keeps the name taken): %s\n", name)
		}
		if err != nil {
			_, _ = fmt.Fprintf(out, "release %s: %s: %v\n", mode, task.release, err)
			log.Error("release failed", zap.Error(err))
			return 1
		}
	}

	// 3 - the owners of the 0.7.1 tombstones, from an export of the cache
	if task.restore {
		stats, err := restoreTombstones(ctx, m, task)
		_, _ = fmt.Fprintf(out, "restore %s: tombstones=%d restored=%d missing=%d\n",
			mode, stats.Tombstones, stats.Restored, len(stats.Missing))
		for _, name := range stats.Missing {
			_, _ = fmt.Fprintf(out, "  missing: %s\n", name)
		}
		if err != nil {
			log.Error("restore failed", zap.Error(err))
			return 1
		}
		if len(stats.Missing) > 0 {
			log.Error("some tombstones have no owner in the export", zap.Int("missing", len(stats.Missing)))
			return 1
		}
	}

	// 4 - every name from the contracts
	if task.refresh {
		stats, err := m.RefreshAll(ctx, task.apply, task.interval)
		_, _ = fmt.Fprintf(out, "refresh %s: total=%d unchanged=%d updated=%d lapsed=%d not-on-chain=%d failed=%d non-canonical=%d\n",
			mode, stats.Total, stats.Unchanged, stats.Updated, stats.Lapsed, stats.NotOnChain, stats.Failed, stats.NonCanonical)
		if err != nil {
			log.Error("refresh failed", zap.Error(err))
			return 1
		}
		if stats.Failed > 0 {
			log.Error("some names could not be refreshed", zap.Int("failed", stats.Failed))
			return 1
		}
	}
	return 0
}

func restoreTombstones(ctx context.Context, m cache.Maintainer, task cacheTask) (cache.RestoreStats, error) {
	if task.restoreFrom == "" {
		return cache.RestoreStats{}, errors.New("-restore-tombstones needs -restore-from <mongoexport file>")
	}
	f, err := os.Open(task.restoreFrom)
	if err != nil {
		return cache.RestoreStats{}, err
	}
	defer func() { _ = f.Close() }()
	return m.RestoreTombstones(ctx, f, task.apply)
}

// rpcHost: the host of a provider URL, for the logs (the URL itself can carry an API key)
func rpcHost(rpcURL string) string {
	u, err := url.Parse(rpcURL)
	if err != nil || u.Host == "" {
		return "(unparsable)"
	}
	return u.Host
}
