-- +goose Up
-- +goose StatementBegin
-- Add new UUID column
ALTER TABLE fasilitas ADD COLUMN category_id UUID;

-- Backfill: copy distinct non-NULL/empty type values into fasilitas_categories
INSERT INTO fasilitas_categories (name)
    SELECT DISTINCT type FROM fasilitas WHERE type IS NOT NULL AND type <> ''
    ON CONFLICT (name) DO NOTHING;

-- Backfill category_id from name match. NULL/empty rows get the seeded "Lainnya" fallback.
UPDATE fasilitas f
SET category_id = COALESCE(
    (SELECT id FROM fasilitas_categories WHERE name = f.type LIMIT 1),
    (SELECT id FROM fasilitas_categories WHERE name = 'Lainnya' LIMIT 1)
);

-- Drop the old `type` column
ALTER TABLE fasilitas DROP COLUMN type;

-- Rename category_id to category for API ergonomics
ALTER TABLE fasilitas RENAME COLUMN category_id TO category;

-- Add FK with ON DELETE RESTRICT (delete blocked if any row references the category).
-- Column stays NULLABLE so existing semantics are preserved.
ALTER TABLE fasilitas ADD CONSTRAINT fk_fasilitas_category
    FOREIGN KEY (category) REFERENCES fasilitas_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_fasilitas_category ON fasilitas(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fasilitas DROP CONSTRAINT IF EXISTS fk_fasilitas_category;
DROP INDEX IF EXISTS idx_fasilitas_category;
ALTER TABLE fasilitas ADD COLUMN type_name VARCHAR(100);
UPDATE fasilitas f SET type_name = fc.name
    FROM fasilitas_categories fc WHERE fc.id = f.category;
ALTER TABLE fasilitas DROP COLUMN category;
ALTER TABLE fasilitas RENAME COLUMN type_name TO type;
-- +goose StatementEnd
