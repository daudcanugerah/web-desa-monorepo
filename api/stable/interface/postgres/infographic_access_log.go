package postgres

import (
	"context"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/infographic"
)

// InfographicAccessLogRepository implements infographic.AccessLogRepository
// against PostgreSQL.
type InfographicAccessLogRepository struct {
	db *sqlx.DB
}

// NewInfographicAccessLogRepository creates a new repository.
func NewInfographicAccessLogRepository(db *sqlx.DB) *InfographicAccessLogRepository {
	return &InfographicAccessLogRepository{db: db}
}

func (r *InfographicAccessLogRepository) Create(ctx context.Context, log *infographic.AccessLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	if log.TokenIssuedAt.IsZero() {
		log.TokenIssuedAt = now
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO infographic_access_log
		    (id, infographic_id, component_id, component_type, endpoint,
		     ip_address, user_agent, referer, token_issued_at, token_expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		log.ID, log.InfographicID, log.ComponentID, log.ComponentType, log.Endpoint,
		nullIfEmptyString(log.IPAddress), nullIfEmptyString(log.UserAgent), nullIfEmptyString(log.Referer),
		log.TokenIssuedAt, log.TokenExpiresAt, log.CreatedAt,
	)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *InfographicAccessLogRepository) ListByInfographic(ctx context.Context, infographicID uuid.UUID, offset, limit int) ([]*infographic.AccessLog, int, error) {
	var logs []*infographic.AccessLog
	err := r.db.SelectContext(ctx, &logs, `
		SELECT id, infographic_id, component_id, component_type, endpoint,
		       COALESCE(ip_address::text, '') AS ip_address,
		       COALESCE(user_agent, '') AS user_agent,
		       COALESCE(referer, '') AS referer,
		       token_issued_at, token_expires_at, created_at
		FROM infographic_access_log
		WHERE infographic_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, infographicID, limit, offset)
	if err != nil {
		return nil, 0, errtrace.Wrap(err)
	}
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM infographic_access_log WHERE infographic_id = $1`,
		infographicID); err != nil {
		return nil, 0, errtrace.Wrap(err)
	}
	return logs, total, nil
}

func (r *InfographicAccessLogRepository) ListByIP(ctx context.Context, ip string, offset, limit int) ([]*infographic.AccessLog, int, error) {
	var logs []*infographic.AccessLog
	err := r.db.SelectContext(ctx, &logs, `
		SELECT id, infographic_id, component_id, component_type, endpoint,
		       COALESCE(ip_address::text, '') AS ip_address,
		       COALESCE(user_agent, '') AS user_agent,
		       COALESCE(referer, '') AS referer,
		       token_issued_at, token_expires_at, created_at
		FROM infographic_access_log
		WHERE ip_address = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, ip, limit, offset)
	if err != nil {
		return nil, 0, errtrace.Wrap(err)
	}
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM infographic_access_log WHERE ip_address = $1`, ip); err != nil {
		return nil, 0, errtrace.Wrap(err)
	}
	return logs, total, nil
}

func (r *InfographicAccessLogRepository) CountByInfographicAndIPSince(ctx context.Context, infographicID uuid.UUID, ip string, since time.Time) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, `
		SELECT COUNT(*) FROM infographic_access_log
		WHERE infographic_id = $1 AND ip_address = $2 AND created_at >= $3`,
		infographicID, ip, since)
	if err != nil {
		return 0, errtrace.Wrap(err)
	}
	return n, nil
}

func (r *InfographicAccessLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM infographic_access_log WHERE created_at < $1`, before)
	if err != nil {
		return 0, errtrace.Wrap(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, errtrace.Wrap(err)
	}
	return rows, nil
}

func nullIfEmptyIP(ip []byte) interface{} {
	if len(ip) == 0 {
		return nil
	}
	return ip
}

func nullIfEmptyString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
