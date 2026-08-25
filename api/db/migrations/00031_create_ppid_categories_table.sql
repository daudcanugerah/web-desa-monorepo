-- +goose Up
-- +goose StatementBegin
CREATE TABLE ppid_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ppid_categories_name ON ppid_categories(name);

-- Seed a default category so existing ppid rows with NULL category can
-- optionally backfill to a non-null FK during the conversion migration.
-- The "Tanpa Kategori" row is the conventional "uncategorized" placeholder.
INSERT INTO ppid_categories (name) VALUES ('Tanpa Kategori')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ppid_categories;
-- +goose StatementEnd
