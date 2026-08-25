-- +goose Up
-- +goose StatementBegin
CREATE TABLE umkm_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_umkm_categories_name ON umkm_categories(name);

-- Seed a default category so existing rows with empty category can backfill to a non-null FK.
INSERT INTO umkm_categories (name) VALUES ('Umum')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS umkm_categories;
-- +goose StatementEnd
