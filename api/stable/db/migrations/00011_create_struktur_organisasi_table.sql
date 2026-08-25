-- +goose Up
-- +goose StatementBegin
CREATE TABLE struktur_organisasi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    position VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    profile_image_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_struktur_name ON struktur_organisasi(name);
CREATE INDEX idx_struktur_position ON struktur_organisasi(position);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_struktur_position;
DROP INDEX IF EXISTS idx_struktur_name;
DROP TABLE IF EXISTS struktur_organisasi;
-- +goose StatementEnd
