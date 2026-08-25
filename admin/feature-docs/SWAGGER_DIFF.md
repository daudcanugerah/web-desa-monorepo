# Swagger vs Implementation — Diff Summary

Generated after `swagger.yaml` was rewritten from OpenAPI 3.0 → **Swagger 2.0** (`basePath: /api/v1`, `host: localhost:8080`, `schemes: [http]`) and substantially expanded (2152 → 5691 lines).

---

## Top-line numbers (updated after fixes)

| Scope | Swagger 2.0 | Frontend impl | Coverage |
|-------|-------------|---------------|----------|
| Total operations | **113** | 91 | **80.5%** |
| Admin / non-public | **88** | 78 | **88.6%** |
| Public (`/public/*`, `/auth/*`, `/health`, `/files/*`) | 25 | 13 (4 auth + 9 public-* for category list/datalist) | 52% |

**No divergences** — every impl endpoint has a matching swagger spec.
**10 admin endpoints unused** — all are category-management endpoints for resources without a category field, plus 3 helpers.

---

## Breaking / blocking divergences (fix these first)

### 🔴 1. PPID download — singular vs plural (still broken)

| | Path |
|---|------|
| **Impl** | `GET /ppid/documents/{documentId}/download` |
| **Swagger** | `GET /ppid/document/{documentId}/download` (singular) |

- Source: `src/services/ppid.service.js:25`.
- Swagger location: `swagger.yaml:2614`.
- **Effect**: download will 404. Same bug I already flagged in `feature-docs/00-architecture.md:262` and `feature-docs/10-ppid.md:155` — the new swagger confirms it is still wrong on the **frontend** side.

### 🔴 2. PPID categories — impl expects GET, swagger defines only POST

| | Path |
|---|------|
| **Impl** | `GET /ppid/categories` |
| **Swagger** | `POST /ppid/categories` (line 2506), `DELETE /ppid/categories/{id}` (line 2562). **No `GET`.** |

- Source: `src/services/ppid.service.js:17` (`getCategories`).
- Used by `PpidListView` to populate the category filter dropdown.
- **Effect**: the filter dropdown will be empty (request will 405/404).
- **Fix options**:
  - Add a `GET /ppid/categories` to the swagger spec and backend.
  - Or remove the category filter from `PpidListView` and collect categories client-side (like `NewsListView` does).

### 🟡 3. Facility image deletion — two designs, only one used

| | Approach |
|---|----------|
| **Swagger** | `DELETE /fasilitas/{id}/images/{imageIndex}` (line 1479) — remove one image by zero-based index |
| **Impl** | sends `removed_images: JSON.stringify([...])` as a `FormData` field on `PUT /fasilitas/{id}` (`FacilityFormView.vue:241-244`) |

- The swagger endpoint is implemented in the backend but **never called from the admin**.
- `facility.service.js` has no per-image delete method.
- Either is valid, but the current state means the swagger endpoint is **dead code** and the admin depends on a backend behavior that the swagger doesn't document. Pick one and document it.

---

## Spec gaps the admin needs but the backend doesn't expose

These were already noted in feature-docs as "spec doesn't define" — the new swagger confirms they still don't exist.

| Wanted endpoint | Used by | Swagger status |
|----------------|---------|----------------|
| `GET /permissions` | Could replace the hard-coded `KNOWN_PERMISSIONS` list in `RoleListView` | ✅ **Defined** at swagger.yaml:2184 (was missing from prior summary) |

---

## Spec endpoints the admin doesn't use (informational, post-fix)

After the previous round of fixes, the remaining **10 admin endpoints not consumed** fall into 3 buckets. They are all **informational**, not blocking — every one of them is a backend capability without a corresponding admin UI surface.

### Bucket A — category CRUD for resources that have no `category` field

The admin frontend exposes a `category` input + `Kelola Kategori` modal only on **Berita, UMKM, PPID**. The backend also has category CRUD for **Banners, Fasilitas, Infographic** but no admin form has a category input on those resources (Banners has none, Fasilitas uses `type`, Infographic uses `component_type`).

| Endpoint | RBAC | Purpose |
|----------|------|---------|
| `POST   /banners/categories` | `banners:write` | Create banner category |
| `GET    /banners/categories` | `banners:read`  | Paginated list with usage counts + search |
| `DELETE /banners/categories/{id}` | `banners:write` | Delete (409 if in use) |
| `POST   /fasilitas/categories` | `fasilitas:write` | Create fasilitas category |
| `DELETE /fasilitas/categories/{id}` | `fasilitas:write` | Delete (409 if in use) |
| `POST   /infographic/categories` | `infographic:write` | Create infographic category |
| `DELETE /infographic/categories/{id}` | `infographic:write` | Delete (409 if in use) |

> **To wire these up you'd need to first add a `category` field to `BannerFormView`, `FacilityFormView`, `InfographicFormView`** — that's net-new functionality, not a divergence fix. Backend is ready.

### Bucket B — public-facing helper

| Endpoint | RBAC | Purpose |
|----------|------|---------|
| `GET /banners/active` | public | Lists up to 100 active banners. Consumed by the village public site, not the admin. Could be surfaced in a future "preview public site" admin view. |

