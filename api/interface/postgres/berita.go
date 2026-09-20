package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/berita"
	beritaUsecase "webdesa/api/usecase/berita"
)

// BeritaRepository implements usecase/berita.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/berita where it's USED, not here where it's implemented.
type BeritaRepository struct {
	db *sqlx.DB
}

// NewBeritaRepository creates a new BeritaRepository instance
func NewBeritaRepository(db *sqlx.DB) beritaUsecase.Repository {
	return &BeritaRepository{db: db}
}

// Create creates a new berita in the database
// Generates UUID, sets timestamps, and inserts the berita record
// Validates: Requirements 9.3, 20.2, 20.3
func (r *BeritaRepository) Create(ctx context.Context, b *berita.Berita) error {
	// Generate UUID if not provided
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	query := `
		INSERT INTO berita (id, title, content, category, image_media_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		b.ID,
		b.Title,
		b.Content,
		b.Category,
		b.ImageMediaID,
		b.Status,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create berita: %w", err))
	}

	return nil
}

// FindByID retrieves a berita by ID
// Returns error if berita not found
// Validates: Requirements 9.4
func (r *BeritaRepository) FindByID(ctx context.Context, id string) (*berita.Berita, error) {
	var b berita.Berita

	query := `
		SELECT berita.id, berita.title, berita.content, berita.category, berita.image_media_id, berita.status, berita.created_at, berita.updated_at,
		       berita_categories.name AS category_name
		FROM berita
		LEFT JOIN berita_categories ON berita.category = berita_categories.id
		WHERE berita.id = $1
	`

	err := r.db.GetContext(ctx, &b, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("berita not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find berita by ID: %w", err))
	}

	return &b, nil
}

// List retrieves paginated berita with optional filtering
// Returns berita slice, total count, and error
// Uses ILIKE for search queries, BETWEEN for date range
// Validates: Requirements 9.1, 9.2, 17.2, 20.5
func (r *BeritaRepository) List(ctx context.Context, query *string, category *string, status *string, since *time.Time, until *time.Time, sort, order string, offset, limit int) ([]*berita.Berita, int, error) {
	// Build count query with filters
	countQuery := `SELECT COUNT(*) FROM berita WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if query != nil && *query != "" {
		countQuery += fmt.Sprintf(` AND (title ILIKE $%d OR content ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	if category != nil && *category != "" {
		countQuery += fmt.Sprintf(` AND category = $%d`, argIndex)
		countArgs = append(countArgs, *category)
		argIndex++
	}

	if status != nil && *status != "" {
		countQuery += fmt.Sprintf(` AND status = $%d`, argIndex)
		countArgs = append(countArgs, *status)
		argIndex++
	}

	if since != nil {
		countQuery += fmt.Sprintf(` AND created_at >= $%d`, argIndex)
		countArgs = append(countArgs, *since)
		argIndex++
	}

	if until != nil {
		countQuery += fmt.Sprintf(` AND created_at <= $%d`, argIndex)
		countArgs = append(countArgs, *until)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count berita: %w", err))
	}

	// Build list query with filters
	listQuery := `
		SELECT berita.id, berita.title, berita.content, berita.category, berita.image_media_id, berita.status, berita.created_at, berita.updated_at,
		       berita_categories.name AS category_name
		FROM berita
		LEFT JOIN berita_categories ON berita.category = berita_categories.id
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex = 1

	if query != nil && *query != "" {
		listQuery += fmt.Sprintf(` AND (berita.title ILIKE $%d OR berita.content ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		listArgs = append(listArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	if category != nil && *category != "" {
		listQuery += fmt.Sprintf(` AND berita.category = $%d`, argIndex)
		listArgs = append(listArgs, *category)
		argIndex++
	}

	if status != nil && *status != "" {
		listQuery += fmt.Sprintf(` AND berita.status = $%d`, argIndex)
		listArgs = append(listArgs, *status)
		argIndex++
	}

	if since != nil {
		listQuery += fmt.Sprintf(` AND berita.created_at >= $%d`, argIndex)
		listArgs = append(listArgs, *since)
		argIndex++
	}

	if until != nil {
		listQuery += fmt.Sprintf(` AND berita.created_at <= $%d`, argIndex)
		listArgs = append(listArgs, *until)
		argIndex++
	}

	// Validate sort + order against whitelist to prevent SQL injection
	sortColumn := "created_at"
	switch sort {
	case "created_at", "title":
		sortColumn = sort
	}
	sortDirection := "DESC"
	switch order {
	case "asc", "ASC":
		sortDirection = "ASC"
	case "desc", "DESC":
		sortDirection = "DESC"
	}

	listQuery += fmt.Sprintf(` ORDER BY berita.%s %s LIMIT $%d OFFSET $%d`, sortColumn, sortDirection, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated berita
	var beritas []*berita.Berita
	err = r.db.SelectContext(ctx, &beritas, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list berita: %w", err))
	}

	return beritas, total, nil
}

// Update updates an existing berita
// Sets updated_at timestamp automatically
// Validates: Requirements 9.5, 20.3
func (r *BeritaRepository) Update(ctx context.Context, b *berita.Berita) error {
	query := `
		UPDATE berita
		SET title = $1, content = $2, category = $3, image_media_id = $4, status = $5, updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		b.Title,
		b.Content,
		b.Category,
		b.ImageMediaID,
		b.Status,
		b.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update berita: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("berita not found: %s", b.ID))
	}

	return nil
}

// Delete removes a berita
// Validates: Requirements 9.6
func (r *BeritaRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM berita WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete berita: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("berita not found: %s", id))
	}

	return nil
}
