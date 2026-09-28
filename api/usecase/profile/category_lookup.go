package profile

import (
	"context"

	"webdesa/api/domain/profilecategory"
)

// CategoryLookup is the minimal interface the Profile service needs from the
// category subsystem to validate that a category ID exists (when one is
// provided; profiles may have no category).
//
// It is defined here (where it is USED), not in usecase/profilecategory,
// following the project's interface placement convention. The concrete
// *profilecategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*profilecategory.Category, error)
}
