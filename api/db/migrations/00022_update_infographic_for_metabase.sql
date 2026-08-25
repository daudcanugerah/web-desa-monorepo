-- +goose Up
-- +goose StatementBegin
-- Add component_id and component_type columns
ALTER TABLE infographic 
ADD COLUMN component_id VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN component_type VARCHAR(50) NOT NULL DEFAULT 'dashboard';

-- Migrate existing dashboard_id to component_id
UPDATE infographic SET component_id = dashboard_id WHERE component_id = '';

-- Drop the old dashboard_id column
ALTER TABLE infographic DROP COLUMN dashboard_id;

-- Add index on component_type for filtering
CREATE INDEX idx_infographic_component_type ON infographic(component_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate dashboard_id column from component_id
ALTER TABLE infographic 
ADD COLUMN dashboard_id VARCHAR(255) NOT NULL DEFAULT '';

UPDATE infographic SET dashboard_id = component_id;

-- Drop new columns
ALTER TABLE infographic DROP COLUMN component_id;
ALTER TABLE infographic DROP COLUMN component_type;

-- Drop the new index
DROP INDEX IF EXISTS idx_infographic_component_type;
-- +goose StatementEnd
