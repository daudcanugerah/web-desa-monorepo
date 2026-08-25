package auth

import (
	"context"
	"time"

	"webdesa/api/domain/auth"
)

// Repository defines what the auth service needs from persistence layer.
// This interface is defined HERE in the usecase layer where it is USED,
// not in the domain or interface layers (following Go best practices).
//
// The interface/postgres package will implement this interface.
type Repository interface {
	// CreateResetToken creates a new password reset token with metadata
	CreateResetToken(ctx context.Context, userID, token string, expiresAt time.Time, metadata auth.ResetMetadata) error

	// FindResetToken retrieves a reset token by its token string
	FindResetToken(ctx context.Context, token string) (*auth.ResetToken, error)

	// InvalidateUserResetTokens marks all reset tokens for a user as used
	InvalidateUserResetTokens(ctx context.Context, userID string) error

	// UpdatePassword updates a user's hashed password
	UpdatePassword(ctx context.Context, userID, hashedPassword string) error

	// CountResetRequestsByEmail counts reset requests for an email since a given time
	CountResetRequestsByEmail(ctx context.Context, email string, since time.Time) (int, error)
}
