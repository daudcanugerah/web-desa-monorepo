-- +goose Up
-- +goose StatementBegin
-- Profile section categories: a managed vocabulary that groups profile
-- information sections (Sejarah, Visi, Misi, …) into named groups for the
-- public Profil page navigation.
CREATE TABLE profile_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_profile_categories_name ON profile_categories(name);

-- Seed the default groups.
INSERT INTO profile_categories (name) VALUES
    ('Tentang Desa'),
    ('Potensi & Ekonomi'),
    ('Visi & Misi'),
    ('Pemerintahan'),
    ('Lainnya')
ON CONFLICT (name) DO NOTHING;

-- Add a nullable category FK to profile.
ALTER TABLE profile ADD COLUMN category UUID;

-- Backfill existing rows by section name; anything unmatched -> 'Lainnya'.
UPDATE profile AS p SET category = c.id
FROM profile_categories c
WHERE c.name = CASE
    WHEN p.section_name IN ('Sejarah', 'Lokasi Desa', 'Geografi', 'Demografi') THEN 'Tentang Desa'
    WHEN p.section_name IN ('Potensi Desa', 'Keuangan') THEN 'Potensi & Ekonomi'
    WHEN p.section_name IN ('Visi', 'Misi') THEN 'Visi & Misi'
    WHEN p.section_name IN ('Pemerintahan', 'Struktur Organisasi') THEN 'Pemerintahan'
    ELSE 'Lainnya'
END;

-- Any remaining NULLs -> 'Lainnya'.
UPDATE profile SET category = (SELECT id FROM profile_categories WHERE name = 'Lainnya' LIMIT 1)
WHERE category IS NULL;

ALTER TABLE profile ADD CONSTRAINT fk_profile_category
    FOREIGN KEY (category) REFERENCES profile_categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_profile_category ON profile(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE profile DROP CONSTRAINT IF EXISTS fk_profile_category;
DROP INDEX IF EXISTS idx_profile_category;
ALTER TABLE profile DROP COLUMN IF EXISTS category;
DROP TABLE IF EXISTS profile_categories;
-- +goose StatementEnd