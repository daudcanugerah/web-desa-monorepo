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
	Name              string    `json:"name"`
	Description       *string   `json:"description,omitempty"`
	Address           *string   `json:"address,omitempty"`
	Phone             *string   `json:"phone,omitempty"`
	Email             *string   `json:"email,omitempty"`
	Website           *string   `json:"website,omitempty"`
	VisionMission     *string   `json:"vision_mission,omitempty"`
	KepalaDesa        *string   `json:"kepala_desa,omitempty"`
	KepalaDesaMessage *string   `json:"kepala_desa_message,omitempty"`
	KepalaDesaMediaID *string   `json:"kepala_desa_media_id,omitempty"`
	Motto             *string   `json:"motto,omitempty"`
	Kecamatan         *string   `json:"kecamatan,omitempty"`
	Kabupaten         *string   `json:"kabupaten,omitempty"`
	Provinsi          *string   `json:"provinsi,omitempty"`
	JumlahPenduduk    *int      `json:"jumlah_penduduk,omitempty"`
	JumlahKK          *int      `json:"jumlah_kk,omitempty"`
	JumlahDusun       *int      `json:"jumlah_dusun,omitempty"`
	JumlahRT          *int      `json:"jumlah_rt,omitempty"`
	JumlahRW          *int      `json:"jumlah_rw,omitempty"`
	JumlahUMKM        *int      `json:"jumlah_umkm,omitempty"`
	SocialMedia       []SocialLink `json:"social_media,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SocialLink is one official social-media channel for the village, stored in
// the desa_profile settings JSON. Platform is a free-form slug (e.g.
// "facebook", "instagram", "youtube", "tiktok", "whatsapp", "twitter") the
// front-end maps to an icon; URL must be http(s).
type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
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

	if d.KepalaDesa != nil {
		if err := validateOptionalText("kepala desa", *d.KepalaDesa, 255); err != nil {
			return err
		}
	}

	if d.KepalaDesaMessage != nil {
		if err := validateOptionalText("kepala desa message", *d.KepalaDesaMessage, 10000); err != nil {
			return err
		}
	}

	if d.Motto != nil {
		if err := validateOptionalText("motto", *d.Motto, 500); err != nil {
			return err
		}
	}

	if d.Kecamatan != nil {
		if err := validateOptionalText("kecamatan", *d.Kecamatan, 255); err != nil {
			return err
		}
	}

	if d.Kabupaten != nil {
		if err := validateOptionalText("kabupaten", *d.Kabupaten, 255); err != nil {
			return err
		}
	}

	if d.Provinsi != nil {
		if err := validateOptionalText("provinsi", *d.Provinsi, 255); err != nil {
			return err
		}
	}

	for i, link := range d.SocialMedia {
		if strings.TrimSpace(link.Platform) == "" {
			return fmt.Errorf("social_media[%d] platform is required", i)
		}
		if len(link.Platform) > 50 {
			return fmt.Errorf("social_media[%d] platform must not exceed 50 characters", i)
		}
		if !strings.HasPrefix(link.URL, "http://") && !strings.HasPrefix(link.URL, "https://") {
			return fmt.Errorf("social_media[%d] url must start with http:// or https://", i)
		}
		if len(link.URL) > 500 {
			return fmt.Errorf("social_media[%d] url must not exceed 500 characters", i)
		}
	}

	for _, counter := range []struct {
		name  string
		value *int
	}{
		{"jumlah penduduk", d.JumlahPenduduk},
		{"jumlah kk", d.JumlahKK},
		{"jumlah dusun", d.JumlahDusun},
		{"jumlah rt", d.JumlahRT},
		{"jumlah rw", d.JumlahRW},
		{"jumlah umkm", d.JumlahUMKM},
	} {
		if counter.value != nil && *counter.value < 0 {
			return fmt.Errorf("%s must not be negative", counter.name)
		}
	}

	return nil
}

// validateOptionalText checks an optional free-text field: when present it
// must be non-blank after trim and within maxLen runes.
func validateOptionalText(field, value string, maxLen int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty string", field)
	}
	if len([]rune(value)) > maxLen {
		return fmt.Errorf("%s must not exceed %d characters", field, maxLen)
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
