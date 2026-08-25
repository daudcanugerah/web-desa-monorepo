# OpenAPI vs Implementation Audit

Source of truth: [`../openapi.yaml`](../openapi.yaml) (5706 lines). Per OpenAPI, public endpoints return data **at the top level** — no `data` wrapper, only `{resource: [], pagination: {}}` for paginated lists and the resource object directly for single-item GETs.

**Update after live test:** the real backend wraps **every** response in `{success: true, data: {...}}`, contradicting the OpenAPI spec. `desaService.js` now uses `unwrapData()` / `unwrapList()` helpers that handle both shapes (envelope and raw), so it works against either implementation. If/when the backend is brought in line with OpenAPI, the helpers will keep working — drop them only after confirming the backend no longer wraps.

The frontend was written against a **different, fictional API shape**. Almost every page will break when wired to a real backend.

---

## A. Response Unwrapping — 8 of 13 Functions Wrong

Service code does `response.data?.X` first. The actual responses have no `data` wrapper. Three functions happen to work because they also try `response.X` as a later fallback. Eight do not.

| Function | Endpoint | Real shape | Service does | Verdict |
|---|---|---|---|---|
| `getPublicBeritaList` | `/public/berita/list` | `{berita:[],pagination:{}}` | tries `response.berita` | OK (accidental) |
| `getPublicFasilitasList` | `/public/fasilitas/list` | `{fasilitas:[],pagination:{}}` | tries `response.fasilitas` | OK (accidental) |
| `getPublicUMKMList` | `/public/umkm/list` | `{umkm:[],pagination:{}}` | tries `response.umkm` | OK (accidental) |
| `getPublicDesa` | `/public/desa` | `desa.DesaResponse` (raw) | `response.data` | **WRONG → `null`** |
| `getPublicStrukturList` | `/public/struktur/list` | `{struktur:[],pagination:{}}` | `response.data?.struktur` | **WRONG → `[]`** |
| `getPublicPPIDList` | `/public/ppid/list` | `{ppid:[],pagination:{}}` | `response.data?.ppid` | **WRONG → `[]`** |
| `getPublicPPIDById` | `/public/ppid/{id}` | `ppid.PPIDResponse` (raw) | `response.data` | **WRONG → `null`** |
| `createPPIDRequest` | `POST /public/ppid/{id}/requests` | `ppid.PPIDRequestResponse` (raw) | `response.data` | **WRONG → `null`** |
| `getPublicInfographicList` | `/public/infographic/list` | `{infographic:[],pagination:{}}` | `response.data?.infographic` | **WRONG → `[]`** |
| `getActiveBanners` | `/banners/active` | raw array | `response.data?.banners` | **WRONG** (would be `[]`) — but unused |
| `getPublicBannerList` | `/public/banner` | `{banners:[],pagination:{}}` | `response.data?.banners` | **WRONG** — unused |
| `getPublicBannerById` | `/public/banner/{id}` | `banner.BannerResponse` (raw) | `response.data` | **WRONG** — unused |
| `getPublicBeritaById` | `/public/berita/{id}` | `berita.BeritaResponse` (raw) | `response.data \|\| response` | **OK by accident** (2nd branch) |
| `getPublicFasilitasById` | `/public/fasilitas/{id}` | raw | `response.data \|\| response` | OK by accident — unused |
| `getPublicUMKMById` | `/public/umkm/{id}` | raw | `response.data \|\| response` | OK by accident — unused |
| `getPublicStrukturById` | `/public/struktur/{id}` | raw | `response.data` | **WRONG** — unused |
| `getPublicInfographicById` | `/public/infographic/{id}` | raw | `response.data` | **WRONG** — unused |
| `getPublicProfileList` | `/public/profile/list` | `{profile:[],pagination:{}}` | `response.data?.profile` | **WRONG** — unused |
| `getPublicProfileById` | `/public/profile/{id}` | raw | `response.data` | **WRONG** — unused |
| `getBeritaCategories` | `/berita/categories` (admin) | `{categories:[],pagination:{}}` | `response.data?.categories` | **WRONG** — unused |
| `getPPIDCategories` | `/ppid/categories` (admin) | same | `response.data?.categories` | **WRONG** — unused |
| `getUMKMCategories` | `/umkm/categories` (admin) | same | `response.data?.categories` | **WRONG** — unused |

**Fix shape** (replace whole `request()` wrapper or each function):

```js
// Paginated list
return response[resourceName] || []
// Single resource
return response || null
```

