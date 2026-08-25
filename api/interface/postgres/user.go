package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/user"
	userUsecase "webdesa/api/usecase/user"
)

// UserRepository implements usecase/user.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/user where it's USED, not here where it's implemented.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sqlx.DB) userUsecase.Repository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
// Generates UUID, sets timestamps, and inserts the user record
// Validates: Requirements 5.1, 20.2, 20.3
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	// Generate UUID if not provided
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	query := `
		INSERT INTO users (id, name, email, hashed_password, profile_image_media_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		u.ID,
		u.Name,
		u.Email,
		u.HashedPassword,
		u.ProfileImageMediaID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create user: %w", err))
	}

	return nil
}

// FindByID retrieves a user by ID
// Returns error if user not found
// Validates: Requirements 5.3
func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User

	query := `
		SELECT id, name, email, hashed_password, profile_image_media_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("user not found: %s", id))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find user by ID: %w", err))
	}

	return &u, nil
}

// FindByEmail retrieves a user by email
// Returns error if user not found
// Validates: Requirements 5.3
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User

	query := `
		SELECT id, name, email, hashed_password, profile_image_media_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("user not found: %s", email))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find user by email: %w", err))
	}

	return &u, nil
}

// List retrieves paginated users
// Returns users slice, total count, and error
// Uses LIMIT/OFFSET for pagination
// Validates: Requirements 5.2, 17.2, 20.5
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to count users: %w", err))
	}

	// Get paginated users
	var users []*user.User
	query := `
		SELECT id, name, email, hashed_password, profile_image_media_id, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err = r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, 0, errtrace.Wrap(fmt.Errorf("failed to list users: %w", err))
	}

	return users, total, nil
}

// Update updates an existing user
// Sets updated_at timestamp automatically
// Validates: Requirements 5.4, 20.3
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, profile_image_media_id = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		u.Name,
		u.Email,
		u.ProfileImageMediaID,
		u.ID,
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update user: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user not found: %s", u.ID))
	}

	return nil
}

// Delete removes a user and cascades to user_roles
// CASCADE delete is handled by database foreign key constraint
// Validates: Requirements 5.5, 20.4
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to delete user: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user not found: %s", id))
	}

	return nil
}

// UpdatePassword updates a user's password
// Sets updated_at timestamp automatically
// Validates: Requirements 5.6, 5.7, 20.3
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	query := `
		UPDATE users
		SET hashed_password = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update password: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user not found: %s", id))
	}

	return nil
}

// UpdateProfileImage updates a user's avatar media id.
// Sets updated_at timestamp automatically.
// Validates: Requirements 5.8, 20.3
func (r *UserRepository) UpdateProfileImage(ctx context.Context, id string, profileImageMediaID *string) error {
	query := `
		UPDATE users
		SET profile_image_media_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, profileImageMediaID, id)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update profile image: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user not found: %s", id))
	}

	return nil
}