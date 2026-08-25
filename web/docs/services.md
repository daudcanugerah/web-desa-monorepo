# Services

## `apiClient.js`

Singleton `APIClient` class exported as `apiClient`. Fetch-based.

```js
import { apiClient } from '@/services/apiClient'
const data = await apiClient.get('/public/berita/list', { params: { page: 1, limit: 10 } })
await apiClient.post('/public/ppid/123/requests', { requester_name, requester_email })
```

### Config

- `API_BASE_URL` = `import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'`
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

Auto-attaches `Authorization: Bearer <token>` if token present and `skipAuth` not set. Use `setToken` / `getToken` / `clearToken` to manage. No login flow wired up.

### Error shape

Throws `Error` with `.status` and `.data` attached. Caller is expected to `try/catch` and degrade gracefully (empty array, `null`, etc.).

## `desaService.js`

All data access against the public REST API. **No legacy mocks** — the service talks only to the backend.

### Normalization

OpenAPI returns `category: {id, name}` (CategoryInfo object) on every list/detail response. The service flattens it to a plain string via `normalize()` so pages can use it as `item.category === 'Kuliner'` without unwrapping manually. Items also pass through unchanged for everything else.

### Live functions

| Function | HTTP | Notes |
|---|---|---|
| `getActiveBanners()` | `GET /banners/active` | raw array (public) |
| `getPublicBannerList(params)` | `GET /public/banner` | paginated `{banners, pagination}` |
| `getPublicBannerById(id)` | `GET /public/banner/:id` | single |
| `getPublicBeritaList(params)` | `GET /public/berita/list` | paginated `{berita, pagination}` |
| `getPublicBeritaById(id)` | `GET /public/berita/:id` | single |
| `getBeritaCategories()` | `GET /berita/categories` | admin, unused |
| `getPublicDesa()` | `GET /public/desa` | single `desa.DesaResponse` |
| `getPublicFasilitasList(params)` | `GET /public/fasilitas/list` | paginated `{fasilitas, pagination}` |
| `getPublicFasilitasById(id)` | `GET /public/fasilitas/:id` | unused |
| `getPublicPPIDList(params)` | `GET /public/ppid/list` | paginated `{ppid, pagination}` |
| `getPublicPPIDById(id)` | `GET /public/ppid/:id` | single (full description) |
| `createPPIDRequest(id, data)` | `POST /public/ppid/:id/requests` | `{requester_name, requester_email, purpose}` |
| `getPPIDCategories()` | `GET /ppid/categories` | admin, unused |
| `getPublicStrukturList(params)` | `GET /public/struktur/list` | paginated `{struktur, pagination}` |
| `getPublicStrukturById(id)` | `GET /public/struktur/:id` | unused |
| `getPublicUMKMList(params)` | `GET /public/umkm/list` | paginated `{umkm, pagination}` |
| `getPublicUMKMById(id)` | `GET /public/umkm/:id` | unused |
| `getUMKMCategories()` | `GET /umkm/categories` | admin, unused |
| `getPublicProfileList(params)` | `GET /public/profile/list` | paginated `{profile, pagination}` |
| `getPublicProfileById(id)` | `GET /public/profile/:id` | unused |
| `getPublicInfographicList(params)` | `GET /public/infographic/list` | paginated `{infographic, pagination}` |
| `getPublicInfographicById(id)` | `GET /public/infographic/:id` | unused |

### Server-side pagination

All `*List(params)` accept `{page, limit}` and forward to the API. Pages currently fetch the full list and paginate client-side — the params are wired but not yet passed by callers.

### Response unwrapping

`apiClient.request()` returns the parsed body directly. List functions return `response[resourceName] || []`. Detail functions return `response || null`. No `data` wrapper.
