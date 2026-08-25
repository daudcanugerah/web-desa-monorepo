-- +goose Up
-- +goose StatementBegin
ALTER TABLE berita ADD COLUMN category VARCHAR(100) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE berita DROP COLUMN category;
-- +goose StatementEnd
