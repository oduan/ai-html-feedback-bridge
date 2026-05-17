package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when an interaction is not found in the database.
var ErrNotFound = errors.New("interaction not found")

// Interaction represents a user-submitted content for a specific interaction.
type Interaction struct {
	InteractionID string
	ContentType   string
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Store provides CRUD operations for interactions backed by SQLite.
type Store struct {
	DB *sql.DB
}

// SaveInteraction upserts an interaction. If the interaction_id already exists,
// it overwrites content and content_type while preserving the original created_at.
// All times are stored as RFC3339 UTC strings.
func (s *Store) SaveInteraction(interaction *Interaction) error {
	// We use UPSERT: on conflict, update content, content_type, and updated_at,
	// but keep the original created_at.
	query := `
		INSERT INTO interactions (interaction_id, content_type, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(interaction_id) DO UPDATE SET
			content_type = excluded.content_type,
			content = excluded.content,
			updated_at = excluded.updated_at
	`

	_, err := s.DB.Exec(query,
		interaction.InteractionID,
		interaction.ContentType,
		interaction.Content,
		interaction.CreatedAt.UTC().Format(time.RFC3339),
		interaction.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("save interaction: %w", err)
	}

	return nil
}

// GetInteraction retrieves an interaction by its ID.
// Returns ErrNotFound if the interaction does not exist.
func (s *Store) GetInteraction(id string) (*Interaction, error) {
	query := `
		SELECT interaction_id, content_type, content, created_at, updated_at
		FROM interactions
		WHERE interaction_id = ?
	`

	var interaction Interaction
	var createdStr, updatedStr string

	err := s.DB.QueryRow(query, id).Scan(
		&interaction.InteractionID,
		&interaction.ContentType,
		&interaction.Content,
		&createdStr,
		&updatedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get interaction: %w", err)
	}

	interaction.CreatedAt, err = time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	interaction.UpdatedAt, err = time.Parse(time.RFC3339, updatedStr)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &interaction, nil
}
