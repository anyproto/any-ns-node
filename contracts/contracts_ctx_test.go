package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	ac "github.com/anyproto/any-ns-node/anytype_crypto"
	"github.com/anyproto/any-ns-node/config"
)

const testNameWrapper = "0x0000000000000000000000000000000000000abc"

// a regression must fail the test, not hang it until the go test timeout
const watchdog = 10 * time.Second

type rpcRequest struct {
	ID     json.RawMessage   `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

// newTestNode starts a local JSON-RPC endpoint that passes every request to handle.
// the test ends every pending request on cleanup, so that a regression can not hang it
func newTestNode(t *testing.T, handle func(w http.ResponseWriter, r *http.Request, req rpcRequest, stop <-chan struct{})) *anynsContracts {
	stop := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		handle(w, r, req, stop)
	}))
	// cleanups run in reverse: release the handlers first, then close the server
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(stop) })

	return &anynsContracts{config: config.Contracts{
		GethUrl:                     srv.URL,
		AddrRegistry:                "0x0000000000000000000000000000000000000456",
		AddrNameWrapper:             testNameWrapper,
		AddrResolver:                "0x0000000000000000000000000000000000000def",
		AddrRegistrarImplementation: "0x0000000000000000000000000000000000000123",
	}}
}

func reply(w http.ResponseWriter, id json.RawMessage, result string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, id, result)
}

// a node that never answers: only the caller's context ends the call
// (eth_getCode answers: the address looks like a contract)
func newStalledNode(t *testing.T) *anynsContracts {
	return newTestNode(t, func(w http.ResponseWriter, r *http.Request, req rpcRequest, stop <-chan struct{}) {
		if req.Method == "eth_getCode" {
			reply(w, req.ID, `"0x01"`)
			return
		}
		select {
		case <-r.Context().Done():
		case <-stop:
		}
	})
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

// the refresh of a cached name calls these on lookups: a stalled provider must not hang them
func TestContracts_StalledNodeHonoursContext(t *testing.T) {
	calls := map[string]func(ctx context.Context, c *anynsContracts) error{
		"LatestBlock": func(ctx context.Context, c *anynsContracts) error {
			_, err := c.LatestBlock(ctx)
			return err
		},
		"FinalizedBlock": func(ctx context.Context, c *anynsContracts) error {
			_, err := c.FinalizedBlock(ctx)
			return err
		},
		"GetOwnerForNamehash": func(ctx context.Context, c *anynsContracts) error {
			_, err := c.GetOwnerForNamehash(ctx, [32]byte{1}, testBlock)
			return err
		},
		"GetNameExpires": func(ctx context.Context, c *anynsContracts) error {
			_, err := c.GetNameExpires(ctx, "hello.any", testBlock)
			return err
		},
		// owned by the NameWrapper: asks it for the real owner first
		"GetAdditionalNameInfo": func(ctx context.Context, c *anynsContracts) error {
			_, _, _, err := c.GetAdditionalNameInfo(ctx, common.HexToAddress(testNameWrapper), "hello.any", testBlock)
			return err
		},
		// eth_getCode answers, the owner() call stalls
		"GetScwOwner": func(ctx context.Context, c *anynsContracts) error {
			_, err := c.GetScwOwner(ctx, common.HexToAddress("0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51"), testBlock)
			return err
		},
	}

	nodes := map[string]func(t *testing.T) *anynsContracts{
		"stalled calls":        newStalledNode,
		"stalled ws handshake": newStalledWsNode,
	}
	for node, newNode := range nodes {
		for name, call := range calls {
			t.Run(node+"/"+name, func(t *testing.T) {
				c := newNode(t)

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				var err error
				within(t, name, func() { err = call(ctx, c) })
				require.Error(t, err)
			})
		}
	}
}

// a ws:// provider that accepts the TCP connection, but never answers the handshake
func newStalledWsNode(t *testing.T) *anynsContracts {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	var (
		mu    sync.Mutex
		conns []net.Conn
	)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
		}
	}()
	t.Cleanup(func() {
		_ = l.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, c := range conns {
			_ = c.Close()
		}
	})

	c := newTestNode(t, func(http.ResponseWriter, *http.Request, rpcRequest, <-chan struct{}) {})
	c.config.GethUrl = "ws://" + l.Addr().String()
	return c
}

var testBlock = common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000b10c4")

// every read of one refresh is pinned to the same block by its hash (EIP-1898): eth_call and
// eth_getCode. every call gets a valid answer, so that every read of the chain is made
func TestContracts_ReadsArePinnedToTheBlockHash(t *testing.T) {
	const (
		scw   = "0x10d5b0e279e5e4c1d1df5f57dfb7e84813920a51"
		eoa   = "0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5"
		anyID = "12D3KooWA8EXV3KjBxEU5EnsPfneLx84vMWAtTBQBeyooN82KSuS"
	)
	answers := map[string][]interface{}{
		"owner(bytes32)":       {common.HexToAddress(testNameWrapper)},
		"nameExpires(uint256)": {big.NewInt(12345)},
		"ownerOf(uint256)":     {common.HexToAddress(scw)},
		"contenthash(bytes32)": {[]byte(anyID)},
		"spaceId(bytes32)":     {[]byte("space")},
		"owner()":              {common.HexToAddress(eoa)},
	}
	methods := map[string]abi.Method{}
	for _, md := range []*bind.MetaData{ac.ENSRegistryMetaData, ac.AnytypeRegistrarImplementationMetaData,
		ac.AnytypeNameWrapperMetaData, ac.AnytypeResolverMetaData, ac.SCWMetaData} {
		parsed, err := md.GetAbi()
		require.NoError(t, err)
		for _, m := range parsed.Methods {
			if _, ok := answers[m.Sig]; ok {
				methods[string(m.ID)] = m
			}
		}
	}

	var (
		mu    sync.Mutex
		calls []string
		pins  []json.RawMessage
	)
	c := newTestNode(t, func(w http.ResponseWriter, _ *http.Request, req rpcRequest, _ <-chan struct{}) {
		mu.Lock()
		defer mu.Unlock()
		require.Len(t, req.Params, 2, req.Method)
		pins = append(pins, req.Params[1])

		switch req.Method {
		case "eth_getCode":
			calls = append(calls, req.Method)
			reply(w, req.ID, `"0x01"`)
		case "eth_call":
			var msg struct {
				Input hexutil.Bytes `json:"input"`
			}
			require.NoError(t, json.Unmarshal(req.Params[0], &msg))
			m, ok := methods[string(msg.Input[:4])]
			require.True(t, ok, "unexpected call %x", msg.Input[:4])
			calls = append(calls, m.Sig)
			out, err := m.Outputs.Pack(answers[m.Sig]...)
			require.NoError(t, err)
			reply(w, req.ID, fmt.Sprintf(`"%s"`, hexutil.Encode(out)))
		default:
			t.Errorf("unexpected method %s", req.Method)
		}
	})

	ctx := context.Background()
	within(t, "the reads", func() {
		owner, err := c.GetOwnerForNamehash(ctx, [32]byte{1}, testBlock)
		require.NoError(t, err)
		require.Equal(t, common.HexToAddress(testNameWrapper), owner)

		exp, err := c.GetNameExpires(ctx, "hello.any", testBlock)
		require.NoError(t, err)
		require.EqualValues(t, 12345, exp.Int64())

		// owned by the NameWrapper: its ownerOf, then the resolver
		ea, aa, si, err := c.GetAdditionalNameInfo(ctx, common.HexToAddress(testNameWrapper), "hello.any", testBlock)
		require.NoError(t, err)
		require.Equal(t, common.HexToAddress(scw).Hex(), ea)
		require.Equal(t, anyID, aa)
		require.Equal(t, "space", si)

		// eth_getCode, then the wallet's owner()
		own, err := c.GetScwOwner(ctx, common.HexToAddress(scw), testBlock)
		require.NoError(t, err)
		require.Equal(t, common.HexToAddress(eoa), own)
	})

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"owner(bytes32)", "nameExpires(uint256)", "ownerOf(uint256)",
		"contenthash(bytes32)", "spaceId(bytes32)", "eth_getCode", "owner()"}, calls)
	for i, pin := range pins {
		var ref struct {
			BlockHash *common.Hash `json:"blockHash"`
		}
		require.NoError(t, json.Unmarshal(pin, &ref), "call %d (%s): %s", i, calls[i], pin)
		require.NotNil(t, ref.BlockHash, "call %d (%s) is not pinned by the hash: %s", i, calls[i], pin)
		require.Equal(t, testBlock, *ref.BlockHash, calls[i])
	}
}

// the hash of a block is taken from the node, never recomputed from the header: this
// go-ethereum does not know the newest header fields, its Header.Hash() is wrong for them
func TestContracts_BlockHashIsTheNodes(t *testing.T) {
	const hash = "0x1111111111111111111111111111111111111111111111111111111111111111"
	block := func(number string) string {
		// a header with a field this go-ethereum does not know (Prague)
		return fmt.Sprintf(`{"number":"%s","hash":"%s","timestamp":"0x64","parentHash":"0x%064x",`+
			`"requestsHash":"0x%064x","miner":"0x0000000000000000000000000000000000000000"}`, number, hash, 1, 2)
	}

	t.Run("latest", func(t *testing.T) {
		var tags []string
		c := newTestNode(t, func(w http.ResponseWriter, _ *http.Request, req rpcRequest, _ <-chan struct{}) {
			var tag string
			_ = json.Unmarshal(req.Params[0], &tag)
			tags = append(tags, tag)
			reply(w, req.ID, block("0x2a"))
		})
		b, err := c.LatestBlock(context.Background())
		require.NoError(t, err)
		require.Equal(t, &Block{Number: 42, Hash: common.HexToHash(hash), Time: 100}, b)
		require.Equal(t, []string{"latest"}, tags)
	})

	t.Run("finalized, falls back to safe", func(t *testing.T) {
		var tags []string
		c := newTestNode(t, func(w http.ResponseWriter, _ *http.Request, req rpcRequest, _ <-chan struct{}) {
			var tag string
			_ = json.Unmarshal(req.Params[0], &tag)
			tags = append(tags, tag)
			if tag == "finalized" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"error":{"code":-32602,"message":"unknown block tag"}}`, req.ID)
				return
			}
			reply(w, req.ID, block("0x29"))
		})
		b, err := c.FinalizedBlock(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, 41, b.Number)
		require.Equal(t, []string{"finalized", "safe"}, tags)
	})

	t.Run("no block", func(t *testing.T) {
		c := newTestNode(t, func(w http.ResponseWriter, _ *http.Request, req rpcRequest, _ <-chan struct{}) {
			reply(w, req.ID, "null")
		})
		_, err := c.LatestBlock(context.Background())
		require.Error(t, err)
	})
}
