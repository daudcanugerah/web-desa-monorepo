package auth

import (
	"time"
)

// TokenPair represents both access and refresh tokens issued during login.
// This is a concrete domain type returned by authentication operations.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// AccessToken represents a short-lived access token for API authentication.
// This is a concrete domain type returned by token refresh operations.
type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}

// ResetTokenInfo provides information about a password reset token's validity.
// This is a concrete domain type returned by token validation operations.
type ResetTokenInfo struct {
	Valid     bool
	ExpiresAt *time.Time
}

// ResetToken represents a password reset token entity in the domain.
// This is a concrete domain entity with validation rules.
type ResetToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	IPAddress string    `db:"ip_address"`
	UserAgent string    `db:"user_agent"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
}

// ResetMetadata contains security-related information about a password reset request.
// This is a concrete domain type used for audit trails.
type ResetMetadata struct {
	IPAddress string
	UserAgent string
}
