-- +goose Up
-- +goose StatementBegin
-- Add new UUID column
ALTER TABLE ppid ADD COLUMN category_id UUID;

-- Backfill: copy distinct non-NULL category names into ppid_categories
INSERT INTO ppid_categories (name)
    SELECT DISTINCT category FROM ppid WHERE category IS NOT NULL AND category <> ''
    ON CONFLICT (name) DO NOTHING;

-- Backfill category_id from name match. NULL/empty rows get NULL FK
-- (category_id is allowed to be NULL because the original column was nullable).
UPDATE ppid p
SET category_id = (
    SELECT id FROM ppid_categories WHERE name = p.category LIMIT 1
)
WHERE p.category IS NOT NULL AND p.category <> '';

-- Drop the old column
ALTER TABLE ppid DROP COLUMN category;

-- Rename new column to keep API ergonomics
ALTER TABLE ppid RENAME COLUMN category_id TO category;

-- Add FK with ON DELETE RESTRICT (delete is blocked if any ppid references the category)
-- Column stays NULLABLE to preserve original semantics.
ALTER TABLE ppid ADD CONSTRAINT fk_ppid_category
    FOREIGN KEY (category) REFERENCES ppid_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_ppid_category ON ppid(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ppid DROP CONSTRAINT IF EXISTS fk_ppid_category;
DROP INDEX IF EXISTS idx_ppid_category;
ALTER TABLE ppid ADD COLUMN category_name VARCHAR(100);
UPDATE ppid p SET category_name = pc.name
    FROM ppid_categories pc WHERE pc.id = p.category;
ALTER TABLE ppid DROP COLUMN category;
ALTER TABLE ppid RENAME COLUMN category_name TO category;
-- +goose StatementEnd
