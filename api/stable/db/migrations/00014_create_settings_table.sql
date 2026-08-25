-- +goose Up
-- +goose StatementBegin
CREATE TABLE settings (
    key         VARCHAR(255) PRIMARY KEY,
    value       JSONB        NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS settings;
-- +goose StatementEnd
