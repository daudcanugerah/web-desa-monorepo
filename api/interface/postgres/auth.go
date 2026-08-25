package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/auth"
	"webdesa/api/domain/user"
	authUsecase "webdesa/api/usecase/auth"
)

// AuthRepository implements usecase/auth.Repository interface.
// This follows the Dependency Rule: interface/postgres → usecase → domain
// The interface is defined in usecase/auth where it's USED, not here where it's implemented.
type AuthRepository struct {
	db *sqlx.DB
}

// NewAuthRepository creates a new AuthRepository instance
func NewAuthRepository(db *sqlx.DB) authUsecase.Repository {
	return &AuthRepository{db: db}
}

// FindUserByEmail retrieves a user by email address
// This implements the UserRepository interface from usecase/auth
// Returns concrete User struct
// Validates: Requirements 3.1
func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User

	query := `
		SELECT id, name, email, hashed_password, created_at, updated_at
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

// CreateResetToken creates a new password reset token with metadata
// Generates UUID for token ID and stores metadata for security auditing
// Returns concrete error
// Validates: Requirements 3.1, 3.3
func (r *AuthRepository) CreateResetToken(ctx context.Context, userID, token string, expiresAt time.Time, metadata auth.ResetMetadata) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, ip_address, user_agent, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

	tokenID := uuid.New().String()

	_, err := r.db.ExecContext(ctx, query,
		tokenID,
		userID,
		token,
		expiresAt,
		metadata.IPAddress,
		metadata.UserAgent,
		false, // used = false initially
	)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to create reset token: %w", err))
	}

	return nil
}

// FindResetToken retrieves a reset token by its token string
// Returns concrete ResetToken struct with all fields
// Validates: Requirements 3.5
func (r *AuthRepository) FindResetToken(ctx context.Context, token string) (*auth.ResetToken, error) {
	var rt auth.ResetToken

	query := `
		SELECT id, user_id, token, expires_at, ip_address, user_agent, used, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	err := r.db.GetContext(ctx, &rt, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("reset token not found: %s", token))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to find reset token: %w", err))
	}

	return &rt, nil
}

// InvalidateUserResetTokens marks all reset tokens for a user as used
// This ensures only 1 active reset token per user at any time
// Validates: Requirements 3.4
func (r *AuthRepository) InvalidateUserResetTokens(ctx context.Context, userID string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE user_id = $1 AND used = false
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to invalidate reset tokens: %w", err))
	}

	return nil
}

// UpdatePassword updates a user's hashed password
// Sets updated_at timestamp automatically
// Validates: Requirements 3.6
func (r *AuthRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	query := `
		UPDATE users
		SET hashed_password = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to update password: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to get rows affected: %w", err))
	}

	if rowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("user not found: %s", userID))
	}

	return nil
}

// CountResetRequestsByEmail counts reset requests for an email since a given time
// Used for rate limiting (5 requests per hour per email)
// Validates: Requirements 3.2
func (r *AuthRepository) CountResetRequestsByEmail(ctx context.Context, email string, since time.Time) (int, error) {
	var count int

	// First get the user ID for this email
	var userID string
	userQuery := `SELECT id FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &userID, userQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist, return 0 count
			return 0, nil
		}
		return 0, errtrace.Wrap(fmt.Errorf("failed to find user: %w", err))
	}

	// Count UNUSED reset requests for this user since the time window.
	// Used tokens must not count toward the rate limit, otherwise legitimate
	// users get locked out after a few successful resets.
	query := `
		SELECT COUNT(*)
		FROM password_reset_tokens
		WHERE user_id = $1 AND created_at >= $2 AND used = false
	`

	err = r.db.GetContext(ctx, &count, query, userID, since)
	if err != nil {
		return 0, errtrace.Wrap(fmt.Errorf("failed to count reset requests: %w", err))
	}

	return count, nil
}
