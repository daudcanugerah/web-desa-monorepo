/**
 * Runtime configuration.
 *
 * Resolution order:
 *   1. window.__ENV__ — injected by the container entrypoint at startup
 *      (docker-entrypoint.d/40-runtime-env.sh) so the same image can be
 *      redeployed with a different API/Metabase URL without a rebuild.
 *   2. import.meta.env.* — the value baked in at build time (Docker build args
 *      or .env), used for local `npm run dev`.
 */
const runtime = (typeof window !== 'undefined' && window.__ENV__) || {}

const pick = (key, fallback) =>
  runtime[key] !== undefined && runtime[key] !== '' ? runtime[key] : fallback

export const env = {
  VITE_API_BASE_URL: pick('VITE_API_BASE_URL', import.meta.env.VITE_API_BASE_URL),
  VITE_METABASE_URL: pick('VITE_METABASE_URL', import.meta.env.VITE_METABASE_URL),
  VITE_FEATURE_GALLERY_PUBLIC: pick('VITE_FEATURE_GALLERY_PUBLIC', import.meta.env.VITE_FEATURE_GALLERY_PUBLIC),
  VITE_FEATURE_BERITA_SEARCH: pick('VITE_FEATURE_BERITA_SEARCH', import.meta.env.VITE_FEATURE_BERITA_SEARCH),
  VITE_FEATURE_INFOGRAPHIC_SORT: pick('VITE_FEATURE_INFOGRAPHIC_SORT', import.meta.env.VITE_FEATURE_INFOGRAPHIC_SORT)
}