-- +goose Up
-- +goose StatementBegin
-- Change component_id from VARCHAR to BIGINT
-- First, drop the default constraint
ALTER TABLE infographic ALTER COLUMN component_id DROP DEFAULT;

-- Convert existing empty strings to 0 (invalid value that will need to be fixed)
UPDATE infographic SET component_id = '0' WHERE component_id = '';

-- Change the column type
ALTER TABLE infographic ALTER COLUMN component_id TYPE BIGINT USING component_id::BIGINT;

-- Set new default for BIGINT (though new records should provide valid IDs)
ALTER TABLE infographic ALTER COLUMN component_id SET DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert component_id back to VARCHAR
ALTER TABLE infographic ALTER COLUMN component_id DROP DEFAULT;
ALTER TABLE infographic ALTER COLUMN component_id TYPE VARCHAR(255) USING component_id::VARCHAR;
ALTER TABLE infographic ALTER COLUMN component_id SET DEFAULT '';
-- +goose StatementEnd
