-- +goose Up
-- +goose StatementBegin
CREATE TABLE berita_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_berita_categories_name ON berita_categories(name);

-- Seed a default category so existing rows with empty category can backfill to a non-null FK.
INSERT INTO berita_categories (name) VALUES ('Lainnya')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS berita_categories;
-- +goose StatementEnd
