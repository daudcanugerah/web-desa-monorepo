// Package profilecategory contains the use case (business logic) for
// Profile section categories. It defines the Repository interface (used here,
// implemented in interface/postgres) and the Service that orchestrates
// Create / Delete / List operations.
//
// Per the category pattern, categories support create + delete only — there is
// no Update. To rename, delete and re-create.
package profilecategory

import (
	"context"

	"webdesa/api/domain/profilecategory"
)

// Repository defines the persistence interface for Profile category management.
// This interface is defined in the usecase layer (where it's USED), not in the
// domain layer, following Go best practices.
//
// Dependency Rule: interface/postgres → usecase → domain.
type Repository interface {
	// Create persists a new category.
	Create(ctx context.Context, c *profilecategory.Category) error

	// Delete removes a category by ID. The service layer is responsible
	// for checking that the category is not in use before calling this.
	Delete(ctx context.Context, id string) error

	// FindByID retrieves a category by ID. Returns an error if not found.
	// Used by the Profile service to validate that a category exists.
	FindByID(ctx context.Context, id string) (*profilecategory.Category, error)

	// List returns a paginated list of categories whose name matches the
	// optional query string (case-insensitive substring match). Results are
	// sorted by name ascending.
	List(ctx context.Context, query *string, offset, limit int) ([]*profilecategory.Category, int, error)

	// CountByCategoryIDs returns a map of category ID -> count of profile
	// records referencing that category. Categories with zero references
	// are omitted from the map.
	CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error)
}
