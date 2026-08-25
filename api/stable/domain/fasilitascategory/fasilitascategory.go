// Package fasilitascategory defines the Fasilitas (village facility) category
// domain entity.
//
// A Fasilitas category is a controlled vocabulary entry that organizes
// village facilities into named groups (e.g., "School", "Health Center",
// "Mosque"). Categories are first-class domain objects with their own
// lifecycle (create + delete) and are referenced by Fasilitas entities
// via FK.
//
// Note: Repository interfaces are defined in the usecase layer where they
// are USED, following Go best practices. See usecase/fasilitascategory for
// the Repository interface definition.
package fasilitascategory

import (
	"fmt"
	"strings"
	"time"
)

// Category represents a Fasilitas category in the domain layer.
// Categories are a separate, managed vocabulary — they are NOT free text.
// They are created and deleted by users with the appropriate permission
// (intentionally no Update operation; to rename, delete and re-create).
type Category struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate checks if the Category entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (c *Category) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("category ID is required")
	}

	if err := validateName(c.Name); err != nil {
		return err
	}

	if c.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}

	return nil
}

// validateName checks if the name meets domain requirements.
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("category name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("category name must not exceed 100 characters")
	}

	return nil
}
