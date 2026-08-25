package berita

import (
	"context"

	"webdesa/api/domain/beritacategory"
)

// CategoryLookup is the minimal interface the berita service needs from
// the category subsystem to validate that a category ID exists.
//
// It is defined here (where it is USED), not in the usecase/beritacategory
// package, following the project's interface placement convention.
// The concrete *beritacategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*beritacategory.Category, error)
}
