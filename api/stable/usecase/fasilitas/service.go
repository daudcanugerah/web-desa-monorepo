package fasilitas

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/fasilitas"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// Service implements Fasilitas (facility) management business logic.
// It accepts interfaces (Repository, FileHandler, CategoryLookup) and returns concrete structs (Fasilitas).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileHandler  FileHandler
	categoryRepo CategoryLookup
	clock        clock.Clock
}

// NewService creates a new Fasilitas management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileHandler FileHandler, clk clock.Clock, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileHandler:  fileHandler,
		categoryRepo: categoryRepo,
		clock:        clk,
	}
}

// CreateFasilitasInput represents the input for creating a Fasilitas
type CreateFasilitasInput struct {
	Name        string
	Category    *string
	Latitude    float64
	Longitude   float64
	Description *string
	ImageFiles  []ImageInput // For file uploads
}

// ImageInput represents an image file input
type ImageInput struct {
	Filename    string
	Content     io.Reader
	Size        int64
	ContentType string
}

// UpdateFasilitasInput represents the input for updating a Fasilitas
type UpdateFasilitasInput struct {
	Name        string
	Category    *string
	Latitude    float64
	Longitude   float64
	Description *string
	ImageFiles  []ImageInput // For file uploads
}

// ListFasilitasInput represents the input for listing Fasilitas
type ListFasilitasInput struct {
	BBox     *BoundingBox // Optional bounding box filter [minLon, minLat, maxLon, maxLat]
	Query    *string      // Optional search across name
	Category *string      // Optional UUID FK to fasilitas_categories
	Page     int
	Limit    int
}

// Create creates a new Fasilitas.
// Returns concrete Fasilitas struct.
//
// Validates: Requirements 9.3, 11.3
func (s *Service) Create(ctx context.Context, input CreateFasilitasInput) (*fasilitas.Fasilitas, error) {
	// Validate the category exists, if one is provided. The FK is enforced
	// at the DB level, but checking here provides a friendlier 400 response.
	var categoryName *string
	if input.Category != nil && *input.Category != "" {
		cat, err := s.categoryRepo.FindByID(ctx, *input.Category)
		if err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		name := cat.Name
		categoryName = &name
	}

	// Validate coordinates before creating entity
	if err := validateCoordinates(input.Latitude, input.Longitude); err != nil {
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	// Image URLs: handler has already saved files; we just store the URLs.
	imageURLs := make([]string, 0)
	if len(input.ImageFiles) > 0 {
		imageURLs = make([]string, len(input.ImageFiles))
		for i, img := range input.ImageFiles {
			imageURLs[i] = img.Filename
		}
	}

	// Create Fasilitas entity
	now := s.clock.Now()
	f := &fasilitas.Fasilitas{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Category:    input.Category,
		CategoryName: categoryName,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Description: input.Description,
		Images:      imageURLs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		// Clean up uploaded images on validation failure
		for _, img := range imageURLs {
			_ = s.fileHandler.Delete(ctx, img)
		}
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist Fasilitas
	if err := s.repo.Create(ctx, f); err != nil {
		// Clean up uploaded images on persistence failure
		for _, img := range imageURLs {
			_ = s.fileHandler.Delete(ctx, img)
		}
		return nil, fmt.Errorf("failed to create fasilitas: %w", err)
	}

	return f, nil
}

// GetByID retrieves a Fasilitas by ID.
// Returns concrete Fasilitas struct.
//
// Validates: Requirements 9.4, 11.2
func (s *Service) GetByID(ctx context.Context, id string) (*fasilitas.Fasilitas, error) {
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fasilitas not found: %w", err)
	}

	return f, nil
}

// List retrieves paginated Fasilitas with optional bounding box filtering.
// Returns concrete Fasilitas structs and pagination metadata.
//
// Validates: Requirements 9.1, 9.2, 11.1, 17.2, 17.3
func (s *Service) List(ctx context.Context, input ListFasilitasInput) ([]*fasilitas.Fasilitas, pagination.Result, error) {
	// Validate bounding box if provided
	if input.BBox != nil {
		if err := validateBoundingBox(input.BBox); err != nil {
			return nil, pagination.Result{}, fmt.Errorf("invalid bounding box: %w", err)
		}
	}

	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve Fasilitas from repository with filters
	fasilitasList, total, err := s.repo.List(ctx, input.BBox, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list fasilitas: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return fasilitasList, paginationResult, nil
}

// Update updates an existing Fasilitas.
// Returns concrete Fasilitas struct.
//
// Validates: Requirements 9.5, 11.4
func (s *Service) Update(ctx context.Context, id string, input UpdateFasilitasInput) (*fasilitas.Fasilitas, error) {
	// Validate the new category exists, if one is provided.
	var newCategoryName *string
	if input.Category != nil && *input.Category != "" {
		cat, err := s.categoryRepo.FindByID(ctx, *input.Category)
		if err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		name := cat.Name
		newCategoryName = &name
	}

	// Validate coordinates before updating
	if err := validateCoordinates(input.Latitude, input.Longitude); err != nil {
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	// Retrieve existing Fasilitas
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fasilitas not found: %w", err)
	}

	// Handle image updates
	imageURLs := f.Images

	// If new image files are provided, delete old images and use new URLs
	if len(input.ImageFiles) > 0 {
		// Delete existing image files (best-effort)
		for _, oldImage := range f.Images {
			_ = s.fileHandler.Delete(ctx, oldImage)
		}
		// Handler has already saved new files; just collect URLs.
		imageURLs = make([]string, len(input.ImageFiles))
		for i, img := range input.ImageFiles {
			imageURLs[i] = img.Filename
		}
	}

	// Update fields
	f.Name = input.Name
	f.Category = input.Category
	if newCategoryName != nil {
		f.CategoryName = newCategoryName
	}
	f.Latitude = input.Latitude
	f.Longitude = input.Longitude
	f.Description = input.Description
	f.Images = imageURLs
	f.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		// Clean up uploaded images on validation failure
		if len(input.ImageFiles) > 0 {
			for _, img := range imageURLs {
				_ = s.fileHandler.Delete(ctx, img)
			}
		}
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, f); err != nil {
		// Clean up uploaded images on persistence failure
		if len(input.ImageFiles) > 0 {
			for _, img := range imageURLs {
				_ = s.fileHandler.Delete(ctx, img)
			}
		}
		return nil, fmt.Errorf("failed to update fasilitas: %w", err)
	}

	return f, nil
}

// Delete removes a Fasilitas.
//
// Validates: Requirements 9.6, 11.5
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve Fasilitas to ensure it exists and get its images
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fasilitas not found: %w", err)
	}

	// Delete associated image files
	for _, image := range f.Images {
		_ = s.fileHandler.Delete(ctx, image)
	}

	// Delete Fasilitas from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete fasilitas: %w", err)
	}

	return nil
}

