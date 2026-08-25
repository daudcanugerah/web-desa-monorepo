package umkm

import (
	"fmt"
	"strings"
	"time"
)

// UMKM represents a small and medium business entity in the domain layer.
// UMKM (Usaha Mikro, Kecil, dan Menengah) listings promote local businesses in the village.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/umkm for the Repository interface definition.
type UMKM struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Category    string    `db:"category"`    // FK to umkm_categories.id (UUID)
	CategoryName *string `db:"category_name"`
	Description string    `db:"description"` // Business description
	Owner       *string   `db:"owner"`       // Optional business owner name
	Address     *string   `db:"address"`     // Optional business address
	Phone       *string   `db:"phone"`       // Optional contact phone
	Email       *string   `db:"email"`       // Optional contact email
	Website     *string   `db:"website"`     // Optional business website
	Images      []string  `db:"images"`      // Array of image URLs
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Validate checks if the UMKM entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (u *UMKM) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("umkm ID is required")
	}

	if err := validateName(u.Name); err != nil {
		return err
	}

	if strings.TrimSpace(u.Category) == "" {
		return fmt.Errorf("category is required")
	}

	if err := validateDescription(u.Description); err != nil {
		return err
	}

	if u.Owner != nil {
		if err := validateOwner(*u.Owner); err != nil {
			return err
		}
	}

	if u.Address != nil {
		if err := validateAddress(*u.Address); err != nil {
			return err
		}
	}

	if u.Phone != nil {
		if err := validatePhone(*u.Phone); err != nil {
			return err
		}
	}

	if u.Email != nil {
		if err := validateEmail(*u.Email); err != nil {
			return err
		}
	}

	if u.Website != nil {
		if err := validateWebsite(*u.Website); err != nil {
			return err
		}
	}

	if u.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if u.UpdatedAt.IsZero() {
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

// validateOwner checks if the owner name meets domain requirements
func validateOwner(owner string) error {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("owner cannot be empty string")
	}

	if len(owner) > 255 {
		return fmt.Errorf("owner must not exceed 255 characters")
	}

	return nil
}

// validateAddress checks if the address meets domain requirements
func validateAddress(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return fmt.Errorf("address cannot be empty string")
	}

	// Address is TEXT type in database, no strict length limit
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

// validateWebsite checks if the website meets domain requirements
func validateWebsite(website string) error {
	website = strings.TrimSpace(website)
	if website == "" {
		return fmt.Errorf("website cannot be empty string")
	}

	if len(website) > 255 {
		return fmt.Errorf("website must not exceed 255 characters")
	}

	return nil
}

// validateDescription checks if the description meets domain requirements
func validateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("description is required")
	}

	// Description is TEXT type in database, no strict length limit
	return nil
}
