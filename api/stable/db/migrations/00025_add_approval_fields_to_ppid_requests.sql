-- +goose Up
-- +goose StatementBegin
ALTER TABLE ppid_requests ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';
ALTER TABLE ppid_requests ADD COLUMN approved_at TIMESTAMP;
ALTER TABLE ppid_requests ADD COLUMN approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE ppid_requests ADD COLUMN revoked_at TIMESTAMP;
ALTER TABLE ppid_requests ADD COLUMN revoked_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_ppid_requests_status ON ppid_requests(status);
CREATE INDEX idx_ppid_requests_approved_at ON ppid_requests(approved_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ppid_requests_approved_at;
DROP INDEX IF EXISTS idx_ppid_requests_status;
ALTER TABLE ppid_requests DROP COLUMN revoked_by;
ALTER TABLE ppid_requests DROP COLUMN revoked_at;
ALTER TABLE ppid_requests DROP COLUMN approved_by;
ALTER TABLE ppid_requests DROP COLUMN approved_at;
ALTER TABLE ppid_requests DROP COLUMN status;
-- +goose StatementEnd
