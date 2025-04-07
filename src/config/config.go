package config

import (
	"os"
)

type Config struct {
	MocoDomain string
	MocoAPIKey string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		MocoDomain: os.Getenv("MOCO_DOMAIN"),
		MocoAPIKey: os.Getenv("MOCO_API_KEY"),
	}

	if cfg.MocoDomain == "" || cfg.MocoAPIKey == "" {
		return nil, ErrMissingEnvVars
	}

	return cfg, nil
}

var ErrMissingEnvVars = &ConfigError{"MOCO_DOMAIN and MOCO_API_KEY environment variables must be set"}

type ConfigError struct {
	message string
}

func (e *ConfigError) Error() string {
	return e.message
}