`apiClient.request()` returns `data` directly (it does `return data` after parsing). The `data?.X` unwrap assumes the backend wraps every response in `{data: ...}`, but it never does.

### Peta debug noise

`getPublicFasilitasList` has `console.log` debug statements (`console.log('getPublicFasilitasList raw response:', response)`, `'Returning from response.data.fasilitas'`, `'response.data:'`, `'Returning empty array'`). Strip.

---

## B. Field Mismatches — Most Pages Will Show Nothing

### B1. `desa.DesaResponse` (Home + Profil)

**OpenAPI:** `{address, description, email, name, phone, updated_at, vision_mission, website}`

**Frontend expects** (Profil + Home):
- Profil: `history`, `vision`, `mission[]`, `structure`
- Home: `population`, `area`, `neighborhoods`, `businesses`, `statistics`

**Overlap:** zero. Every fallback string ("Data sejarah tidak tersedia", etc.) will fire on Profil. Home stats will all be `undefined`; `stats.population.toLocaleString` will throw.

**Possible mapping:**
- `vision_mission` (string) → split into vision + mission
- `description` → could be history
- `name`, `address`, `phone`, `email`, `website` → not consumed

Stats (`population`, `area`, `neighborhoods`, `businesses`, `statistics`) **do not exist in any backend schema**. The Home page is built against an imagined `desa` endpoint. Either:
- Add these fields to `desa.DesaResponse`, or
- Pull them from a separate `/infographic` section, or
- Use a different endpoint

### B2. `berita.BeritaListResponse` (Berita list)

**OpenAPI item:** `{category: {id,name}, category_id, created_at, id, image_url, title, updated_at}`

**Frontend uses:** `title` ✓, `excerpt` ✗, `content` (in search) ✗, `category` (as string, real is object) ✗, `date` → falls back to `created_at` ✓, `slug` ✗, `image_url` (ignored, gradient used)

**Bug effects:**
- `news.category` rendered → `[object Object]`
- RouterLink `to="/berita/${news.slug}"` → `/berita/undefined` (broken navigation)
- Cards show no excerpt, no image
- Search `news.excerpt`/`news.content` always `undefined` → silent no-op

### B3. `berita.BeritaResponse` (Berita detail)

**OpenAPI:** `{category: {id,name}, category_id, content, created_at, id, image_url, title, updated_at}`

**Frontend uses:** `title` ✓, `content` ✓, `category` (as string, real is object) ✗, `author` ✗, `date` → falls back to `created_at` ✓, `image_url` (ignored)

**Bug effects:**
- Category label shows `[object Object]`
- "Oleh ..." author line never shows
- No hero image

### B4. `umkm.UMKMResponse` (UMKM)

**OpenAPI:** `{address, category: {id,name}, category_id, created_at, description, email, id, images[], name, owner, phone, updated_at, website}`

**Frontend uses:** `name`, `description`, `phone` (tel: link), `address` — all ✓
**But:** `category` is object → filter `umkm.category === 'Kuliner'` **never matches**. Dropdown built from data also broken.

**Ignored:** `email`, `owner`, `images[]` (could populate card UI), `website` (could be link)

### B5. `ppid.PPIDResponse` (PPID list + detail)

**OpenAPI:** `{category: {id,name}, category_id, created_at, description, document_url, id, publication_at, thumbnail_url, title, updated_at}`

**Frontend uses:** `title`, `description` (v-html), `publication_at` ✓, `thumbnail_url` ✓, `category` (as string, real is object) ✗
**Detail uses:** `category` ✗, `publication_at` ✓, `description` (v-html) ✓, `created_at` ✓, `updated_at` ✓

**Bug effects:**
- `selectedCategory` filter never matches
- Categories dropdown shows `[object Object]` and value is object reference
- `v-html` on `description` is XSS risk if backend doesn't sanitize (it should, but verify)

### B6. `fasilitas.FasilitasResponse` (Peta)

**OpenAPI:** `{category: {id,name}, category_id, created_at, description, id, images[], latitude, longitude, name, updated_at}`

**Frontend uses:** `id`, `name`, `description`, `latitude`, `longitude`, plus heuristic `type` / `location` (always empty)
**Ignored:** `images[]` (could be marker popup gallery), `category` object (Peta uses `getCategoryFromName` heuristic, so `type` and `category.name` never used)

**Bug effects:**
- "Location" line on Peta sidebar always empty
- Category badge uses name substring matching (e.g. "kantor" → Pemerintahan) — only works for ID-language names. If categories are localized, the heuristic fails.
- Markers without `latitude`/`longitude` get random points inside polygon (`Peta.vue:430`) — silent data degradation.

