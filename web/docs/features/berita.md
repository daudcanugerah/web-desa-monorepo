# Berita

Two pages: list (`pages/Berita.vue` → `/berita`) and detail (`pages/BeritaDetail.vue` → `/berita/:slug`).

## List (`Berita.vue`)

### Filters

| Ref | Type | Default |
|---|---|---|
| `searchQuery` | text | `''` |
| `dateFilter` | `'all' \| 'last30days' \| 'last90days' \| 'thisYear'` | `'all'` |
| `currentPage` | number | `1` |
| `itemsPerPage` | const | `9` |

### `filteredBerita` computed

- Search: matches `title`, `excerpt`, `content`, or `category` (lowercased substring)
- Date: `news.date || news.publication_at || news.created_at` against cutoff
- Returns full filtered list; pagination slices it

### Computed

- `totalPages`, `startIndex`, `endIndex`, `paginatedBerita` — standard pagination pattern
- `getDate(berita)` helper: tries `date` → `publication_at` → `created_at`

### UI

- 3-col grid on `lg:`, 2-col on `md:`, 1-col mobile
- Card → `RouterLink` to `/berita/${news.slug}`
- Pagination: prev + page numbers + next
- "Reset Filter" button shows only when any filter is active

### API

- `getPublicBeritaList()` → array of `{ id, slug, title, excerpt, content, date, publication_at, created_at, category, author, thumbnail, featuredImage }`

## Detail (`BeritaDetail.vue`)

### Route

`/berita/:slug` — uses `route.params.slug` as the ID. **Caveat:** the param is called `slug` but `getPublicBeritaById` hits `/public/berita/:id`. If backend expects UUID, the route would be a mismatch. Treat slug and ID as interchangeable until backend clarifies.

### State

- `data` — single article
- `loading` — toggled in `onMounted` `finally`

### UI

- Back link to `/berita`
- 96-tall gradient hero (no real image, placeholder)
- Category pill + date (long Indonesian format)
- Author line
- `data.content` rendered as plain text (no `v-html`, no markdown)

### API

- `getPublicBeritaById(route.params.slug)` → full article

## Notes

- `formatDate` differs between pages: list uses `month: 'short'`, detail uses `month: 'long'`.
- Both pages tolerate missing dates via `getDate` fallback chain.
- `EmptyState` shown when list empty or detail not found.
