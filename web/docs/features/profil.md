# Profil

`pages/Profil.vue` → route `/profil`. Village profile with sidebar category groups.

## Data sources

Fully API-driven — no hardcoded section/category list.

| Source | Function | Endpoint |
|---|---|---|
| Sections | `getPublicProfileAll()` | `GET /public/profile/list` (walks all pages) |
| Categories (group nav) | `getPublicProfileCategories()` | `GET /public/profile/categories` |
| Perangkat / officials | `getPublicStrukturAll()` | `GET /public/struktur/list` (walks all pages) |

Each profile row carries `category: {id, name}` (nullable) and `section_name`.
Categories carry `sort_order` + `usage_count`.

## Grouping & ordering

- Section → group = the category the API attached to the row (`category.name`).
  No frontend inference: an uncategorized section groups under its own name, so
  Visi and Misi stay independent use cases and are never merged.
- Perangkat (officials) is not a profile row; its data comes from the struktur
  API and the nav entry only renders when struktur data exists. Since no API row
  carries a category for it, it groups under its own name (`Perangkat`).
- Group display order = API `sort_order` ascending (`categoryOrder` /
  `categoryRank`); unknown groups sort last, then alphabetically.
- `EXCLUDED_SECTIONS = ['Struktur Organisasi']` hides one legacy section.

## Views

| View | Condition | Content |
|---|---|---|
| Perangkat Desa | `isPerangkat` | officials grid (photo/name/position/phone/description) + pagination (6/page) |
| Misi | `isMisi` | numbered list, `stripHtml` |
| Generic section | else | rich text via `sanitize(s.content)` (allowlist HTML) |

## State

| Ref | Default | Notes |
|---|---|---|
| `loading` | `true` | |
| `selectedSection` | first nav entry | set after fetch |
| `showSidebar` | `false` | mobile bottom-sheet toggle |
| `isMobile` | `false` | `innerWidth < 768` |
| `searchText` | `''` | filters nav groups (section + group name) |
| `currentPage` | `1` | Perangkat pagination |
| `itemsPerPage` | `6` (const) | |
| `profiles` / `officials` / `apiCategories` | `[]` | API results |

## Computed

- `activeSections` — rows for `selectedSection` where `state !== false`
- `apiSectionNames` — distinct section names minus `EXCLUDED_SECTIONS`
- `categoryOrder` / `categoryRank` — API-sorted categories
- `navGroups` / `filteredGroups` / `sectionCount` / `activeGroupName`
- `totalPages` / `paginatedOfficials` — Perangkat slice

## Methods

- `selectSection(name)` — sets selection, resets `currentPage`
- `checkMobile()` — updates `isMobile` on resize (listener removed in `onBeforeUnmount`)
- `sanitize(html)` — allowlist HTML sanitizer for rich content
- `officialPhoto(item)` — `mediaUrl(item, 'single')`

## Layout

`h-[calc(100dvh-116px)]` split-pane (content + sidebar). Mobile bottom-sheet via
`<Transition name="slide-up">`.

## Notes

- Officials fall back to a person-icon avatar when `mediaUrl` is empty.
- Pagination shown only for Perangkat and only when `totalPages > 1`.
