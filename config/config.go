package config

import (
	"os"

	commonaccount "github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/net/rpc"
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
	Account          commonaccount.Config   `yaml:"account"`
	Network          nodeconf.Configuration `yaml:"network"`
	NetworkStorePath string                 `yaml:"networkStorePath"`
	Yamux            yamux.Config           `yaml:"yamux"`
	Mongo            Mongo                  `yaml:"mongo"`
	Contracts        Contracts              `yaml:"contracts"`
	Queue            Queue                  `yaml:"queue"`
	Nonce            Nonce                  `yaml:"nonce"`
	Aa               AA                     `yaml:"accountAbstraction"`
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

func (c *Config) GetAccount() commonaccount.Config {
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

func (c *Config) GetQueue() Queue {
	return c.Queue
}

func (c *Config) GetAA() AA {
	return c.Aa
}
