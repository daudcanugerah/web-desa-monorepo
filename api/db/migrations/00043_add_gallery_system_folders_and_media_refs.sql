-- +goose Up
-- +goose StatementBegin

-- Add 'document' to the gallery media type ENUM so the PPID feature can
-- store PDFs and Office documents alongside the existing image and video
-- types. ENUM value additions are non-destructive.
ALTER TYPE gallery_media_type ADD VALUE IF NOT EXISTS 'document';

-- System folder flags: system folders back uploads for every feature
-- (banner/berita/struktur/umkm/fasilitas/user/ppid). They default to private
-- and are not exposed to public endpoints.
ALTER TABLE gallery_folders
    ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN feature_slug VARCHAR(32);

CREATE UNIQUE INDEX uq_gallery_folders_feature_slug
    ON gallery_folders (feature_slug)
    WHERE feature_slug IS NOT NULL;

CREATE INDEX idx_gallery_folders_is_system
    ON gallery_folders (is_system, created_at DESC);

-- Per-feature media_id columns. New rows store the gallery media UUID;
-- legacy rows keep their existing image_url/document_url/profile_image_url
-- strings for one release. The *_media_id columns are nullable so the
-- back-compat static /uploads/* path keeps rendering for legacy rows.
ALTER TABLE users
    ADD COLUMN profile_image_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL;

ALTER TABLE struktur_organisasi
    ADD COLUMN profile_image_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL;

ALTER TABLE berita
    ADD COLUMN image_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL;

ALTER TABLE banners
    ADD COLUMN image_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL;

ALTER TABLE umkm
    ADD COLUMN images_media_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE fasilitas
    ADD COLUMN images_media_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE ppid
    ADD COLUMN document_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL,
    ADD COLUMN thumbnail_media_id UUID REFERENCES gallery_media(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE ppid
    DROP COLUMN IF EXISTS thumbnail_media_id,
    DROP COLUMN IF EXISTS document_media_id;

ALTER TABLE fasilitas
    DROP COLUMN IF EXISTS images_media_ids;

ALTER TABLE umkm
    DROP COLUMN IF EXISTS images_media_ids;

ALTER TABLE banners
    DROP COLUMN IF EXISTS image_media_id;

ALTER TABLE berita
    DROP COLUMN IF EXISTS image_media_id;

ALTER TABLE struktur_organisasi
    DROP COLUMN IF EXISTS profile_image_media_id;

ALTER TABLE users
    DROP COLUMN IF EXISTS profile_image_media_id;

DROP INDEX IF EXISTS idx_gallery_folders_is_system;
DROP INDEX IF EXISTS uq_gallery_folders_feature_slug;

ALTER TABLE gallery_folders
    DROP COLUMN IF EXISTS feature_slug,
    DROP COLUMN IF EXISTS is_system;

-- +goose StatementEnd
