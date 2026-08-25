-- +goose Up
-- +goose StatementBegin
CREATE TABLE gallery_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    cover_media_id UUID,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_gallery_folders_name ON gallery_folders (LOWER(name));
CREATE INDEX idx_gallery_folders_created_at ON gallery_folders(created_at DESC);
CREATE INDEX idx_gallery_folders_public_created ON gallery_folders(is_public, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_gallery_folders_public_created;
DROP INDEX IF EXISTS idx_gallery_folders_created_at;
DROP INDEX IF EXISTS uq_gallery_folders_name;
DROP TABLE IF EXISTS gallery_folders;
-- +goose StatementEnd
