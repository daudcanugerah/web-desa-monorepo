-- Migration 00047: composite + pg_trgm indexes for hot list endpoints.
--
-- Resolves performance.md#4 (missing indexes). Most list endpoints filter
-- by category + sort by created_at DESC; a composite index lets Postgres
-- satisfy both from one B-tree scan. ILIKE search on berita.title / umkm.name
-- / fasilitas.name benefits from pg_trgm GIN indexes.
--
-- pg_trgm extension is created idempotently; CREATE EXTENSION IF NOT EXISTS
-- requires superuser on managed Postgres instances. On AWS RDS / Cloud SQL
-- use the `pg_trgm` parameter group instead.

-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- berita: list by category sorted by created_at; search by title.
CREATE INDEX IF NOT EXISTS idx_berita_category_created_at
    ON berita (category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_berita_title_trgm
    ON berita USING gin (title gin_trgm_ops);

-- umkm: same shape.
CREATE INDEX IF NOT EXISTS idx_umkm_category_created_at
    ON umkm (category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_umkm_name_trgm
    ON umkm USING gin (name gin_trgm_ops);

-- fasilitas: composite (category, latitude, longitude) supports the
-- admin bbox filter. Public list still scans without bbox.
CREATE INDEX IF NOT EXISTS idx_fasilitas_category_created_at
    ON fasilitas (category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_fasilitas_name_trgm
    ON fasilitas USING gin (name gin_trgm_ops);

-- struktur: list is currently unsorted-by-category but category lookup
-- is the only filter; existing idx_struktur_created_at remains.
CREATE INDEX IF NOT EXISTS idx_struktur_created_at
    ON struktur_organisasi (created_at DESC);

-- ppid_requests: list is filtered by status + ordered by created_at.
CREATE INDEX IF NOT EXISTS idx_ppid_requests_status_created_at
    ON ppid_requests (status, created_at DESC);

-- ppid documents: same composite shape.
CREATE INDEX IF NOT EXISTS idx_ppid_category_created_at
    ON ppid (category, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_ppid_category_created_at;
DROP INDEX IF EXISTS idx_ppid_requests_status_created_at;
DROP INDEX IF EXISTS idx_struktur_created_at;
DROP INDEX IF EXISTS idx_fasilitas_name_trgm;
DROP INDEX IF EXISTS idx_fasilitas_category_created_at;
DROP INDEX IF EXISTS idx_umkm_name_trgm;
DROP INDEX IF EXISTS idx_umkm_category_created_at;
DROP INDEX IF EXISTS idx_berita_title_trgm;
DROP INDEX IF EXISTS idx_berita_category_created_at;
