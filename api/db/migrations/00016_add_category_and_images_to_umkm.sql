-- +goose Up
-- +goose StatementBegin
ALTER TABLE umkm ADD COLUMN category VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE umkm ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE umkm ADD COLUMN images JSONB DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE umkm DROP COLUMN images;
ALTER TABLE umkm DROP COLUMN description;
ALTER TABLE umkm DROP COLUMN category;
-- +goose StatementEnd
