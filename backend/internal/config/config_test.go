package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	// Ensure no env vars are set
	os.Unsetenv("PORT")
	os.Unsetenv("SQLITE_PATH")
	os.Unsetenv("MAX_CONTENT_SIZE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.PORT != "8080" {
		t.Errorf("expected PORT 8080, got %q", cfg.PORT)
	}
	expectedPath := expectedDefaultPath(t)
	if cfg.SQLitePath != expectedPath {
		t.Errorf("expected SQLITE_PATH %q, got %q", expectedPath, cfg.SQLitePath)
	}
	if cfg.MaxContentSize != 1048576 {
		t.Errorf("expected MAX_CONTENT_SIZE 1048576, got %d", cfg.MaxContentSize)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("SQLITE_PATH", "/tmp/test.db")
	os.Setenv("MAX_CONTENT_SIZE", "512000")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("SQLITE_PATH")
		os.Unsetenv("MAX_CONTENT_SIZE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.PORT != "9090" {
		t.Errorf("expected PORT 9090, got %q", cfg.PORT)
	}
	if cfg.SQLitePath != "/tmp/test.db" {
		t.Errorf("expected SQLITE_PATH /tmp/test.db, got %q", cfg.SQLitePath)
	}
	if cfg.MaxContentSize != 512000 {
		t.Errorf("expected MAX_CONTENT_SIZE 512000, got %d", cfg.MaxContentSize)
	}
}

func TestInvalidMaxContentSize(t *testing.T) {
	os.Setenv("MAX_CONTENT_SIZE", "not-a-number")
	defer os.Unsetenv("MAX_CONTENT_SIZE")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid MAX_CONTENT_SIZE, got nil")
	}
}

func TestInvalidMaxContentSizeNegative(t *testing.T) {
	os.Setenv("MAX_CONTENT_SIZE", "-100")
	defer os.Unsetenv("MAX_CONTENT_SIZE")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for negative MAX_CONTENT_SIZE, got nil")
	}
}

func TestInvalidMaxContentSizeZero(t *testing.T) {
	os.Setenv("MAX_CONTENT_SIZE", "0")
	defer os.Unsetenv("MAX_CONTENT_SIZE")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for zero MAX_CONTENT_SIZE, got nil")
	}
}

func TestPartialOverride(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Unsetenv("SQLITE_PATH")
	os.Unsetenv("MAX_CONTENT_SIZE")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.PORT != "3000" {
		t.Errorf("expected PORT 3000, got %q", cfg.PORT)
	}
	expectedPath := expectedDefaultPath(t)
	if cfg.SQLitePath != expectedPath {
		t.Errorf("expected default SQLITE_PATH %q, got %q", expectedPath, cfg.SQLitePath)
	}
	if cfg.MaxContentSize != 1048576 {
		t.Errorf("expected default MAX_CONTENT_SIZE 1048576, got %d", cfg.MaxContentSize)
	}
}

// expectedDefaultPath replicates defaultSQLitePath logic for test assertions.
func expectedDefaultPath(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "data", "app.db")
}
