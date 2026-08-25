package user

import (
	"fmt"
	"strings"
	"time"
)

// User represents a system user entity in the domain layer.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/user for the Repository interface definition.
type User struct {
	ID              string    `db:"id"`
	Name            string    `db:"name"`
	Email           string    `db:"email"`
	HashedPassword  string    `db:"hashed_password"`
	ProfileImageURL *string   `db:"profile_image_url"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Validate checks if the User entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (u *User) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("user ID is required")
	}

	if err := validateName(u.Name); err != nil {
		return err
	}

	if err := validateEmail(u.Email); err != nil {
		return err
	}

	if u.HashedPassword == "" {
		return fmt.Errorf("hashed password is required")
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

// validateEmail checks if the email meets domain requirements
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
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
