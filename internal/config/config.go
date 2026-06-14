package config

import (
	"fmt"
	"os"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		Addr:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("Database URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
