package config

type AA struct {
	AlchemyApiKey     string `yaml:"alchemy_api_key"`
	AlchemyRpcUrl     string `yaml:"alchemy_rpc_url"`
	AccountFactory    string `yaml:"account_factory"`
	EntryPoint        string `yaml:"entry_point"`
	GasPolicyId       string `yaml:"gas_policy_id"`
	ChainID           int    `yaml:"chain_id"`
	NameTokensPerName uint8  `yaml:"name_tokens_per_name"`
}
