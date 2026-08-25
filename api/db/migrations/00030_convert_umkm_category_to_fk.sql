-- +goose Up
-- +goose StatementBegin
-- Add new UUID column
ALTER TABLE umkm ADD COLUMN category_id UUID;

-- Backfill: copy distinct non-empty category names into umkm_categories
INSERT INTO umkm_categories (name)
    SELECT DISTINCT category FROM umkm WHERE category <> ''
    ON CONFLICT (name) DO NOTHING;

-- Backfill category_id from name match. Empty rows get the seeded 'Umum' fallback.
UPDATE umkm u
SET category_id = COALESCE(
    (SELECT id FROM umkm_categories WHERE name = u.category LIMIT 1),
    (SELECT id FROM umkm_categories WHERE name = 'Umum' LIMIT 1)
);

-- Drop the old column and rename new one to keep API ergonomics
ALTER TABLE umkm DROP COLUMN category;
ALTER TABLE umkm RENAME COLUMN category_id TO category;

-- Enforce NOT NULL now that all rows are backfilled
ALTER TABLE umkm ALTER COLUMN category SET NOT NULL;

-- Add FK with ON DELETE RESTRICT (delete is blocked if any row references the category)
ALTER TABLE umkm ADD CONSTRAINT fk_umkm_category
    FOREIGN KEY (category) REFERENCES umkm_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_umkm_category ON umkm(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE umkm DROP CONSTRAINT IF EXISTS fk_umkm_category;
DROP INDEX IF EXISTS idx_umkm_category;
ALTER TABLE umkm ADD COLUMN category_name VARCHAR(100) NOT NULL DEFAULT '';
UPDATE umkm u SET category_name = uc.name
    FROM umkm_categories uc WHERE uc.id = u.category;
ALTER TABLE umkm DROP COLUMN category;
ALTER TABLE umkm RENAME COLUMN category_name TO category;
-- +goose StatementEnd
