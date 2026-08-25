-- +goose Up
-- +goose StatementBegin
ALTER TABLE fasilitas ADD COLUMN images JSONB DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fasilitas DROP COLUMN images;
-- +goose StatementEnd