### B7. `struktur.StrukturResponse` (Profil Perangkat)

**OpenAPI:** `{created_at, description, email, id, name, phone, position, profile_image_url, updated_at}`

**Frontend uses:** `id`, `name`, `position`, `phone` — all ✓
**Ignored:** `profile_image_url` (Profil shows SVG placeholder), `email`, `description`

Works. Note: service unwrap is wrong (always returns `[]`) — see A. **Profil Perangkat will be empty in production** because of the unwrap bug, not the field mismatch.

### B8. `banner.BannerResponse` (Home hero)

**OpenAPI:** `{category, category_id, created_at, description, id, image_url, link, metadata, status, title, updated_at}`

**Home.vue:** hero slides are 3 hardcoded `slides` array (title, subtitle, gradient, image). **`getActiveBanners` is imported but never called** — dead. If called with the current service code, returns `[]` (wrong unwrap, see A) and would replace the hardcoded slides if unwrap were fixed.

`link` field exists — no click-through implemented.

---

## C. Route / URL Param Mismatches

### C1. Berita slug vs UUID

- Router: `path: 'berita/:slug'`
- Page: `const id = route.params.slug`
- API: `/public/berita/{id}` where `{id}` is **UUID** (OpenAPI line 3502: `description: Berita UUID`)

If backend only accepts UUID, navigating to `/berita/foo-slug` will 404. Either:
- Rename router param to `:id` and treat the route value as UUID, or
- Add `slug` to the Berita list response (currently not in `BeritaListResponse` schema) and have backend resolve it, or
- Use a separate `/public/berita/slug/:slug` endpoint (does not exist)

### C2. Admin endpoints used by service

- `getBeritaCategories` → `/berita/categories` (admin, requires auth)
- `getPPIDCategories` → `/ppid/categories` (admin, requires auth)
- `getUMKMCategories` → `/umkm/categories` (admin, requires auth)

All are unused in pages. If they were wired, they would 401 because no token flow exists. Public equivalents exist at `/public/berita/categories`, `/public/ppid/categories`, `/public/umkm/categories` — but service doesn't call them.

---

## D. Pagination: Client-side vs Server-side

All public list endpoints support `page` + `limit` query params. None of the service functions accept them. Pages fetch the entire list and paginate client-side (9 per page).

- Wastes bandwidth
- Defeats API pagination metadata (`response.pagination`)
- `Peta` filters by `q` / category client-side; backend supports them
- `Berita` doesn't use `q` (searches client-side); backend has no `q` param on `/public/berita/list` (only category)
- `PPID` accepts `q` + `category` on backend — could push search down

---

## E. Auth — All Dead Code

`apiClient` has `setToken` / `getToken` / `clearToken` but **no code path calls them**. No login page, no logout, no router guard. The `Authorization: Bearer` header is never sent (no token to send).

`getActiveBanners` is imported by `Home.vue` but not called. Same for several other admin-endpoint functions.

The whole auth scaffolding is forward-looking for an admin panel that doesn't exist in this repo.

---

## F. Image / File URLs — Mostly Wasted

| Field | Source | Consumed? |
|---|---|---|
| Berita list/detail `image_url` | API | NO — gradient placeholder always shown |
| UMKM `images[]` | API | NO — static SVG used |
| Fasilitas `images[]` | API | NO — stored in `locations[i].images` but never rendered |
| Struktur `profile_image_url` | API | NO — SVG avatar used |
| PPID `thumbnail_url` | API | YES — combined with `/files/{filename}` endpoint |
| Banner `image_url` | API | NO — Hero uses hardcoded picsum URLs |

Only PPID thumbnails are actually rendered. The other endpoints expose image data that the UI never uses.

---

## G. Infografik — Completely Disconnected

`Infografik.vue` calls `getPublicInfographicList()` but **discards the response**. Page renders from a hardcoded `MOCK_DATA` constant.

`infographic.InfographicResponse` shape: `{category, category_id, component_id, component_type, section_endpoint, section_name, state, token, updated_at, created_at}`

- `component_id`, `component_type`, `section_endpoint`, `section_name`, `token` → these are Metabase-embedding parameters (`token` is a JWT for `https://[metabase]/embed/dashboard/{token}`)
- No `data`, `sectors`, `income`, `ageGroups`, `levels`, `literacy`, `facilities` fields

