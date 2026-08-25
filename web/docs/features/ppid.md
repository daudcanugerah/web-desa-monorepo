# PPID

`pages/PPID.vue` → route `/ppid`. Public Information Disclosure (Pejabat Pengelola Informasi dan Dokumentasi).

Most complex feature: list view + detail view + request form modal + success dialog, all in one file.

## View Modes

Single page toggles between two views via `selectedPPID`:

- `selectedPPID === null` → **List view**
- `selectedPPID` set → **Detail view** (same file, different template branch)

## List View

### Filters (3)

| Ref | Type | Default |
|---|---|---|
| `searchQuery` | text | `''` |
| `dateFilter` | `'all' \| 'last30days' \| 'last90days' \| 'thisYear'` | `'all'` |
| `selectedCategory` | string | `'Semua'` |
| `currentPage` | number | `1` |
| `itemsPerPage` | const | `9` |

### Computed

- `categories` — `['Semua', ...new Set(data.value.map(doc => doc.category))]`
- `filteredDocuments` — category → search (`title` / `description`) → date (vs `publication_at`)
- Standard `totalPages` / `startIndex` / `endIndex` / `paginatedDocuments`

### UI

- List rows (not cards): thumbnail (or placeholder icon) + title + category pill + description (HTML) + publication date + "Lihat Detail" button
- 3-col filter bar (search / date / category)
- "Reset Filter" when any filter active
- Pagination when `totalPages > 1`

### Thumbnail URL

```js
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
:src="`${apiBaseUrl}/files/${doc.thumbnail_url}`"
@error="handleImageError"   // hides broken image
```

## Detail View

### Flow

1. `viewDetail(doc)` — sets `selectedPPID = doc` (lightweight), `currentPage = 1`, then `fetchFullDetail(doc.id)` to swap in full description.
2. Detail shows title, category pill, `publication_at` date, `description` (HTML), then a 2×2 info grid (Kategori, Tanggal Publikasi, Dibuat, Diperbarui).
3. Action buttons: **Minta Dokumen** → opens request modal. **Kembali** → `backToList()`.

## Request Modal (`showRequestDialog`)

Fields:

| Field | Validation |
|---|---|
| `requesterName` | required, trim-checked |
| `requesterEmail` | required, `^[^\s@]+@[^\s@]+\.[^\s@]+$` |
| `requestPurpose` | optional textarea, 3 rows |

Submission:

```js
await createPPIDRequest(selectedPPID.value.id, {
  requester_name: requesterName.value,
  requester_email: requesterEmail.value,
  purpose: requestPurpose.value || undefined
})
```

`isSubmitting` disables inputs/buttons, shows spinner on submit. Errors render in `requestError` ref (red box).

## Success Dialog (`showSuccessDialog`)

Shows after successful POST. "Tutup" button → `closeSuccessDialog()` → `backToList()` (clears `selectedPPID`, closes request modal).

## API

- `getPublicPPIDList()` → list
- `getPublicPPIDById(id)` → full doc (with description)
- `createPPIDRequest(id, { requester_name, requester_email, purpose })` → POST

## Notes

- The list view's `description` is HTML — rendered with `v-html` (XSS risk if backend doesn't sanitize).
- `description` in list view comes from list response, not from `getPublicPPIDById`. Detail view overwrites it with the full version. List description may be truncated.
- No real file download — request flow only triggers email follow-up (per `openapi.yaml` token-based approval).
- `apiBaseUrl` is computed in `setup()` but also computed once for thumbnail — duplicated logic with `apiClient` (which doesn't expose baseURL).
