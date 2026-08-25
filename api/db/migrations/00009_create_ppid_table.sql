-- +goose Up
-- +goose StatementBegin
CREATE TABLE ppid (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    file_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ppid_category ON ppid(category);
CREATE INDEX idx_ppid_title ON ppid(title);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ppid_title;
DROP INDEX IF EXISTS idx_ppid_category;
DROP TABLE IF EXISTS ppid;
-- +goose StatementEnd
