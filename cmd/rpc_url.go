package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// chainIDTimeout bounds the chain id read of a provider
const chainIDTimeout = 10 * time.Second

// chainIDFunc reads eth_chainId of a provider
type chainIDFunc func(ctx context.Context, rpcURL string) (*big.Int, error)

// chainIDOf reads eth_chainId of the provider. its errors never carry the URL (it can carry an
// API key): only the host
func chainIDOf(ctx context.Context, rpcURL string) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(ctx, chainIDTimeout)
	defer cancel()
	c, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, redact(err, rpcURL)
	}
	defer c.Close()
	id, err := c.ChainID(ctx)
	if err != nil {
		return nil, redact(err, rpcURL)
	}
	return id, nil
}

// checkRPCURL: -rpc-url must be a provider of the chain the node runs on. the expected chain id
// is the config's (accountAbstraction.chainID) if it has one, otherwise the one of
// contracts.gethUrl (one read)
func checkRPCURL(ctx context.Context, rpcURL, configURL string, configChainID int64, chainID chainIDFunc) error {
	expected := big.NewInt(configChainID)
	source := "accountAbstraction.chainID of the config"
	if configChainID <= 0 {
		id, err := chainID(ctx, configURL)
		if err != nil {
			return fmt.Errorf("the chain id of contracts.gethUrl (%s): %w", rpcHost(configURL), err)
		}
		expected, source = id, "contracts.gethUrl ("+rpcHost(configURL)+")"
	}
	got, err := chainID(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("the chain id of -rpc-url (%s): %w", rpcHost(rpcURL), err)
	}
	if got.Cmp(expected) != 0 {
		return fmt.Errorf("-rpc-url (%s) is on chain %s, %s on chain %s", rpcHost(rpcURL), got, source, expected)
	}
	return nil
}

// redact: the error without the URL (and without its escaped form)
func redact(err error, rpcURL string) error {
	msg := err.Error()
	host := rpcHost(rpcURL)
	for _, s := range []string{rpcURL, url.QueryEscape(rpcURL)} {
		if s != "" {
			msg = strings.ReplaceAll(msg, s, host)
		}
	}
	return errors.New(msg)
}
