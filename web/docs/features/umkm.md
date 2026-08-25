# UMKM

`pages/UMKM.vue` → route `/umkm`. Local business (SME) directory.

## Card

Each card shows: name, category pill, description (line-clamp 3), contact person, phone (`tel:` link), address (line-clamp 1). Static SVG icon (no logo image).

## Filters

| Ref | Type | Default |
|---|---|---|
| `searchQuery` | text | `''` |
| `selectedCategory` | string | `'Semua'` |
| `currentPage` | number | `1` |
| `itemsPerPage` | const | `9` |

### `categories` computed

```js
return ['Semua', ...new Set(data.value.map(umkm => umkm.category))]
```

Derived from data, not hardcoded.

### `filteredUMKM` computed

- Category: `umkm.category === selectedCategory` (skipped when `'Semua'`)
- Search: matches `name`, `description`, `category`, `contact`, `address` (lowercased substring)

### Computed

- `totalPages`, `startIndex`, `endIndex`, `paginatedUMKM`

### `resetFilters()` — clears all 3, resets page to 1.

## UI

- 3-col grid on `lg:`, 2-col on `md:`, 1-col mobile
- "Reset Filter" shows when search or category is active
- Pagination shown when `totalPages > 1`
- Results count: `Menampilkan X-Y dari Z UMKM (Total: N)`

## API

- `getPublicUMKMList()` → array of `{ id, name, category, description, contact, phone, address, logo }`

## Notes

- `logo` field is in the data shape (legacy mock) but not rendered.
- `getUMKMCategories()` exists in service but unused — categories are derived locally instead.
- Phone numbers link to `tel:business.phone` — works on mobile.
