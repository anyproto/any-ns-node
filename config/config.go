package config

import (
	"os"

	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/rpc"
	"github.com/anyproto/any-sync/net/rpc/limiter"
	"github.com/anyproto/any-sync/net/transport/quic"
	"github.com/anyproto/any-sync/net/transport/yamux"

	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/nodeconf"
	"gopkg.in/yaml.v3"

	"github.com/anyproto/any-sync/app/logger"
)

const CName = "config"

func NewFromFile(path string) (c *Config, err error) {
	c = &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return
}

type Config struct {
	Drpc             rpc.Config             `yaml:"drpc"`
	Log              logger.Config          `yaml:"log"`
	Account          accountservice.Config  `yaml:"account"`
	Network          nodeconf.Configuration `yaml:"network"`
	NetworkStorePath string                 `yaml:"networkStorePath"`
	Quic             quic.Config            `yaml:"quic"`
	Yamux            yamux.Config           `yaml:"yamux"`
	Mongo            Mongo                  `yaml:"mongo"`
	Contracts        Contracts              `yaml:"contracts"`
	Aa               AA                     `yaml:"accountAbstraction"`
	Metric           metric.Config          `yaml:"metric"`
	Nonce            Nonce                  `yaml:"nonce"`
	Queue            Queue                  `yaml:"queue"`
	Limiter          limiter.Config         `yaml:"limiter"`
	Sentry           Sentry                 `yaml:"sentry"`
	// use mongo cache to read data from
	ReadFromCache bool `yaml:"readFromCache"`

	// if false -> will use old-school ENSIP1
	//
	// 1. ENSIP1 standard: ens-go v3.6.0 (current) is using it
	// 2. ENSIP15 standard: that is an another standard for ENS namehashes
	// that was accepted in June 2023.
	//
	// Current AnyNS (as of June 2024) implementation supports ENSIP1, ENSIP15
	Ensip15Validation bool `yaml:"ensip15validation"`
}

func (c *Config) Init(a *app.App) (err error) {
	return
}

func (c *Config) Name() (name string) {
	return CName
}

func (c *Config) GetContracts() Contracts {
	return c.Contracts
}

func (c *Config) GetNodeConf() nodeconf.Configuration {
	return c.Network
}

func (c *Config) GetAccount() accountservice.Config {
	return c.Account
}

func (c *Config) GetNodeConfStorePath() string {
	return c.NetworkStorePath
}

func (c *Config) GetYamux() yamux.Config {
	return c.Yamux
}

func (c *Config) GetDrpc() rpc.Config {
	return c.Drpc
}

func (c *Config) GetAA() AA {
	return c.Aa
}

func (c *Config) GetQuic() quic.Config {
	return c.Quic
}

func (c *Config) GetMetric() metric.Config {
	return c.Metric
}

func (c *Config) GetNonce() Nonce {
	return c.Nonce
}

func (c *Config) GetQueue() Queue {
	return c.Queue
}

func (c *Config) GetLimiterConf() limiter.Config {
	return c.Limiter
}

func (c *Config) GetSentry() Sentry {
	return c.Sentry
}
