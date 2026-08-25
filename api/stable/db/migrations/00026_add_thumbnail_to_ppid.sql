-- +goose Up
-- +goose StatementBegin
ALTER TABLE ppid ADD COLUMN thumbnail_url VARCHAR(500);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ppid DROP COLUMN IF EXISTS thumbnail_url;
-- +goose StatementEnd
