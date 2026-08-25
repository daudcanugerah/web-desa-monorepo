package fasilitas

import (
	"context"

	"webdesa/api/domain/fasilitas"
)

// Repository defines the persistence interface for Fasilitas (facility) management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/fasilitas_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new Fasilitas in the database
	Create(ctx context.Context, f *fasilitas.Fasilitas) error

	// FindByID retrieves a Fasilitas by ID
	FindByID(ctx context.Context, id string) (*fasilitas.Fasilitas, error)

	// List retrieves paginated Fasilitas with optional bounding box, query, and category filtering
	// Returns Fasilitas slice, total count, and error
	// Filters:
	// - bbox: bounding box filter [minLon, minLat, maxLon, maxLat]
	//   If nil, no bounding box filter is applied
	// - query: optional ILIKE search across name
	// - category: optional UUID FK to fasilitas_categories.id
	List(ctx context.Context, bbox *BoundingBox, query *string, category *string, offset, limit int) ([]*fasilitas.Fasilitas, int, error)

	// Update updates an existing Fasilitas
	Update(ctx context.Context, f *fasilitas.Fasilitas) error

	// Delete removes a Fasilitas
	Delete(ctx context.Context, id string) error
}

// BoundingBox represents a geographic bounding box for filtering facilities
// Format: [minLon, minLat, maxLon, maxLat]
type BoundingBox struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}
