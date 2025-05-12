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

type DurationMs time.Duration

func (d *DurationMs) UnmarshalJSON(b []byte) error {
	var ms int64
	if err := json.Unmarshal(b, &ms); err != nil {
		return fmt.Errorf("DurationMs: expected number of milliseconds, got %s: %w", string(b), err)
	}
	*d = DurationMs(time.Duration(ms))
	return nil
}

func (d DurationMs) AsDuration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

type LoadBalancerConfig struct {
	Algorithm             string     `json:"algorithm"`
	HealthCheckIntervalMS DurationMs `json:"health_check_interval_ms"`
}

type TokenBucketLimiterOptions struct {
	DefaultCapacity         int        `json:"default_capacity"`
	DefaultRefillIntervalMS DurationMs `json:"refill_interval_ms"`
}

type RateLimitConfig struct {
	Algorithm string `json:"algorithm"`
	Options   any    `json:"options"`
}

type rawRateLimitConfig struct {
	Algorithm string          `json:"algorithm"`
	Options   json.RawMessage `json:"options"`
}

func (rl *RateLimitConfig) UnmarshalJSON(data []byte) error {
	var raw rawRateLimitConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal rate limiter object: %w", err)
	}

	switch raw.Algorithm {
	case "token_bucket":
		var cfg TokenBucketLimiterOptions
		if err := json.Unmarshal(raw.Options, &cfg); err != nil {
			return fmt.Errorf("failed to unmarshal rate limiter object: %w", err)
		}
		rl.Options = cfg
	default:
		return fmt.Errorf("unknown algorithm %q", raw.Algorithm)
	}

	rl.Algorithm = raw.Algorithm
	return nil
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
