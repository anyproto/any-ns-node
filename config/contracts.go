package config

type Contracts struct {
	GethUrl               string `yaml:"geth_url"`
	AddrRegistry          string `yaml:"registry"`
	AddrResolver          string `yaml:"resolver"`
	AddrRegistrar         string `yaml:"registrar"`
	AddrPrivateController string `yaml:"private_controller"`
	AddrNameWrapper       string `yaml:"name_wrapper"`

	AddrAdmin string `yaml:"admin"`
	AdminPk   string `yaml:"admin_pk"`

	// when tx is sent, we will first try to get it N times
	// each time waiting for X seconds. If we will not get it -> we will think that TX
	// was immediately rejected without mining, which is a "high nonce" sign
	// (probabilistic, but it's ok)
	WaitMiningRetryCount uint `yaml:"wait_mining_retry_count"`
}
