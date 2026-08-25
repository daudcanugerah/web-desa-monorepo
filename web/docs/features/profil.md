# Profil

`pages/Profil.vue` → route `/profil`. Village profile with sidebar categories.

## Categories

Hardcoded list of 4:

| ID | Name | Icon color | Purpose |
|---|---|---|---|
| `Sejarah` | Sejarah | blue | `data.history` (falls back to `description`) |
| `Visi Misi` | Visi Misi | emerald | `data.vision` + `data.mission[]` (numbered list) |
| `Struktur` | Struktur | purple | `data.structure` |
| `Perangkat` | Perangkat | amber | `data.officials[]` (grid + pagination) |

## State

| Ref | Default | Notes |
|---|---|---|
| `loading` | `true` | |
| `selectedCategory` | `''` | empty → "Pilih kategori dari sidebar" |
| `showSidebar` | `false` | mobile bottom-sheet toggle |
| `isMobile` | `false` | `innerWidth < 768` |
| `searchText` | `''` | filters category list |
| `currentPage` | `1` | Perangkat pagination |
| `itemsPerPage` | `6` (const) | |
| `data` | `{ history, vision, mission, structure, officials }` | |

## Computed

- `filteredCategories` — text search (`name` / `description`, lowercased) + selectedCategory filter
- `totalPages` — `Math.ceil(officials.length / 6)`
- `paginatedOfficials` — slice by `startIndex`/`endIndex`

## Methods

- `selectCategory(id)` — sets selection, resets `currentPage` to 1
- `checkMobile()` — updates `isMobile` on resize

## API

- `getPublicDesa()` → history/vision/mission/structure
- `getPublicStrukturList()` → officials array (each `{ id, name, position, phone }`)

## Layout

`h-[calc(100vh-80px)]` split-pane (sidebar + content). Mobile bottom-sheet via `<Transition name="slide-up">`. Scoped `<style>` defines the slide-up transform.

## Notes

- `onBeforeUnmount` is **missing** for the `resize` listener — minor leak. Same in Infografik.
- Officials show static avatar placeholder (no `photo` field used in template).
- Pagination only shown for Perangkat view, only when `totalPages > 1`.
