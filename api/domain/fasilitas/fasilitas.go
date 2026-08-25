package fasilitas

import (
	"fmt"
	"strings"
	"time"
)

// Fasilitas represents a village facility entity with geographic location in the domain layer.
// Facilities include public infrastructure like schools, health centers, mosques, etc.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/fasilitas for the Repository interface definition.
type Fasilitas struct {
	ID             string    `db:"id"`
	Name           string    `db:"name"`
	Category       *string   `db:"category"` // Optional FK to fasilitas_categories.id (UUID)
	CategoryName   *string   `db:"category_name"`
	Latitude       float64   `db:"latitude"`         // Geographic latitude coordinate (DECIMAL(10, 8))
	Longitude      float64   `db:"longitude"`        // Geographic longitude coordinate (DECIMAL(11, 8))
	Description    *string   `db:"description"`      // Optional facility description
	ImagesMediaIDs []string  `db:"images_media_ids"` // Gallery media UUIDs (preferred going forward)
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// Validate checks if the Fasilitas entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (f *Fasilitas) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("fasilitas ID is required")
	}

	if err := validateName(f.Name); err != nil {
		return err
	}

	if err := validateLatitude(f.Latitude); err != nil {
		return err
	}

	if err := validateLongitude(f.Longitude); err != nil {
		return err
	}

	if f.Category != nil {
		if err := validateCategory(*f.Category); err != nil {
			return err
		}
	}

	if f.Description != nil {
		if err := validateDescription(*f.Description); err != nil {
			return err
		}
	}


	if err := validateImagesMediaIDs(f.ImagesMediaIDs); err != nil {
		return err
	}

	if f.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if f.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}

	return nil
}

// validateName checks if the name meets domain requirements
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("name must not exceed 255 characters")
	}

	return nil
}

// validateLatitude checks if the latitude is within valid geographic range
func validateLatitude(lat float64) error {
	if lat < -90.0 || lat > 90.0 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees")
	}

	return nil
}

// validateLongitude checks if the longitude is within valid geographic range
func validateLongitude(lon float64) error {
	if lon < -180.0 || lon > 180.0 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees")
	}

	return nil
}

// validateCategory checks if the category (UUID string) meets domain requirements.
// The DB enforces the FK; this is just a basic non-empty check.
func validateCategory(category string) error {
	category = strings.TrimSpace(category)
	if category == "" {
		return fmt.Errorf("category cannot be empty string")
	}

	return nil
}

// validateDescription checks if the description meets domain requirements
func validateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("description cannot be empty string")
	}

	// Description is TEXT type in database, no strict length limit
	return nil
}

// validateImagesMediaIDs checks if the images_media_ids array meets domain requirements.

// validateImagesMediaIDs checks if the images_media_ids array meets domain requirements.
// Each entry must be a non-empty UUID-shaped string.
func validateImagesMediaIDs(ids []string) error {
	if ids == nil {
		return nil
	}

	for i, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("images_media_ids at index %d cannot be empty string", i)
		}

		if len(id) > 64 {
			return fmt.Errorf("images_media_ids at index %d must not exceed 64 characters", i)
		}
	}

	return nil
}
