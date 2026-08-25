-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS infographic_access_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    infographic_id UUID NOT NULL,
    component_id BIGINT NOT NULL,
    component_type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(50) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    token_issued_at TIMESTAMP NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_infographic_access_log_infographic_id
    ON infographic_access_log (infographic_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_infographic_access_log_ip
    ON infographic_access_log (ip_address, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_infographic_access_log_created
    ON infographic_access_log (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_infographic_access_log_created;
DROP INDEX IF EXISTS idx_infographic_access_log_ip;
DROP INDEX IF EXISTS idx_infographic_access_log_infographic_id;
DROP TABLE IF EXISTS infographic_access_log;
-- +goose StatementEnd
