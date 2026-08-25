-- +goose Up
-- +goose StatementBegin
-- Create banner_categories table (managed vocabulary)
CREATE TABLE banner_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT idx_banner_categories_name UNIQUE (name)
);

-- Seed placeholder category for legacy banners (or banners without category)
INSERT INTO banner_categories (name)
VALUES ('Lainnya')
ON CONFLICT (name) DO NOTHING;

-- Add category column to banners (nullable for backward compat with existing rows)
ALTER TABLE banners
    ADD COLUMN category UUID REFERENCES banner_categories(id) ON DELETE RESTRICT;

-- Backfill existing banners with the placeholder category
UPDATE banners
SET category = (SELECT id FROM banner_categories WHERE name = 'Lainnya')
WHERE category IS NULL;

-- Index for category lookups
CREATE INDEX idx_banners_category ON banners (category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_banners_category;
ALTER TABLE banners DROP COLUMN IF EXISTS category;
DROP TABLE IF EXISTS banner_categories;
-- +goose StatementEnd