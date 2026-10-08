package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// a JSON-RPC provider that answers eth_chainId
func chainServer(t *testing.T, chainID int64) string {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.Contains(t, string(body), "eth_chainId")
		_, _ = fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":"0x%x"}`, chainID)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/v2/SECRET-KEY"
}

func TestCheckRPCURL(t *testing.T) {
	ctx := context.Background()
	sepolia, mainnet := chainServer(t, 11155111), chainServer(t, 1)

	// the config's chain id, or the one of gethUrl
	require.NoError(t, checkRPCURL(ctx, sepolia, mainnet, 11155111, chainIDOf))
	require.NoError(t, checkRPCURL(ctx, sepolia, chainServer(t, 11155111), 0, chainIDOf))

	err := checkRPCURL(ctx, mainnet, sepolia, 11155111, chainIDOf)
	require.ErrorContains(t, err, "is on chain 1")
	err = checkRPCURL(ctx, mainnet, sepolia, 0, chainIDOf)
	require.ErrorContains(t, err, "on chain 11155111")
	require.NotContains(t, err.Error(), "SECRET-KEY")

	// a provider that can not be read: refused, the URL is not in the error
	failing := func(_ context.Context, u string) (*big.Int, error) {
		return nil, redact(errors.New(`Post "`+u+`": connection refused`), u)
	}
	err = checkRPCURL(ctx, "https://eth.example.com/v2/SECRET-KEY", sepolia, 11155111, failing)
	require.ErrorContains(t, err, "eth.example.com")
	require.False(t, strings.Contains(err.Error(), "SECRET-KEY"), err.Error())
}
