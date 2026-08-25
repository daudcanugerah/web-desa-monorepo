# Components

## Layout (`components/layout/`)

| File | Role |
|---|---|
| `Layout.vue` | Vertical stack: `<Navbar />` + `<main><RouterView/></main>` + `<Footer />` |
| `Navbar.vue` | Sticky top bar. Brand "Desa Sukamaju" (initials `DS`), 7 links, active-route highlight (`text-emerald-600 border-b-2 border-emerald-600`), mobile hamburger toggle. Active state: exact match for `/`, `startsWith` for others. |
| `Footer.vue` | 3-col grid (about / contact / quick links). Dark `bg-gray-900`. Static content, not data-driven. |

## Common (`components/common/`)

| File | Props | Notes |
|---|---|---|
| `LoadingSpinner.vue` | — | Spinning SVG circle. `min-h-[200px]` flex-center. |
| `EmptyState.vue` | `message: String` (default `'Data tidak tersedia'`) | Icon + label. |
| `Card.vue` | `className: String` | White rounded panel. **Unused** — kept for future. |

## Shared Patterns Across Pages

### Sidebar (desktop) + bottom-sheet (mobile)

Profil, Infografik, Peta all duplicate this layout. Identical structure:

```html
<div class="h-[calc(100vh-80px)] flex overflow-hidden relative">
  <div class="flex-1 overflow-y-auto">…content…</div>
  <div class="hidden md:flex w-80 bg-white border-l …">…sidebar…</div>
  <button @click="showSidebar = !showSidebar" class="md:hidden …">☰</button>
  <Transition name="slide-up">
    <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] … max-h-[70vh] …">…bottom-sheet…</div>
  </Transition>
</div>
```

**DRY candidate** — extract to `<CategorySidebar>` + `<MobileSheet>`.

### State for mobile detection

```js
const isMobile = ref(false)
const checkMobile = () => { isMobile.value = window.innerWidth < 768 }
onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile) })
onBeforeUnmount(() => window.removeEventListener('resize', checkMobile))
```

Profil omits the `onBeforeUnmount` cleanup — minor leak risk on route thrash.

### Pagination control

Identical 3-button block (`prev` + page numbers + `next`) repeated in Berita, UMKM, PPID, Profil. DRY candidate.
