package fasilitas

import (
	"context"

	"webdesa/api/domain/fasilitascategory"
)

// CategoryLookup is the minimal interface the Fasilitas service needs from
// the category subsystem to validate that a category ID exists (when one
// is provided; facilities may have no category).
//
// It is defined here (where it is USED), not in the usecase/fasilitascategory
// package, following the project's interface placement convention.
// The concrete *fasilitascategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*fasilitascategory.Category, error)
}
