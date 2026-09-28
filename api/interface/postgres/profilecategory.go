package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/profilecategory"
	profilecategoryUsecase "webdesa/api/usecase/profilecategory"
)

// ProfileCategoryRepository implements usecase/profilecategory.Repository.
// Dependency Rule: interface/postgres → usecase → domain.
type ProfileCategoryRepository struct {
	db *sqlx.DB
}

// NewProfileCategoryRepository creates a new ProfileCategoryRepository instance.
func NewProfileCategoryRepository(db *sqlx.DB) profilecategoryUsecase.Repository {
	return &ProfileCategoryRepository{db: db}
}

// Create persists a new Profile category.
func (r *ProfileCategoryRepository) Create(ctx context.Context, c *profilecategory.Category) error {
	query := `
		INSERT INTO profile_categories (id, name, sort_order, created_at, updated_at)
		VALUES ($1, $2, (SELECT COALESCE(MAX(sort_order), 0) + 10 FROM profile_categories), $3, $4)
		RETURNING sort_order
	`
	if err := r.db.QueryRowxContext(ctx, query, c.ID, c.Name, c.CreatedAt, c.UpdatedAt).Scan(&c.SortOrder); err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create profile category: %w", err))
	}
	return nil
}

// Delete removes a Profile category by ID.
func (r *ProfileCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM profile_categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete profile category: %w", err))
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

// FindByID retrieves a Profile category by ID.
// Returns sql.ErrNoRows if the category is not found.
func (r *ProfileCategoryRepository) FindByID(ctx context.Context, id string) (*profilecategory.Category, error) {
	var c profilecategory.Category
	query := `SELECT id, name, sort_order, created_at, updated_at FROM profile_categories WHERE id = $1`
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errNotFound
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find profile category: %w", err))
	}
	return &c, nil
}

// List returns a paginated, optionally name-filtered list of categories
// sorted by name ascending.
func (r *ProfileCategoryRepository) List(ctx context.Context, query *string, offset, limit int) ([]*profilecategory.Category, int, error) {
	countQuery := `SELECT COUNT(*) FROM profile_categories WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if query != nil && strings.TrimSpace(*query) != "" {
		countQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		countArgs = append(countArgs, "%"+*query+"%")
		argIndex++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count profile categories: %w", err))
	}

	listQuery := `SELECT id, name, sort_order, created_at, updated_at FROM profile_categories WHERE 1=1`
	var listArgs []interface{}
	argIndex = 1

	if query != nil && strings.TrimSpace(*query) != "" {
		listQuery += fmt.Sprintf(` AND name ILIKE $%d`, argIndex)
		listArgs = append(listArgs, "%"+*query+"%")
		argIndex++
	}

	listQuery += fmt.Sprintf(` ORDER BY sort_order ASC, name ASC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	var categories []*profilecategory.Category
	if err := r.db.SelectContext(ctx, &categories, listQuery, listArgs...); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list profile categories: %w", err))
	}

	return categories, total, nil
}

// CountByCategoryIDs returns a map of category ID -> count of profile rows
// referencing it. Categories with zero references are omitted.
func (r *ProfileCategoryRepository) CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error) {
	if len(ids) == 0 {
		return map[string]int{}, nil
	}

	query := `
		SELECT category, COUNT(*) AS usage_count
		FROM profile
		WHERE category = ANY($1)
		GROUP BY category
	`

	type row struct {
		Category   *string `db:"category"`
		UsageCount int     `db:"usage_count"`
	}

	rows := []row{}
	if err := r.db.SelectContext(ctx, &rows, query, ids); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to count profile by category: %w", err))
	}

	result := make(map[string]int, len(rows))
	for _, r := range rows {
		if r.Category != nil {
			result[*r.Category] = r.UsageCount
		}
	}
	return result, nil
}
