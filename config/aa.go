package config

type AA struct {
	AlchemyApiKey  string `yaml:"alchemyApiKey"`
	AlchemyRpcUrl  string `yaml:"alchemyRpcUrl"`
	AccountFactory string `yaml:"accountFactory"`
	EntryPoint     string `yaml:"entryPoint"`
	GasPolicyId    string `yaml:"gasPolicyId"`
	ChainID        int    `yaml:"chainId"`
}
