package user

import (
	"context"
	"fmt"

	"webdesa/api/domain/user"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"
	"webdesa/api/pkg/password"
	galleryUsecase "webdesa/api/usecase/gallery"

	"github.com/google/uuid"
)

// Service implements user management business logic.
// It accepts interfaces (Repository, FileStore) and returns concrete structs (User).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo      Repository
	fileStore galleryUsecase.FileStore
	clock     clock.Clock
}

// NewService creates a new user management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock) *Service {
	return &Service{
		repo:      repo,
		fileStore: fileStore,
		clock:     clk,
	}
}

const featureSlug = galleryUsecase.FeatureUser

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Name     string
	Email    string
	Password string
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name  string
	Email string
}

// UpdatePasswordInput represents the input for updating a password
type UpdatePasswordInput struct {
	OldPassword string
	NewPassword string
}

// Create creates a new user account.
// Returns concrete User struct.
//
// Validates: Requirements 5.1
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*user.User, error) {
	// Check if email already exists
	existingUser, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password using bcrypt with cost factor 12
	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	now := s.clock.Now()
	u := &user.User{
		ID:             uuid.New().String(),
		Name:           input.Name,
		Email:          input.Email,
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Persist user
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

// List retrieves paginated users.
// Returns concrete User structs and pagination metadata.
//
// Validates: Requirements 5.2, 17.2, 17.3
func (s *Service) List(ctx context.Context, page, limit int) ([]*user.User, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(page, limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve users from repository
	users, total, err := s.repo.List(ctx, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list users: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(page, validatedLimit, total)

	return users, paginationResult, nil
}

// GetByID retrieves a user by ID.
// Returns concrete User struct.
//
// Validates: Requirements 5.3
func (s *Service) GetByID(ctx context.Context, id string) (*user.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return u, nil
}

// Update updates an existing user.
// Returns concrete User struct.
//
// Validates: Requirements 5.4
func (s *Service) Update(ctx context.Context, id string, input UpdateUserInput) (*user.User, error) {
	// Retrieve existing user
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if email is being changed and if new email already exists
	if input.Email != u.Email {
		existingUser, err := s.repo.FindByEmail(ctx, input.Email)
		if err == nil && existingUser != nil {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Update fields
	u.Name = input.Name
	u.Email = input.Email
	u.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := u.Validate(); err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return u, nil
}

// Delete removes a user account and cascades to user_roles.
//
// Validates: Requirements 5.5
func (s *Service) Delete(ctx context.Context, id string) error {
	// Check if user exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Delete user (cascade to user_roles handled by database)
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password.
// For non-admin users, requires old password verification.
// For admin users, old password is not required.
//
// Validates: Requirements 5.6, 5.7
func (s *Service) UpdatePassword(ctx context.Context, id string, input UpdatePasswordInput, isAdmin bool) error {
	// Retrieve existing user
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// If not admin, verify old password
	if !isAdmin {
		if !password.Verify(input.OldPassword, u.HashedPassword) {
			return fmt.Errorf("invalid old password")
		}
	}

	// Hash new password using bcrypt with cost factor 12
	hashedPassword, err := password.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, id, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateProfileImageMediaID attaches a media already uploaded via
// POST /users/upload-media or /users/me/upload-media to the user, removing
// the previous avatar. The media must exist (validated via FileStore.Open).
func (s *Service) UpdateProfileImageMediaID(ctx context.Context, userID, mediaID string) (*user.User, error) {
	if _, err := s.fileStore.Open(ctx, mediaID); err != nil {
		return nil, fmt.Errorf("invalid profile_image_media_id: %w", err)
	}

	// Retrieve existing user
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	oldMediaID := ""
	if u.ProfileImageMediaID != nil {
		oldMediaID = *u.ProfileImageMediaID
	}

	u.ProfileImageMediaID = &mediaID
	u.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateProfileImage(ctx, userID, &mediaID); err != nil {
		return nil, fmt.Errorf("failed to update profile image: %w", err)
	}

	// Delete the previous avatar only when it actually changed; re-attaching
	// the current media id must not delete it.
	if oldMediaID != "" && oldMediaID != mediaID {
		_ = s.fileStore.Delete(ctx, oldMediaID)
	}

	// Refresh from DB so callers see the persisted state.
	u, err = s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	return u, nil
}

// UpdateProfileImage saves a new avatar through the gallery FileStore,
// persists the new media_id on the user, and removes the previous avatar
// from gallery storage. Returns the refreshed user.
//
// Validates: Requirements 5.8
func (s *Service) UpdateProfileImage(ctx context.Context, userID string, in galleryUsecase.FileInput) (*user.User, error) {
	// Retrieve existing user
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	oldMediaID := ""
	if u.ProfileImageMediaID != nil {
		oldMediaID = *u.ProfileImageMediaID
	}

	saved, err := s.fileStore.SaveImage(ctx, featureSlug, in)
	if err != nil {
		return nil, fmt.Errorf("failed to save profile image: %w", err)
	}

	mediaID := saved.MediaID
	u.ProfileImageMediaID = &mediaID
	u.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateProfileImage(ctx, userID, &mediaID); err != nil {
		_ = s.fileStore.Delete(ctx, mediaID)
		return nil, fmt.Errorf("failed to update profile image: %w", err)
	}

	if oldMediaID != "" {
		_ = s.fileStore.Delete(ctx, oldMediaID)
	}

	// Refresh from DB so callers see the persisted state.
	u, err = s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	return u, nil
}
