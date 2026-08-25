package umkm

import (
	"context"

	"webdesa/api/domain/umkmcategory"
)

// CategoryLookup is the minimal interface the UMKM service needs from
// the category subsystem to validate that a category ID exists.
//
// It is defined here (where it is USED), not in the usecase/umkmcategory
// package, following the project's interface placement convention.
// The concrete *umkmcategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*umkmcategory.Category, error)
}
