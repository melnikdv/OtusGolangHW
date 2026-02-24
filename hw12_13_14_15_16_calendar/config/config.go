// config/config.go
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

// === Существующая структура (для calendar) ===

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

// SchedulerConfig — конфиг для calendar_scheduler
type SchedulerConfig struct {
	Logger struct {
		Level string `yaml:"level"`
	} `yaml:"logger"`
	Storage struct {
		Type StorageType `yaml:"type"`
		SQL  struct {
			DSN string `yaml:"dsn"`
		} `yaml:"sql"`
	} `yaml:"storage"`
	RMQ struct {
		URL   string `yaml:"url"`
		Queue string `yaml:"queue"`
	} `yaml:"rmq"`
	Interval string `yaml:"interval"` // e.g. "1m", "30s"
}

func LoadSchedulerConfig(path string) (*SchedulerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scheduler config: %w", err)
	}
	var cfg SchedulerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse scheduler config: %w", err)
	}
	return &cfg, nil
}

// SenderConfig — конфиг для calendar_sender
type SenderConfig struct {
	Logger struct {
		Level string `yaml:"level"`
	} `yaml:"logger"`
	RMQ struct {
		URL   string `yaml:"url"`
		Queue string `yaml:"queue"`
	} `yaml:"rmq"`
}

func LoadSenderConfig(path string) (*SenderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sender config: %w", err)
	}
	var cfg SenderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse sender config: %w", err)
	}
	return &cfg, nil
}
