package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MocoDomain string
	MocoAPIKey string
}

func LoadConfig() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

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
