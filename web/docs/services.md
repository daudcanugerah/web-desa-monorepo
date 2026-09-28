# Services

## `apiClient.js`

Singleton `APIClient` class exported as `apiClient`. Fetch-based.

```js
import { apiClient } from '@/services/apiClient'
const data = await apiClient.get('/public/berita/list', { params: { page: 1, limit: 10 } })
await apiClient.post('/public/ppid/123/requests', { requester_name, requester_email })
```

### Config

- `API_BASE_URL` = `import.meta.env.VITE_API_BASE_URL` (default `http://localhost:8081/api/v1`),
  passed through `normalizeBaseUrl()` which guarantees a `/api/v1` suffix — so
  `VITE_API_BASE_URL` may be set either as `http://host:port` or
  `http://host:port/api/v1`.
- `TOKEN_KEY` = `'auth_token'` (localStorage)

### Methods

- `get(endpoint, { params, skipAuth } = {})`
- `post(endpoint, data, options)`
- `put(endpoint, data, options)`
- `patch(endpoint, data, options)`
- `delete(endpoint, options)`

### Request options

- `params` → appended as query string
- `data` → JSON body (skipped on GET)
- `isMultipart` → switches to FormData via `createFormData()`; arrays become `key[index]`
- `skipAuth` → suppresses Bearer header
- `headers` → merged into default headers

### Auth

Auto-attaches `Authorization: Bearer <token>` if token present and `skipAuth`
not set. Use `setToken` / `getToken` / `clearToken` to manage. The public site
has no login flow.

### Error shape

Throws `Error` with `.status` and `.data` attached. Caller is expected to
`try/catch` and degrade gracefully (empty array, `null`, etc.).

## `desaService.js`

All data access against the public REST API. **No legacy mocks** — the service
talks only to the backend.

### Helpers

- `unwrapData(response)` → `response.data` if present, else `response`.
- `unwrapList(response, key)` → array under `data[key]` / `data` / `data.data[key]` / `[]`.
- `unwrapPaginated(response, key)` → `{ items, pagination }`.
- `fetchAllPages(endpoint, key, params, { pageSize=100, maxPages=50 })` → walks
  `total_pages` (server `MaxLimit` is 100) and concatenates all items.
- `normalize(item)` / `normalizeList(items)` → flattens `category: {id, name}`
  to the plain name string. Every list/detail response passes through this.

### Live functions

| Function | HTTP | Returns / Notes |
|---|---|---|
| `getActiveBanners()` | `GET /banners/active` | raw array |
| `getPublicBeritaList(params)` | `GET /public/berita/list` | `[{id,title,category,created_at,media,…}]`; supports `q`, `category` (UUID) |
| `getPublicBeritaCategoryIdByName(name)` | `GET /public/berita/categories` | UUID of a category by name (case-insensitive), or `null` |
| `getPublicBeritaById(id)` | `GET /public/berita/:id` | single |
| `getPublicDesa()` | `GET /public/desa` | single `DesaResponse` |
| `getPublicFasilitasList(params)` | `GET /public/fasilitas/list` | flat array (paginated API, items only) |
| `getPublicFasilitasAll(params)` | `GET /public/fasilitas/list` | `{items, pagination}` — walks all pages |
| `getPublicFasilitasCategories()` | `GET /public/fasilitas/categories` | `[{id,name,usage_count}]` |
| `getPublicPPIDList(params)` | `GET /public/ppid/list` | flat array |
| `getPublicPPIDById(id)` | `GET /public/ppid/:id` | single (full description) |
| `createPPIDRequest(id, data)` | `POST /public/ppid/:id/requests` | `{requester_name, requester_email, purpose}` |
| `getPublicStrukturList(params)` | `GET /public/struktur/list` | flat array (single page) |
| `getPublicStrukturAll(params)` | `GET /public/struktur/list` | flat array — walks all pages (Perangkat) |
| `getPublicUMKMList(params)` | `GET /public/umkm/list` | flat array |
| `getPublicProfileList(params)` | `GET /public/profile/list` | flat array (single page) |
| `getPublicProfileAll(params)` | `GET /public/profile/list` | flat array — walks all pages |
| `getPublicProfileCategories(params)` | `GET /public/profile/categories` | `[{id,name,sort_order,usage_count}]` — drives Profil group order |
| `getPublicInfographicList({page,limit})` | `GET /public/infographic/list` | `{infographics, pagination}` |
| `getPublicInfographicWithToken(id)` | `GET /public/infographic/:id` | `{infographic, token}`; throws `.code='rate_limited'` (429) / `'not_found'` (404) |
| `getInfographicEmbedUrl(base, infographic, token)` | — | builds `/embed/dashboard|question/{token}` |
| `getPublicGalleryFolders(params)` | `GET /public/gallery/folders` | `{folders, pagination}` (public folders with public media only) |
| `getPublicGalleryFolderById(id)` | `GET /public/gallery/folders/:id` | `{folder, media, pagination}` (public media only) |

### URL helpers

- `resolveGalleryAssetUrl(value)` → absolute URL. Signed media paths
  (`/api/v1/media/...?jwt=`) are resolved against the **API origin**, not the
  SPA origin. Use this for every API-emitted media URL.
- `mediaUrl(item, variant)` / `thumbnailUrl(item)` / `documentUrl(item)`
  → tolerant field reads for feature media.
- `resolveMediaThumbnailUrl(media)` / `resolveMediaContentUrl(media)` / `resolveFolderCoverUrl(folder)`.
- `fileSrc(filename)` → `${origin}/files/{filename}` (mounted at API root, not under `/api/v1`).
- `decodeJwtExp(token)` → ms epoch, used to schedule infographic token refresh.
- `formatMediaDate(dateString)` → `id-ID` short date.

### Pagination

`*List(params)` forward `{page, limit}` to the API but currently return only the
first page's items. `getPublicFasilitasAll` is the exception — it walks all
pages (used by Peta).

### Response unwrapping

`apiClient.request()` returns the parsed body directly. The API wraps responses
as `{ success, data }`; the `unwrap*` helpers peel that. No `data` wrapper is
returned to callers.
