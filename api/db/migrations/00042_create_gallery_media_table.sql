-- +goose Up
-- +goose StatementBegin
CREATE TYPE gallery_media_type AS ENUM ('image', 'video');

CREATE TABLE gallery_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    folder_id UUID NOT NULL REFERENCES gallery_folders(id) ON DELETE CASCADE,
    media_type gallery_media_type NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    thumbnail_failed BOOLEAN NOT NULL DEFAULT FALSE,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    width INTEGER,
    height INTEGER,
    duration_seconds DOUBLE PRECISION,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_gallery_media_size CHECK (file_size > 0)
);

ALTER TABLE gallery_folders
    ADD CONSTRAINT fk_gallery_folders_cover_media
    FOREIGN KEY (cover_media_id) REFERENCES gallery_media(id) ON DELETE SET NULL;

CREATE INDEX idx_gallery_media_folder_id     ON gallery_media(folder_id);
CREATE INDEX idx_gallery_media_is_public     ON gallery_media(is_public);
CREATE INDEX idx_gallery_media_folder_public ON gallery_media(folder_id, is_public, created_at DESC);
CREATE INDEX idx_gallery_media_created_at    ON gallery_media(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_gallery_media_created_at;
DROP INDEX IF EXISTS idx_gallery_media_folder_public;
DROP INDEX IF EXISTS idx_gallery_media_is_public;
DROP INDEX IF EXISTS idx_gallery_media_folder_id;
ALTER TABLE gallery_folders DROP CONSTRAINT IF EXISTS fk_gallery_folders_cover_media;
DROP TABLE IF EXISTS gallery_media;
DROP TYPE IF EXISTS gallery_media_type;
-- +goose StatementEnd
