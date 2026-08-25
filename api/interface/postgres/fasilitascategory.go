package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/fasilitascategory"
	fasilitascategoryUsecase "webdesa/api/usecase/fasilitascategory"
)

// FasilitasCategoryRepository implements usecase/fasilitascategory.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/fasilitascategory where it's USED, not here where it's implemented.
type FasilitasCategoryRepository struct {
	db *sqlx.DB
}

// NewFasilitasCategoryRepository creates a new FasilitasCategoryRepository instance.
func NewFasilitasCategoryRepository(db *sqlx.DB) fasilitascategoryUsecase.Repository {
	return &FasilitasCategoryRepository{db: db}
}

// Create persists a new Fasilitas category.
func (r *FasilitasCategoryRepository) Create(ctx context.Context, c *fasilitascategory.Category) error {
	query := `
		INSERT INTO fasilitas_categories (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Name, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create fasilitas category: %w", err))
	}
	return nil
}

// Delete removes a Fasilitas category by ID.
func (r *FasilitasCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM fasilitas_categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete fasilitas category: %w", err))
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}
	if rows == 0 {
		return errtrace.Wrap(sql.ErrNoRows)
	}
	return nil
}

// FindByID retrieves a Fasilitas category by ID.
// Returns sql.ErrNoRows if the category is not found.
func (r *FasilitasCategoryRepository) FindByID(ctx context.Context, id string) (*fasilitascategory.Category, error) {
	var c fasilitascategory.Category
	query := `SELECT id, name, created_at, updated_at FROM fasilitas_categories WHERE id = $1`
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNotFound
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find fasilitas category: %w", err))
	}
	return &c, nil
}

// List returns a paginated, optionally name-filtered list of categories
// sorted by name ascending. query, when non-nil and non-empty, performs
// a case-insensitive substring match (ILIKE %query%) suitable for autocomplete.
func (r *FasilitasCategoryRepository) List(ctx context.Context, query *string, offset, limit int) ([]*fasilitascategory.Category, int, error) {
	countQuery := `SELECT COUNT(*) FROM fasilitas_categories WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if query != nil && strings.TrimSpace(*query) != "" {
		countQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		countArgs = append(countArgs, "%"+*query+"%")
		argIndex++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count fasilitas categories: %w", err))
	}

	listQuery := `SELECT id, name, created_at, updated_at FROM fasilitas_categories WHERE 1=1`
	var listArgs []interface{}
	argIndex = 1

	if query != nil && strings.TrimSpace(*query) != "" {
		listQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		listArgs = append(listArgs, "%"+*query+"%")
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	var categories []*fasilitascategory.Category
	if err := r.db.SelectContext(ctx, &categories, listQuery, listArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list fasilitas categories: %w", err))
	}

	return categories, total, nil
}

// CountByCategoryIDs returns a map of category ID -> count of fasilitas referencing it.
// Categories with zero references are omitted from the map.
func (r *FasilitasCategoryRepository) CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error) {
	if len(ids) == 0 {
		return map[string]int{}, nil
	}

	query := `
		SELECT category, COUNT(*) AS usage_count
		FROM fasilitas
		WHERE category = ANY($1)
		GROUP BY category
	`

	type row struct {
		Category   *string `db:"category"`
		UsageCount int     `db:"usage_count"`
	}

	rows := []row{}
	if err := r.db.SelectContext(ctx, &rows, query, ids); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to count fasilitas by category: %w", err))
	}

	result := make(map[string]int, len(rows))
	for _, r := range rows {
		if r.Category != nil {
			result[*r.Category] = r.UsageCount
		}
	}
	return result, nil
}
