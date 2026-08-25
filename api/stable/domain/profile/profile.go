package profile

import (
	"fmt"
	"strings"
	"time"
)

// Profile represents a profile subsection entity in the domain layer.
// Profile sections contain information like keuangan (finance), lokasi desa (village location), potensi (potential).
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/profile for the Repository interface definition.
type Profile struct {
	ID              string    `db:"id"`
	Content         string    `db:"content"`
	SectionName     string    `db:"section_name"`
	SectionEndpoint string    `db:"section_endpoint"`
	State           bool      `db:"state"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Validate checks if the Profile entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (p *Profile) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("profile ID is required")
	}

	if err := validateContent(p.Content); err != nil {
		return err
	}

	if err := validateSectionName(p.SectionName); err != nil {
		return err
	}

	if err := validateSectionEndpoint(p.SectionEndpoint); err != nil {
		return err
	}

	if p.CreatedAt.IsZero() {
		return fmt.Errorf("profile created_at is required")
	}

	if p.UpdatedAt.IsZero() {
		return fmt.Errorf("profile updated_at is required")
	}

	return nil
}

func validateContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("profile content is required")
	}
	if len(content) < 1 {
		return fmt.Errorf("profile content must be at least 1 character")
	}
	if len(content) > 10000 {
		return fmt.Errorf("profile content must not exceed 10000 characters")
	}
	return nil
}

func validateSectionName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("profile section_name is required")
	}
	if len(name) < 1 {
		return fmt.Errorf("profile section_name must be at least 1 character")
	}
	if len(name) > 100 {
		return fmt.Errorf("profile section_name must not exceed 100 characters")
	}
	return nil
}

func validateSectionEndpoint(endpoint string) error {
	if strings.TrimSpace(endpoint) == "" {
		return fmt.Errorf("profile section_endpoint is required")
	}
	if len(endpoint) < 1 {
		return fmt.Errorf("profile section_endpoint must be at least 1 character")
	}
	if len(endpoint) > 255 {
		return fmt.Errorf("profile section_endpoint must not exceed 255 characters")
	}
	if !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("profile section_endpoint must start with /")
	}
	return nil
}
