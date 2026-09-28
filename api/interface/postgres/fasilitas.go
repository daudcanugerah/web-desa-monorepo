package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"log"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/fasilitas"
	fasilitasUsecase "webdesa/api/usecase/fasilitas"
)

// FasilitasRepository implements usecase/fasilitas.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/fasilitas where it's USED, not here where it's implemented.
type FasilitasRepository struct {
	db *sqlx.DB
}

// NewFasilitasRepository creates a new FasilitasRepository instance
func NewFasilitasRepository(db *sqlx.DB) fasilitasUsecase.Repository {
	return &FasilitasRepository{db: db}
}

// Create creates a new fasilitas in the database
// Generates UUID, sets timestamps, and inserts the fasilitas record
// Validates: Requirements 11.2, 20.2, 20.3
func (r *FasilitasRepository) Create(ctx context.Context, f *fasilitas.Fasilitas) error {
	// Generate UUID if not provided
	if f.ID == "" {
		f.ID = uuid.New().String()
	}

	imagesMediaIDsJSON, err := marshalStringSlice(f.ImagesMediaIDs)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO fasilitas (id, name, category, latitude, longitude, description, images_media_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err = r.db.ExecContext(ctx, query,
		f.ID,
		f.Name,
		f.Category,
		f.Latitude,
		f.Longitude,
		f.Description,
		imagesMediaIDsJSON,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create fasilitas: %w", err))
	}

	return nil
}

// FindByID retrieves a fasilitas by ID
// Returns error if fasilitas not found
// Validates: Requirements 11.3
func (r *FasilitasRepository) FindByID(ctx context.Context, id string) (*fasilitas.Fasilitas, error) {
	var f fasilitas.Fasilitas
	var imagesMediaIDsJSON []byte

	query := `
		SELECT fasilitas.id, fasilitas.name, fasilitas.category, fasilitas_categories.name AS category_name, fasilitas.latitude, fasilitas.longitude, fasilitas.description, fasilitas.images_media_ids, fasilitas.created_at, fasilitas.updated_at
		FROM fasilitas
		LEFT JOIN fasilitas_categories ON fasilitas.category = fasilitas_categories.id
		WHERE fasilitas.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID,
		&f.Name,
		&f.Category,
		&f.CategoryName,
		&f.Latitude,
		&f.Longitude,
		&f.Description,
		&imagesMediaIDsJSON,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("fasilitas not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find fasilitas by ID: %w", err))
	}

	if err := unmarshalStringSlice(imagesMediaIDsJSON, &f.ImagesMediaIDs); err != nil {
		return nil, err
	}

	return &f, nil
}

// List retrieves paginated fasilitas with optional bounding box filtering
// Returns fasilitas slice, total count, and error
// Uses latitude/longitude range queries for bbox filtering
// Validates: Requirements 10.3, 10.4, 11.1, 17.2, 20.5
func (r *FasilitasRepository) List(ctx context.Context, bbox *fasilitasUsecase.BoundingBox, query *string, category *string, offset, limit int) ([]*fasilitas.Fasilitas, int, error) {
	// Build count query with optional filters
	countQuery := `SELECT COUNT(*) FROM fasilitas WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if bbox != nil {
		countQuery += fmt.Sprintf(` AND fasilitas.latitude BETWEEN $%d AND $%d AND fasilitas.longitude BETWEEN $%d AND $%d`,
			argIndex, argIndex+1, argIndex+2, argIndex+3)
		countArgs = append(countArgs, bbox.MinLat, bbox.MaxLat, bbox.MinLon, bbox.MaxLon)
		argIndex += 4
	}

	if query != nil && *query != "" {
		searchPattern := "%" + *query + "%"
		countQuery += fmt.Sprintf(` AND fasilitas.name ILIKE $%d`, argIndex)
		countArgs = append(countArgs, searchPattern)
		argIndex++
	}

	if category != nil && *category != "" {
		countQuery += fmt.Sprintf(` AND fasilitas.category = $%d`, argIndex)
		countArgs = append(countArgs, *category)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count fasilitas: %w", err))
	}

	// Build list query with optional filters
	listQuery := `
		SELECT fasilitas.id, fasilitas.name, fasilitas.category, fasilitas_categories.name AS category_name, fasilitas.latitude, fasilitas.longitude, fasilitas.description, fasilitas.images_media_ids, fasilitas.created_at, fasilitas.updated_at
		FROM fasilitas
		LEFT JOIN fasilitas_categories ON fasilitas.category = fasilitas_categories.id
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex = 1

	if bbox != nil {
		listQuery += fmt.Sprintf(` AND fasilitas.latitude BETWEEN $%d AND $%d AND fasilitas.longitude BETWEEN $%d AND $%d`,
			argIndex, argIndex+1, argIndex+2, argIndex+3)
		listArgs = append(listArgs, bbox.MinLat, bbox.MaxLat, bbox.MinLon, bbox.MaxLon)
		argIndex += 4
	}

	if query != nil && *query != "" {
		searchPattern := "%" + *query + "%"
		listQuery += fmt.Sprintf(` AND fasilitas.name ILIKE $%d`, argIndex)
		listArgs = append(listArgs, searchPattern)
		argIndex++
	}

	if category != nil && *category != "" {
		listQuery += fmt.Sprintf(` AND fasilitas.category = $%d`, argIndex)
		listArgs = append(listArgs, *category)
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY fasilitas.created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated fasilitas
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list fasilitas: %w", err))
	}
	defer rows.Close()

	var fasilitases []*fasilitas.Fasilitas
	for rows.Next() {
		var f fasilitas.Fasilitas
		var imagesMediaIDsJSON []byte

		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Category,
			&f.CategoryName,
			&f.Latitude,
			&f.Longitude,
			&f.Description,
			&imagesMediaIDsJSON,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errtrace.Wrap(fmt.Errorf("failed to scan fasilitas: %w", err))
		}

		if err := unmarshalStringSlice(imagesMediaIDsJSON, &f.ImagesMediaIDs); err != nil {
			return nil, 0, err
		}

		fasilitases = append(fasilitases, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to iterate fasilitas: %w", err))
	}

	return fasilitases, total, nil
}

// Update updates an existing fasilitas
// Sets updated_at timestamp automatically
// Validates: Requirements 11.4, 20.3
func (r *FasilitasRepository) Update(ctx context.Context, f *fasilitas.Fasilitas) error {
	imagesMediaIDsJSON, err := marshalStringSlice(f.ImagesMediaIDs)
	if err != nil {
		return err
	}

	query := `
		UPDATE fasilitas
		SET name = $1, category = $2, latitude = $3, longitude = $4, description = $5, images_media_ids = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		f.Name,
		f.Category,
		f.Latitude,
		f.Longitude,
		f.Description,
		imagesMediaIDsJSON,
		f.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update fasilitas: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("fasilitas not found: %s", f.ID))
	}

	return nil
}

// Delete removes a fasilitas
// Validates: Requirements 11.5
func (r *FasilitasRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM fasilitas WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete fasilitas: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("fasilitas not found: %s", id))
	}

 	return nil
}


// OnMediaDeleted scrubs a deleted gallery media id out of every
// Fasilitas row's images_media_ids array (Task 5.1 on-delete sweep).
// Errors are returned for logging; the gallery service treats the
// sweep as best-effort.
func (r *FasilitasRepository) OnMediaDeleted(ctx context.Context, mediaID string) {
	if mediaID == "" {
		return
	}
	query := `UPDATE fasilitas SET images_media_ids = images_media_ids - $1 WHERE images_media_ids @> $1`
	encoded, err := marshalStringSlice([]string{mediaID})
	if err != nil {
		log.Printf("umkm: failed to encode media id %s: %v", mediaID, err)
		return
	}
	if _, err := r.db.ExecContext(ctx, query, encoded); err != nil {
		log.Printf("fasilitas: failed to sweep media id %s: %v", mediaID, err)
	}
}
