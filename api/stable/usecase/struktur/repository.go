package struktur

import (
	"context"

	"webdesa/api/domain/struktur"
)

// Repository defines the persistence interface for struktur (organizational structure) management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/struktur_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new struktur member in the database
	Create(ctx context.Context, s *struktur.Struktur) error

	// FindByID retrieves a struktur member by ID
	FindByID(ctx context.Context, id string) (*struktur.Struktur, error)

	// List retrieves paginated struktur members with optional filtering
	// Returns struktur slice, total count, and error
	// Filters:
	// - query: search in name or position (ILIKE)
	List(ctx context.Context, query *string, offset, limit int) ([]*struktur.Struktur, int, error)

	// Update updates an existing struktur member
	Update(ctx context.Context, s *struktur.Struktur) error

	// Delete removes a struktur member
	Delete(ctx context.Context, id string) error
}
