package banner

import (
	"context"

	"webdesa/api/domain/banner"
)

// Repository defines the persistence interface for banner management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/banner_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new banner in the database
	Create(ctx context.Context, b *banner.Banner) error

	// FindByID retrieves a banner by ID
	FindByID(ctx context.Context, id string) (*banner.Banner, error)

	// List retrieves paginated banners with optional status, query, and category filtering
	// Returns banners slice, total count, and error
	// Filters:
	// - status: optional "active" or "inactive"
	// - query: optional ILIKE search across title and description
	// - category: optional UUID FK to banner_categories.id
	List(ctx context.Context, status *string, query *string, category *string, offset, limit int) ([]*banner.Banner, int, error)

	// Update updates an existing banner
	Update(ctx context.Context, b *banner.Banner) error

	// Delete removes a banner
	Delete(ctx context.Context, id string) error

	// CountActivebanners returns the count of banners with status "active"
	CountActiveBanners(ctx context.Context) (int, error)
}
