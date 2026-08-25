package ppid

import (
	"context"

	"webdesa/api/domain/ppidcategory"
)

// CategoryLookup is the minimal interface the PPID service needs from
// the category subsystem to validate that a category ID exists (when one
// is provided; PPID documents may have no category).
//
// It is defined here (where it is USED), not in the usecase/ppidcategory
// package, following the project's interface placement convention.
// The concrete *ppidcategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*ppidcategory.Category, error)
}
