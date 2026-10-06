package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anyproto/any-ns-node/cache"
)

type fakeMaintainer struct {
	refreshStats cache.RefreshStats
	refreshErr   error
	indexStats   cache.NameIndexStats
	indexErr     error
	aliasStats   cache.AliasStats
	aliasErr     error
	purgeStats   cache.PurgeStats
	purgeErr     error

	calls []string
}

func (f *fakeMaintainer) PurgeTombstones(_ context.Context, _ bool) (cache.PurgeStats, error) {
	f.calls = append(f.calls, "purge")
	return f.purgeStats, f.purgeErr
}

func (f *fakeMaintainer) RefreshAll(_ context.Context, _ bool, _ time.Duration) (cache.RefreshStats, error) {
	f.calls = append(f.calls, "refresh")
	return f.refreshStats, f.refreshErr
}

func (f *fakeMaintainer) MigrateAliases(_ context.Context, _ bool) (cache.AliasStats, error) {
	f.calls = append(f.calls, "aliases")
	return f.aliasStats, f.aliasErr
}

func (f *fakeMaintainer) VerifyNameIndex(_ context.Context, _ bool) (cache.NameIndexStats, error) {
	f.calls = append(f.calls, "dedupe")
	return f.indexStats, f.indexErr
}

func TestMaintainCache_ExitCode(t *testing.T) {
	refresh := cacheTask{refresh: true, apply: true, interval: time.Millisecond}

	t.Run("success", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshStats: cache.RefreshStats{Total: 2, Updated: 1, Unchanged: 1}}
		require.Equal(t, 0, maintainCache(context.Background(), f, refresh, &out))
		require.Contains(t, out.String(), "refresh APPLIED: total=2")
	})

	t.Run("not final is not a failure", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshStats: cache.RefreshStats{Total: 2, NotFinal: 2}}
		require.Equal(t, 0, maintainCache(context.Background(), f, refresh, &out))
		require.Contains(t, out.String(), "not-final=2 failed=0")
	})

	t.Run("some names failed (incl. finalized-read failures): exit 1, the summary is printed", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshStats: cache.RefreshStats{Total: 2, Updated: 1, Failed: 1}}
		require.Equal(t, 1, maintainCache(context.Background(), f, refresh, &out))
		require.Contains(t, out.String(), "failed=1")
	})

	t.Run("dry run: the same exit code", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshStats: cache.RefreshStats{Total: 2, Failed: 1}}
		task := refresh
		task.apply = false
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Contains(t, out.String(), "refresh DRY RUN")
	})

	t.Run("refresh failed", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshErr: errors.New("mongo is down")}
		require.Equal(t, 1, maintainCache(context.Background(), f, refresh, &out))
	})

	t.Run("dedupe runs first, its failure stops the rest", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{indexErr: cache.ErrNameIndexNotUnique}
		task := refresh
		task.dedupe = true
		task.purge = true
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe"}, f.calls)
	})

	t.Run("dedupe accepts the existing index, then purge, then refresh", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{indexStats: cache.NameIndexStats{NameIndex: "exists"}, purgeStats: cache.PurgeStats{Tombstones: 3}}
		task := refresh
		task.dedupe = true
		task.purge = true
		require.Equal(t, 0, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe", "aliases", "purge", "refresh"}, f.calls)
		require.Contains(t, out.String(), "unique name index=exists")
		require.Contains(t, out.String(), "purge tombstones APPLIED: tombstones=3 incomplete records=0")
	})

	t.Run("the alias migration is reported, kept ones listed, a failure stops the rest", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{indexStats: cache.NameIndexStats{NameIndex: "exists"},
			aliasStats: cache.AliasStats{Aliases: 3, Renamed: 1, Deleted: 1, Kept: []string{"Foo.any"}}}
		task := cacheTask{dedupe: true}
		require.Equal(t, 0, maintainCache(context.Background(), f, task, &out))
		require.Contains(t, out.String(), "non-canonical records=3 renamed=1 deleted=1 kept=1\n  kept: Foo.any\n")

		f = &fakeMaintainer{aliasErr: errors.New("mongo is down")}
		task = refresh
		task.dedupe = true
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe", "aliases"}, f.calls)
	})

	t.Run("purge refused (incomplete records): they are listed, exit 1, no refresh", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{purgeErr: cache.ErrIncompleteRecords, purgeStats: cache.PurgeStats{Incomplete: []string{"a.any", "b.any"}}}
		task := refresh
		task.purge = true
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"purge"}, f.calls)
		require.Contains(t, out.String(), "incomplete records=2")
		require.Contains(t, out.String(), "  incomplete: a.any\n  incomplete: b.any\n")
	})
}
