package infographic

import (
	"time"

	"github.com/google/uuid"
)

// AccessLog represents a single token-issuance event for an infographic.
// Used for audit trails and abuse detection.
type AccessLog struct {
	ID             uuid.UUID `db:"id"`
	InfographicID  uuid.UUID `db:"infographic_id"`
	ComponentID    int64     `db:"component_id"`
	ComponentType  string    `db:"component_type"`
	Endpoint       string    `db:"endpoint"`
	IPAddress      string    `db:"ip_address"` // INET text representation
	UserAgent      string    `db:"user_agent"`
	Referer        string    `db:"referer"`
	TokenIssuedAt  time.Time `db:"token_issued_at"`
	TokenExpiresAt time.Time `db:"token_expires_at"`
	CreatedAt      time.Time `db:"created_at"`
}
