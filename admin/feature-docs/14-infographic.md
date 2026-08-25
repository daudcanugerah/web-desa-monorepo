# 14 — Infographic Dashboard

Embed Metabase dashboards / questions into the admin (and surface them on the public site). Uses a preview-token pattern.

---

## 1. Purpose

Manage a list of embedded Metabase components (dashboards or questions). Admins create an entry by selecting a Metabase component type + id, then preview it before publishing.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/infographic` | `Infographic` | Infographic Dashboard | `requiresAuth` |
| `/infographic/create` | `InfographicCreate` | Tambah Dashboard | `requiresAuth` |
| `/infographic/:id/edit` | `InfographicEdit` | Edit Dashboard | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/content/InfographicListView.vue` | Filterable list with modal preview |
| `src/views/content/InfographicFormView.vue` | Create + edit with live preview (split layout) |
| `src/services/infographic.service.js` | 6 endpoint wrappers + preview-token |
| `src/composables/useMetabase.js` | Lazy-loads Metabase embed.js + sets `window.metabaseConfig` |
| `src/components/common/AppTable.vue` | List table |
| `src/components/common/AppModal.vue` | Preview modal |
| `src/components/common/AppPagination.vue` | Pager |
| `src/components/common/AppBackButton.vue` | Back-to-list |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/infographic?page&limit&search&component_type&state` | — | List view |
| `GET`    | `/infographic/{id}`                                    | — | Form view (edit) |
| `POST`   | `/infographic`                                         | `{ title, component_type, component_id, state? }` | Form view (create) |
| `PUT`    | `/infographic/{id}`                                    | `{ title?, component_type?, component_id?, state? }` | Form view (edit) |
| `DELETE` | `/infographic/{id}`                                    | — | List view (delete) |
| `POST`   | `/infographic/preview/token`                           | `{ component_id, component_type }` | Form view (live preview) → returns `{ token }` |

Swagger also defines `/infographic/sections/names` but it is not currently consumed.

---

## 5. Data Model

```ts
type Infographic = {
  id: string
  title: string
  component_type: 'dashboard' | 'question'
  component_id: number              // Metabase internal id
  state: 'draft' | 'published'     // confirm exact enum
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 Metabase Bootstrap

`useMetabase()` composable:

1. On first call, injects `<script src="${VITE_METABASE_URL}/app/embed.js">` into the document head.
2. Sets `window.metabaseConfig = { isGuest: true }`.
3. The Metabase script then registers the custom elements `<metabase-dashboard>` and `<metabase-question>`.

### 6.2 List (`InfographicListView.vue`)

- Filter row: search input, component_type dropdown, state dropdown.
- Table columns: Title | Component Type | State | Aksi.
- **Preview action**: opens an `AppModal` containing either `<metabase-dashboard>` or `<metabase-question>` with a freshly fetched preview token.

### 6.3 Form (`InfographicFormView.vue`)

- Split layout: form left, preview right (or below on small screens).
- Fields:
  - `title` — required.
  - `component_type` — radio (dashboard / question).
  - `component_id` — number input (the Metabase id).
  - `state` — radio (draft / published).
- **Live preview**:
  - On any of `component_type` / `component_id` change (debounced), calls `POST /infographic/preview/token` with `{ component_id, component_type }`.
  - On success, renders `<metabase-dashboard :token="...">` or `<metabase-question :token="...">` on the right.
  - Shows a loading spinner while waiting for the token.
- Submit → service call → success toast + navigate to list.

### 6.4 Delete

- Confirm dialog → `DELETE /infographic/{id}`.

---

## 7. UI Notes

- The preview pane uses the Metabase web component, which auto-sizes itself.
- A loading skeleton shows inside the preview area while the token is being fetched.

---

## 8. Notes & Gotchas

- **Metabase script is loaded lazily** but only once per session — there is no cleanup.
- **Token TTL** is determined by the backend; long-running edit sessions may need to re-fetch.
- **`window.metabaseConfig.isGuest = true`** — Metabase treats the embed as a guest. If guest access is restricted on the Metabase server, previews will fail.
- **No image / fallback** — if Metabase is down, the preview area shows a Metabase-native error, not a friendly one.
- **Pagination** is computed locally (`Math.ceil(total / limit)`) rather than using `pagination.total_pages`. See `00-architecture.md` §10.