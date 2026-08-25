# Infografik

`pages/Infografik.vue` → route `/infografik`. Statistics dashboard with 4 categories.

## Categories

| ID | Name | Visual |
|---|---|---|
| `Demografi` | Demografi | Gender bars (blue/pink) + age group cards |
| `Ekonomi` | Ekonomi | Sector bars (emerald) + monthly income cards (Rp X.XM) |
| `Pendidikan` | Pendidikan | Education level cards + literacy bars (green/red) |
| `Infrastruktur` | Infrastruktur | Facility value/unit cards |

## Data Source — **MOCK ONLY**

Page calls `getPublicInfographicList()` in `onMounted` but **discards the result**. Renders from local `MOCK_DATA` constant via `currentData` computed.

```js
const currentData = computed(() => {
  const keyMap = { Demografi: 'demographics', Ekonomi: 'economy', Pendidikan: 'education', Infrastruktur: 'infrastructure' }
  return keyMap[selectedCategory.value] ? MOCK_DATA[keyMap[selectedCategory.value]] : null
})
```

`MOCK_DATA` is module-scope (`pages/Infografik.vue:260-318`) with the same shape as the legacy `getInfografikData()` mock in `desaService.js`.

## State

| Ref | Default |
|---|---|
| `loading` | `true` |
| `selectedCategory` | `''` |
| `showSidebar` | `false` |
| `isMobile` | `false` |
| `searchText` | `''` |

No pagination. No `currentPage`.

## Computed

- `filteredCategories` — text search on `name` + `description`
- `currentData` — maps selected category → mock dataset

## Methods

- `selectCategory(id)`
- `checkMobile()` + `resize` listener (no `onBeforeUnmount` cleanup)

## Layout

Same sidebar pattern as Profil. Mobile bottom-sheet for category nav.

## Bug / TODO

The API endpoint exists (`/public/infographic/list`) and the page already calls it, but the response is thrown away. Wire `currentData` to fetch result when backend populates the endpoint, or remove the dead fetch.
