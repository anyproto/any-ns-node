package config

type Contracts struct {
	GethUrl                string `yaml:"gethUrl"`
	AddrRegistry           string `yaml:"registry"`
	AddrResolver           string `yaml:"resolver"`
	AddrRegistrarConroller string `yaml:"registrarController"`
	AddrToken              string `yaml:"token"`
	TokenDecimals          uint8  `yaml:"tokenDecimals"`
	AddrPrivateController  string `yaml:"privateController"`
	AddrNameWrapper        string `yaml:"nameWrapper"`

	AddrAdmin string `yaml:"admin"`
	AdminPk   string `yaml:"adminPk"`

	// when tx is sent, we will first try to get it N times
	// each time waiting for X seconds. If we will not get it -> we will think that TX
	// was immediately rejected without mining, which is a "high nonce" sign
	// (probabilistic, but it's ok)
	WaitMiningRetryCount uint `yaml:"waitMintingRetryCount"`
}
