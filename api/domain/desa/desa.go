package desa

import (
	"fmt"
	"strings"
	"time"
)

// Desa represents a village profile entity in the domain layer.
// It is stored as JSON under the key "desa_profile" in the settings table.
// This is the core domain entity with no external dependencies.
//
// Note: Repository interfaces are defined in the usecase layer where they are USED,
// following Go best practices. See usecase/desa for the Repository interface definition.
type Desa struct {
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	Address       *string   `json:"address,omitempty"`
	Phone         *string   `json:"phone,omitempty"`
	Email         *string   `json:"email,omitempty"`
	Website       *string   `json:"website,omitempty"`
	VisionMission *string   `json:"vision_mission,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate checks if the Desa entity satisfies domain invariants.
func (d *Desa) Validate() error {
	if err := validateName(d.Name); err != nil {
		return err
	}

	if d.Description != nil {
		if err := validateDescription(*d.Description); err != nil {
			return err
		}
	}

	if d.Address != nil {
		if err := validateAddress(*d.Address); err != nil {
			return err
		}
	}

	if d.Phone != nil {
		if err := validatePhone(*d.Phone); err != nil {
			return err
		}
	}

	if d.Email != nil {
		if err := validateEmail(*d.Email); err != nil {
			return err
		}
	}

	if d.Website != nil {
		if err := validateWebsite(*d.Website); err != nil {
			return err
		}
	}

	if d.VisionMission != nil {
		if err := validateVisionMission(*d.VisionMission); err != nil {
			return err
		}
	}

	return nil
}

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

func validateDescription(description string) error {
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description cannot be empty string")
	}
	return nil
}

func validateAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return fmt.Errorf("address cannot be empty string")
	}
	return nil
}

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

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email cannot be empty string")
	}
	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("email must be a valid format")
	}
	return nil
}

func validateWebsite(website string) error {
	website = strings.TrimSpace(website)
	if website == "" {
		return fmt.Errorf("website cannot be empty string")
	}
	if len(website) > 255 {
		return fmt.Errorf("website must not exceed 255 characters")
	}
	if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
		return fmt.Errorf("website must be a valid URL starting with http:// or https://")
	}
	return nil
}

func validateVisionMission(visionMission string) error {
	if strings.TrimSpace(visionMission) == "" {
		return fmt.Errorf("vision/mission cannot be empty string")
	}
	return nil
}
