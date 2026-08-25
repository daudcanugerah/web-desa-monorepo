package infographic

import (
	"context"

	"webdesa/api/domain/infographiccategory"
)

// CategoryLookup is the minimal interface the Infographic service needs from
// the category subsystem to validate that a category ID exists (when one
// is provided; infographics may have no category).
//
// It is defined here (where it is USED), not in the usecase/infographiccategory
// package, following the project's interface placement convention.
// The concrete *infographiccategory.Service satisfies this interface implicitly.
type CategoryLookup interface {
	FindByID(ctx context.Context, id string) (*infographiccategory.Category, error)
}
