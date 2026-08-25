package auth

import (
	"context"

	"webdesa/api/domain/user"
)

// UserRepository defines what the auth service needs from user persistence.
// This interface is defined HERE in the usecase layer where it is USED,
// not in the domain or interface layers (following Go best practices).
//
// The interface/postgres package will implement this interface.
type UserRepository interface {
	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*user.User, error)
}
