package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Database    Database    `yaml:"database"`
	EthereumRPC EthereumRPC `yaml:"ethereum-rpc"`
}

type Database struct {
	URL string `yaml:"url"`
}

type EthereumRPC struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api-key"`
}

func Init() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("config.yml", &cfg)
	return &cfg, err
}
