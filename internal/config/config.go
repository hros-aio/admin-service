// Package config provides application configuration loading and validation.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config represents the global application configuration.
type Config struct {
	AppName      string   `env:"APP_NAME,required"`
	Env          string   `env:"ENV,required"`
	Port         int      `env:"PORT,required"`
	DBURL        string   `env:"DB_URL,required"`
	RedisURL     string   `env:"REDIS_URL,required"`
	KafkaBrokers []string `env:"KAFKA_BROKERS,required"`
	LogLevel     string   `env:"LOG_LEVEL" envDefault:"info"`
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// Validate performs basic validation on the configuration.
func (c *Config) Validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	// Additional validation can be added here
	return nil
}
