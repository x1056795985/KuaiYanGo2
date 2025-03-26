package config

type CORS struct {
	Mode      string          `mapstructure:"mode" json:"mode" `
	Whitelist []CORSWhitelist `mapstructure:"whitelist" json:"whitelist" `
}

type CORSWhitelist struct {
	AllowOrigin      string `mapstructure:"allow-origin" json:"allow-origin" `
	AllowMethods     string `mapstructure:"allow-methods" json:"allow-methods" `
	AllowHeaders     string `mapstructure:"allow-headers" json:"allow-headers" `
	ExposeHeaders    string `mapstructure:"expose-headers" json:"expose-headers" `
	AllowCredentials bool   `mapstructure:"allow-credentials" json:"allow-credentials" `
}
