// Package ppidcategory contains the use case (business logic) for
// PPID (public information disclosure) categories. It defines the
// Repository interface (used here, implemented in interface/repository)
// and the Service that orchestrates Create / Delete / List operations.
//
// Per requirements, categories support create + delete only — there is
// no Update. To rename, delete and re-create.
package ppidcategory

import (
	"context"

	"webdesa/api/domain/ppidcategory"
)

// Repository defines the persistence interface for PPID category management.
// This interface is defined in the usecase layer (where it's USED), not in the
// domain layer, following Go best practices.
//
// The implementation will be in interface/repository/ppidcategory_postgres.go,
// which depends on this interface. This maintains the Dependency Rule:
// interface/postgres → usecase → domain.
type Repository interface {
	// Create persists a new category.
	Create(ctx context.Context, c *ppidcategory.Category) error

	// Delete removes a category by ID. The service layer is responsible
	// for checking that the category is not in use before calling this.
	Delete(ctx context.Context, id string) error

	// FindByID retrieves a category by ID. Returns an error if not found.
	// Used by the PPID service to validate that a category exists.
	FindByID(ctx context.Context, id string) (*ppidcategory.Category, error)

	// List returns a paginated list of categories whose name matches the
	// optional query string (case-insensitive substring match). Results are
	// sorted by name ascending.
	// Returns categories, total count, and error.
	List(ctx context.Context, query *string, offset, limit int) ([]*ppidcategory.Category, int, error)

	// CountByCategoryIDs returns a map of category ID -> count of ppid
	// records referencing that category. Used by the service to enrich
	// list responses with usage counts. Categories with zero references
	// are omitted from the map.
	CountByCategoryIDs(ctx context.Context, ids []string) (map[string]int, error)
}
