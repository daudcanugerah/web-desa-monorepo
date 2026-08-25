# Swagger v3 — Incomplete Implementation Fix Plan

Generated from a diff of `swagger.yaml` (v3) against the current frontend code. Captures every site that is broken, partially wired, or otherwise out of sync with the spec.

**Current state**: 78 / 88 admin endpoints consumed (88.6%). 0 endpoint divergences. Several **code-level** divergences remain.

---

## Priority 1 — NewsListView is non-functional (5 bugs)

`src/views/content/NewsListView.vue` has the worst divergence in the app. Every filter and search returns empty because all four query-param names are wrong.

| Line | Frontend sends | Swagger expects |
|---|---|---|
| 118 | `params.search = ...` | `q` |
| 119 | `params.date_from = ...` | `since` |
| 120 | `params.date_to = ...` | `until` |
| 121 | `params.category = <name>` | `category` = `<UUID>` |
| 23  | dropdown `value` is category name | dropdown `value` must be category UUID |

Plus line 129 adds `item.category` (now `{id, name}` object) to a `Set<string>` — won't dedupe, and the column cell renders `[object Object]`.

**Acceptance criteria**:
- Search input filters results via `?q=`.
- Category dropdown sends `?category=<UUID>`, not the name.
- Date pickers send `?since=YYYY-MM-DD&until=YYYY-MM-DD`.
- Category column cell shows `row.category.name`.

---

## Priority 2 — UmkmListView category cell

`src/views/content/UmkmListView.vue` line 66 declares a `category` column but the cell template is missing — falls through to default rendering which prints the raw `{id, name}` object.

**Acceptance**: column cell shows `row.category?.name`.

---

## Priority 3 — Verify response envelope shape

Swagger v3 defines every response with top-level fields (no `{data: {...}}` wrapper). The frontend is inconsistent:

| File | Line | Code | Assumes |
|---|---|---|---|
| `services/api.js` (refresh) | 35 | `setTokens(data)` where `data = axios_response.data` | flat |
| `views/auth/LoginView.vue` | 91 | `authStore.setTokens(response.data.data)` | wrapped |
| `views/auth/ResetPasswordView.vue` | 135 | `if (!res.data?.data?.valid)` | wrapped |
| `stores/auth.js` | 35 | `this.currentUser = response.data.data` | wrapped |

**Step 1**: `curl -X POST <api>/auth/login -d '{"email":"...","password":"..."}'` and inspect the actual response body.
**Step 2**: Apply the matching assumption everywhere (patch the 3 wrong-envelope sites).

**Note**: every list view has `res.data.data.X` accesses (~23 sites). If the backend is actually flat, those break too — same fix applies to all.

---

## Priority 4 — Patch envelope assumption (gated on P3 result)

Three frontend sites are wrong depending on backend shape. After verifying via curl:

**If backend returns wrapped** (e.g. `{success, data: {users: [...]}}`):
- `api.js:35` `setTokens(data)` → `setTokens(data.data)` (and same for the auth store sync)
- No changes to list views

**If backend returns flat** (e.g. `{users: [...]}`):
- `LoginView.vue:91` `response.data.data` → `response.data`
- `ResetPasswordView.vue:135` `res.data?.data?.valid` → `res.data?.valid`
- `stores/auth.js:35` `response.data.data` → `response.data`
- `services/api.js` is already correct
- Same fix for `response.data.data.X` in all list views (~23 sites)

---

## Priority 5 — Add missing `security: BearerAuth` to 5 admin endpoints in swagger

| Endpoint | Method |
|---|---|
| `/banners/{id}/status` | PATCH |
| `/ppid/document/{documentId}/download` | GET |
| `/users/{id}/password` | PUT |
| `/users/{id}/roles` | POST |
| `/users/{id}/roles/{role}` | DELETE |

Also add explicit `security: []` to `/banners/active` to document that it's intentionally public.

---

## Priority 6 — Wire `GET /permissions` to replace hard-coded list

`src/views/users/RoleListView.vue:227` (`KNOWN_PERMISSIONS`) is a 12-entry hard-coded array of `(resource, action)` pairs. Swagger v3 exposes `GET /permissions` which returns the de-duplicated set from the backend (Casbin catalog). Replace the hard-coded list with a live fetch — keeps the role editor in sync with what the backend actually enforces.

---

## Priority 7 — Decide on category fields for banner / fasilitas / infographic

These resources have category CRUD endpoints in the backend but no `category` field in the admin form:

- **Banner**: no category input at all
- **Fasilitas**: uses `type` field instead of `category`
- **Infographic**: uses `component_type` instead of `category`

This is **architectural**, not a bug. Decide whether to:
- (a) Add a `category` field to each form + wire the existing `CategoryManagerModal` pattern (3 services × `{getCategories, createCategory, deleteCategory}` + 3 form-view wirings)
- (b) Leave as-is and document that admin manages categories only for berita/umkm/ppid

---

## 10 endpoints unused by admin (informational, not bugs)

| Endpoint | Notes |
|---|---|
| `GET/POST/DELETE /banners/categories{,/{id}}` | blocked on P7 |
| `POST/DELETE /fasilitas/categories{,/{id}}` | blocked on P7 |
| `POST/DELETE /infographic/categories{,/{id}}` | blocked on P7 |
| `GET /banners/active` | public — consumed by village site |
| `GET /infographic/sections/names` | optional dropdown for InfographicFormView |
| `GET /permissions` | see P6 |

---

## Execution order

1. **P1**: Fix `NewsListView` (single file, high impact).
2. **P2**: Fix `UmkmListView` category cell (single file, low effort).
3. **P3**: Curl backend to verify envelope shape.
4. **P4**: Patch the 3 (or 26) wrong-envelope sites.
5. **P5**: Add `security: BearerAuth` to 5 swagger endpoints (spec hygiene, no code change).
6. **P6**: Wire `GET /permissions` in `RoleListView`.
7. **P7**: Defer — requires product decision.

**Commit per wave**. Lint + build after each wave.

---

## Tracking

See the todo list in the conversation for live status.