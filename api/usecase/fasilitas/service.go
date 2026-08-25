package fasilitas

import (
	"context"
	"fmt"
	"io"

	"webdesa/api/domain/fasilitas"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"
	galleryUsecase "webdesa/api/usecase/gallery"

	"github.com/google/uuid"
)

// BulkLimits exposes the gallery bulk-upload caps (max files per record,
// max total bytes) enforced on multi-image create/update flows. The
// config.GalleryConfig type satisfies this interface.
type BulkLimits interface {
	GetBulkUploadMaxFiles() int
	GetBulkUploadMaxTotalMB() int
}

// Service implements Fasilitas (facility) management business logic.
// It accepts interfaces (Repository, FileStore, CategoryLookup) and returns concrete structs (Fasilitas).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileStore    galleryUsecase.FileStore
	categoryRepo CategoryLookup
	clock        clock.Clock
	limits       BulkLimits
}

// NewService creates a new Fasilitas management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock, categoryRepo CategoryLookup, limits BulkLimits) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
		categoryRepo: categoryRepo,
		clock:        clk,
		limits:       limits,
	}
}

// enforceBulkLimits rejects a multi-image batch that exceeds the gallery
// bulk caps (20 files / 250MB total by default). Refs and inline files are
// counted together because together they form the record's image set.
func (s *Service) enforceBulkLimits(refs []string, files []ImageInput) error {
	maxFiles := s.limits.GetBulkUploadMaxFiles()
	if len(refs)+len(files) > maxFiles {
		return fmt.Errorf("%w: max %d images per record", galleryUsecase.ErrBulkTooManyFiles, maxFiles)
	}
	var total int64
	for _, f := range files {
		total += f.Size
	}
	if total > int64(s.limits.GetBulkUploadMaxTotalMB())*1024*1024 {
		return fmt.Errorf("%w: max %dMB total", galleryUsecase.ErrBulkTotalTooLarge, s.limits.GetBulkUploadMaxTotalMB())
	}
	return nil
}

const featureSlug = galleryUsecase.FeatureFasilitas

