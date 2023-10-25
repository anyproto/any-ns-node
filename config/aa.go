package config

type AA struct {
	AlchemyApiKey     string `yaml:"alchemyApiKey"`
	AlchemyRpcUrl     string `yaml:"alchemyRpcUrl"`
	AccountFactory    string `yaml:"accountFactory"`
	EntryPoint        string `yaml:"entryPoint"`
	GasPolicyId       string `yaml:"gasPolicyID"`
	ChainID           int    `yaml:"chainID"`
	NameTokensPerName uint8  `yaml:"nameTokensPerName"`
}
