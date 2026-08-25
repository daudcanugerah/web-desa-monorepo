package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/umkm"
	umkmUsecase "webdesa/api/usecase/umkm"
)

// UMKMRepository implements usecase/umkm.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/umkm where it's USED, not here where it's implemented.
type UMKMRepository struct {
	db *sqlx.DB
}

// NewUMKMRepository creates a new UMKMRepository instance
func NewUMKMRepository(db *sqlx.DB) umkmUsecase.Repository {
	return &UMKMRepository{db: db}
}

// Create creates a new UMKM in the database
// Generates UUID, sets timestamps, and inserts the UMKM record
// Validates: Requirements 10.5, 20.2, 20.3
func (r *UMKMRepository) Create(ctx context.Context, u *umkm.UMKM) error {
	// Generate UUID if not provided
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	// Marshal images array to JSON for array column
	var imagesJSON []byte
	var err error
	if len(u.Images) > 0 {
		imagesJSON, err = json.Marshal(u.Images)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to marshal images: %w", err))
		}
	}

	query := `
		INSERT INTO umkm (id, name, owner, address, phone, email, website, category, description, images, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		u.ID,
		u.Name,
		u.Owner,
		u.Address,
		u.Phone,
		u.Email,
		u.Website,
		u.Category,
		u.Description,
		imagesJSON,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create umkm: %w", err))
	}

	return nil
}

// FindByID retrieves a UMKM by ID
// Returns error if UMKM not found
// Validates: Requirements 10.6
func (r *UMKMRepository) FindByID(ctx context.Context, id string) (*umkm.UMKM, error) {
	var u umkm.UMKM
	var imagesJSON []byte

	query := `
		SELECT umkm.id, umkm.name, umkm.owner, umkm.address, umkm.phone, umkm.email, umkm.website, umkm.category, umkm_categories.name AS category_name, umkm.description, umkm.images, umkm.created_at, umkm.updated_at
		FROM umkm
		LEFT JOIN umkm_categories ON umkm.category = umkm_categories.id
		WHERE umkm.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Owner,
		&u.Address,
		&u.Phone,
		&u.Email,
		&u.Website,
		&u.Category,
		&u.CategoryName,
		&u.Description,
		&imagesJSON,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("umkm not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find umkm by ID: %w", err))
	}

	// Unmarshal images from JSON
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &u.Images); err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("failed to unmarshal images: %w", err))
		}
	} else {
		// Initialize as empty array if no images
		u.Images = []string{}
	}

	return &u, nil
}

// List retrieves paginated UMKM with optional search filtering
// Returns UMKM slice, total count, and error
// Uses ILIKE for search queries
// Validates: Requirements 10.1, 10.2, 17.2, 20.5
func (r *UMKMRepository) List(ctx context.Context, query *string, category *string, offset, limit int) ([]*umkm.UMKM, int, error) {
	// Build count query with optional search filter
	countQuery := `SELECT COUNT(*) FROM umkm WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if query != nil && *query != "" {
		countQuery += fmt.Sprintf(` AND (name ILIKE $%d OR description ILIKE $%d OR owner ILIKE $%d)`, argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	if category != nil && *category != "" {
		countQuery += fmt.Sprintf(` AND category = $%d`, argIndex)
		countArgs = append(countArgs, *category)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count umkm: %w", err))
	}

	// Build list query with optional search filter
	listQuery := `
		SELECT umkm.id, umkm.name, umkm.owner, umkm.address, umkm.phone, umkm.email, umkm.website, umkm.category, umkm_categories.name AS category_name, umkm.description, umkm.images, umkm.created_at, umkm.updated_at
		FROM umkm
		LEFT JOIN umkm_categories ON umkm.category = umkm_categories.id
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex = 1

	if query != nil && *query != "" {
		listQuery += fmt.Sprintf(` AND (umkm.name ILIKE $%d OR umkm.description ILIKE $%d OR umkm.owner ILIKE $%d)`, argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + *query + "%"
		listArgs = append(listArgs, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	if category != nil && *category != "" {
		listQuery += fmt.Sprintf(` AND umkm.category = $%d`, argIndex)
		listArgs = append(listArgs, *category)
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY umkm.created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated UMKM
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list umkm: %w", err))
	}
	defer rows.Close()

	var umkms []*umkm.UMKM
	for rows.Next() {
		var u umkm.UMKM
		var imagesJSON []byte

		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Owner,
			&u.Address,
			&u.Phone,
			&u.Email,
			&u.Website,
			&u.Category,
			&u.CategoryName,
			&u.Description,
			&imagesJSON,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errtrace.Wrap(fmt.Errorf("failed to scan umkm: %w", err))
		}

		// Unmarshal images from JSON
		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &u.Images); err != nil {
				return nil, 0, errtrace.Wrap(fmt.Errorf("failed to unmarshal images: %w", err))
			}
		} else {
			// Initialize as empty array if no images
			u.Images = []string{}
		}

		umkms = append(umkms, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to iterate umkm: %w", err))
	}

	return umkms, total, nil
}

// Update updates an existing UMKM
// Sets updated_at timestamp automatically
// Validates: Requirements 10.7, 20.3
func (r *UMKMRepository) Update(ctx context.Context, u *umkm.UMKM) error {
	// Marshal images array to JSON for array column
	var imagesJSON []byte
	var err error
	if len(u.Images) > 0 {
		imagesJSON, err = json.Marshal(u.Images)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to marshal images: %w", err))
		}
	}

	query := `
		UPDATE umkm
		SET name = $1, owner = $2, address = $3, phone = $4, email = $5, website = $6, category = $7, description = $8, images = $9, updated_at = NOW()
		WHERE id = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		u.Name,
		u.Owner,
		u.Address,
		u.Phone,
		u.Email,
		u.Website,
		u.Category,
		u.Description,
		imagesJSON,
		u.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update umkm: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("umkm not found: %s", u.ID))
	}

	return nil
}

// Delete removes a UMKM
// Validates: Requirements 10.8
func (r *UMKMRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM umkm WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete umkm: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("umkm not found: %s", id))
	}

	return nil
}
