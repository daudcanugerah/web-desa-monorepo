# Peta

`pages/Peta.vue` → route `/peta`. Leaflet map of village polygon + facility markers.

## Map Setup

- **Lib:** `leaflet` 1.9.4 + `leaflet-minimap` + `leaflet.markercluster`
- **Tiles:** OpenStreetMap (`https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png`)
- **Default icon fix:** `L.Icon.Default.mergeOptions({ iconRetinaUrl, iconUrl, shadowUrl })` overriding with unpkg URLs (Vite strips Leaflet's default asset paths).
- **Polygon:** Hardcoded at `pages/Peta.vue:235` (line truncated by preview tools but contains ~140 `[lng, lat]` pairs around lat -6.71, lng 107.68). Color `#10B981` (emerald-600), `fillOpacity: 0.1`.
- **Minimap:** bottom-left, 150×150, toggleable.
- **Marker cluster:** `maxClusterRadius: 50`, custom green cluster icon with child count.

## Init Order (`onMounted`)

1. `checkMobile()` + resize listener
2. `getPublicFasilitasList()` → transform items
3. `await nextTick()` + 100ms wait
4. `initMap()` — map, tiles, polygon, minimap, cluster group
5. `updateMarkers()` — render markers from `filteredLocations`

If a facility lacks `latitude` / `longitude`, the page **injects a random point** inside the polygon bounds (`center.lat + (random-0.5)*0.005`, etc.) — this is a fallback, not a real bug, but it means data without coords will look like a cluster around the village center.

## State

| Ref | Default |
|---|---|
| `locations` | `[]` |
| `loading` | `true` |
| `mapReady` | `false` |
| `showLegend` | `true` |
| `showSidebar` | `false` |
| `selectedCategory` | `'Semua'` |
| `searchText` | `''` |
| `isMobile` | `false` |

Non-ref handles (module-scope `let`): `map`, `polygon`, `markerCluster`, `locationMarkers`.

## Legend / Categories (hardcoded)

6-color map: Pemerintahan / Pendidikan / Kesehatan / Ibadah / Ekonomi / Umum. `getCategoryFromName(name)` infers category from name substrings when API doesn't supply `type`.

## Computed

- `categories` — `Set` of unique `location.category` values from data
- `filteredLocations` — category filter + search (name / description / category / location)

## Methods

- `initMap()` — async map setup
- `updateMarkers()` — clears cluster, rebuilds from `filteredLocations`, stores `locationMarkers[id] = marker` for lookups
- `focusOnMarker(id)` — `setView(latlng, 18)` + open popup
- `getCategoryColor(category)` — color map
- `getCategoryFromName(name)` — heuristic category assignment
- `createCustomIcon(category)` — `L.divIcon` with colored circle

## API

- `getPublicFasilitasList()` → facilities (each `{ id, name, description, type, location, latitude, longitude, images }`)

`watch(filteredLocations, updateMarkers)` — marker cluster updates on filter change.

## Cleanup (`onBeforeUnmount`)

- Removes resize listener
- `map.remove()` + nulls reference

## Known Issues

- Polygon coords hardcoded in component; `public/area-desa-poly.json` (GeoJSON) ships but unused. Single source of truth needed.
- `getPublicFasilitasList` has `console.log` debug statements in the service — clean up.
- `<style>` (non-scoped) overrides Leaflet z-index. Risky if other pages also use Leaflet.
- No `<style scoped>` — global selectors like `.leaflet-popup-content-wrapper` leak.
