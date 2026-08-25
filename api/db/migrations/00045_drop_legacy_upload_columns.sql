-- Migration 00045: drop legacy upload columns
--
-- Phase 4 of draft/002-upload-migration. Every feature now writes
-- uploads exclusively through the gallery media system (see Task 4.3
-- response shape, the `*_media_id` columns are the source of truth).
-- Drop the legacy `image_url` / `images` / `profile_image_url` /
-- `thumbnail_url` columns that predate the gallery port.
--
-- This migration is irreversible: callers that still send a legacy URL
-- via form/JSON silently lose that field, and any pre-existing legacy
-- rows are migrated only conceptually — the `*_media_id` columns are
-- already the source of truth for newly created rows.

-- +goose Up
ALTER TABLE banners             DROP COLUMN IF EXISTS image_url;
ALTER TABLE berita              DROP COLUMN IF EXISTS image_url;
ALTER TABLE users               DROP COLUMN IF EXISTS profile_image_url;
ALTER TABLE struktur_organisasi DROP COLUMN IF EXISTS profile_image_url;
ALTER TABLE ppid                DROP COLUMN IF EXISTS file_url;
ALTER TABLE ppid                DROP COLUMN IF EXISTS thumbnail_url;
ALTER TABLE umkm                DROP COLUMN IF EXISTS images;
ALTER TABLE fasilitas           DROP COLUMN IF EXISTS images;

-- +goose Down
-- Irreversible: the legacy columns were removed permanently. Nothing to
-- restore; a down-migration would require re-creating columns that no
-- longer exist in the application schema.
SELECT 1;
