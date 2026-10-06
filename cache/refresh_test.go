package cache

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	nsp "github.com/anyproto/any-sync/nameservice/nameserviceproto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"

	"github.com/anyproto/any-ns-node/contracts"
	mock_contracts "github.com/anyproto/any-ns-node/contracts/mock"
)

const (
	testScw      = "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51"
	testEoa      = "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5"
	testAnyID    = "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
	otherScw     = "0xaab27b150451726ec7738aa1d0a94505c8729bd1"
	otherEoa     = "0xe595e2ba3f0ce990d8037e07250c5c78ce40f8ff"
	otherAnyID   = "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio"
	nameWrapper  = "0x0000000000000000000000000000000000000abc"
	testFullName = "test.any"

	// a regression must fail the test, not hang it until the go test timeout
	watchdog = 10 * time.Second
)

// expires in the past, but the name is still in the grace period
var inGrace = time.Now().Add(-30 * 24 * time.Hour).Unix()

// expired and past the grace period: anyone can register it again
var lapsed = time.Now().Add(-100 * 24 * time.Hour).Unix()

// --- contracts ---

// the contracts answer for a registered name (any block)
func expectRegistered(m *mock_contracts.MockContractsService, scw string, eoa string, anyID string, expires int64) {
	m.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
	m.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(expires), nil)
	m.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(scw, anyID, "", nil)
	m.EXPECT().GetScwOwner(gomock.Any(), common.HexToAddress(scw), gomock.Any()).Return(common.HexToAddress(eoa), nil)
}

// the contracts answer for a lapsed name: the registry still points to the NameWrapper.
// read twice: at the latest block, then (a cached record is removed) at the finalized one
func expectLapsed(m *mock_contracts.MockContractsService) {
	m.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).Times(2)
	m.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(lapsed), nil).Times(2)
}

// pause makes a mocked call block until release (or until the test ends)
type pause struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newPause(t *testing.T) *pause {
	p := &pause{entered: make(chan struct{}), release: make(chan struct{})}
	t.Cleanup(p.Release)
	return p
}

func (p *pause) Release() { p.once.Do(func() { close(p.release) }) }

func (p *pause) wait() {
	close(p.entered)
	<-p.release
}

func (p *pause) Entered(t *testing.T) {
	t.Helper()
	select {
	case <-p.entered:
	case <-time.After(watchdog):
		t.Fatal("the paused call was never made")
	}
}

// within runs fn and fails the test if it does not return in time
func within(t *testing.T, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(watchdog):
		t.Fatalf("%s did not return within %s", what, watchdog)
	}
}

// --- cache ---

func cachedItem(t *testing.T, fx *fixture) *NameDataItem {
	t.Helper()
	item, err := fx.getNameData(ctx, testFullName)
	require.NoError(t, err)
	return item
}

func countRecords(t *testing.T, fx *fixture) int64 {
	t.Helper()
	n, err := fx.itemColl.CountDocuments(ctx, bson.M{"name": testFullName})
	require.NoError(t, err)
	return n
}

// seedItem: a cached registration of testFullName, read at the block
func seedItem(t *testing.T, fx *fixture, expires int64, block int64, observedAt int64) {
	t.Helper()
	require.NoError(t, fx.setNameData(ctx, &NameDataItem{
		FullName: testFullName, OwnerEthAddress: testEoa, OwnerScwEthAddress: testScw, OwnerAnyAddress: testAnyID,
		SpaceId: "space", NameExpires: expires, ObservedBlock: block, ObservedBlockHash: blockHash(block).Hex(), ObservedAt: observedAt,
	}))
}

// the seeded record, as a taken name
func requireCachedAsTaken(t *testing.T, out *nsp.NameAvailableResponse, expires int64) {
	t.Helper()
	require.False(t, out.Available)
	require.Equal(t, testEoa, out.OwnerEthAddress)
	require.Equal(t, testScw, out.OwnerScwEthAddress)
	require.Equal(t, testAnyID, out.OwnerAnyAddress)
	require.Equal(t, expires, out.NameExpires)
}

