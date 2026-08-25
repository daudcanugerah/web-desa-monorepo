# Backend Spec Changes Request

Generated from openapi.yaml reaudit on 2026-08-25. Frontend code on `fix/openapi-drift` branch is now decoupled from these gaps where possible. Items below require backend + spec updates.

---

## 1. Add public gallery endpoints (P0 — Galeri page 100% broken)

The current `openapi.yaml` only declares admin `/gallery/...` routes and `/media/{id}/{content,thumbnail}?jwt=...`. The public site calls these and gets 404:

```
GET /public/gallery/folders
GET /public/gallery/folders/{id}
GET /public/gallery/media/{id}
```

Add these paths to `openapi.yaml` under the `public` tag with `security: []`. Schemas already exist (`gallery.FolderListResponse`, `gallery.FolderDetailResponse`, `gallery.MediaResponse`).

Behavior:
- `/public/gallery/folders` — query params: `q, page, limit` (consistent with other public lists). Returns only folders where `is_public = true`.
- `/public/gallery/folders/{id}` — returns `{ folder, media: [], pagination }`. Only media where `is_public = true`.
- `/public/gallery/media/{id}` — returns `MediaResponse`. Only if `is_public = true` and `folder_id` belongs to a public folder.

Fix the JWT signing gap: the existing `/media/{id}/content?jwt=...` and `/media/{id}/thumbnail?jwt=...` require a signed JWT. Either:
- (a) have `/public/gallery/folders/{id}` return pre-signed URLs in `media[].thumbnail_url` / `media[].content_url`, OR
- (b) expose a `/public/media/sign?id={id}&type=content|thumbnail` endpoint that mints short-lived JWTs.

Without one of these, every gallery thumbnail and full-size image returns 401.

## 2. Extend `desa.DesaResponse` (P0 — Home stats all fallback)

Spec currently has only `address, description, email, name, phone, updated_at, vision_mission, website`. The frontend reads fields that aren't there, so every Home stats card and the Sambutan section hit the FAKE_* fallback strings.

Add to `desa.DesaResponse`:

```yaml
kepala_desa:
  type: string
  description: Full name of the village head.
kepala_desa_message:
  type: string
  description: Sambutan / welcome message shown on the Home page.
motto:
  type: string
  description: Village motto shown in the ribbon.
kecamatan:
  type: string
kabupaten:
  type: string
provinsi:
  type: string
jumlah_penduduk:
  type: integer
jumlah_kk:
  type: integer
jumlah_dusun:
  type: integer
jumlah_rt:
  type: integer
jumlah_rw:
  type: integer
jumlah_umkm:
  type: integer
social_links:
  type: object
  properties:
    facebook: { type: string, format: uri }
    instagram: { type: string, format: uri }
    youtube: { type: string, format: uri }
    twitter: { type: string, format: uri }
    tiktok: { type: string, format: uri }
```

`social_links` is optional; `Footer.vue` currently hardcodes `#` for social icons.

## 3. Fix `/files/{filename}` parameter naming (P1 — spec self-inconsistency)

`openapi.yaml:2574-2582` declares the path as `/files/{filename}` but the parameter block names it `id` (`description: Resource ID`). Swagger 2.0 doesn't enforce this, but Go-style handlers binding on `{id}` will mismatch `{filename}`.

Rename the parameter to `filename` so the spec matches the URL template.

## 4. Add `q` query param to public list endpoints (P2 — bandwidth)

Public list endpoints currently paginate/filter client-side because the spec doesn't declare `q`. Backend may already support it via the admin handlers — if so, add to spec.

Endpoints:
- `/public/berita/list` — add `q: string` (full-text on title)
- `/public/berita/categories` — already has unused endpoint, declare `q, page, limit` for consistency
- `/public/ppid/list` — already declared ✓
- `/public/umkm/list` — already declared ✓

## 5. Add `sort`/`order` to `/public/infographic/list` (P2)

Spec declares only `page, limit` (`openapi.yaml:5188-5197`). Frontend used to send `sort`/`order` and was dropped in this branch. If backend supports it, declare it so callers can rely on stable ordering:

```yaml
- name: sort
  in: query
  type: string
  enum: [created_at, updated_at, section_name]
  default: created_at
- name: order
  in: query
  type: string
  enum: [asc, desc]
  default: desc
```

If backend does NOT support it, leave the spec as is — frontend already stops sending them.

## 6. Add `expires_at_unix` to `infographic.InfographicResponse` (P2)

Frontend decodes the JWT `exp` claim to schedule auto-refresh. Adding an explicit `expires_at_unix: integer` field would let clients avoid JWT parsing and survive token-format changes. Optional; current workaround works.

---

## Reference: what frontend now does

- `src/services/desaService.js:mediaUrl(item, variant)` — reads `item.media?.url || item.image_url` for single, `item.media?.[0]?.url || item.images?.[0]` for array. Drops `images[]` as a documented field; will fall back if backend keeps emitting it.
- `thumbnailUrl(item)` / `documentUrl(item)` — read `item.thumbnail?.url || item.thumbnail_url` and `item.document?.url || item.document_url` respectively. Frontend tolerates either name for one release, then we will drop the legacy.
- `getPublicProfileList` reads response key `profile` (singular) per spec; frontend ignores any `profiles` plural.
- `getPublicInfographicList` no longer sends `sort`/`order`.
- `getInfographicEmbedUrl` builds `${base}/embed/dashboard|question/${token}#bordered=false&titled=false`. Token only in path. If Metabase actually requires the token in the hash fragment for your version, confirm and we revert.

## Feature flags (frontend gating)

Until backend ships the items above, the frontend gates related UI behind env flags read by `src/composables/useFeatureFlags.js`:

| Flag | Default | When true |
|---|---|---|
| `VITE_FEATURE_GALLERY_PUBLIC` | `false` | Galeri pages and nav link render; calls hit `/public/gallery/*` |
| `VITE_FEATURE_DESA_EXTENDED` | `false` | Home Sambutan/stats rely on `desa.kepala_desa`, `motto`, `jumlah_*`, etc. (else fall back to FAKE_*) |
| `VITE_FEATURE_BERITA_SEARCH` | `false` | Berita list sends `?q=` to `/public/berita/list` (else client-side filter only) |
| `VITE_FEATURE_INFOGRAPHIC_SORT` | `false` | Infografik list sends `?sort=&order=` (else backend-default order) |

Flip a flag by setting the env var in `.env.local` or CI, then redeploy. No code change needed.

## Non-actionable notes (just FYI)

- The frontend has unused service functions (`getPublicBeritaCategories`, `getPublicFasilitasById`, etc.) removed in this branch. If you intend them for a future admin panel, do not add frontend callers until then.
- `useDesaInfo` caches the village info module-globally. If you ever add auth, invalidate this cache on login/logout.
- `Peta.vue` uses hardcoded polygon coords (`src/pages/Peta.vue:235`) and ignores `public/area-desa-poly.json`. Frontend cleanup, not backend.