// validateCoordinates validates latitude and longitude values
func validateCoordinates(lat, lon float64) error {
	if lat < -90.0 || lat > 90.0 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees")
	}

	if lon < -180.0 || lon > 180.0 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees")
	}

	return nil
}

// validateBoundingBox validates bounding box coordinates
func validateBoundingBox(bbox *BoundingBox) error {
	if bbox == nil {
		return nil
	}

	// Validate individual coordinates
	if err := validateCoordinates(bbox.MinLat, bbox.MinLon); err != nil {
		return fmt.Errorf("invalid min coordinates: %w", err)
	}

	if err := validateCoordinates(bbox.MaxLat, bbox.MaxLon); err != nil {
		return fmt.Errorf("invalid max coordinates: %w", err)
	}

	// Validate that min < max
	if bbox.MinLat >= bbox.MaxLat {
		return fmt.Errorf("minLat must be less than maxLat")
	}

	if bbox.MinLon >= bbox.MaxLon {
		return fmt.Errorf("minLon must be less than maxLon")
	}

	return nil
}

// RemoveImage removes a specific image from the Fasilitas images array by index.
// Returns the updated Fasilitas struct.
//
// Validates: Requirements 11.4
func (s *Service) RemoveImage(ctx context.Context, id string, imageIndex int) (*fasilitas.Fasilitas, error) {
	// Retrieve existing Fasilitas
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fasilitas not found: %w", err)
	}

	// Validate image index
	if imageIndex < 0 || imageIndex >= len(f.Images) {
		return nil, fmt.Errorf("image index out of range: index %d, total images %d", imageIndex, len(f.Images))
	}

	// Remove image from array
	f.Images = append(f.Images[:imageIndex], f.Images[imageIndex+1:]...)
	f.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, fmt.Errorf("failed to update fasilitas: %w", err)
	}

	return f, nil
}
