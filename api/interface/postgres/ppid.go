package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/ppid"
	ppidUsecase "webdesa/api/usecase/ppid"
)

// PPIDRepository implements usecase/ppid.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/ppid where it's USED, not here where it's implemented.
type PPIDRepository struct {
	db *sqlx.DB
}

// NewPPIDRepository creates a new PPIDRepository instance
func NewPPIDRepository(db *sqlx.DB) ppidUsecase.Repository {
	return &PPIDRepository{db: db}
}

// Create creates a new PPID document in the database
// Generates UUID, sets timestamps, and inserts the PPID record
// Validates: Requirements 12.4, 20.2, 20.3
func (r *PPIDRepository) Create(ctx context.Context, p *ppid.PPID) error {
	// Generate UUID if not provided
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	query := `
		INSERT INTO ppid (id, title, category, document_media_id, thumbnail_media_id, description, publication_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.Title,
		p.Category,
		p.DocumentMediaID,
		p.ThumbnailMediaID,
		p.Description,
		p.PublicationAt,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create ppid: %w", err))
	}

	return nil
}

// FindByID retrieves a PPID document by ID
// Returns error if PPID not found
// Validates: Requirements 12.5
func (r *PPIDRepository) FindByID(ctx context.Context, id string) (*ppid.PPID, error) {
	var p ppid.PPID

	query := `
		SELECT ppid.id, ppid.title, ppid.category,
		       ppid.document_media_id, ppid.thumbnail_media_id,
		       ppid.description, ppid.publication_at, ppid.created_at, ppid.updated_at,
		       ppid_categories.name AS category_name
		FROM ppid
		LEFT JOIN ppid_categories ON ppid.category = ppid_categories.id
		WHERE ppid.id = $1
	`

	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("ppid not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find ppid by ID: %w", err))
	}

	return &p, nil
}

// List retrieves paginated PPID documents with optional filtering
// Returns PPID slice, total count, and error
// Uses ILIKE for search queries
// Validates: Requirements 12.1, 12.2, 12.3, 17.2, 20.5
func (r *PPIDRepository) List(ctx context.Context, category *string, query *string, offset, limit int) ([]*ppid.PPID, int, error) {
	// Build count query with filters
	countQuery := `SELECT COUNT(*) FROM ppid WHERE 1=1`
	var countArgs []interface{}
	argIndex := 1

	if category != nil && *category != "" {
		countQuery += fmt.Sprintf(` AND category = $%d`, argIndex)
		countArgs = append(countArgs, *category)
		argIndex++
	}

	if query != nil && *query != "" {
		countQuery += fmt.Sprintf(` AND (title ILIKE $%d OR description ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count ppid: %w", err))
	}

	// Build list query with filters
	listQuery := `
		SELECT ppid.id, ppid.title, ppid.category,
		       ppid.document_media_id, ppid.thumbnail_media_id,
		       ppid.description, ppid.publication_at, ppid.created_at, ppid.updated_at,
		       ppid_categories.name AS category_name
		FROM ppid
		LEFT JOIN ppid_categories ON ppid.category = ppid_categories.id
		WHERE 1=1
	`
	var listArgs []interface{}
	argIndex = 1

	if category != nil && *category != "" {
		listQuery += fmt.Sprintf(` AND ppid.category = $%d`, argIndex)
		listArgs = append(listArgs, *category)
		argIndex++
	}

	if query != nil && *query != "" {
		listQuery += fmt.Sprintf(` AND (ppid.title ILIKE $%d OR ppid.description ILIKE $%d)`, argIndex, argIndex+1)
		searchPattern := "%" + *query + "%"
		listArgs = append(listArgs, searchPattern, searchPattern)
		argIndex += 2
	}

	listQuery += fmt.Sprintf(` ORDER BY ppid.created_at DESC LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	listArgs = append(listArgs, limit, offset)

	// Get paginated PPID documents
	var ppids []*ppid.PPID
	err = r.db.SelectContext(ctx, &ppids, listQuery, listArgs...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list ppid: %w", err))
	}

	return ppids, total, nil
}

// Update updates an existing PPID document
// Sets updated_at timestamp automatically
// Validates: Requirements 12.6, 20.3
func (r *PPIDRepository) Update(ctx context.Context, p *ppid.PPID) error {
	query := `
		UPDATE ppid
		SET title = $1, category = $2, document_media_id = $3, thumbnail_media_id = $4, description = $5, publication_at = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		p.Title,
		p.Category,
		p.DocumentMediaID,
		p.ThumbnailMediaID,
		p.Description,
		p.PublicationAt,
		p.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update ppid: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("ppid not found: %s", p.ID))
	}

	return nil
}

// Delete removes a PPID document
// Validates: Requirements 12.7
func (r *PPIDRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ppid WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete ppid: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("ppid not found: %s", id))
	}

	return nil
}

// CreateRequest creates a new PPID document request
// Generates UUID, sets timestamp, and inserts the request record
// Validates: Requirements 12.8, 20.2, 20.3
func (r *PPIDRepository) CreateRequest(ctx context.Context, req *ppid.PPIDRequest) error {
	// Generate UUID if not provided
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	query := `
		INSERT INTO ppid_requests (id, ppid_id, requester_name, requester_email, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.PPIDId,
		req.RequesterName,
		req.RequesterEmail,
		req.Purpose,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create ppid request: %w", err))
	}

	return nil
}

// ListRequests retrieves paginated PPID requests with optional status filtering
// Returns PPIDRequest slice, total count, and error
// Validates: Requirements 12.9, 17.2, 20.5
func (r *PPIDRepository) ListRequests(ctx context.Context, statuses []string, offset, limit int) ([]*ppid.PPIDRequest, int, error) {
	var args []interface{}
	argIndex := 1

	whereClause := ""
	if len(statuses) > 0 {
		placeholders := make([]string, len(statuses))
		for i, status := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		whereClause = fmt.Sprintf(" WHERE status IN (%s)", strings.Join(placeholders, ", "))
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM ppid_requests" + whereClause
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count ppid requests: %w", err))
	}

	var requests []*ppid.PPIDRequest
	listQuery := fmt.Sprintf(`
		SELECT id, ppid_id, requester_name, requester_email, notes, status, approved_at, approved_by, revoked_at, revoked_by, created_at
		FROM ppid_requests
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	err = r.db.SelectContext(ctx, &requests, listQuery, args...)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list ppid requests: %w", err))
	}

	return requests, total, nil
}

// UpdateRequest updates an existing PPID request
// Sets updated_at timestamp automatically
func (r *PPIDRepository) UpdateRequest(ctx context.Context, req *ppid.PPIDRequest) error {
	query := `
		UPDATE ppid_requests
		SET status = $1, approved_at = $2, approved_by = $3, revoked_at = $4, revoked_by = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		req.Status,
		req.ApprovedAt,
		req.ApprovedBy,
		req.RevokedAt,
		req.RevokedBy,
		req.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update ppid request: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("ppid request not found: %s", req.ID))
	}

	return nil
}

// FindRequestByID retrieves a PPID request by ID
// Returns error if request not found
func (r *PPIDRepository) FindRequestByID(ctx context.Context, id string) (*ppid.PPIDRequest, error) {
	var req ppid.PPIDRequest

	query := `
		SELECT id, ppid_id, requester_name, requester_email, notes, status, approved_at, approved_by, revoked_at, revoked_by, created_at
		FROM ppid_requests
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &req, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("ppid request not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find ppid request by ID: %w", err))
	}

	return &req, nil
}
