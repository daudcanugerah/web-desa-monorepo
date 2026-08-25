package profile

import (
	"context"

	"webdesa/api/domain/profile"
)

// Repository defines the persistence interface for profile management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/profile_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new profile in the database
	Create(ctx context.Context, p *profile.Profile) error

	// FindByID retrieves a profile by ID
	FindByID(ctx context.Context, id string) (*profile.Profile, error)

	// List retrieves paginated profiles with optional filtering
	// Returns profile slice, total count, and error
	// Filters:
	// - sectionName: filter by section name (exact match)
	// - state: filter by state (true/false)
	// - query: search in section_name or content (ILIKE)
	List(ctx context.Context, sectionName *string, state *bool, query *string, offset, limit int) ([]*profile.Profile, int, error)

	// Update updates an existing profile
	Update(ctx context.Context, p *profile.Profile) error

	// Delete removes a profile
	Delete(ctx context.Context, id string) error

	// GetSectionNames retrieves all unique section names from profile
	GetSectionNames(ctx context.Context) ([]string, error)
}
