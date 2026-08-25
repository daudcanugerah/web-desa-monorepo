package struktur

import (
	"fmt"
	"strings"
	"time"
)

// Struktur represents an organizational structure member entity in the domain layer.
// This entity represents village leadership and organizational hierarchy members
// that citizens can view to understand the village governance structure.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/struktur for the Repository interface definition.
type Struktur struct {
	ID              string    `db:"id"`
	Name            string    `db:"name"`
	Position        *string   `db:"position"`          // Optional position/title (e.g., "Village Head", "Secretary")
	Email           *string   `db:"email"`             // Optional contact email
	Phone           *string   `db:"phone"`             // Optional contact phone
	ProfileImageURL *string   `db:"profile_image_url"` // Optional profile image URL
	Description     *string   `db:"description"`       // Optional long-form biography/description
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Validate checks if the Struktur entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (s *Struktur) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("struktur ID is required")
	}

	if err := validateName(s.Name); err != nil {
		return err
	}

	if s.Position != nil {
		if err := validatePosition(*s.Position); err != nil {
			return err
		}
	}

	if s.Email != nil {
		if err := validateEmail(*s.Email); err != nil {
			return err
		}
	}

	if s.Phone != nil {
		if err := validatePhone(*s.Phone); err != nil {
			return err
		}
	}

	if s.ProfileImageURL != nil {
		if err := validateProfileImageURL(*s.ProfileImageURL); err != nil {
			return err
		}
	}

	if s.Description != nil {
		if err := validateDescription(*s.Description); err != nil {
			return err
		}
	}

	if s.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if s.UpdatedAt.IsZero() {
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

// validatePosition checks if the position meets domain requirements
func validatePosition(position string) error {
	position = strings.TrimSpace(position)
	if position == "" {
		return fmt.Errorf("position cannot be empty string")
	}

	if len(position) > 255 {
		return fmt.Errorf("position must not exceed 255 characters")
	}

	return nil
}

// validateEmail checks if the email meets domain requirements
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email cannot be empty string")
	}

	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}

	// Basic email format validation (domain layer keeps it simple)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("email must be a valid format")
	}

	return nil
}

// validatePhone checks if the phone meets domain requirements
func validatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return fmt.Errorf("phone cannot be empty string")
	}

	if len(phone) > 50 {
		return fmt.Errorf("phone must not exceed 50 characters")
	}

	return nil
}

// validateProfileImageURL checks if the profile image URL meets domain requirements
func validateProfileImageURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("profile image URL cannot be empty string")
	}

	if len(url) > 500 {
		return fmt.Errorf("profile image URL must not exceed 500 characters")
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
