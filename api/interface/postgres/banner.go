package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/banner"
	bannerUsecase "webdesa/api/usecase/banner"
)

// BannerRepository implements usecase/banner.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/banner where it's USED, not here where it's implemented.
type BannerRepository struct {
	db *sqlx.DB
}

// NewBannerRepository creates a new BannerRepository instance
func NewBannerRepository(db *sqlx.DB) bannerUsecase.Repository {
	return &BannerRepository{db: db}
}

// Create creates a new banner in the database
// Generates UUID, sets timestamps, and inserts the banner record
// Validates: Requirements 7.2, 20.2, 20.3
func (r *BannerRepository) Create(ctx context.Context, b *banner.Banner) error {
	// Generate UUID if not provided
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	// Marshal metadata to JSON for JSONB column
	var metadataJSON []byte
	var err error
	if b.Metadata != nil && len(b.Metadata) > 0 {
		metadataJSON, err = json.Marshal(b.Metadata)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to marshal metadata: %w", err))
		}
	} else {
		// Store empty object for nil or empty metadata
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO banners (id, title, description, link, image_media_id, status, category, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		b.ID,
		b.Title,
		b.Description,
		b.Link,
		b.ImageMediaID,
		b.Status,
		b.Category,
		metadataJSON,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create banner: %w", err))
	}

	return nil
}

