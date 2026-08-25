package infographic

import (
	"context"
	"time"

	"webdesa/api/domain/infographic"

	"github.com/google/uuid"
)

// AccessLogRepository persists infographic access logs.
type AccessLogRepository interface {
	// Create persists a new access log entry.
	Create(ctx context.Context, log *infographic.AccessLog) error

	// ListByInfographic returns the most recent access logs for an infographic,
	// paginated. Use cursor-based pagination for high-volume lists.
	ListByInfographic(ctx context.Context, infographicID uuid.UUID, offset, limit int) ([]*infographic.AccessLog, int, error)

	// ListByIP returns access logs for a given IP address, paginated.
	ListByIP(ctx context.Context, ip string, offset, limit int) ([]*infographic.AccessLog, int, error)

	// CountByInfographicAndIPSince counts access logs for a single
	// (infographic, IP) tuple within a time window. Used by the per-component
	// rate limiter.
	CountByInfographicAndIPSince(ctx context.Context, infographicID uuid.UUID, ip string, since time.Time) (int, error)

	// DeleteOlderThan removes audit records older than the given time.
	// Used by the retention job.
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
