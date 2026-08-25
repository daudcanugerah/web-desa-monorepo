package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/infographic"
	infographicUsecase "webdesa/api/usecase/infographic"
)

// InfographicRepository implements usecase/infographic.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/infographic where it's USED, not here where it's implemented.
type InfographicRepository struct {
	db *sqlx.DB
}

// NewInfographicRepository creates a new InfographicRepository instance
func NewInfographicRepository(db *sqlx.DB) infographicUsecase.Repository {
	return &InfographicRepository{db: db}
}

// Create creates a new infographic in the database
func (r *InfographicRepository) Create(ctx context.Context, i *infographic.Infographic) error {
	// Generate UUID if not provided
	if i.ID == "" {
		i.ID = uuid.New().String()
	}

	query := `
		INSERT INTO infographic (id, component_id, component_type, section_name, section_endpoint, category, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		i.ID,
		i.ComponentID,
		string(i.ComponentType),
		i.SectionName,
		i.SectionEndpoint,
		i.Category,
		i.State,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create infographic: %w", err))
	}

	return nil
}

// FindByID retrieves an infographic by ID
func (r *InfographicRepository) FindByID(ctx context.Context, id string) (*infographic.Infographic, error) {
	var i infographic.Infographic

	query := `
		SELECT infographic.id, infographic.component_id, infographic.component_type,
		       infographic.section_name, infographic.section_endpoint, infographic.category,
		       infographic.state, infographic.created_at, infographic.updated_at,
		       infographic_categories.name AS category_name
		FROM infographic
		LEFT JOIN infographic_categories ON infographic.category = infographic_categories.id
		WHERE infographic.id = $1
	`

	err := r.db.GetContext(ctx, &i, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("infographic not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find infographic by ID: %w", err))
	}

	return &i, nil
}

// List retrieves paginated infographics with optional filtering
func (r *InfographicRepository) List(ctx context.Context, sectionName *string, state *bool, query *string, category *string, offset, limit int) ([]*infographic.Infographic, int, error) {
	// Build count query with filters
	countQuery := `SELECT COUNT(*) FROM infographic WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if sectionName != nil && *sectionName != "" {
		countQuery += fmt.Sprintf(` AND section_name = $%d`, argIndex)
		countArgs = append(countArgs, *sectionName)
		argIndex++
	}

	if state != nil {
		countQuery += fmt.Sprintf(` AND state = $%d`, argIndex)
		countArgs = append(countArgs, *state)
		argIndex++
	}

	if query != nil && *query != "" {
		countQuery += fmt.Sprintf(` AND (section_name ILIKE $%d OR component_id::text ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
		argIndex += 2
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
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count infographics: %w", err))
	}

	// Build data query with filters
	dataQuery := `
		SELECT infographic.id, infographic.component_id, infographic.component_type,
		       infographic.section_name, infographic.section_endpoint, infographic.category,
		       infographic.state, infographic.created_at, infographic.updated_at,
		       infographic_categories.name AS category_name
		FROM infographic
		LEFT JOIN infographic_categories ON infographic.category = infographic_categories.id
		WHERE 1=1
	`
	var dataArgs []interface{}
	argIndex = 1

	if sectionName != nil && *sectionName != "" {
		dataQuery += fmt.Sprintf(` AND section_name = $%d`, argIndex)
		dataArgs = append(dataArgs, *sectionName)
		argIndex++
	}

	if state != nil {
		dataQuery += fmt.Sprintf(` AND state = $%d`, argIndex)
		dataArgs = append(dataArgs, *state)
		argIndex++
	}

	if query != nil && *query != "" {
		dataQuery += fmt.Sprintf(` AND (infographic.section_name ILIKE $%d OR infographic.component_id::text ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		dataArgs = append(dataArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	if category != nil && *category != "" {
		dataQuery += fmt.Sprintf(` AND infographic.category = $%d`, argIndex)
		dataArgs = append(dataArgs, *category)
		argIndex++
	}

	dataQuery += ` ORDER BY infographic.created_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	dataArgs = append(dataArgs, limit, offset)

	// Query infographics
	var infographics []*infographic.Infographic
	err = r.db.SelectContext(ctx, &infographics, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list infographics: %w", err))
	}

	return infographics, total, nil
}

// Update updates an existing infographic
func (r *InfographicRepository) Update(ctx context.Context, i *infographic.Infographic) error {
	query := `
		UPDATE infographic
		SET component_id = $1, component_type = $2, section_name = $3, section_endpoint = $4, category = $5, state = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		i.ComponentID,
		string(i.ComponentType),
		i.SectionName,
		i.SectionEndpoint,
		i.Category,
		i.State,
		i.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update infographic: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("infographic not found: %s", i.ID))
	}

	return nil
}

// Delete removes an infographic
func (r *InfographicRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM infographic WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete infographic: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("infographic not found: %s", id))
	}

	return nil
}

// GetSectionNames retrieves all unique section names
func (r *InfographicRepository) GetSectionNames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT section_name FROM infographic ORDER BY section_name`

	var names []string
	err := r.db.SelectContext(ctx, &names, query)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get section names: %w", err))
	}

	return names, nil
}
