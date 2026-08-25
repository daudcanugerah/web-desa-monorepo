package infographic

import (
	"context"

	"webdesa/api/domain/infographic"
)

// Repository defines the persistence interface for infographic management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/infographic_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new infographic in the database
	Create(ctx context.Context, i *infographic.Infographic) error

	// FindByID retrieves an infographic by ID
	FindByID(ctx context.Context, id string) (*infographic.Infographic, error)

	// List retrieves paginated infographics with optional filtering
	// Returns infographic slice, total count, and error
	// Filters:
	// - sectionName: filter by section name (exact match)
	// - state: filter by state (true/false)
	// - query: search in section_name or component_id (ILIKE)
	// - category: filter by category UUID (exact match)
	List(ctx context.Context, sectionName *string, state *bool, query *string, category *string, offset, limit int) ([]*infographic.Infographic, int, error)

	// Update updates an existing infographic
	Update(ctx context.Context, i *infographic.Infographic) error

	// Delete removes an infographic
	Delete(ctx context.Context, id string) error

	// GetSectionNames retrieves all unique section names from infographic
	GetSectionNames(ctx context.Context) ([]string, error)
}
