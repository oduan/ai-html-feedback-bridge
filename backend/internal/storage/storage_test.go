package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Open() failed: %v", err)
	}
	cleanup := func() {
		db.Close()
		os.RemoveAll(dir)
	}
	return db, cleanup
}

func TestInitCreatesTable(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Verify the interactions table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='interactions'").Scan(&tableName)
	if err != nil {
		t.Fatalf("interactions table not found: %v", err)
	}
	if tableName != "interactions" {
		t.Errorf("expected table name 'interactions', got %q", tableName)
	}
}

func TestMigrationIdempotent(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate() failed: %v", err)
	}
	// Second run should not error
}

func TestTableColumns(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	rows, err := db.Query("PRAGMA table_info(interactions)")
	if err != nil {
		t.Fatalf("query table info failed: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan column info failed: %v", err)
		}
		columns[name] = ctype
	}

	expected := map[string]string{
		"interaction_id": "TEXT",
		"content_type":   "TEXT",
		"content":        "TEXT",
		"created_at":     "TEXT",
		"updated_at":     "TEXT",
	}

	for name, ctype := range expected {
		got, ok := columns[name]
		if !ok {
			t.Errorf("missing column: %s", name)
			continue
		}
		if got != ctype {
			t.Errorf("column %s: expected type %s, got %s", name, ctype, got)
		}
	}
}
