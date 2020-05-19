package config

type Config struct {
	LoggerDBUrl  string `toml:"logger_database"`
	AccountDBUrl string `toml:"account_database"`
	HTTPPort     string `toml:"http_port"`
	GRPCPort     string `toml:"grpc_port"`
	ConsulPort   string `toml:"consul_port"`
}

func NewConfig() *Config {
	return &Config{}
}
