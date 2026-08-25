-- +goose Up
-- +goose StatementBegin
CREATE TABLE banners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    image_url VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_banner_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX idx_banners_status ON banners(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_banners_status;
DROP TABLE IF EXISTS banners;
-- +goose StatementEnd