// CreateFasilitasInput represents the input for creating a Fasilitas
type CreateFasilitasInput struct {
	Name        string
	Category    *string
	Latitude    float64
	Longitude   float64
	Description *string
	ImageFiles  []ImageInput // For file uploads
	// ImagesMediaIDs references pre-uploaded gallery media (POST
	// /fasilitas/upload-media), merged after ImageFiles.
	ImagesMediaIDs []string
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
	// ImagesMediaIDs references pre-uploaded gallery media (POST
	// /fasilitas/upload-media). Replaces the image set on update.
	ImagesMediaIDs []string
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

	// Validate pre-uploaded media refs before touching the file store.
	if err := s.validateMediaRefs(ctx, input.ImagesMediaIDs); err != nil {
		return nil, err
	}

	// Enforce gallery bulk caps (file count, total bytes) on the combined
	// image set before any file is saved.
	if err := s.enforceBulkLimits(input.ImagesMediaIDs, input.ImageFiles); err != nil {
		return nil, err
	}

	// Persist uploaded images via the gallery FileStore. The handler has
	// already parsed the multipart bodies; the service is the only
	// place that talks to FileStore. Cleanup on failure is restricted to
	// the files created here — pre-uploaded refs are left untouched.
	savedIDs, err := s.saveImageFiles(ctx, input.ImageFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to upload images: %w", err)
	}
	mediaIDs := append(input.ImagesMediaIDs, savedIDs...)

	// Create Fasilitas entity
	now := s.clock.Now()
	f := &fasilitas.Fasilitas{
		ID:             uuid.New().String(),
		Name:           input.Name,
		Category:       input.Category,
		CategoryName:   categoryName,
		Latitude:       input.Latitude,
		Longitude:      input.Longitude,
		Description:    input.Description,
		ImagesMediaIDs: mediaIDs,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		for _, id := range savedIDs {
			_ = s.fileStore.Delete(ctx, id)
		}
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist Fasilitas
	if err := s.repo.Create(ctx, f); err != nil {
		for _, id := range savedIDs {
			_ = s.fileStore.Delete(ctx, id)
		}
		return nil, fmt.Errorf("failed to create fasilitas: %w", err)
	}

	return f, nil
}

// saveImageFiles iterates the inputs through the gallery FileStore and
// returns the resulting media UUIDs in input order. Files are validated
// upfront (per-file type/size caps via FileStore.ValidateFiles) so nothing
// is persisted for a batch with a bad member. Any partial-failure state is
// cleaned up before returning.
func (s *Service) saveImageFiles(ctx context.Context, inputs []ImageInput) ([]string, error) {
	if len(inputs) == 0 {
		return []string{}, nil
	}

	if errs := s.fileStore.ValidateFiles(ctx, featureSlug, toFileInputs(inputs)); len(errs) > 0 {
		return nil, fmt.Errorf("invalid image: %w", errs[0])
	}

	ids := make([]string, 0, len(inputs))
	for i, in := range inputs {
		saved, err := s.fileStore.SaveImage(ctx, featureSlug, galleryUsecase.FileInput{
			OriginalName: in.Filename,
			Content:      in.Content,
			Size:         in.Size,
			ContentType:  in.ContentType,
		})
		if err != nil {
			for _, id := range ids {
				_ = s.fileStore.Delete(ctx, id)
			}
			return nil, fmt.Errorf("failed to save image at index %d: %w", i, err)
		}
		ids = append(ids, saved.MediaID)
	}
	return ids, nil
}

func toFileInputs(files []ImageInput) []galleryUsecase.FileInput {
	out := make([]galleryUsecase.FileInput, len(files))
	for i, f := range files {
		out[i] = galleryUsecase.FileInput{
			OriginalName: f.Filename,
			ContentType:  f.ContentType,
			Size:         f.Size,
		}
	}
	return out
}

// validateMediaRefs checks that every pre-uploaded media ref still exists
// in the gallery. Refs were validated at upload time; this guards against
// GC'd or hand-crafted ids.
func (s *Service) validateMediaRefs(ctx context.Context, refs []string) error {
	for _, ref := range refs {
		if _, err := s.fileStore.Open(ctx, ref); err != nil {
			return fmt.Errorf("invalid media id %q: %w", ref, err)
		}
	}
	return nil
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

	f.ImagesMediaIDs, err = galleryUsecase.PruneDanglingMediaIDs(ctx, s.fileStore, f.ImagesMediaIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to prune media ids: %w", err)
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

	// Prune dangling media ids (Task 5.1) — deleted gallery rows must
	// never appear in API responses.
	for _, f := range fasilitasList {
		f.ImagesMediaIDs, err = galleryUsecase.PruneDanglingMediaIDs(ctx, s.fileStore, f.ImagesMediaIDs)
		if err != nil {
			return nil, pagination.Result{}, fmt.Errorf("failed to prune media ids: %w", err)
		}
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

	// Handle image updates: when new files or pre-uploaded media refs are
	// supplied, save/attach them through the gallery FileStore and queue
	// the previous media IDs for deletion. Failure cleanup is restricted
	// to the files created here — pre-uploaded refs are left untouched.
	var oldMediaIDs []string
	var createdIDs []string
	switch {
	case len(input.ImagesMediaIDs) > 0:
		if err := s.validateMediaRefs(ctx, input.ImagesMediaIDs); err != nil {
			return nil, err
		}
		if err := s.enforceBulkLimits(input.ImagesMediaIDs, nil); err != nil {
			return nil, err
		}
		oldMediaIDs = append([]string{}, f.ImagesMediaIDs...)
		f.ImagesMediaIDs = input.ImagesMediaIDs
	case len(input.ImageFiles) > 0:
		if err := s.enforceBulkLimits(nil, input.ImageFiles); err != nil {
			return nil, err
		}
		oldMediaIDs = append([]string{}, f.ImagesMediaIDs...)

		ids, err := s.saveImageFiles(ctx, input.ImageFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to upload images: %w", err)
		}
		createdIDs = ids
		f.ImagesMediaIDs = ids
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
	f.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		for _, id := range createdIDs {
			_ = s.fileStore.Delete(ctx, id)
		}
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, f); err != nil {
		for _, id := range createdIDs {
			_ = s.fileStore.Delete(ctx, id)
		}
		return nil, fmt.Errorf("failed to update fasilitas: %w", err)
	}

	// Best-effort cleanup of the previous gallery media after a successful
	// replacement.
	for _, id := range oldMediaIDs {
		_ = s.fileStore.Delete(ctx, id)
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

	// Delete associated gallery media rows. Best-effort: the DB row is
	// the source of truth, media cleanup failures are logged by the
	// FileStore implementation.
	for _, mediaID := range f.ImagesMediaIDs {
		_ = s.fileStore.Delete(ctx, mediaID)
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

// RemoveImage removes a specific image from the Fasilitas images_media_ids
// array by index. The corresponding gallery media row is also deleted via
// the FileStore. Returns the updated Fasilitas struct.
//
// Validates: Requirements 11.4
func (s *Service) RemoveImage(ctx context.Context, id string, imageIndex int) (*fasilitas.Fasilitas, error) {
	// Retrieve existing Fasilitas
	f, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fasilitas not found: %w", err)
	}

	// Validate image index
	if imageIndex < 0 || imageIndex >= len(f.ImagesMediaIDs) {
		return nil, fmt.Errorf("image index out of range: index %d, total images %d", imageIndex, len(f.ImagesMediaIDs))
	}

	// Capture the media ID before splicing so we can delete it after a
	// successful DB update.
	removedID := f.ImagesMediaIDs[imageIndex]

	// Remove image from array
	f.ImagesMediaIDs = append(f.ImagesMediaIDs[:imageIndex], f.ImagesMediaIDs[imageIndex+1:]...)
	f.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("invalid fasilitas data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, fmt.Errorf("failed to update fasilitas: %w", err)
	}

	// Best-effort gallery media cleanup after the DB row was updated.
	_ = s.fileStore.Delete(ctx, removedID)

	return f, nil
}
