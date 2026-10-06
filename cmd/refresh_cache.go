package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/anyproto/any-ns-node/cache"
	"github.com/anyproto/any-ns-node/contracts"
	"github.com/anyproto/any-sync/app"
)

// cacheTask is what the cache maintenance mode of the node runs
type cacheTask struct {
	dedupe   bool
	purge    bool
	refresh  bool
	apply    bool
	interval time.Duration
}

// runCacheMaintenance runs -dedupe-cache, -purge-tombstones and/or -refresh-cache (in that order)
// as a one-off process (no RPC server, no background refresh).
// a dry run unless -refresh-apply is set. returns the process exit code: not 0 if anything failed
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
	}

	// 2 - no tombstones (before a rollback)
	if task.purge {
		stats, err := m.PurgeTombstones(ctx, task.apply)
		_, _ = fmt.Fprintf(out, "purge tombstones %s: tombstones=%d incomplete records=%d\n",
			mode, stats.Tombstones, len(stats.Incomplete))
		for _, name := range stats.Incomplete {
			_, _ = fmt.Fprintf(out, "  incomplete: %s\n", name)
		}
		if err != nil {
			log.Error("purge failed", zap.Error(err))
			return 1
		}
	}

	// 3 - every name from the contracts
	if task.refresh {
		stats, err := m.RefreshAll(ctx, task.apply, task.interval)
		_, _ = fmt.Fprintf(out, "refresh %s: total=%d unchanged=%d updated=%d removed=%d not-final=%d failed=%d non-canonical=%d\n",
			mode, stats.Total, stats.Unchanged, stats.Updated, stats.Removed, stats.NotFinal, stats.Failed, stats.NonCanonical)
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
