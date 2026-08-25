package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"webdesa/api/domain/bannercategory"

	"github.com/jmoiron/sqlx"
	"braces.dev/errtrace"
)

type BannerCategoryRepository struct {
	db *sqlx.DB
}

func NewBannerCategoryRepository(db *sqlx.DB) *BannerCategoryRepository {
	return &BannerCategoryRepository{db: db}
}

func (r *BannerCategoryRepository) Create(ctx context.Context, c *bannercategory.Category) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO banner_categories (id, name, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())`,
		c.ID, c.Name,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create banner category: %w", err))
	}
	return nil
}

func (r *BannerCategoryRepository) FindByID(ctx context.Context, id string) (*bannercategory.Category, error) {
	var c bannercategory.Category
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at FROM banner_categories WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errtrace.Wrap(fmt.Errorf("banner category not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find banner category: %w", err))
	}
	return &c, nil
}

func (r *BannerCategoryRepository) List(ctx context.Context, query string, offset, limit int) ([]bannercategory.Category, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM banner_categories WHERE 1=1`+listWhereClause(query),
		listArgs(query)...,
	).Scan(&total); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count banner categories: %w", err))
	}

	var out []bannercategory.Category
	if err := r.db.SelectContext(ctx, &out,
		`SELECT id, name, created_at, updated_at FROM banner_categories WHERE 1=1`+
			listWhereClause(query)+
			` ORDER BY created_at ASC LIMIT $`+fmt.Sprintf("%d", len(listArgs(query))+1)+
			` OFFSET $`+fmt.Sprintf("%d", len(listArgs(query))+2),
		append(listArgs(query), limit, offset)...,
	); err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list banner categories: %w", err))
	}
	return out, total, nil
}

func (r *BannerCategoryRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM banner_categories WHERE id = $1`, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete banner category: %w", err))
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errtrace.Wrap(fmt.Errorf("banner category not found: %s", id))
	}
	return nil
}

func (r *BannerCategoryRepository) CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error) {
	out := map[string]int{}
	if len(ids) == 0 {
		return out, nil
	}
	q, args, err := sqlx.In(`SELECT category, COUNT(*) FROM banners WHERE category IN (?) GROUP BY category`, ids)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	q = r.db.Rebind(q)
	rows, err := r.db.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cat sql.NullString
		var count int
		if err := rows.Scan(&cat, &count); err != nil {
			return nil, errtrace.Wrap(err)
		}
		if cat.Valid {
			out[cat.String] = count
		}
	}
	return out, nil
}

// listWhereClause + listArgs are helpers for optional q search
func listWhereClause(query string) string {
	if query == "" {
		return ""
	}
	return " AND name ILIKE '%' || $1 || '%'"
}
func listArgs(query string) []any {
	if query == "" {
		return nil
	}
	return []any{query}
}