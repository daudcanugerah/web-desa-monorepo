package umkm

import (
	"context"

	"webdesa/api/domain/umkm"
)

// Repository defines the persistence interface for UMKM (small business) management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/umkm_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new UMKM in the database
	Create(ctx context.Context, u *umkm.UMKM) error

	// FindByID retrieves a UMKM by ID
	FindByID(ctx context.Context, id string) (*umkm.UMKM, error)

	// List retrieves paginated UMKM with optional filtering
	// Returns UMKM slice, total count, and error
	// Filters:
	// - query: search in name, description, owner (ILIKE)
	// - category: filter by category (exact match)
	List(ctx context.Context, query *string, category *string, offset, limit int) ([]*umkm.UMKM, int, error)

	// Update updates an existing UMKM
	Update(ctx context.Context, u *umkm.UMKM) error

	// Delete removes a UMKM
	Delete(ctx context.Context, id string) error

	// OnMediaDeleted is the gallery deletion-listener hook (Task 5.1).
	// Called by the gallery service after a media row is removed so
	// dangling UUIDs can be swept out of images_media_ids. Best-effort:
	// failures are logged but never propagated.
	OnMediaDeleted(ctx context.Context, mediaID string)
}
