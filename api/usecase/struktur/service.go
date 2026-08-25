package struktur

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/struktur"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"
	galleryUsecase "webdesa/api/usecase/gallery"

	"github.com/google/uuid"
)

// Service implements struktur (organizational structure) management business logic.
// It accepts interfaces (Repository, FileStore) and returns concrete structs (Struktur).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo      Repository
	fileStore galleryUsecase.FileStore
	clock     clock.Clock
}

// NewService creates a new struktur management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock) *Service {
	return &Service{
		repo:      repo,
		fileStore: fileStore,
		clock:     clk,
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
	// ProfileImageMediaID references a media already uploaded via
	// POST /struktur/upload-media. Mutually exclusive with ImageFile.
	ProfileImageMediaID *string
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
	// ProfileImageMediaID references a media already uploaded via
	// POST /struktur/upload-media. Mutually exclusive with ImageFile.
	ProfileImageMediaID *string
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
	profileImageMediaID, err := s.resolveProfileImageMediaID(ctx, input.ProfileImageMediaID, input.ImageFile, func() (galleryUsecase.SavedFile, error) {
		return s.fileStore.SaveImage(ctx, galleryUsecase.FeatureStruktur, galleryUsecase.FileInput{
			OriginalName: input.ImageName,
			Content:      input.ImageFile,
			Size:         input.ImageSize,
			ContentType:  input.ContentType,
		})
	})
	if err != nil {
		return nil, err
	}

	// Create struktur entity
	now := s.clock.Now()
	st := &struktur.Struktur{
		ID:                  uuid.New().String(),
		Name:                input.Name,
		Position:            input.Position,
		Email:               input.Email,
		Phone:               input.Phone,
		Description:         input.Description,
		ProfileImageMediaID: profileImageMediaID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Validate domain invariants
	if err := st.Validate(); err != nil {
		// Clean up the inline upload on validation failure; a
		// pre-uploaded media ref is left for the caller/GC.
		if input.ImageFile != nil && profileImageMediaID != nil {
			_ = s.fileStore.Delete(ctx, *profileImageMediaID)
		}
		return nil, fmt.Errorf("invalid struktur data: %w", err)
	}

	// Persist struktur
	if err := s.repo.Create(ctx, st); err != nil {
		// Clean up the inline upload on persistence failure.
		if input.ImageFile != nil && profileImageMediaID != nil {
			_ = s.fileStore.Delete(ctx, *profileImageMediaID)
		}
		return nil, fmt.Errorf("failed to create struktur: %w", err)
	}

	return st, nil
}

// resolveProfileImageMediaID returns the media id to attach to a struktur
// row. A pre-uploaded media ref (from POST /struktur/upload-media) wins;
// otherwise the inline image file is saved. Passing both is an error.
func (s *Service) resolveProfileImageMediaID(ctx context.Context, mediaID *string, imageFile io.Reader, saveFile func() (galleryUsecase.SavedFile, error)) (*string, error) {
	if mediaID != nil && *mediaID != "" {
		if imageFile != nil {
			return nil, fmt.Errorf("provide either profile_image_media_id or an image file, not both")
		}
		if _, err := s.fileStore.Open(ctx, *mediaID); err != nil {
			return nil, fmt.Errorf("invalid profile_image_media_id: %w", err)
		}
		mid := *mediaID
		return &mid, nil
	}
	if imageFile != nil {
		saved, err := saveFile()
		if err != nil {
			return nil, fmt.Errorf("failed to save profile image: %w", err)
		}
		mid := saved.MediaID
		return &mid, nil
	}
	return nil, nil
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

	var oldMediaID string
	if st.ProfileImageMediaID != nil {
		oldMediaID = *st.ProfileImageMediaID
	}

	var newMediaID string
	switch {
	case input.ProfileImageMediaID != nil && *input.ProfileImageMediaID != "":
		if input.ImageFile != nil {
			return nil, fmt.Errorf("provide either profile_image_media_id or an image file, not both")
		}
		if _, err := s.fileStore.Open(ctx, *input.ProfileImageMediaID); err != nil {
			return nil, fmt.Errorf("invalid profile_image_media_id: %w", err)
		}
		newMediaID = *input.ProfileImageMediaID
	case input.ImageFile != nil:
		saved, err := s.fileStore.SaveImage(ctx, galleryUsecase.FeatureStruktur, galleryUsecase.FileInput{
			OriginalName: input.ImageName,
			Content:      input.ImageFile,
			Size:         input.ImageSize,
			ContentType:  input.ContentType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save new profile image: %w", err)
		}
		newMediaID = saved.MediaID
	}

	// Update fields
	st.Name = input.Name
	st.Position = input.Position
	st.Email = input.Email
	st.Phone = input.Phone
	st.Description = input.Description
	if newMediaID != "" {
		id := newMediaID
		st.ProfileImageMediaID = &id
	}
	st.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := st.Validate(); err != nil {
		// Clean up the inline upload on validation failure; a
		// pre-uploaded media ref is left for the caller/GC.
		if input.ImageFile != nil && newMediaID != "" {
			_ = s.fileStore.Delete(ctx, newMediaID)
		}
		return nil, fmt.Errorf("invalid struktur data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, st); err != nil {
		// Clean up the inline upload on persistence failure.
		if input.ImageFile != nil && newMediaID != "" {
			_ = s.fileStore.Delete(ctx, newMediaID)
		}
		return nil, fmt.Errorf("failed to update struktur: %w", err)
	}

	// Delete old image if a new one was uploaded successfully
	if newMediaID != "" && oldMediaID != "" {
		_ = s.fileStore.Delete(ctx, oldMediaID)
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
	if st.ProfileImageMediaID != nil && *st.ProfileImageMediaID != "" {
		_ = s.fileStore.Delete(ctx, *st.ProfileImageMediaID)
	}

	return nil
}
