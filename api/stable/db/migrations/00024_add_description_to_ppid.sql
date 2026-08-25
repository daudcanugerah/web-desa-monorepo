-- +goose Up
-- +goose StatementBegin
ALTER TABLE ppid ADD COLUMN description TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ppid DROP COLUMN description;
-- +goose StatementEnd
