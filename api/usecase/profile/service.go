package profile

import (
	"context"
	"fmt"

	"webdesa/api/domain/profile"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements profile management business logic.
// It accepts interfaces (Repository) and returns concrete structs (Profile).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo  Repository
	clock clock.Clock
}

// NewService creates a new profile management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, clk clock.Clock) *Service {
	return &Service{
		repo:  repo,
		clock: clk,
	}
}

// CreateProfileInput represents the input for creating a profile
type CreateProfileInput struct {
	Content         string
	SectionName     string
	SectionEndpoint string
	State           bool
}

// UpdateProfileInput represents the input for updating a profile
type UpdateProfileInput struct {
	Content         string
	SectionName     string
	SectionEndpoint string
	State           bool
}

// ListProfilesInput represents the input for listing profiles
type ListProfilesInput struct {
	SectionName *string
	State       *bool
	Query       *string
	Page        int
	Limit       int
}

// Create creates a new profile.
// Returns concrete Profile struct.
func (s *Service) Create(ctx context.Context, input CreateProfileInput) (*profile.Profile, error) {
	now := s.clock.Now()
	p := &profile.Profile{
		ID:              uuid.New().String(),
		Content:         input.Content,
		SectionName:     input.SectionName,
		SectionEndpoint: input.SectionEndpoint,
		State:           input.State,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Validate domain invariants
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile data: %w", err)
	}

	// Persist to database
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return p, nil
}

// GetByID retrieves a profile by ID.
// Returns error if profile not found.
func (s *Service) GetByID(ctx context.Context, id string) (*profile.Profile, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return p, nil
}

// List retrieves paginated profiles with optional filtering.
// Returns profiles, pagination info, and error.
func (s *Service) List(ctx context.Context, input ListProfilesInput) ([]*profile.Profile, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Query repository
	profiles, total, err := s.repo.List(ctx, input.SectionName, input.State, input.Query, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return profiles, paginationResult, nil
}

// Update updates an existing profile.
// Returns updated Profile struct.
func (s *Service) Update(ctx context.Context, id string, input UpdateProfileInput) (*profile.Profile, error) {
	// Get existing profile
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Update fields
	p.Content = input.Content
	p.SectionName = input.SectionName
	p.SectionEndpoint = input.SectionEndpoint
	p.State = input.State
	p.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile data: %w", err)
	}

	// Persist to database
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return p, nil
}

// Delete removes a profile.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// GetSectionNames retrieves all unique section names.
func (s *Service) GetSectionNames(ctx context.Context) ([]string, error) {
	names, err := s.repo.GetSectionNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get section names: %w", err)
	}
	return names, nil
}
