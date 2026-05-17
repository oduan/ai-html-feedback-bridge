package config

import (
	"fmt"
	"os"
	"path/filepath"
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
		SQLitePath: getEnv("SQLITE_PATH", defaultSQLitePath()),
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

// defaultSQLitePath returns the default SQLite database path
// relative to the executable's directory, so the db file is always
// alongside the binary regardless of the working directory.
func defaultSQLitePath() string {
	exe, err := os.Executable()
	if err != nil {
		// Fallback: if we can't determine the exe path, use working dir
		return "./data/app.db"
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "data", "app.db")
}
