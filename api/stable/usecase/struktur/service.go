package struktur

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/struktur"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements struktur (organizational structure) management business logic.
// It accepts interfaces (Repository, FileHandler) and returns concrete structs (Struktur).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo        Repository
	fileHandler FileHandler
	clock       clock.Clock
}

// NewService creates a new struktur management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileHandler FileHandler, clk clock.Clock) *Service {
	return &Service{
		repo:        repo,
		fileHandler: fileHandler,
		clock:       clk,
	}
}

// CreateStrukturInput represents the input for creating a struktur member
type CreateStrukturInput struct {
	Name        string
	Position    *string   // Optional
	Email       *string   // Optional
	Phone       *string   // Optional
	Description *string   // Optional
	ImageFile   io.Reader // Optional
	ImageName   string
	ImageSize   int64
	ContentType string
}

// UpdateStrukturInput represents the input for updating a struktur member
type UpdateStrukturInput struct {
	Name        string
	Position    *string   // Optional
	Email       *string   // Optional
	Phone       *string   // Optional
	Description *string   // Optional
	ImageFile   io.Reader // Optional - only if updating image
	ImageName   string
	ImageSize   int64
	ContentType string
}

// ListStrukturInput represents the input for listing struktur members
type ListStrukturInput struct {
	Query *string // Optional search query (name or position)
	Page  int
	Limit int
}

// Create creates a new struktur member with an optional profile image.
// Returns concrete Struktur struct.
//
// Validates: Requirements 13.1, 13.4
func (s *Service) Create(ctx context.Context, input CreateStrukturInput) (*struktur.Struktur, error) {
	var profileImageURL *string

	// Save profile image file if provided
	if input.ImageFile != nil {
		savedPath, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save profile image: %w", err)
		}
		profileImageURL = &savedPath
	}

	// Create struktur entity
	now := s.clock.Now()
	st := &struktur.Struktur{
		ID:              uuid.New().String(),
		Name:            input.Name,
		Position:        input.Position,
		Email:           input.Email,
		Phone:           input.Phone,
		Description:     input.Description,
		ProfileImageURL: profileImageURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Validate domain invariants
	if err := st.Validate(); err != nil {
		// Clean up uploaded file on validation failure
		if profileImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *profileImageURL)
		}
		return nil, fmt.Errorf("invalid struktur data: %w", err)
	}

	// Persist struktur
	if err := s.repo.Create(ctx, st); err != nil {
		// Clean up uploaded file on persistence failure
		if profileImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *profileImageURL)
		}
		return nil, fmt.Errorf("failed to create struktur: %w", err)
	}

	return st, nil
}

// GetByID retrieves a struktur member by ID.
// Returns concrete Struktur struct.
//
// Validates: Requirements 13.3
func (s *Service) GetByID(ctx context.Context, id string) (*struktur.Struktur, error) {
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("struktur not found: %w", err)
	}

	return st, nil
}

// List retrieves paginated struktur members with optional filtering.
// Returns concrete Struktur structs and pagination metadata.
//
// Validates: Requirements 13.1, 13.2, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListStrukturInput) ([]*struktur.Struktur, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve struktur members from repository with filters
	strukturList, total, err := s.repo.List(ctx, input.Query, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list struktur: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return strukturList, paginationResult, nil
}

// Update updates an existing struktur member.
// If ImageFile is provided, updates the profile image; otherwise keeps existing image.
// Returns concrete Struktur struct.
//
// Validates: Requirements 13.5
func (s *Service) Update(ctx context.Context, id string, input UpdateStrukturInput) (*struktur.Struktur, error) {
	// Retrieve existing struktur
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("struktur not found: %w", err)
	}

	var oldImageURL *string
	if st.ProfileImageURL != nil {
		oldImageURL = new(string)
		*oldImageURL = *st.ProfileImageURL
	}

	// Update profile image if provided
	if input.ImageFile != nil {
		newImageURL, err := s.fileHandler.SaveImage(ctx, input.ImageName, input.ImageFile, input.ImageSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save new profile image: %w", err)
		}
		st.ProfileImageURL = &newImageURL
	}

	// Update fields
	st.Name = input.Name
	st.Position = input.Position
	st.Email = input.Email
	st.Phone = input.Phone
	st.Description = input.Description
	st.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := st.Validate(); err != nil {
		// Clean up new image if validation fails
		if input.ImageFile != nil && st.ProfileImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *st.ProfileImageURL)
		}
		return nil, fmt.Errorf("invalid struktur data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, st); err != nil {
		// Clean up new image if persistence fails
		if input.ImageFile != nil && st.ProfileImageURL != nil {
			_ = s.fileHandler.Delete(ctx, *st.ProfileImageURL)
		}
		return nil, fmt.Errorf("failed to update struktur: %w", err)
	}

	// Delete old image if a new one was uploaded successfully
	if input.ImageFile != nil && oldImageURL != nil && *oldImageURL != "" {
		_ = s.fileHandler.Delete(ctx, *oldImageURL)
	}

	return st, nil
}

// Delete removes a struktur member and its associated profile image file.
//
// Validates: Requirements 13.6
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve struktur to get profile image URL
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("struktur not found: %w", err)
	}

	// Delete struktur from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete struktur: %w", err)
	}

	// Delete associated profile image file if it exists
	if st.ProfileImageURL != nil && *st.ProfileImageURL != "" {
		if err := s.fileHandler.Delete(ctx, *st.ProfileImageURL); err != nil {
			// Log error but don't fail the operation since struktur is already deleted
			// In production, this should be logged properly
			_ = err
		}
	}

	return nil
}