func isNameAvailable(t *testing.T, cs *cacheService) *nsp.NameAvailableResponse {
	t.Helper()
	out, err := cs.IsNameAvailable(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
	require.NoError(t, err)
	return out
}

func TestIsLapsed(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	expires := now.Unix() - gracePeriodSec

	// the registrar: available() is nameExpires + GRACE_PERIOD < block.timestamp
	require.False(t, isLapsed(expires, now))
	require.True(t, isLapsed(expires-1, now))
	require.False(t, isLapsed(now.Unix(), now))
	// whole seconds: half a second later it is still the same block.timestamp
	require.False(t, isLapsed(expires, now.Add(500*time.Millisecond)))
	require.True(t, isLapsed(expires, now.Add(time.Second)))
}

func TestCacheService_UpdateInCache_PinnedRead(t *testing.T) {
	t.Run("all reads of one refresh are pinned to one block, read once, the block is stored", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		at := time.Now().Add(-time.Minute)
		var heads int
		fx.setHeadFunc(func() *contracts.Block {
			heads++
			return testBlock(777, at)
		})
		block := gomock.Eq(blockHash(777))

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), block).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, block).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, block).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), block).Return(common.HexToAddress(testEoa), nil)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.Equal(t, 1, heads)

		item := cachedItem(t, fx)
		require.Equal(t, int64(777), item.ObservedBlock)
		require.Equal(t, blockHash(777).Hex(), item.ObservedBlockHash)
		require.Equal(t, at.Unix(), item.ObservedBlockTime)
		require.False(t, item.RefreshNeeded)
	})

	t.Run("lapse is decided by the block timestamp, not by the local clock", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		// lapsed by the local clock, but at the block (20 days ago) still in the grace period
		fx.setHead(500, time.Now().Add(-20*24*time.Hour))
		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, lapsed)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		require.Equal(t, lapsed, cachedItem(t, fx).NameExpires)
	})

	t.Run("renewal updates nameExpires of a cached name", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, inGrace, 0, 0)

		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, notExpired)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		require.Equal(t, testEoa, item.OwnerEthAddress)
	})

	t.Run("registration by another owner replaces all owner fields", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 0, 0)

		expectRegistered(fx.contracts, otherScw, otherEoa, otherAnyID, notExpired)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		item := cachedItem(t, fx)
		require.Equal(t, otherEoa, item.OwnerEthAddress)
		require.Equal(t, otherScw, item.OwnerScwEthAddress)
		require.Equal(t, otherAnyID, item.OwnerAnyAddress)
		require.Empty(t, item.SpaceId)
		require.Equal(t, notExpired, item.NameExpires)
	})

	t.Run("name in the grace period is still cached as taken", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		expectRegistered(fx.contracts, testScw, testEoa, testAnyID, inGrace)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))
		out := isNameAvailable(t, fx.cacheService)
		require.False(t, out.Available)
		require.Equal(t, inGrace, out.NameExpires)
	})

	t.Run("lapsed at the latest block: ErrNameNotRegistered, a single read never removes the record", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)
		seedItem(t, fx, lapsed, 0, 0)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil).AnyTimes()
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(lapsed), nil).AnyTimes()
		// only the finalized block could confirm it: not readable here
		fx.setFinalizedErr(errors.New("no finalized block"))

		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
		require.Error(t, err)
		requireCachedAsTaken(t, isNameAvailable(t, fx.cacheService), lapsed)
	})

	t.Run("wallet owner lookup fails for a new name: the confirmed part is cached as taken, marked", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(testScw, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, errors.New("rpc is down"))

		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
		require.ErrorIs(t, err, ErrNameDataIncomplete)

		item := cachedItem(t, fx)
		require.Empty(t, item.OwnerEthAddress)
		require.Equal(t, testScw, item.OwnerScwEthAddress)
		require.Equal(t, notExpired, item.NameExpires)
		require.True(t, item.RefreshNeeded)
		require.False(t, isNameAvailable(t, fx.cacheService).Available)
	})

	t.Run("owner read times out: the confirmed registration is cached (own budget), nothing is guessed", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		old := enrichTimeout
		enrichTimeout = 50 * time.Millisecond
		defer func() { enrichTimeout = old }()

		p := newPause(t)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).DoAndReturn(
			func(ctx context.Context, _ common.Address, _ string, _ common.Hash) (string, string, string, error) {
				select {
				case <-ctx.Done():
					return "", "", "", ctx.Err()
				case <-p.release:
					return "", "", "", errors.New("released by the test")
				}
			})

		var err error
		within(t, "UpdateInCache", func() {
			err = fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
		})
		require.ErrorIs(t, err, ErrNameDataIncomplete)

		item := cachedItem(t, fx)
		require.Equal(t, notExpired, item.NameExpires)
		require.Empty(t, item.OwnerEthAddress)
		require.Empty(t, item.OwnerScwEthAddress)
		require.Empty(t, item.OwnerAnyAddress)
		require.True(t, item.RefreshNeeded)
	})

	t.Run("name owned by an EOA (no code at the address)", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(testEoa, testAnyID, "", nil)
		fx.contracts.EXPECT().GetScwOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.Address{}, contracts.ErrNotAContract)

		require.NoError(t, fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName}))

		item := cachedItem(t, fx)
		require.Equal(t, testEoa, item.OwnerEthAddress)
		require.Empty(t, item.OwnerScwEthAddress)
		require.False(t, item.RefreshNeeded)
	})

	t.Run("unknown name owner: no owner is guessed, the record is marked", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).Return(common.HexToAddress(nameWrapper), nil)
		fx.contracts.EXPECT().GetNameExpires(gomock.Any(), testFullName, gomock.Any()).Return(big.NewInt(notExpired), nil)
		fx.contracts.EXPECT().GetAdditionalNameInfo(gomock.Any(), gomock.Any(), testFullName, gomock.Any()).Return(common.Address{}.Hex(), testAnyID, "", nil)

		err := fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
		require.ErrorIs(t, err, ErrNameDataIncomplete)

		item := cachedItem(t, fx)
		require.Empty(t, item.OwnerEthAddress)
		require.Empty(t, item.OwnerScwEthAddress)
		require.True(t, item.RefreshNeeded)
	})

	t.Run("a stalled confirmation read ends with the timeout, nothing is written", func(t *testing.T) {
		fx := newFixture(t)
		defer fx.finish(t)

		old := confirmTimeout
		confirmTimeout = 50 * time.Millisecond
		defer func() { confirmTimeout = old }()

		p := newPause(t)
		fx.contracts.EXPECT().GetOwnerForNamehash(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ [32]byte, _ common.Hash) (common.Address, error) {
			select {
			case <-ctx.Done():
				return common.Address{}, ctx.Err()
			case <-p.release:
				return common.Address{}, errors.New("released by the test")
			}
		})

		var err error
		within(t, "UpdateInCache", func() {
			err = fx.UpdateInCache(ctx, &nsp.NameAvailableRequest{FullName: testFullName})
		})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Nil(t, cachedItem(t, fx))
	})
}
