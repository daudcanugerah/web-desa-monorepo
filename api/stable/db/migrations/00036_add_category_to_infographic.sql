-- +goose Up
-- +goose StatementBegin
-- Add new nullable UUID column
ALTER TABLE infographic ADD COLUMN category UUID;

-- Backfill all existing rows with the seeded "Lainnya" fallback
UPDATE infographic SET category = (
    SELECT id FROM infographic_categories WHERE name = 'Lainnya' LIMIT 1
);

-- Add FK with ON DELETE RESTRICT (delete blocked if any infographic references the category)
ALTER TABLE infographic ADD CONSTRAINT fk_infographic_category
    FOREIGN KEY (category) REFERENCES infographic_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_infographic_category ON infographic(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE infographic DROP CONSTRAINT IF EXISTS fk_infographic_category;
DROP INDEX IF EXISTS idx_infographic_category;
ALTER TABLE infographic DROP COLUMN category;
-- +goose StatementEnd
