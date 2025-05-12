package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
)

type ServerConfig struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

type LoadBalancerConfig struct {
	Algorithm           string `json:"algorithm"`
	HealthCheckInterval int    `json:"health_check_interval_sec"`
}

type DurationMs time.Duration

func (d *DurationMs) UnmarshalJSON(b []byte) error {
	var ms int64
	if err := json.Unmarshal(b, &ms); err != nil {
		return fmt.Errorf("DurationMs: expected number of milliseconds, got %s: %w", string(b), err)
	}
	*d = DurationMs(time.Duration(ms) * time.Millisecond)
	return nil
}

func (d DurationMs) AsDuration() time.Duration {
	return time.Duration(d)
}

type RateLimitConfig struct {
	Algorithm           string     `json:"algorithm"`
	DefaultCapacity     int        `json:"default_capacity"`
	DefaultRefillPeriod DurationMs `json:"refill_pepiod_ms"`
}

type DBConfig struct {
	DSN string `json:"dsn"`
}

type BackendConfig struct {
	URL *url.URL
}

type rawBackendConfig struct {
	URL string `json:"url"`
}

func (b *BackendConfig) UnmarshalJSON(data []byte) error {
	var raw rawBackendConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal backend object: %w", err)
	}
	parsed, err := url.Parse(raw.URL)
	if err != nil {
		return fmt.Errorf("invalid backend URL %q: %w", raw.URL, err)
	}
	b.URL = parsed
	return nil
}

type Config struct {
	Server       ServerConfig       `json:"server"`
	Backends     []BackendConfig    `json:"backends"`
	LoadBalancer LoadBalancerConfig `json:"balancer"`
	RateLimit    RateLimitConfig    `json:"rate_limit"`
	DB           DBConfig           `json:"db"`
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
