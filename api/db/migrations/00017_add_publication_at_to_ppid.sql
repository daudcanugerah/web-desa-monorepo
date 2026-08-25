-- +goose Up
-- +goose StatementBegin
ALTER TABLE ppid ADD COLUMN publication_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ppid DROP COLUMN publication_at;
-- +goose StatementEnd
