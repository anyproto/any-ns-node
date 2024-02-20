package config

type Contracts struct {
	GethUrl                        string `yaml:"gethUrl"`
	AddrRegistry                   string `yaml:"ensRegistry"`
	AddrResolver                   string `yaml:"resolver"`
	AddrRegistrarImplementation    string `yaml:"registrarImplementation"`
	AddrRegistrarConroller         string `yaml:"registrarController"`
	AddrRegistrarPrivateController string `yaml:"registrarControllerPrivate"`
	AddrToken                      string `yaml:"nameToken"`
	TokenDecimals                  uint8  `yaml:"tokenDecimals"`
	AddrNameWrapper                string `yaml:"nameWrapper"`

	AddrAdmin string `yaml:"admin"`
	AdminPk   string `yaml:"adminPk"`

	// when tx is sent, we will first try to get it N times
	// each time waiting for X seconds. If we will not get it -> we will think that TX
	// was immediately rejected without mining, which is a "high nonce" sign
	// (probabilistic, but it's ok)
	WaitMiningRetryCount uint `yaml:"waitMintingRetryCount"`
}
