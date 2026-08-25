-- +goose Up
-- +goose StatementBegin
CREATE TABLE ppid_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ppid_id UUID NOT NULL REFERENCES ppid(id) ON DELETE CASCADE,
    requester_name VARCHAR(255) NOT NULL,
    requester_email VARCHAR(255) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ppid_requests_ppid_id ON ppid_requests(ppid_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ppid_requests_ppid_id;
DROP TABLE IF EXISTS ppid_requests;
-- +goose StatementEnd
