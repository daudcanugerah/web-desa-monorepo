# Home (Beranda)

`pages/Home.vue` → route `/`. Landing page. Fully API-driven — no FAKE_* data.

## Sections (top → bottom)

1. **Pengumuman ticker** — latest articles in the berita **"Pengumuman"**
   category (`getPublicBeritaCategoryIdByName('Pengumuman')` → `getPublicBeritaList({ category })`).
   Each links to `/berita/:id`. Empty state when none.
2. **Hero Banner Slider** — active banners (`getActiveBanners()`), 5s auto-rotate,
   manual prev/next + dots. Falls back to a single title slide when no banners.
3. **Motto Ribbon** — `desa.motto`; hidden when empty.
4. **Sambutan Kepala Desa** — `desa.kepala_desa`, `desa.kepala_desa_message`,
   and `desa.kepala_desa_media` (photo, signed URL; falls back to an icon when
   absent). Whole section hidden when name and message are both empty.
5. **Statistik** — 6 cards from `desa.jumlah_*`; hidden when all empty.
6. **Berita & Informasi** — latest 4 from `getPublicBeritaList()`.
7. **APBDesa** — link-only card to `/infografik` (no hardcoded figures).
8. **Mini Peta** — schematic SVG + `desa.jumlah_dusun`, `wilayah`, live
   fasilitas count (`getPublicFasilitasAll`); links to `/peta`.
9. **Jelajahi** — static nav cards.
10. **Contact CTA** — real `desa.phone` / `desa.email`; PPID link.

## Data sources

| Field | Source |
|---|---|
| name, phone, email, address | `getPublicDesa()` (via `useDesaInfo`) |
| kepala_desa, kepala_desa_message, kepala_desa_media, motto | `getPublicDesa()` |
| kecamatan, kabupaten, provinsi | `getPublicDesa()` |
| jumlah_penduduk, jumlah_kk, jumlah_dusun, jumlah_rt, jumlah_rw, jumlah_umkm | `getPublicDesa()` |
| banners | `getActiveBanners()` |
| pengumuman | `getPublicBeritaList({ category })` |
| latest news | `getPublicBeritaList()` |
| fasilitas count | `getPublicFasilitasAll()` |

## Notes

- Every section degrades to hidden/empty state rather than showing placeholder
  data. There are no `FAKE_*` constants anymore.
- `desa` extended fields are edited in admin **Info Umum** (`/profile`).