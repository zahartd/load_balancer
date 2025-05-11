package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Host                string `json:"host"`
	Port                uint16 `json:"port"`
	Algorithm           string `json:"algorithm"`
	HealthCheckInterval int    `json:"health_check_interval_sec"`
}

type RateLimitConfig struct {
	DefaultCapacity   int `json:"default_capacity"`
	DefaultRatePerSec int `json:"default_rate_per_sec"`
}

type DBConfig struct {
	DSN string `json:"dsn"`
}

type Config struct {
	Server    ServerConfig    `json:"server"`
	Backends  []string        `json:"backends"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	DB        DBConfig        `json:"db"`
}

func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config

	// TODO: make it more abstract (JSON, YAML, ...)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	return &cfg, nil
}
