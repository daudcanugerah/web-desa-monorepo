package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/struktur"
	strukturUsecase "webdesa/api/usecase/struktur"
)

// StrukturRepository implements usecase/struktur.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/struktur where it's USED, not here where it's implemented.
type StrukturRepository struct {
	db *sqlx.DB
}

// NewStrukturRepository creates a new StrukturRepository instance
func NewStrukturRepository(db *sqlx.DB) strukturUsecase.Repository {
	return &StrukturRepository{db: db}
}

// Create creates a new struktur member in the database
// Generates UUID, sets timestamps, and inserts the struktur record
// Validates: Requirements 13.3, 20.2, 20.3
func (r *StrukturRepository) Create(ctx context.Context, s *struktur.Struktur) error {
	// Generate UUID if not provided
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	query := `
		INSERT INTO struktur_organisasi (id, name, position, email, phone, profile_image_media_id, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		s.ID,
		s.Name,
		s.Position,
		s.Email,
		s.Phone,

		s.ProfileImageMediaID,
		s.Description,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create struktur: %w", err))
	}

	return nil
}

// FindByID retrieves a struktur member by ID
// Returns error if struktur not found
// Validates: Requirements 13.4
func (r *StrukturRepository) FindByID(ctx context.Context, id string) (*struktur.Struktur, error) {
	var s struktur.Struktur

	query := `
		SELECT id, name, position, email, phone, profile_image_media_id, description, created_at, updated_at
		FROM struktur_organisasi
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &s, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("struktur not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find struktur by ID: %w", err))
	}

	return &s, nil
}

// List retrieves paginated struktur members with optional search filtering
// Returns struktur slice, total count, and error
// Uses ILIKE for search queries on name and position
// Validates: Requirements 13.1, 13.2, 17.2, 20.5
func (r *StrukturRepository) List(ctx context.Context, query *string, offset, limit int) ([]*struktur.Struktur, int, error) {
	// Build count query with optional search filter
	countQuery := `SELECT COUNT(*) FROM struktur_organisasi WHERE 1=1`
	var countArgs []interface{}

	if query != nil && *query != "" {
		countQuery += ` AND (name ILIKE $1 OR position ILIKE $1)`
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern)
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count struktur: %w", err))
	}

	// Build list query with optional search filter
	listQuery := `
		SELECT id, name, position, email, phone, profile_image_media_id, description, created_at, updated_at
		FROM struktur_organisasi
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex := 1

	if query != nil && *query != "" {
		listQuery += fmt.Sprintf(` AND (name ILIKE $%d OR position ILIKE $%d)`, argIndex, argIndex)
		searchPattern := "%" + *query + "%"
		listArgs = append(listArgs, searchPattern)
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated struktur members
	var strukturs []*struktur.Struktur
	err = r.db.SelectContext(ctx, &strukturs, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list struktur: %w", err))
	}

	return strukturs, total, nil
}

// Update updates an existing struktur member
// Sets updated_at timestamp automatically
// Validates: Requirements 13.5, 20.3
func (r *StrukturRepository) Update(ctx context.Context, s *struktur.Struktur) error {
	query := `
		UPDATE struktur_organisasi
		SET name = $1, position = $2, email = $3, phone = $4, profile_image_media_id = $5, description = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		s.Name,
		s.Position,
		s.Email,
		s.Phone,
		s.ProfileImageMediaID,
		s.Description,
		s.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update struktur: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("struktur not found: %s", s.ID))
	}

	return nil
}

// Delete removes a struktur member
// Validates: Requirements 13.6
func (r *StrukturRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM struktur_organisasi WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete struktur: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("struktur not found: %s", id))
	}

	return nil
}
