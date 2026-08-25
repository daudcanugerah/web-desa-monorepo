package user

import (
	"context"

	"webdesa/api/domain/user"
)

// Repository defines the persistence interface for user management.
// This interface is defined in the usecase layer (where it's USED), not in the domain layer.
// This follows Go best practices: "interfaces belong in the package that uses them."
//
// The implementation will be in interface/repository/user_postgres.go, which depends on this interface.
// This maintains the Dependency Rule: interface/postgres → usecase → domain
type Repository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, u *user.User) error

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id string) (*user.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*user.User, error)

	// List retrieves paginated users
	// Returns users slice, total count, and error
	List(ctx context.Context, offset, limit int) ([]*user.User, int, error)

	// Update updates an existing user
	Update(ctx context.Context, u *user.User) error

	// Delete removes a user and cascades to user_roles
	Delete(ctx context.Context, id string) error

	// UpdatePassword updates a user's password
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error

	// UpdateProfile updates a user's name and profile image
	UpdateProfile(ctx context.Context, id string, name string, profileImageURL *string) error
}
