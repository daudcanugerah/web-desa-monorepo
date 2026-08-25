package berita

import (
	"context"
	"time"

	"webdesa/api/domain/berita"
)

// Repository defines the persistence interface for berita (news) management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/berita_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new berita in the database
	Create(ctx context.Context, b *berita.Berita) error

	// FindByID retrieves a berita by ID
	FindByID(ctx context.Context, id string) (*berita.Berita, error)

	// List retrieves paginated berita with optional filtering
	// Returns berita slice, total count, and error
	// Filters:
	// - query: search in title or content (ILIKE)
	// - category: filter by category (exact match)
	// - since: filter articles created after this date
	// - until: filter articles created before this date
	// - sort: column to sort by ("created_at" or "title"); defaults to "created_at"
	// - order: "asc" or "desc"; defaults to "desc"
	List(ctx context.Context, query *string, category *string, since *time.Time, until *time.Time, sort, order string, offset, limit int) ([]*berita.Berita, int, error)

	// Update updates an existing berita
	Update(ctx context.Context, b *berita.Berita) error

	// Delete removes a berita
	Delete(ctx context.Context, id string) error
}
