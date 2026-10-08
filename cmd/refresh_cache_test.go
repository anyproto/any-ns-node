package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	releaseStats cache.ReleaseStats
	releaseErr   error
	restoreStats cache.RestoreStats
	restoreErr   error

	calls []string
	// what RestoreTombstones read
	export string
}

func (f *fakeMaintainer) ReleaseName(_ context.Context, name string, _ bool) (cache.ReleaseStats, error) {
	f.calls = append(f.calls, "release "+name)
	return f.releaseStats, f.releaseErr
}

func (f *fakeMaintainer) RestoreTombstones(_ context.Context, export io.Reader, _ bool) (cache.RestoreStats, error) {
	f.calls = append(f.calls, "restore")
	b, err := io.ReadAll(export)
	if err != nil {
		return cache.RestoreStats{}, err
	}
	f.export = string(b)
	return f.restoreStats, f.restoreErr
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

	t.Run("lapsed and not on chain are not failures", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{refreshStats: cache.RefreshStats{Total: 5, Unchanged: 1, Updated: 1, Lapsed: 2, NotOnChain: 1, NonCanonical: 1}}
		require.Equal(t, 0, maintainCache(context.Background(), f, refresh, &out))
		require.Equal(t, "refresh APPLIED: total=5 unchanged=1 updated=1 lapsed=2 not-on-chain=1 failed=0 non-canonical=1\n", out.String())
	})

	t.Run("some names failed: exit 1, the summary is printed", func(t *testing.T) {
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
		task.release = "x.any"
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe"}, f.calls)
	})

	t.Run("the order: dedupe, release, restore, refresh", func(t *testing.T) {
		var out bytes.Buffer
		export := writeExport(t, "{}\n")
		f := &fakeMaintainer{indexStats: cache.NameIndexStats{NameIndex: "exists"},
			releaseStats: cache.ReleaseStats{Name: "x.any", Record: &cache.NameDataItem{FullName: "x.any"}},
			restoreStats: cache.RestoreStats{Tombstones: 3, Restored: 3}}
		task := refresh
		task.dedupe, task.release, task.restore, task.restoreFrom = true, "X.any", true, export
		require.Equal(t, 0, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe", "aliases", "release X.any", "restore", "refresh"}, f.calls)
		require.Contains(t, out.String(), "unique name index=exists")
		require.Contains(t, out.String(), "restore APPLIED: tombstones=3 restored=3 missing=0\n")
		require.Equal(t, "{}\n", f.export)
	})

	t.Run("the alias migration is reported, kept ones listed, a failure stops the rest", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{indexStats: cache.NameIndexStats{NameIndex: "exists"},
			aliasStats: cache.AliasStats{Aliases: 3, Renamed: 1, Deleted: 1, Kept: []string{"Foo.any"}}}
		task := cacheTask{dedupe: true}
		require.Equal(t, 0, maintainCache(context.Background(), f, task, &out))
		require.Contains(t, out.String(), "non-canonical records=3 renamed=1 deleted=1 kept=1 canon set=0\n  kept: Foo.any\n")

		f = &fakeMaintainer{aliasErr: errors.New("mongo is down")}
		task = refresh
		task.dedupe = true
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"dedupe", "aliases"}, f.calls)
	})

	t.Run("release: the record and the aliases are printed; a dry run says so", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{releaseStats: cache.ReleaseStats{Name: "x.any", Aliases: []string{"X.any"},
			Record: &cache.NameDataItem{FullName: "x.any", OwnerAnyAddress: "A-owner", NameExpires: 7, Lapsed: true}}}
		require.Equal(t, 0, maintainCache(context.Background(), f, cacheTask{release: "x.any"}, &out))
		require.Contains(t, out.String(), `release DRY RUN (nothing was written, use -refresh-apply to write): name=x.any owner_any_address="A-owner"`)
		require.Contains(t, out.String(), "name_expires=7 lapsed=true removed=false deleted=false\n")
		require.Contains(t, out.String(), "  alias (not deleted, it keeps the name taken): X.any\n")
	})

	t.Run("release of a name that is not cached: exit 1, nothing else runs", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{releaseStats: cache.ReleaseStats{Name: "x.any"}, releaseErr: cache.ErrNameNotCached}
		task := refresh
		task.release = "x.any"
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"release x.any"}, f.calls)
		require.Contains(t, out.String(), "not in the cache")
	})

	t.Run("restore: missing ones are listed, exit 1, no refresh", func(t *testing.T) {
		var out bytes.Buffer
		f := &fakeMaintainer{restoreStats: cache.RestoreStats{Tombstones: 3, Restored: 1, Missing: []string{"a.any", "b.any"}}}
		task := refresh
		task.restore, task.restoreFrom = true, writeExport(t, "")
		require.Equal(t, 1, maintainCache(context.Background(), f, task, &out))
		require.Equal(t, []string{"restore"}, f.calls)
		require.Contains(t, out.String(), "restore APPLIED: tombstones=3 restored=1 missing=2\n  missing: a.any\n  missing: b.any\n")
	})

	t.Run("restore without a readable export: exit 1, the maintainer is not called", func(t *testing.T) {
		for _, from := range []string{"", filepath.Join(t.TempDir(), "nope.jsonl")} {
			var out bytes.Buffer
			f := &fakeMaintainer{}
			require.Equal(t, 1, maintainCache(context.Background(), f, cacheTask{restore: true, restoreFrom: from}, &out))
			require.Empty(t, f.calls)
		}
	})
}

func writeExport(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cache.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestRpcHost(t *testing.T) {
	require.Equal(t, "eth-sepolia.g.alchemy.com", rpcHost("https://eth-sepolia.g.alchemy.com/v2/SECRET"))
	require.Equal(t, "(unparsable)", rpcHost("SECRET"))
}
