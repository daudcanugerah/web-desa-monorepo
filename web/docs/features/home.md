# Home

`pages/Home.vue` → route `/`. Landing page.

## Sections (top → bottom)

1. **Hero Banner Slider** — 3 hardcoded slides, 5s auto-rotate (`setInterval`), manual prev/next + dot nav. Gradient bg + picsum image overlay. Fades in via `opacity-100`/`opacity-0` per slide.
2. **Stats** — 4 stat cards (Penduduk, Luas Wilayah, Dusun, UMKM). Pulls from `getPublicDesa()`; falls back to hardcoded `{ population: 5420, area: '12.5 km²', neighborhoods: 8, businesses: 45 }` on miss/error. Colors bind via `bg-${stat.color}-100` (Tailwind purge risk).
3. **Village Head Welcome** — static. "Budi Santoso, S.Sos / Kepala Desa Sukamaju". Avatar placeholder (`BS` initials).
4. **Latest News** — 3 most recent from `getPublicBeritaList().slice(0, 3)`. Card → `RouterLink` to `/berita/:slug`.
5. **Explore Sections** — 3 static cards (Profil / Infografik / UMKM) with custom SVG icons.
6. **Contact CTA** — `tel:02112345678` link + "Layanan Publik" → PPID.

## State

| Ref | Type | Notes |
|---|---|---|
| `data` | `ref(null)` | `{ latestNews, stats }` |
| `loading` | `ref(true)` | toggled in `onMounted` `finally` |
| `currentSlide` | `ref(0)` | hero index |
| `slideTimer` | `let` (not ref) | `setInterval` handle, cleared `onUnmounted` |

## Computed

- `statItems` — maps raw stats → `{ value, label, sublabel, color, icon }`. `population` formatted via `toLocaleString('id-ID')`.

## Effects

- `nextSlide` / `previousSlide` — mod-arithmetic index update.
- `startSlideTimer` — starts `setInterval(nextSlide, 5000)` after data load.

## API

- `getPublicBeritaList()` → list, slice(0, 3)
- `getPublicDesa()` → object; `population` / `area` / `neighborhoods` / `businesses` / `statistics` keys consumed

## Notes

- Active banners API (`getActiveBanners`) imported but **not called** — hero slides are still hardcoded.
- `getActiveBanners` import is dead — safe to remove.
- `<LoadingSpinner>` rendered while loading. Section-level, not page-level.
