-- Migration 00046: track full vs db-only backup kind (Task 5.2)
--
-- Phase 5 of draft/002-upload-migration. Backups produced by
-- `backup create --with-files` are tar.gz archives that bundle the
-- pg_dump output plus the on-disk uploads/ tree; the metadata column
-- lets the restore command tell the two apart.

-- +goose Up
ALTER TABLE backups ADD COLUMN IF NOT EXISTS kind VARCHAR(16) NOT NULL DEFAULT 'db';

-- Backfill: any existing rows are db-only.
UPDATE backups SET kind = 'db' WHERE kind IS NULL;

-- +goose Down
ALTER TABLE backups DROP COLUMN IF EXISTS kind;
