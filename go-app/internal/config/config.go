package config

import (
	"errors"

	"github.com/caarlos0/env/v10"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port        int    `env:"PORT" envDefault:"3002"`
	DatabaseURL string `env:"DATABASE_URL,required"`
}

// Load parses environment variables into Config.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}
