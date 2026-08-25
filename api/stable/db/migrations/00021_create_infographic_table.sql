-- +goose Up
-- +goose StatementBegin
CREATE TABLE infographic (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id VARCHAR(255) NOT NULL,
    section_name VARCHAR(100) NOT NULL,
    section_endpoint VARCHAR(255) NOT NULL,
    state BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_infographic_section_name ON infographic(section_name);
CREATE INDEX idx_infographic_state ON infographic(state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_infographic_state;
DROP INDEX IF EXISTS idx_infographic_section_name;
DROP TABLE IF EXISTS infographic;
-- +goose StatementEnd
