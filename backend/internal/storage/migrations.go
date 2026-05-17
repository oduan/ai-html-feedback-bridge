package storage

import (
	"database/sql"
	"fmt"
)

// Migrate runs all pending database migrations.
// It is idempotent and safe to call multiple times.
func Migrate(db *sql.DB) error {
	// Create migrations tracking table if not exists
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			name TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "001_create_interactions",
			sql: `
				CREATE TABLE IF NOT EXISTS interactions (
					interaction_id TEXT PRIMARY KEY,
					content_type   TEXT NOT NULL,
					content        TEXT NOT NULL,
					created_at     TEXT NOT NULL,
					updated_at     TEXT NOT NULL
				)
			`,
		},
	}

	for _, m := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE name = ?", m.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if count > 0 {
			continue
		}

		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}

		if _, err := db.Exec("INSERT INTO _migrations (name, applied_at) VALUES (?, datetime('now'))", m.name); err != nil {
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}

	return nil
}
