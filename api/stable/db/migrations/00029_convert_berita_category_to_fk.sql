-- +goose Up
-- +goose StatementBegin
-- Add new UUID column
ALTER TABLE berita ADD COLUMN category_id UUID;

-- Backfill: copy distinct non-empty category names into berita_categories
INSERT INTO berita_categories (name)
    SELECT DISTINCT category FROM berita WHERE category <> ''
    ON CONFLICT (name) DO NOTHING;

-- Backfill category_id from name match. Empty rows get the seeded 'Lainnya' fallback.
UPDATE berita b
SET category_id = COALESCE(
    (SELECT id FROM berita_categories WHERE name = b.category LIMIT 1),
    (SELECT id FROM berita_categories WHERE name = 'Lainnya' LIMIT 1)
);

-- Drop the old column and rename new one to keep API ergonomics
ALTER TABLE berita DROP COLUMN category;
ALTER TABLE berita RENAME COLUMN category_id TO category;

-- Enforce NOT NULL now that all rows are backfilled
ALTER TABLE berita ALTER COLUMN category SET NOT NULL;

-- Add FK with ON DELETE RESTRICT (delete is blocked if any row references the category)
ALTER TABLE berita ADD CONSTRAINT fk_berita_category
    FOREIGN KEY (category) REFERENCES berita_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_berita_category ON berita(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE berita DROP CONSTRAINT IF EXISTS fk_berita_category;
DROP INDEX IF EXISTS idx_berita_category;
ALTER TABLE berita ADD COLUMN category_name VARCHAR(100) NOT NULL DEFAULT '';
UPDATE berita b SET category_name = bc.name
    FROM berita_categories bc WHERE bc.id = b.category;
ALTER TABLE berita DROP COLUMN category;
ALTER TABLE berita RENAME COLUMN category_name TO category;
-- +goose StatementEnd
