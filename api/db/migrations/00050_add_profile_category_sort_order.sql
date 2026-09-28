-- +goose Up
-- +goose StatementBegin
-- Explicit display order for profile section categories. The public Profil
-- navigation renders category groups in this order (ascending), so ordering is
-- server-driven instead of hardcoded in the frontend.
ALTER TABLE profile_categories ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 100;

-- Seed deterministic order for the default groups; anything else keeps 100.
UPDATE profile_categories SET sort_order = 10 WHERE name = 'Tentang Desa';
UPDATE profile_categories SET sort_order = 20 WHERE name = 'Potensi & Ekonomi';
UPDATE profile_categories SET sort_order = 30 WHERE name = 'Visi & Misi';
UPDATE profile_categories SET sort_order = 40 WHERE name = 'Pemerintahan';
UPDATE profile_categories SET sort_order = 50 WHERE name = 'Lainnya';

CREATE INDEX idx_profile_categories_sort_order ON profile_categories(sort_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_profile_categories_sort_order;
ALTER TABLE profile_categories DROP COLUMN IF EXISTS sort_order;
-- +goose StatementEnd
