package config

type Config struct {
	IndexMappings map[string]int `json:"indexMappings" yaml:"indexMappings"`
}

var config *Config

func InitConfig(cfg *Config) *Config {
	config = cfg
	return config
}

func GetConfig() *Config {
	return config
}
