package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Database    Database    `yaml:"database" env-required:""`
	EthereumRPC EthereumRPC `yaml:"ethereum-rpc" env-required:""`
	GCP         GCP         `yaml:"gcp" env-required:""`
}

type Database struct {
	URL string `yaml:"url" env-required:""`
}

type EthereumRPC struct {
	URL    string `yaml:"url" env-required:""`
	APIKey string `yaml:"api-key" env-required:""`
}

type GCP struct {
	ProjectID      string   `yaml:"project-id" env-required:""`
	ServiceAccount string   `yaml:"service-account" env-required:""`
	BigQuery       BigQuery `yaml:"bigquery" env-required:""`
}

type BigQuery struct {
	Location string `yaml:"location" env-required:""`
	Dataset  string `yaml:"dataset" env-required:""`
}

func Load() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("config.yml", &cfg)
	return &cfg, err
}
