package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all service configuration values.
type Config struct {
	PORT           string
	SQLitePath     string
	MaxContentSize int
}

// Load reads configuration from environment variables,
// applying defaults where values are not set.
func Load() (*Config, error) {
	cfg := &Config{
		PORT:       getEnv("PORT", "8080"),
		SQLitePath: getEnv("SQLITE_PATH", "./data/app.db"),
	}

	maxSizeStr := getEnv("MAX_CONTENT_SIZE", "1048576")
	maxSize, err := strconv.Atoi(maxSizeStr)
	if err != nil {
		return nil, fmt.Errorf("MAX_CONTENT_SIZE: invalid integer %q", maxSizeStr)
	}
	if maxSize <= 0 {
		return nil, fmt.Errorf("MAX_CONTENT_SIZE: must be positive, got %d", maxSize)
	}
	cfg.MaxContentSize = maxSize

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
