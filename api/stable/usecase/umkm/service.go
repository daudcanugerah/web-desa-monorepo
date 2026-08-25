package umkm

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/umkm"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements UMKM (small business) management business logic.
// It accepts interfaces (Repository, FileHandler, CategoryLookup) and
// returns concrete structs (UMKM).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileHandler  FileHandler
	clock        clock.Clock
	categoryRepo CategoryLookup
}

// NewService creates a new UMKM management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileHandler FileHandler, clk clock.Clock, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileHandler:  fileHandler,
		clock:        clk,
		categoryRepo: categoryRepo,
	}
}

// CreateUMKMInput represents the input for creating a UMKM
type CreateUMKMInput struct {
	Name        string
	Category    string
	Description string
	Owner       *string
	Address     *string
	Phone       *string
	Email       *string
	Website     *string
	Images      []string
	ImageFiles  []ImageInput // For file uploads
}

// ImageInput represents an image file input
type ImageInput struct {
	Filename    string
	Content     io.Reader
	Size        int64
	ContentType string
}

// UpdateUMKMInput represents the input for updating a UMKM
type UpdateUMKMInput struct {
	Name        string
	Category    string
	Description string
	Owner       *string
	Address     *string
	Phone       *string
	Email       *string
	Website     *string
	Images      []string
	ImageFiles  []ImageInput // For new file uploads
}

// ListUMKMInput represents the input for listing UMKM
type ListUMKMInput struct {
	Query    *string // Optional search query (name, description, owner)
	Category *string // Optional filter by category
	Page     int
	Limit    int
}

// Create creates a new UMKM.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.6, 11.3
func (s *Service) Create(ctx context.Context, input CreateUMKMInput) (*umkm.UMKM, error) {
	// Validate the category exists. The FK is enforced at the DB level,
	// but checking here provides a friendlier 400 response.
	var categoryName *string
	cat, err := s.categoryRepo.FindByID(ctx, input.Category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}
	name := cat.Name
	categoryName = &name

	// Image URLs: handler has already saved files; just collect URLs.
	imageURLs := input.Images
	if len(input.ImageFiles) > 0 {
		imageURLs = make([]string, len(input.ImageFiles))
		for i, img := range input.ImageFiles {
			imageURLs[i] = img.Filename
		}
	}

	// Create UMKM entity
	now := s.clock.Now()
	u := &umkm.UMKM{
		ID:           uuid.New().String(),
		Name:         input.Name,
		Category:     input.Category,
		CategoryName: categoryName,
		Description:  input.Description,
		Owner:        input.Owner,
		Address:      input.Address,
		Phone:        input.Phone,
		Email:        input.Email,
		Website:      input.Website,
		Images:       imageURLs,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		// Clean up uploaded images on validation failure
		for _, img := range imageURLs {
			_ = s.fileHandler.Delete(ctx, img)
		}
		return nil, fmt.Errorf("invalid umkm data: %w", err)
	}

	// Persist UMKM
	if err := s.repo.Create(ctx, u); err != nil {
		// Clean up uploaded images on persistence failure
		for _, img := range imageURLs {
			_ = s.fileHandler.Delete(ctx, img)
		}
		return nil, fmt.Errorf("failed to create umkm: %w", err)
	}

	return u, nil
}

// GetByID retrieves a UMKM by ID.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.5, 11.2
func (s *Service) GetByID(ctx context.Context, id string) (*umkm.UMKM, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("umkm not found: %w", err)
	}

	return u, nil
}

// List retrieves paginated UMKM with optional filtering.
// Returns concrete UMKM structs and pagination metadata.
//
// Validates: Requirements 10.1, 10.2, 11.1, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListUMKMInput) ([]*umkm.UMKM, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve UMKM from repository with filters
	umkmList, total, err := s.repo.List(ctx, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list umkm: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return umkmList, paginationResult, nil
}

// Update updates an existing UMKM.
// Returns concrete UMKM struct.
//
// Validates: Requirements 10.7, 11.4
func (s *Service) Update(ctx context.Context, id string, input UpdateUMKMInput) (*umkm.UMKM, error) {
	// Validate the new category exists before mutating the UMKM.
	if _, err := s.categoryRepo.FindByID(ctx, input.Category); err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}

	// Retrieve existing UMKM
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("umkm not found: %w", err)
	}

	// Handle image updates
	imageURLs := input.Images

	// If new image URLs provided, delete old images and use new URLs
	if len(input.ImageFiles) > 0 {
		// Delete existing image files (best-effort)
		for _, oldImage := range u.Images {
			_ = s.fileHandler.Delete(ctx, oldImage)
		}
		// Handler has already saved new files (multipart) or passed URLs (JSON);
		// just collect them.
		imageURLs = make([]string, len(input.ImageFiles))
		for i, img := range input.ImageFiles {
			imageURLs[i] = img.Filename
		}
	}

	// Update fields
	u.Name = input.Name
	u.Category = input.Category
	u.Description = input.Description
	u.Owner = input.Owner
	u.Address = input.Address
	u.Phone = input.Phone
	u.Email = input.Email
	u.Website = input.Website
	u.Images = imageURLs
	u.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		// Clean up uploaded images on validation failure
		if len(input.ImageFiles) > 0 {
			for _, img := range imageURLs {
				_ = s.fileHandler.Delete(ctx, img)
			}
		}
		return nil, fmt.Errorf("invalid umkm data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, u); err != nil {
		// Clean up uploaded images on persistence failure
		if len(input.ImageFiles) > 0 {
			for _, img := range imageURLs {
				_ = s.fileHandler.Delete(ctx, img)
			}
		}
		return nil, fmt.Errorf("failed to update umkm: %w", err)
	}

	return u, nil
}

// Delete removes a UMKM.
//
// Validates: Requirements 10.8, 11.5
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve UMKM to ensure it exists and get its images
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("umkm not found: %w", err)
	}

	// Delete associated image files
	for _, image := range u.Images {
		_ = s.fileHandler.Delete(ctx, image)
	}

	// Delete UMKM from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete umkm: %w", err)
	}

	return nil
}