// FindByID retrieves a banner by ID
// Returns error if banner not found
// Validates: Requirements 7.3
// FindByID retrieves a banner by ID
// Returns error if banner not found
// Validates: Requirements 7.3
func (r *BannerRepository) FindByID(ctx context.Context, id string) (*banner.Banner, error) {
	var b banner.Banner
	var metadataJSON []byte

	query := `
		SELECT banners.id AS id, banners.title AS title, banners.description AS description, banners.link AS link, banners.image_media_id AS image_media_id, banners.status AS status, banners.category AS category, banner_categories.name AS category_name, banners.metadata AS metadata, banners.created_at AS created_at, banners.updated_at AS updated_at
		FROM banners
		LEFT JOIN banner_categories ON banners.category = banner_categories.id
		WHERE banners.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID,
		&b.Title,
		&b.Description,
		&b.Link,
		&b.ImageMediaID,
		&b.Status,
		&b.Category,
		&b.CategoryName,
		&metadataJSON,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("banner not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find banner by ID: %w", err))
	}

	// Unmarshal metadata from JSON
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &b.Metadata); err != nil {
			return nil, errtrace.Wrap(fmt.Errorf("failed to unmarshal metadata: %w", err))
		}
	}

	return &b, nil
}

// List retrieves paginated banners with optional status filtering
// Returns banners slice, total count, and error
// Uses LIMIT/OFFSET for pagination
// Validates: Requirements 7.1, 7.7, 17.2, 20.5
// Uses LIMIT/OFFSET for pagination
// Validates: Requirements 7.1, 7.7, 17.2, 20.5
func (r *BannerRepository) List(ctx context.Context, status *string, query *string, category *string, offset, limit int) ([]*banner.Banner, int, error) {
	// Build count query with optional filters
	countQuery := `SELECT COUNT(*) FROM banners WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if status != nil && *status != "" {
		countQuery += fmt.Sprintf(` AND banners.status = $%d`, argIndex)
		countArgs = append(countArgs, *status)
		argIndex++
	}

	if query != nil && *query != "" {
		searchPattern := "%" + *query + "%"
		countQuery += fmt.Sprintf(` AND (banners.title ILIKE $%d OR banners.description ILIKE $%d)`, argIndex, argIndex+1)
		countArgs = append(countArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	if category != nil && *category != "" {
		countQuery += fmt.Sprintf(` AND banners.category = $%d`, argIndex)
		countArgs = append(countArgs, *category)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count banners: %w", err))
	}

	// Build list query with optional filters
	listQuery := `
		SELECT banners.id AS id, banners.title AS title, banners.description AS description, banners.link AS link, banners.image_media_id AS image_media_id, banners.status AS status, banners.category AS category, banner_categories.name AS category_name, banners.metadata AS metadata, banners.created_at AS created_at, banners.updated_at AS updated_at
		FROM banners
		LEFT JOIN banner_categories ON banners.category = banner_categories.id
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex = 1

	if status != nil && *status != "" {
		listQuery += fmt.Sprintf(` AND banners.status = $%d`, argIndex)
		listArgs = append(listArgs, *status)
		argIndex++
	}

	if query != nil && *query != "" {
		searchPattern := "%" + *query + "%"
		listQuery += fmt.Sprintf(` AND (banners.title ILIKE $%d OR banners.description ILIKE $%d)`, argIndex, argIndex+1)
		listArgs = append(listArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	if category != nil && *category != "" {
		listQuery += fmt.Sprintf(` AND banners.category = $%d`, argIndex)
		listArgs = append(listArgs, *category)
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY banners.created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated banners
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list banners: %w", err))
	}
	defer rows.Close()

	var banners []*banner.Banner
	for rows.Next() {
		var b banner.Banner
		var metadataJSON []byte

		err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.Description,
			&b.Link,
			&b.ImageMediaID,
			&b.Status,
			&b.Category,
			&b.CategoryName,
			&metadataJSON,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errtrace.Wrap(fmt.Errorf("failed to scan banner: %w", err))
		}

		// Unmarshal metadata from JSON
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &b.Metadata); err != nil {
				return nil, 0, errtrace.Wrap(fmt.Errorf("failed to unmarshal metadata: %w", err))
			}
		}

		banners = append(banners, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("error iterating banners: %w", err))
	}

	return banners, total, nil
}

// Update updates an existing banner
// Sets updated_at timestamp automatically
// Validates: Requirements 7.4, 20.3
// Validates: Requirements 7.4, 20.3
func (r *BannerRepository) Update(ctx context.Context, b *banner.Banner) error {
	// Marshal metadata to JSON for JSONB column
	var metadataJSON []byte
	var err error
	if b.Metadata != nil && len(b.Metadata) > 0 {
		metadataJSON, err = json.Marshal(b.Metadata)
		if err != nil {
			return errtrace.Wrap(fmt.Errorf("failed to marshal metadata: %w", err))
		}
	} else {
		// Store empty object for nil or empty metadata
		metadataJSON = []byte("{}")
	}

	query := `
		UPDATE banners
		SET title = $1, description = $2, link = $3, image_media_id = $4, status = $5, category = $6, metadata = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		b.Title,
		b.Description,
		b.Link,
		b.ImageMediaID,
		b.Status,
		b.Category,
		metadataJSON,
		b.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update banner: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("banner not found: %s", b.ID))
	}

	return nil
}

// Delete removes a banner
// Validates: Requirements 7.6
func (r *BannerRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM banners WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete banner: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("banner not found: %s", id))
	}

	return nil
}

// CountActiveBanners returns the count of banners with status "active"
// Used to enforce the limit of 20 active banners
// Validates: Requirements 7.5
func (r *BannerRepository) CountActiveBanners(ctx context.Context) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM banners WHERE status = $1`

	err := r.db.GetContext(ctx, &count, query, banner.StatusActive)
	if err != nil {
		return 0, errtrace.Wrap(fmt.Errorf("failed to count active banners: %w", err))
	}

	return count, nil
}

// CountByImageMediaID returns how many banners still reference the given
// image media id. Banners.image_media_id is ON DELETE SET NULL, so a shared
// media row must survive a single banner's deletion.
func (r *BannerRepository) CountByImageMediaID(ctx context.Context, mediaID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM banners WHERE image_media_id = $1`
	if err := r.db.GetContext(ctx, &count, query, mediaID); err != nil {
		return 0, errtrace.Wrap(fmt.Errorf("failed to count banners by image media: %w", err))
	}
	return count, nil
}
