package storage

import (
	"errors"
	"testing"
	"time"
)

func TestSaveAndGetJSON(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}
	now := time.Now().UTC().Truncate(time.Second)

	interaction := &Interaction{
		InteractionID: "test-001",
		ContentType:   "application/json",
		Content:       `{"feedback":"需要补充项目排期","approved":true}`,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.SaveInteraction(interaction); err != nil {
		t.Fatalf("SaveInteraction() failed: %v", err)
	}

	got, err := store.GetInteraction("test-001")
	if err != nil {
		t.Fatalf("GetInteraction() failed: %v", err)
	}

	if got.InteractionID != "test-001" {
		t.Errorf("expected interaction_id test-001, got %q", got.InteractionID)
	}
	if got.ContentType != "application/json" {
		t.Errorf("expected content_type application/json, got %q", got.ContentType)
	}
	if got.Content != `{"feedback":"需要补充项目排期","approved":true}` {
		t.Errorf("content mismatch: got %q", got.Content)
	}
}

func TestSaveAndGetPlainText(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}
	now := time.Now().UTC().Truncate(time.Second)

	interaction := &Interaction{
		InteractionID: "text-001",
		ContentType:   "text/plain; charset=utf-8",
		Content:       "This is plain text feedback",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.SaveInteraction(interaction); err != nil {
		t.Fatalf("SaveInteraction() failed: %v", err)
	}

	got, err := store.GetInteraction("text-001")
	if err != nil {
		t.Fatalf("GetInteraction() failed: %v", err)
	}

	if got.Content != "This is plain text feedback" {
		t.Errorf("content mismatch: got %q", got.Content)
	}
}

func TestOverwritePreservesCreatedAt(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}
	createdAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)

	// First save
	first := &Interaction{
		InteractionID: "overwrite-001",
		ContentType:   "text/plain",
		Content:       "original content",
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
	if err := store.SaveInteraction(first); err != nil {
		t.Fatalf("first SaveInteraction() failed: %v", err)
	}

	// Overwrite with a later time
	updatedAt := time.Date(2026, 5, 16, 10, 0, 1, 0, time.UTC)
	second := &Interaction{
		InteractionID: "overwrite-001",
		ContentType:   "application/json",
		Content:       `{"updated": true}`,
		CreatedAt:     createdAt, // Should be ignored, original preserved
		UpdatedAt:     updatedAt,
	}
	if err := store.SaveInteraction(second); err != nil {
		t.Fatalf("second SaveInteraction() failed: %v", err)
	}

	got, err := store.GetInteraction("overwrite-001")
	if err != nil {
		t.Fatalf("GetInteraction() failed: %v", err)
	}

	if got.Content != `{"updated": true}` {
		t.Errorf("expected updated content, got %q", got.Content)
	}
	if got.ContentType != "application/json" {
		t.Errorf("expected updated content_type application/json, got %q", got.ContentType)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("created_at should not change on overwrite: expected %v, got %v", createdAt, got.CreatedAt)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("updated_at should be %v, got %v", updatedAt, got.UpdatedAt)
	}
}

func TestGetNotFound(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}

	_, err := store.GetInteraction("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent interaction, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestContentPreservedExactly(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}
	now := time.Now().UTC().Truncate(time.Second)

	content := `<html><body><h1>Title</h1><p>段落内容</p></body></html>`
	interaction := &Interaction{
		InteractionID: "html-001",
		ContentType:   "text/html",
		Content:       content,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.SaveInteraction(interaction); err != nil {
		t.Fatalf("SaveInteraction() failed: %v", err)
	}

	got, err := store.GetInteraction("html-001")
	if err != nil {
		t.Fatalf("GetInteraction() failed: %v", err)
	}

	if got.Content != content {
		t.Errorf("content not preserved exactly:\nexpected: %q\ngot:      %q", content, got.Content)
	}
}

func TestTimesAreUTC(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := &Store{DB: db}
	now := time.Now().UTC()

	interaction := &Interaction{
		InteractionID: "time-001",
		ContentType:   "text/plain",
		Content:       "test",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.SaveInteraction(interaction); err != nil {
		t.Fatalf("SaveInteraction() failed: %v", err)
	}

	got, err := store.GetInteraction("time-001")
	if err != nil {
		t.Fatalf("GetInteraction() failed: %v", err)
	}

	// Verify the location is UTC
	if got.CreatedAt.Location() != time.UTC {
		t.Errorf("created_at should be UTC, got %v", got.CreatedAt.Location())
	}
	if got.UpdatedAt.Location() != time.UTC {
		t.Errorf("updated_at should be UTC, got %v", got.UpdatedAt.Location())
	}
}
