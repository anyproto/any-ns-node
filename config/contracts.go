package config

type Contracts struct {
	GethUrl               string `yaml:"geth_url"`
	AddrRegistry          string `yaml:"registry"`
	AddrResolver          string `yaml:"resolver"`
	AddrPrivateController string `yaml:"private_controller"`
	AddrNameWrapper       string `yaml:"name_wrapper"`

	AddrAdmin string `yaml:"admin"`
	AdminPk   string `yaml:"admin_pk"`
}
