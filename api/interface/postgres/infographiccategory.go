package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/infographiccategory"
	infographiccategoryUsecase "webdesa/api/usecase/infographiccategory"
)

// InfographicCategoryRepository implements usecase/infographiccategory.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/infographiccategory where it's USED, not here where it's implemented.
type InfographicCategoryRepository struct {
	db *sqlx.DB
}

// NewInfographicCategoryRepository creates a new InfographicCategoryRepository instance.
func NewInfographicCategoryRepository(db *sqlx.DB) infographiccategoryUsecase.Repository {
	return &InfographicCategoryRepository{db: db}
}

// Create persists a new Infographic category.
func (r *InfographicCategoryRepository) Create(ctx context.Context, c *infographiccategory.Category) error {
	query := `
		INSERT INTO infographic_categories (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Name, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create infographic category: %w", err))
	}
	return nil
}

// Delete removes an Infographic category by ID.
func (r *InfographicCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM infographic_categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete infographic category: %w", err))
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

// FindByID retrieves an Infographic category by ID.
// Returns sql.ErrNoRows if the category is not found.
func (r *InfographicCategoryRepository) FindByID(ctx context.Context, id string) (*infographiccategory.Category, error) {
	var c infographiccategory.Category
	query := `SELECT id, name, created_at, updated_at FROM infographic_categories WHERE id = $1`
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNotFound
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find infographic category: %w", err))
	}
	return &c, nil
}

// List returns a paginated, optionally name-filtered list of categories
// sorted by name ascending. query, when non-nil and non-empty, performs
// a case-insensitive substring match (ILIKE %query%) suitable for autocomplete.
func (r *InfographicCategoryRepository) List(ctx context.Context, query *string, offset, limit int) ([]*infographiccategory.Category, int, error) {
	countQuery := `SELECT COUNT(*) FROM infographic_categories WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if query != nil && strings.TrimSpace(*query) != "" {
		countQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		countArgs = append(countArgs, "%"+*query+"%")
		argIndex++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count infographic categories: %w", err))
	}

	listQuery := `SELECT id, name, created_at, updated_at FROM infographic_categories WHERE 1=1`
	var listArgs []interface{}
	argIndex = 1

	if query != nil && strings.TrimSpace(*query) != "" {
		listQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		listArgs = append(listArgs, "%"+*query+"%")
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	var categories []*infographiccategory.Category
	if err := r.db.SelectContext(ctx, &categories, listQuery, listArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list infographic categories: %w", err))
	}

	return categories, total, nil
}

// CountByCategoryIDs returns a map of category ID -> count of infographic referencing it.
// Categories with zero references are omitted from the map.
func (r *InfographicCategoryRepository) CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error) {
	if len(ids) == 0 {
		return map[string]int{}, nil
	}

	query := `
		SELECT category, COUNT(*) AS usage_count
		FROM infographic
		WHERE category = ANY($1)
		GROUP BY category
	`

	type row struct {
		Category   *string `db:"category"`
		UsageCount int     `db:"usage_count"`
	}

	rows := []row{}
	if err := r.db.SelectContext(ctx, &rows, query, ids); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to count infographic by category: %w", err))
	}

	result := make(map[string]int, len(rows))
	for _, r := range rows {
		if r.Category != nil {
			result[*r.Category] = r.UsageCount
		}
	}
	return result, nil
}
