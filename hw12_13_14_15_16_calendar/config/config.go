package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type StorageType string

const (
	InMemory StorageType = "inmemory"
	SQL      StorageType = "sql"
)

type Config struct {
	Server struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		GRPCPort int    `yaml:"grpc_port"`
	} `yaml:"server"`
	Logger struct {
		Level string `yaml:"level"` // debug, info, warn, error
	} `yaml:"logger"`
	Storage struct {
		Type StorageType `yaml:"type"`
		SQL  struct {
			DSN string `yaml:"dsn"`
		} `yaml:"sql"`
	} `yaml:"storage"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
