package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/profile"
	profileUsecase "webdesa/api/usecase/profile"
)

// ProfileRepository implements usecase/profile.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/profile where it's USED, not here where it's implemented.
type ProfileRepository struct {
	db *sqlx.DB
}

// NewProfileRepository creates a new ProfileRepository instance
func NewProfileRepository(db *sqlx.DB) profileUsecase.Repository {
	return &ProfileRepository{db: db}
}

// Create creates a new profile in the database
func (r *ProfileRepository) Create(ctx context.Context, p *profile.Profile) error {
	// Generate UUID if not provided
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	query := `
		INSERT INTO profile (id, content, section_name, section_endpoint, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.Content,
		p.SectionName,
		p.SectionEndpoint,
		p.State,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create profile: %w", err))
	}

	return nil
}

// FindByID retrieves a profile by ID
func (r *ProfileRepository) FindByID(ctx context.Context, id string) (*profile.Profile, error) {
	var p profile.Profile

	query := `
		SELECT id, content, section_name, section_endpoint, state, created_at, updated_at
		FROM profile
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("profile not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find profile by ID: %w", err))
	}

	return &p, nil
}

// List retrieves paginated profiles with optional filtering
func (r *ProfileRepository) List(ctx context.Context, sectionName *string, state *bool, query *string, offset, limit int) ([]*profile.Profile, int, error) {
	// Build count query with filters
	countQuery := `SELECT COUNT(*) FROM profile WHERE 1=1`
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
		countQuery += fmt.Sprintf(` AND (section_name ILIKE $%d OR content ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count profiles: %w", err))
	}

	// Build data query with filters
	dataQuery := `
		SELECT id, content, section_name, section_endpoint, state, created_at, updated_at
		FROM profile
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
		dataQuery += fmt.Sprintf(` AND (section_name ILIKE $%d OR content ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		dataArgs = append(dataArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	dataQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)
	dataArgs = append(dataArgs, limit, offset)

	// Query profiles
	var profiles []*profile.Profile
	err = r.db.SelectContext(ctx, &profiles, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list profiles: %w", err))
	}

	return profiles, total, nil
}

// Update updates an existing profile
func (r *ProfileRepository) Update(ctx context.Context, p *profile.Profile) error {
	query := `
		UPDATE profile
		SET content = $1, section_name = $2, section_endpoint = $3, state = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		p.Content,
		p.SectionName,
		p.SectionEndpoint,
		p.State,
		p.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update profile: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("profile not found: %s", p.ID))
	}

	return nil
}

// Delete removes a profile
func (r *ProfileRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM profile WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete profile: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("profile not found: %s", id))
	}

	return nil
}

// GetSectionNames retrieves all unique section names
func (r *ProfileRepository) GetSectionNames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT section_name FROM profile ORDER BY section_name`

	var names []string
	err := r.db.SelectContext(ctx, &names, query)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to get section names: %w", err))
	}

	return names, nil
}
