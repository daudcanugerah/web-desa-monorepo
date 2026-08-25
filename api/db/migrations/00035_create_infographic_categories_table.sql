-- +goose Up
-- +goose StatementBegin
CREATE TABLE infographic_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_infographic_categories_name ON infographic_categories(name);

-- Seed a default "Lainnya" fallback.
INSERT INTO infographic_categories (name) VALUES ('Lainnya')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS infographic_categories;
-- +goose StatementEnd
