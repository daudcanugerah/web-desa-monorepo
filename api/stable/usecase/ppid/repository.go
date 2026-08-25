package ppid

import (
	"context"

	"webdesa/api/domain/ppid"
)

// Repository defines the persistence interface for PPID (public information disclosure) management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/ppid_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new PPID document in the database
	Create(ctx context.Context, p *ppid.PPID) error

	// FindByID retrieves a PPID document by ID
	FindByID(ctx context.Context, id string) (*ppid.PPID, error)

	// List retrieves paginated PPID documents with optional filtering
	// Returns PPID slice, total count, and error
	// Filters:
	// - category: filter by document category
	// - query: search in title (ILIKE)
	List(ctx context.Context, category *string, query *string, offset, limit int) ([]*ppid.PPID, int, error)

	// Update updates an existing PPID document
	Update(ctx context.Context, p *ppid.PPID) error

	// Delete removes a PPID document
	Delete(ctx context.Context, id string) error

	// CreateRequest creates a new PPID document request
	CreateRequest(ctx context.Context, r *ppid.PPIDRequest) error

	// ListRequests retrieves paginated PPID requests
	// Returns PPIDRequest slice, total count, and error
	// Filters:
	// - statuses: filter by request status (e.g., []string{"pending", "approved"})
	ListRequests(ctx context.Context, statuses []string, offset, limit int) ([]*ppid.PPIDRequest, int, error)

	// UpdateRequest updates an existing PPID request
	UpdateRequest(ctx context.Context, r *ppid.PPIDRequest) error

	// FindRequestByID retrieves a PPID request by ID
	FindRequestByID(ctx context.Context, id string) (*ppid.PPIDRequest, error)
}
