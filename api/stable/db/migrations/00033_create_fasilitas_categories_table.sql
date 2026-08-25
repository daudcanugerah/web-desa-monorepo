-- +goose Up
-- +goose StatementBegin
CREATE TABLE fasilitas_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_fasilitas_categories_name ON fasilitas_categories(name);

-- Seed a default "Lainnya" fallback for any rows with NULL/empty type.
INSERT INTO fasilitas_categories (name) VALUES ('Lainnya')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS fasilitas_categories;
-- +goose StatementEnd
