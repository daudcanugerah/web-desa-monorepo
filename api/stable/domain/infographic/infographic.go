package infographic

import (
	"fmt"
	"time"
)

// ComponentType represents the type of Metabase component
type ComponentType string

const (
	ComponentTypeQuestion  ComponentType = "question"
	ComponentTypeDashboard ComponentType = "dashboard"
)

// Infographic represents an infographic dashboard entity in the domain layer.
// Infographics display Metabase dashboards or questions for data visualization.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/infographic for the Repository interface definition.
type Infographic struct {
	ID              string        `db:"id"`
	ComponentID     int64         `db:"component_id"`
	ComponentType   ComponentType `db:"component_type"`
	SectionName     string        `db:"section_name"`
	SectionEndpoint string        `db:"section_endpoint"`
	Category        *string       `db:"category"` // Optional FK to infographic_categories.id (UUID)
	CategoryName    *string       `db:"category_name"`
	State           bool          `db:"state"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
}

// Validate checks if the Infographic entity satisfies domain invariants.
// This ensures the entity is in a valid state before persistence.
func (i *Infographic) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("infographic ID is required")
	}

	if err := validateComponentID(i.ComponentID); err != nil {
		return err
	}

	if err := validateComponentType(i.ComponentType); err != nil {
		return err
	}

	if err := validateSectionName(i.SectionName); err != nil {
		return err
	}

	if err := validateSectionEndpoint(i.SectionEndpoint); err != nil {
		return err
	}

	if i.CreatedAt.IsZero() {
		return fmt.Errorf("infographic created_at is required")
	}

	if i.UpdatedAt.IsZero() {
		return fmt.Errorf("infographic updated_at is required")
	}

	return nil
}

func validateComponentID(id int64) error {
	if id <= 0 {
		return fmt.Errorf("infographic component_id must be a positive integer")
	}
	return nil
}

func validateComponentType(ct ComponentType) error {
	if ct != ComponentTypeQuestion && ct != ComponentTypeDashboard {
		return fmt.Errorf("infographic component_type must be 'question' or 'dashboard'")
	}
	return nil
}

func validateSectionName(name string) error {
	if name == "" {
		return fmt.Errorf("infographic section_name is required")
	}
	if len(name) < 1 {
		return fmt.Errorf("infographic section_name must be at least 1 character")
	}
	if len(name) > 100 {
		return fmt.Errorf("infographic section_name must not exceed 100 characters")
	}
	return nil
}

func validateSectionEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("infographic section_endpoint is required")
	}
	if len(endpoint) < 1 {
		return fmt.Errorf("infographic section_endpoint must be at least 1 character")
	}
	if len(endpoint) > 255 {
		return fmt.Errorf("infographic section_endpoint must not exceed 255 characters")
	}
	if endpoint[0] != '/' {
		return fmt.Errorf("infographic section_endpoint must start with /")
	}
	return nil
}
