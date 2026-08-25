-- +goose Up
-- +goose StatementBegin
CREATE TABLE profile (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    section_name VARCHAR(100) NOT NULL,
    section_endpoint VARCHAR(255) NOT NULL,
    state BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profile_section_name ON profile(section_name);
CREATE INDEX idx_profile_state ON profile(state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_profile_state;
DROP INDEX IF EXISTS idx_profile_section_name;
DROP TABLE IF EXISTS profile;
-- +goose StatementEnd
