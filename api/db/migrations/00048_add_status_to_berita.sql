-- Migration 00048: add status column to berita (active/inactive).
--
-- Mirrors the banner status model: public lists only expose 'active'
-- rows, admins can filter and toggle via PATCH /berita/{id}/status.

-- +goose Up
ALTER TABLE berita
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

ALTER TABLE berita
    ADD CONSTRAINT berita_status_check CHECK (status IN ('active', 'inactive'));

CREATE INDEX IF NOT EXISTS idx_berita_status_created_at
    ON berita (status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_berita_status_created_at;
ALTER TABLE berita DROP CONSTRAINT IF EXISTS berita_status_check;
ALTER TABLE berita DROP COLUMN IF EXISTS status;
