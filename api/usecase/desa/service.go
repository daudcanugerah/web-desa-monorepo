package desa

import (
	"context"
	"fmt"

	"webdesa/api/domain/desa"
	"webdesa/api/pkg/clock"
)

// Service implements village profile (desa) management business logic.
// The profile is stored as JSON in the settings table under key "desa_profile".
type Service struct {
	repo  Repository
	clock clock.Clock
}

// NewService creates a new village profile management service.
func NewService(repo Repository, clk clock.Clock) *Service {
	return &Service{repo: repo, clock: clk}
}

// UpdateDesaInput represents the input for updating the village profile.
type UpdateDesaInput struct {
	Name              string
	Description       *string
	Address           *string
	Phone             *string
	Email             *string
	Website           *string
	VisionMission     *string
	KepalaDesa        *string
	KepalaDesaMessage *string
	KepalaDesaMediaID *string
	Motto             *string
	Kecamatan         *string
	Kabupaten         *string
	Provinsi          *string
	JumlahPenduduk    *int
	JumlahKK          *int
	JumlahDusun       *int
	JumlahRT          *int
	JumlahRW          *int
	JumlahUMKM        *int
	SocialMedia       []desa.SocialLink
}

// Get retrieves the village profile from settings.
func (s *Service) Get(ctx context.Context) (*desa.Desa, error) {
	d, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("village profile not found: %w", err)
	}
	return d, nil
}

// Update saves the village profile to settings.
func (s *Service) Update(ctx context.Context, input UpdateDesaInput) (*desa.Desa, error) {
	// Build updated entity (upsert — no need to fetch first)
	d := &desa.Desa{
		Name:              input.Name,
		Description:       input.Description,
		Address:           input.Address,
		Phone:             input.Phone,
		Email:             input.Email,
		Website:           input.Website,
		VisionMission:     input.VisionMission,
		KepalaDesa:        input.KepalaDesa,
		KepalaDesaMessage: input.KepalaDesaMessage,
		KepalaDesaMediaID: input.KepalaDesaMediaID,
		Motto:             input.Motto,
		Kecamatan:         input.Kecamatan,
		Kabupaten:         input.Kabupaten,
		Provinsi:          input.Provinsi,
		JumlahPenduduk:    input.JumlahPenduduk,
		JumlahKK:          input.JumlahKK,
		JumlahDusun:       input.JumlahDusun,
		JumlahRT:          input.JumlahRT,
		JumlahRW:          input.JumlahRW,
		JumlahUMKM:        input.JumlahUMKM,
		SocialMedia:       input.SocialMedia,
		UpdatedAt:         s.clock.Now(),
	}

	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("invalid village profile data: %w", err)
	}

	if err := s.repo.Upsert(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to update village profile: %w", err)
	}

	return d, nil
}