The page is built against a pre-aggregated statistics API that doesn't exist. Backend uses Metabase iframe embeds. The two architectures don't match. Either:
- Replace the page with Metabase iframe embeds using the `token` field, or
- Build a separate statistics aggregation endpoint and call it from a refactored page

---

## H. Other Findings

### H1. PPID request body field names
Frontend sends: `{requester_name, requester_email, purpose}`
Schema: `PPIDRequestResponse` has those exact fields ✓
But schema for the **request body** is just `type: object` (line 4021) — no enforced schema. Backend may not validate. Field names are correct by convention.

### H2. Token storage key
`localStorage['auth_token']` — fine for a public site. If admin panel is added, consider `httpOnly` cookie to avoid XSS theft.

### H3. API base path
`/api/v1` — hardcoded fallback in `apiClient.js:6` and re-hardcoded in `PPID.vue:369` for thumbnail URL building. Single source of truth would help.

### H4. `apiClient.request()` error swallowing
On non-OK response: `throw new Error(data?.error || 'HTTP {status}')`. The thrown `Error` has `.status` and `.data` attached. Service functions catch and `console.error` then return `[]` / `null`. No user-facing error UI anywhere. Pages silently show empty state on failure.

### H5. Peta cleanup
`Peta.vue:235` polygon coords hardcoded, `public/area-desa-poly.json` ships but unused. Mismatch risk if either is updated.

### H6. Berita `getDate` fallback chain
`date || publication_at || created_at` — only `created_at` is in the schema. The other two fields are `null` always. Functionally fine, but the fallback chain is misleading.

### H7. Fasilitas list response field
OpenAPI `/public/fasilitas/list` does not have a `category` query filter — wait, it does (line 3780-3783). But `Peta.vue` doesn't pass it. Category filter is client-side only. If a category UUID existed, server could filter.

### H8. `getPublicBannerList` vs `getPublicBannerById`
Both are unused. If hero should switch to API-driven, use `getActiveBanners` (already imported in `Home.vue`) — fix unwrap first.

---

## I. Summary by Page

| Page | Status | What breaks |
|---|---|---|
| Home | **Broken** | Stats `NaN`/empty (field + unwrap). Hero uses hardcoded slides, ignores `getActiveBanners`. |
| Profil | **Broken** | Unwrap bug → `data.officials = []`. Field names don't exist → history/vision/mission/structure all fallback. |
| Infografik | **Broken** | Renders MOCK_DATA; API result discarded. MOCK_DATA shape incompatible with real endpoint. |
| Peta | **Partial** | Unwrap OK. `location`, `type` always empty (heuristic used). `images[]` ignored. Polygon coords hardcoded, GeoJSON unused. |
| Berita list | **Broken** | Category shows `[object Object]`. Slug missing → `/berita/undefined`. No excerpt, no image. Search hits `excerpt`/`content` always undefined. |
| Berita detail | **Broken** | Category object. No author. No excerpt. No image. UUID vs slug param ambiguity. |
| UMKM | **Broken** | Category filter never matches. Categories dropdown shows object refs. `images[]`, `email`, `owner`, `website` ignored. |
| PPID | **Broken** | Unwrap bug → list always `[]`. Category filter broken. Categories dropdown broken. `v-html` XSS risk. |

**No page is fully functional against the real backend as currently coded.**

---

## J. Suggested Fix Order

1. **Fix all `data?.X` unwraps** in `desaService.js` — use `response[resourceName] || []` for lists, `response || null` for single GET. One-line per function.
2. **Fix field mappings:**
   - Profil: map `desa.vision_mission` → vision+mission, `desa.description` → history. Or push for backend to add fields.
   - Home: same — needs backend schema additions or separate stats endpoint.
   - Berita/UMKM/PPID/Fasilitas: use `item.category?.name` (or `item.category_id`) in templates. Or push backend to flatten category to string.
3. **Rename router param** `berita/:slug` → `berita/:id` (treat as UUID). Or add slug to list response.
4. **Wire `getActiveBanners` in Home** and remove hardcoded slides (after fix #1).
5. **Decide Infografik strategy** — Metabase embed or refactor.
6. **Remove dead service functions** (categories, banner detail, etc.) and dead legacy mocks (lines 1–704).
7. **Use server-side pagination** by passing `{params: {page, limit}}` to `apiClient.get`.
8. **Add user-facing error UI** — pages silently show `EmptyState` on 5xx.
9. **Remove console.log** debug from `getPublicFasilitasList`.
10. **Unify `API_BASE_URL`** — currently duplicated in `apiClient.js` and `PPID.vue`.
