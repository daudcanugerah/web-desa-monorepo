# Peta

`pages/Peta.vue` → route `/peta`. Leaflet map of the village polygon + facility
markers, with a Google-Maps-style side panel (list ⇄ place detail), base-layer
switcher, and API-driven categories.

## Map setup

- **Lib:** `leaflet` 1.9.4 + `leaflet-minimap` + `leaflet.markercluster`
- **Default icon fix:** `L.Icon.Default.mergeOptions({ iconRetinaUrl, iconUrl, shadowUrl })` overriding with unpkg URLs (Vite strips Leaflet's default asset paths).
- **Polygon:** loaded at runtime from `public/area-desa-poly.json` (array of `[lng, lat]` pairs) via `loadPolygon()`. Color `#10B981`, `fillOpacity: 0.1`. If missing, the map falls back to a centered view (`[-6.7128, 107.6925]`, zoom 14).
- **Minimap:** bottom-left, 150×150, toggleable, its own plain OSM layer (does not follow the base-layer choice).
- **Marker cluster:** `maxClusterRadius: 50`, custom green cluster icon with child count.

## Base layers

Defined at module scope in `BASE_LAYERS`. `key` is persisted to
`localStorage['peta_base_layer']`.

| Key | Label | Provider | maxZoom | maxNativeZoom |
|---|---|---|---|---|
| `osm` | Jalan (default) | OpenStreetMap | 19 | 19 |
| `satellite` | Satelit | Esri World Imagery | 19 | 18 |
| `terrain` | Terrain | OpenTopoMap | 17 | 17 |

`maxNativeZoom` caps the deepest zoom with real imagery; past it Leaflet
upscales the last real tile instead of requesting missing tiles (Esri returns a
"map data not yet available" placeholder at z19 over this area).

- All layers are built once in `initMap`; `setBaseLayer(key)` swaps via
  `map.removeLayer`/`addLayer` (no reload) and calls `map.setMaxZoom`.
- `map.setMaxZoom` is also applied on init from the persisted key.
- UI: desktop "Lapisan Peta" button (top-right, next to Legend) + panel; mobile
  third button in the top-center group. Opening one panel closes the other.

## Init order

1. `onMounted` → `checkMobile()` + resize listener → `loadData()`
2. `loadData()` runs in parallel: `loadPolygon()`, `getPublicFasilitasAll()`, `getPublicFasilitasCategories()`
3. Transform facilities, `await nextTick()` + 100ms wait
4. `initMap()` — map, base layer, polygon, minimap, cluster group
5. `updateMarkers()` — render pins from `filteredLocations`

Facilities **without coordinates** are skipped (not plotted) and counted in the
`unmappedCount` notice. They still appear in the list and open their detail view.

## Categories

Categories come from **the API** (`GET /public/fasilitas/categories`), not
hardcoded. `categoryColorMap` assigns each category a stable color from
`CATEGORY_PALETTE` by its API order.

- `getCategoryColor(category)` → palette color, or `FALLBACK_COLOR` (`#6B7280`) for unknown names.
- `legendCategories` (computed) → API categories; falls back to category names present in loaded facilities if the categories request failed.
- `categories` (computed) → dropdown options: API categories first, then any extra names seen in the data.
- `getCategoryFromName(name)` → legacy heuristic used only when a facility has no `category` at all.

## Markers

- **Shape:** Google-Maps-style SVG teardrop pin, colored per category, white outline + center dot. Anchor is the **tip** (`[w/2, h]`).
- **Hover:** scale 1.1× + stronger shadow (also cross-highlights the list row).
- **Active/selected:** recolored to a fixed accent (`#059669`), scale 1.3×, pulsing halo ring (`.marker-halo`), and raised `marker.setZIndexOffset(1000)`. **One location at a time** — selecting another resets the previous.
- Selecting (via pin or list) calls `focusOnMarker(id)` → `selectLocation` + `panToLocation`:
  - `panToLocation` uses `markerCluster.zoomToShowLayer()` when available so a clustered pin reliably declusters, then centers at ≥ zoom 18; otherwise `map.setView`.
  - On mobile the bottom sheet opens and the row scrolls into view.

## Side panel (desktop)

`PlaceDetailPanel.vue` — the panel **swaps between list and detail**:

- **List view:** search (`searchText`), category filter (`selectedCategory`), rows with thumbnail + hover cross-highlight.
- **Detail view** (`selectedLocation`): hero image, **Deskripsi / Foto (N) tabs**, description, coordinates, "Rute" (Google Maps directions) link, and a back button (`clearSelection`).
- **Foto tab:** grid of all `location.media` images; clicking opens a `Teleport`ed lightbox with prev/next.
- Selection is dropped if the location gets filtered out.

## State

| Ref | Default | Notes |
|---|---|---|
| `locations` | `[]` | transformed facilities |
| `loading` | `true` | |
| `mapReady` | `false` | |
| `mapError` | `false` | shows retry UI |
| `showLegend` | `true` | |
| `showLayers` | `false` | base-layer panel |
| `showSidebar` | `false` | mobile bottom sheet |
| `baseLayerKey` | `localStorage` \| `'osm'` | persisted |
| `selectedCategory` | `'Semua'` | |
| `searchText` | `''` | |
| `isMobile` | `false` | `< 768px` |
| `polygonCoords` | `[]` | from JSON |
| `selectedLocationId` | `null` | active pin |
| `hoveredLocationId` | `null` | hover cross-highlight |
| `apiCategories` | `[]` | from API |

Non-ref handles: `map`, `polygon`, `markerCluster`, `minimap`, `locationMarkers`, `invalidateTimer`, `baseLayers`.

## Methods

- `initMap()` — async map setup (base layers, polygon, minimap, cluster)
- `loadData()` — fetch polygon + facilities + categories, then init/markers
- `reload()` / `resetMapRefs()` — teardown for re-init
- `updateMarkers()` — rebuild pins from `filteredLocations`, wire click/hover
- `createCustomIcon(category, { active, hovered })` — SVG teardrop pin
- `refreshMarkerIcon(id, { active, hovered })` — icon + z-order swap
- `focusOnMarker(id)` — select + pan
- `panToLocation(location)` — `zoomToShowLayer` / `setView`
- `selectLocation(id)` / `clearSelection()`
- `setBaseLayer(key)` — swap base layer + persist
- `locationImage(location)` — resolve first media URL (via `resolveGalleryAssetUrl`)
- `scrollListTo(id)` — scroll the active row into view

## API

- `getPublicFasilitasAll()` → `{ items, pagination }` — walks pagination so the map is not capped at the server page size.
- `getPublicFasilitasCategories()` → `[{ id, name, usage_count }]`.

`watch(filteredLocations, ...)` rebuilds markers on filter/search change and
drops a selection that was filtered out.

## Cleanup

`onBeforeUnmount`: removes resize listener, clears `invalidateTimer`,
`map.remove()`, and `resetMapRefs()` (nulls map/polygon/cluster/minimap/markers,
deletes built base layers).

## Notes

- `<style>` is **not** scoped, but all Leaflet overrides are namespaced under
  `#map-container` so they don't leak to other Leaflet maps.
- The `peta` page is the only Leaflet consumer in the app.
- Marker media comes from the gallery-media model; signed URLs
  (`/api/v1/media/{id}/...?jwt=`) are resolved against the API origin by
  `resolveGalleryAssetUrl`.