### Bucket C — admin helpers the frontend hasn't adopted yet

| Endpoint | RBAC | Purpose |
|----------|------|---------|
| `GET /infographic/sections/names` | `infographic:read` | Distinct infographic section names. Useful only if the admin grows a `section_name` input on `InfographicFormView`. |
| `GET /permissions` | `roles:read` | De-duplicated `(resource, action)` pairs. Would replace the hard-coded `KNOWN_PERMISSIONS` list in `RoleListView.vue` — currently the only fully-i18n'd view. |

---

## Notable non-endpoint changes in the new swagger

These are not endpoint-level diffs but show up in the new spec and may matter:

| Change | Detail | Implication |
|--------|--------|------------|
| Spec version | Swagger 2.0 (was OpenAPI 3.0) | If you have any tooling that consumes the spec (codegen, docs), update it. |
| `basePath` is now in the spec | `/api/v1` | Matches the default `VITE_API_BASE_URL`. |
| **RBAC annotations** | Every admin operation has `description: "RBAC: <resource>:<action>."` | The backend enforces permissions via **Casbin**; the admin needs at least one user with the right permission to use each feature. The hard-coded `KNOWN_PERMISSIONS` list in `RoleListView` should mirror this catalog: `banners`, `berita`, `umkm`, `fasilitas`, `ppid`, `struktur`, `desa`, `roles`, `profile`, `infographic` — all with `read` + `write`. ✅ matches. |
| JWT TTLs | access 24h, refresh 7d | No code change; just for awareness. |
| Password reset rate limit | 5 requests / hour / email → `429` | `ForgotPasswordView` should show a friendly "too many attempts" message on 429 (currently shows generic error). |
| General rate limiting | `429 Too Many Requests` documented on many endpoints | The api.js interceptor only handles 401. A 429 will fall through to the per-view `notification.error(err.response?.data?.error)` path. |
| Login rate limit | `429` documented | Same — the login view should ideally surface "too many attempts". |

---

## Recommended fix order

1. **Fix `/ppid/document/{id}/download`** — change impl to singular path (`ppid.service.js:25`). ✅ **DONE**
2. **Fix `/ppid/categories` GET** — either remove the dropdown in `PpidListView` or add a GET to the backend. ✅ **DONE** — switched to `/public/ppid/categories`
3. **Resolve facility image deletion** — either remove `DELETE /fasilitas/{id}/images/{imageIndex}` from the swagger (since it's unused) or switch the frontend to call it. ✅ **DONE** — frontend now calls DELETE per-image with optimistic UI
4. **Surface 429 errors** — add a `notification.warning("Terlalu banyak percobaan")` branch in `LoginView` and `ForgotPasswordView` for `err.response?.status === 429`. ✅ **DONE** — i18n key `auth.rateLimitExceeded` added
5. **Add category CRUD UI to 3 resources** (Berita, UMKM, PPID). ✅ **DONE** — `CategoryManagerModal.vue` + per-service `getCategories` / `createCategory` / `deleteCategory`
6. **Consider `GET /permissions`** — replace `KNOWN_PERMISSIONS` in `RoleListView` with a live fetch from the new endpoint. ⏳ Pending — straightforward 1-line change in `RoleListView`
7. **Decide on category management for Banners / Fasilitas / Infographic** — would require adding a `category` field to each form view first (net-new functionality, not a divergence fix). ⏳ Decision required
8. **Update `ARCHITECTURE.md`** — references OpenAPI 3.0; should now mention Swagger 2.0 / the `basePath: /api/v1` model. ⏳ Pending

---

## Files touched by fixes

| File | What changes |
|------|-------------|
| `src/services/ppid.service.js` | download path singular ✅, getCategories now hits `/public/ppid/categories` ✅, added `createCategory` + `deleteCategory` ✅ |
| `src/services/news.service.js` | added `getCategories` / `createCategory` / `deleteCategory` ✅ |
| `src/services/umkm.service.js` | added `getCategories` / `createCategory` / `deleteCategory` ✅ |
| `src/services/facility.service.js` | added `deleteImage(id, imageIndex)` ✅ |
| `src/views/content/PpidListView.vue` | category filter now hits `getCategories` ✅ |
| `src/views/content/NewsFormView.vue` | category field + datalist + Kelola button ✅ |
| `src/views/content/UmkmFormView.vue` | category field + datalist + Kelola button ✅ |
| `src/views/content/PpidFormView.vue` | category field + datalist + Kelola button ✅ |
| `src/views/content/FacilityFormView.vue` | image removal → DELETE per-image with rollback ✅, dropped `removed_images` ✅ |
| `src/views/auth/LoginView.vue` + `ForgotPasswordView.vue` | 429 branch + i18n key ✅ |
| `src/components/common/CategoryManagerModal.vue` | **NEW** reusable modal ✅ |
| `src/i18n/locales/{id,en}.json` | `auth.rateLimitExceeded` key ✅ |
| `feature-docs/{07,08,09,10,15}-*.md` | Updated for new flows ✅ |