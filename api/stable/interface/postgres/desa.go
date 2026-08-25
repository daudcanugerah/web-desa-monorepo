package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"braces.dev/errtrace"
	"github.com/jmoiron/sqlx"

	"webdesa/api/domain/desa"
	desaUsecase "webdesa/api/usecase/desa"
)

const desaSettingKey = "desa_profile"

// DesaRepository implements usecase/desa.Repository using the settings table.
// The village profile is stored as JSON under the key "desa_profile".
// This follows the Dependency Rule: interface/postgres → usecase → domain.
type DesaRepository struct {
	db *sqlx.DB
}

// NewDesaRepository creates a new DesaRepository instance.
func NewDesaRepository(db *sqlx.DB) desaUsecase.Repository {
	return &DesaRepository{db: db}
}

// settingRow is the DB row shape for the settings table.
type settingRow struct {
	Key       string    `db:"key"`
	Value     []byte    `db:"value"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Get retrieves the village profile from the settings table.
// Returns an empty Desa (zero-value name) if the key does not exist yet.
func (r *DesaRepository) Get(ctx context.Context) (*desa.Desa, error) {
	var row settingRow
	err := r.db.GetContext(ctx, &row,
		`SELECT key, value, updated_at FROM settings WHERE key = $1`, desaSettingKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errtrace.Wrap(fmt.Errorf("village profile not found"))
		}
		return nil, errtrace.Wrap(fmt.Errorf("failed to get village profile: %w", err))
	}

	var d desa.Desa
	if err := json.Unmarshal(row.Value, &d); err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("failed to decode village profile: %w", err))
	}
	d.UpdatedAt = row.UpdatedAt
	return &d, nil
}

// Upsert inserts or updates the village profile JSON in the settings table.
func (r *DesaRepository) Upsert(ctx context.Context, d *desa.Desa) error {
	value, err := json.Marshal(d)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to encode village profile: %w", err))
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE
		  SET value = EXCLUDED.value,
		      updated_at = EXCLUDED.updated_at
	`, desaSettingKey, value)
	if err != nil {
		return errtrace.Wrap(fmt.Errorf("failed to upsert village profile: %w", err))
	}

	return nil
}
