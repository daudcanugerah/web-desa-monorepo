<template>
  <div class="h-[calc(100dvh-116px)] flex overflow-hidden relative">
    <div class="flex-1 relative">
      <div id="map-container" class="absolute inset-0"></div>
      
      <LoadingSpinner v-if="loading" class="absolute top-4 left-1/2 -translate-x-1/2 z-[1000]" />

      <div v-if="mapError && !loading" class="absolute inset-0 flex flex-col items-center justify-center gap-3 bg-gray-100 z-[1] px-6 text-center">
        <p class="text-gray-600">Gagal memuat peta.</p>
        <button @click="reload" class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700 transition-colors">
          Coba lagi
        </button>
      </div>

      <div v-else-if="!mapReady && !loading" class="absolute inset-0 flex items-center justify-center bg-gray-100 z-[1]">
        <p class="text-gray-500">Memuat peta...</p>
      </div>

      <div class="absolute top-4 left-1/2 -translate-x-1/2 z-[500] flex items-center gap-2 md:hidden">
        <button
          @click="() => { showLegend = !showLegend; if (showLegend) showLayers = false }"
          class="bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors"
          title="Legenda"
        >
          <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>

        <button
          @click="() => { showLayers = !showLayers; if (showLayers) showLegend = false }"
          class="bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors"
          title="Lapisan Peta"
        >
          <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6l8-4 8 4-8 4-8-4zm0 6l8 4 8-4M4 18l8 4 8-4" />
          </svg>
        </button>

        <button
          @click="showSidebar = !showSidebar"
          class="bg-emerald-600 text-white rounded-lg shadow-md p-2.5 hover:bg-emerald-700 transition-colors"
          title="Daftar Lokasi"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      <Transition name="fade">
        <div v-if="showLayers && isMobile" class="absolute top-20 left-1/2 -translate-x-1/2 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-2 w-44">
          <div class="flex items-center justify-between mb-1.5 px-1">
            <h3 class="text-xs font-semibold text-gray-900">Lapisan Peta</h3>
            <button @click="showLayers = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <button
            v-for="layer in baseLayersList"
            :key="layer.key"
            type="button"
            class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-left transition-colors"
            :class="baseLayerKey === layer.key ? 'bg-emerald-50 text-emerald-700' : 'text-gray-600 hover:bg-gray-50'"
            @click="setBaseLayer(layer.key)"
          >
            <span
              class="w-3.5 h-3.5 rounded-full border flex items-center justify-center flex-shrink-0"
              :class="baseLayerKey === layer.key ? 'border-emerald-600 bg-emerald-600' : 'border-gray-300'"
            >
              <svg v-if="baseLayerKey === layer.key" class="w-2.5 h-2.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
              </svg>
            </span>
            <span class="text-xs font-medium">{{ layer.label }}</span>
          </button>
        </div>
      </Transition>

      <button
        @click="() => { showLegend = !showLegend; if (showLegend) showLayers = false }"
        class="absolute top-4 right-4 z-[500] bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors hidden md:block"
        title="Legenda"
      >
        <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      <button
        @click="() => { showLayers = !showLayers; if (showLayers) showLegend = false }"
        class="absolute top-4 right-16 z-[500] bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors hidden md:flex items-center gap-2"
        title="Ganti Lapisan Peta"
      >
        <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6l8-4 8 4-8 4-8-4zm0 6l8 4 8-4M4 18l8 4 8-4" />
        </svg>
        <span class="text-xs font-medium text-gray-700">{{ baseLayerLabel }}</span>
      </button>

      <Transition name="fade">
        <div v-if="showLegend && !isMobile" class="absolute top-16 right-4 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-3 w-40">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-xs font-semibold text-gray-900">Legenda</h3>
            <button @click="showLegend = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="space-y-1.5">
            <div v-for="cat in legendCategories" :key="cat.name" class="flex items-center gap-2">
              <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ backgroundColor: cat.color }"></div>
              <span class="text-[11px] text-gray-600">{{ cat.name }}</span>
            </div>
          </div>
        </div>
      </Transition>

      <Transition name="fade">
        <div v-if="showLayers && !isMobile" class="absolute top-16 right-16 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-2 w-44">
          <div class="flex items-center justify-between mb-1.5 px-1">
            <h3 class="text-xs font-semibold text-gray-900">Lapisan Peta</h3>
            <button @click="showLayers = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <button
            v-for="layer in baseLayersList"
            :key="layer.key"
            type="button"
            class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-left transition-colors"
            :class="baseLayerKey === layer.key ? 'bg-emerald-50 text-emerald-700' : 'text-gray-600 hover:bg-gray-50'"
            @click="setBaseLayer(layer.key)"
          >
            <span
              class="w-3.5 h-3.5 rounded-full border flex items-center justify-center flex-shrink-0"
              :class="baseLayerKey === layer.key ? 'border-emerald-600 bg-emerald-600' : 'border-gray-300'"
            >
              <svg v-if="baseLayerKey === layer.key" class="w-2.5 h-2.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
              </svg>
            </span>
            <span class="text-xs font-medium">{{ layer.label }}</span>
          </button>
        </div>
      </Transition>

      <Transition name="fade">
        <div v-if="showLegend && isMobile" class="absolute top-20 left-1/2 -translate-x-1/2 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-3 w-40">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-xs font-semibold text-gray-900">Legenda</h3>
            <button @click="showLegend = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="space-y-1.5">
            <div v-for="cat in legendCategories" :key="cat.name" class="flex items-center gap-2">
              <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ backgroundColor: cat.color }"></div>
              <span class="text-[11px] text-gray-600">{{ cat.name }}</span>
            </div>
          </div>
        </div>
      </Transition>

    </div>

    <!-- Desktop side panel: list view <-> place detail view (Google-Maps style) -->
    <div class="hidden md:flex w-80 bg-white border-l border-gray-200 flex-col overflow-hidden z-[50]">
      <PlaceDetailPanel
        v-if="selectedLocation"
        :location="selectedLocation"
        :has-coords="hasCoords(selectedLocation)"
        @back="clearSelection"
      />

      <template v-else>
        <div class="px-4 py-3 border-b border-gray-200">
          <h2 class="text-sm font-semibold text-gray-900 mb-3">Daftar Lokasi</h2>
          <div class="space-y-2">
            <div class="relative">
              <input
                v-model="searchText"
                type="text"
                placeholder="Cari lokasi..."
                class="w-full pl-9 pr-3 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
              />
              <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select
              v-model="selectedCategory"
              class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            >
              <option value="Semua">Semua Kategori</option>
              <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
            </select>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto">
          <div v-if="filteredLocations.length === 0" class="p-6 text-center text-xs text-gray-500">
            Tidak ada lokasi
          </div>
          <div v-else class="divide-y divide-gray-100">
            <div
              v-for="location in filteredLocations"
              :key="location.id"
              :data-location-id="location.id"
              @click="focusOnMarker(location.id)"
              @mouseenter="hoveredLocationId = location.id"
              @mouseleave="hoveredLocationId = null"
              class="px-3 py-2 cursor-pointer transition-colors border-b border-gray-50 last:border-b-0"
              :class="hoveredLocationId === location.id ? 'bg-emerald-50' : 'hover:bg-emerald-50'"
            >
              <div class="flex items-center gap-2.5">
                <div
                  class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 overflow-hidden"
                  :style="{ backgroundColor: getCategoryColor(location.category) + '20' }"
                >
                  <img
                    v-if="locationImage(location)"
                    :src="locationImage(location)"
                    :alt="location.name"
                    class="w-full h-full object-cover"
                    loading="lazy"
                  />
                  <div v-else class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: getCategoryColor(location.category) }"></div>
                </div>
                <div class="flex-1 min-w-0">
                  <h4 class="text-xs font-medium text-gray-900 truncate">{{ location.name }}</h4>
                  <p class="text-[11px] text-gray-500 mt-0.5 truncate">{{ location.category }}</p>
                </div>
                <svg class="w-4 h-4 text-gray-300 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </div>
          </div>
        </div>

        <div class="px-3 py-2 border-t border-gray-100 bg-gray-50">
          <p class="text-[11px] text-gray-500 text-center">{{ filteredLocations.length }} lokasi</p>
          <p v-if="unmappedCount > 0" class="text-[11px] text-amber-600 text-center mt-0.5">{{ unmappedCount }} lokasi belum memiliki koordinat</p>
        </div>
      </template>
    </div>

    <Transition name="slide-up">
      <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-gray-900">Daftar Lokasi</h2>
          <button @click="showSidebar = false" class="p-1.5 hover:bg-gray-100 rounded-full">
            <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-4 py-3 border-b border-gray-200">
          <div class="space-y-2">
            <div class="relative">
              <input
                v-model="searchText"
                type="text"
                placeholder="Cari lokasi..."
                class="w-full pl-9 pr-3 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
              />
              <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select
              v-model="selectedCategory"
              class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            >
              <option value="Semua">Semua Kategori</option>
              <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
            </select>
          </div>
        </div>
        <div class="flex-1 overflow-y-auto">
          <div v-if="filteredLocations.length === 0" class="p-6 text-center text-xs text-gray-500">
            Tidak ada lokasi
          </div>
          <div v-else class="divide-y divide-gray-100">
            <div
              v-for="location in filteredLocations"
              :key="location.id"
              :data-location-id="location.id"
              @click="focusOnMarker(location.id)"
              class="px-3 py-2 cursor-pointer transition-colors border-b border-gray-50 last:border-b-0"
              :class="selectedLocationId === location.id ? 'bg-emerald-50' : 'hover:bg-emerald-50'"
            >
              <div class="flex items-center gap-2.5">
                <div
                  class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 overflow-hidden"
                  :style="{ backgroundColor: getCategoryColor(location.category) + '20' }"
                >
                  <img
                    v-if="locationImage(location)"
                    :src="locationImage(location)"
                    :alt="location.name"
                    class="w-full h-full object-cover"
                    loading="lazy"
                  />
                  <div v-else class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: getCategoryColor(location.category) }"></div>
                </div>
                <div class="flex-1 min-w-0">
                  <h4 class="text-xs font-medium text-gray-900 truncate">{{ location.name }}</h4>
                  <p class="text-[11px] text-gray-500 mt-0.5 truncate">{{ location.category }}</p>
                </div>
                <svg
                  class="w-4 h-4 text-gray-300 flex-shrink-0 transition-transform"
                  :class="{ 'rotate-90 text-emerald-500': selectedLocationId === location.id }"
                  fill="none" stroke="currentColor" viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>

              <div v-if="selectedLocationId === location.id" class="mt-2 pt-2 border-t border-emerald-100 space-y-1.5">
                <img
                  v-if="locationImage(location)"
                  :src="locationImage(location)"
                  :alt="location.name"
                  class="w-full h-28 object-cover rounded-lg"
                  loading="lazy"
                />
                <p v-if="location.description" class="text-[11px] text-gray-600 leading-relaxed">{{ location.description }}</p>
                <p v-if="hasCoords(location)" class="text-[11px] text-gray-400">
                  {{ Number(location.latitude).toFixed(5) }}, {{ Number(location.longitude).toFixed(5) }}
                </p>
                <p v-else class="text-[11px] text-amber-600">Koordinat belum tersedia</p>
              </div>
            </div>
          </div>
        </div>
        <div class="px-3 py-2 border-t border-gray-100 bg-gray-50">
          <p class="text-[11px] text-gray-500 text-center">{{ filteredLocations.length }} lokasi</p>
          <p v-if="unmappedCount > 0" class="text-[11px] text-amber-600 text-center mt-0.5">{{ unmappedCount }} lokasi belum memiliki koordinat</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { getPublicFasilitasAll, getPublicFasilitasCategories, resolveGalleryAssetUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import PlaceDetailPanel from '../components/common/PlaceDetailPanel.vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-minimap'
import 'leaflet.markercluster'
import 'leaflet.markercluster/dist/MarkerCluster.css'
import 'leaflet.markercluster/dist/MarkerCluster.Default.css'

delete L.Icon.Default.prototype._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
})

const POLYGON_COLOR = '#10B981'
const FILL_COLOR = '#10B981'

const BASE_LAYER_STORAGE_KEY = 'peta_base_layer'

// Keyless public tile providers. `key` is persisted to localStorage.
// `maxNativeZoom` is the deepest zoom the provider actually has imagery for;
// Leaflet upscales the last real tile past it instead of requesting tiles the
// provider doesn't have (which would render "map data not yet available").
const BASE_LAYERS = [
  {
    key: 'osm',
    label: 'Jalan',
    url: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    attribution: '© OpenStreetMap contributors',
    maxZoom: 19,
    maxNativeZoom: 19
  },
  {
    key: 'satellite',
    label: 'Satelit',
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    attribution: '© Esri, Maxar, Earthstar Geographics',
    maxZoom: 19,
    maxNativeZoom: 18
  },
  {
    key: 'terrain',
    label: 'Terrain',
    url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
    attribution: '© OpenStreetMap © OpenTopoMap',
    maxZoom: 17,
    maxNativeZoom: 17
  }
]

export default {
  name: 'Peta',
  components: {
    LoadingSpinner,
    PlaceDetailPanel
  },
  setup() {
    const locations = ref([])
    const loading = ref(true)
    const mapReady = ref(false)
    const mapError = ref(false)
    const showLegend = ref(true)
    const showLayers = ref(false)
    const showSidebar = ref(false)
    const baseLayerKey = ref(
      BASE_LAYERS.some(l => l.key === localStorage.getItem(BASE_LAYER_STORAGE_KEY))
        ? localStorage.getItem(BASE_LAYER_STORAGE_KEY)
        : 'osm'
    )
    const baseLayers = {}
    const selectedCategory = ref('Semua')
    const searchText = ref('')
    const isMobile = ref(false)
    const polygonCoords = ref([])
    const selectedLocationId = ref(null)
    const hoveredLocationId = ref(null)

    let map = null
    let polygon = null
    let markerCluster = null
    let minimap = null
    let locationMarkers = {}
    let invalidateTimer = null

    // Category colors come from a fixed palette, assigned by the category's
    // position in the API-ordered list so a category always maps to the same
    // color. `categoryColorMap` is populated after the categories load.
    const CATEGORY_PALETTE = [
      '#10B981', '#3B82F6', '#EF4444', '#8B5CF6',
      '#F59E0B', '#06B6D4', '#EC4899', '#84CC16',
      '#F97316', '#14B8A6'
    ]
    const FALLBACK_COLOR = '#6B7280'
    const apiCategories = ref([])
    const categoryColorMap = computed(() => {
      const map = {}
      apiCategories.value.forEach((cat, i) => {
        map[cat.name] = CATEGORY_PALETTE[i % CATEGORY_PALETTE.length]
      })
      return map
    })

    const getCategoryColor = (category) =>
      categoryColorMap.value[category] || FALLBACK_COLOR

    const getCategoryFromName = (name) => {
      const lower = name.toLowerCase()
      if (lower.includes('kantor') || lower.includes('desa') || lower.includes('pemerintah')) return 'Pemerintahan'
      if (lower.includes('sekolah') || lower.includes('sd') || lower.includes('smp') || lower.includes('smk')) return 'Pendidikan'
      if (lower.includes('puskesmas') || lower.includes('klinik') || lower.includes('rumah sakit')) return 'Kesehatan'
      if (lower.includes('masjid') || lower.includes('muslim') || lower.includes('gereja') || lower.includes(' pura')) return 'Ibadah'
      if (lower.includes('pasar') || lower.includes('toko') || lower.includes('warung')) return 'Pasar'
      return 'Umum'
    }

    // Legend uses the API categories; falls back to categories present in the
    // loaded facilities if the categories request failed.
    const legendCategories = computed(() => {
      const names = apiCategories.value.length
        ? apiCategories.value.map(c => c.name)
        : [...new Set(locations.value.map(l => l.category).filter(Boolean))]
      return names.map(name => ({ name, color: getCategoryColor(name) }))
    })

    // Dropdown options: API categories first, then any extra names seen in data.
    const categories = computed(() => {
      const names = apiCategories.value.map(c => c.name)
      const set = new Set(names)
      locations.value.forEach(l => { if (l.category) set.add(l.category) })
      return [...set]
    })

    const hasCoords = (location) => {
      const lat = Number(location?.latitude)
      const lng = Number(location?.longitude)
      return Number.isFinite(lat) && Number.isFinite(lng) && !(lat === 0 && lng === 0)
    }

    const filteredLocations = computed(() => {
      let filtered = locations.value

      if (selectedCategory.value !== 'Semua') {
        filtered = filtered.filter(l => l.category === selectedCategory.value)
      }

      if (searchText.value) {
        const query = searchText.value.toLowerCase()
        filtered = filtered.filter(l =>
          l.name?.toLowerCase().includes(query) ||
          l.description?.toLowerCase().includes(query) ||
          l.category?.toLowerCase().includes(query)
        )
      }

      return filtered
    })

    const unmappedCount = computed(() => filteredLocations.value.filter(l => !hasCoords(l)).length)

    const baseLayersList = BASE_LAYERS
    const baseLayerLabel = computed(
      () => BASE_LAYERS.find(l => l.key === baseLayerKey.value)?.label || 'Jalan'
    )

    // The selected place opens a full detail view in the panel. Unlike the
    // map card, it works even for facilities without coordinates.
    const selectedLocation = computed(() => {
      if (!selectedLocationId.value) return null
      return locations.value.find(l => l.id === selectedLocationId.value) || null
    })

    // Google-Maps-style teardrop pin. The active pin is recolored to a fixed
    // accent (ignoring its category color), enlarged, and given a pulsing halo
    // so the chosen location is unmistakable. The anchor stays at the tip so
    // the pin always points at the exact coordinate.
    const PIN_W = 30
    const PIN_H = 42
    const ACTIVE_ACCENT = '#059669'

    const createCustomIcon = (category, { active = false, hovered = false } = {}) => {
      const color = active ? ACTIVE_ACCENT : getCategoryColor(category)
      const scale = active ? 1.3 : hovered ? 1.1 : 1
      const shadow = active
        ? 'drop-shadow(0 4px 6px rgba(0,0,0,0.5))'
        : hovered
          ? 'drop-shadow(0 3px 4px rgba(0,0,0,0.35))'
          : 'drop-shadow(0 2px 3px rgba(0,0,0,0.3))'
      const w = Math.round(PIN_W * scale)
      const h = Math.round(PIN_H * scale)
      const halo = active
        ? `<span class="marker-halo" style="position:absolute;left:50%;top:0;transform:translate(-50%,-30%);width:${w + 22}px;height:${w + 22}px;border-radius:50%;background:${ACTIVE_ACCENT};opacity:0.25;"></span>`
        : ''
      return L.divIcon({
        html: `<div style="position:relative;width:${w}px;height:${h}px;">
          ${halo}
          <svg width="${w}" height="${h}" viewBox="0 0 30 42" xmlns="http://www.w3.org/2000/svg" style="filter:${shadow};display:block;position:relative;">
            <path d="M15 1C7.268 1 1 7.268 1 15c0 9.5 11.5 22.5 13.4 25.6.3.5 1.05.5 1.35 0C17.5 37.5 29 24.5 29 15 29 7.268 22.732 1 15 1z"
              ${active ? 'stroke="#ffffff" stroke-width="2.5"' : 'stroke="#ffffff" stroke-width="2"'} fill="${color}"/>
            <circle cx="15" cy="15" r="5.5" fill="#ffffff"/>
            <circle cx="15" cy="15" r="2.6" fill="${color}"/>
          </svg>
        </div>`,
        className: 'custom-marker-icon',
        iconSize: [w, h],
        iconAnchor: [w / 2, h],
      })
    }

    const loadPolygon = async () => {
      try {
        const res = await fetch('/area-desa-poly.json')
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        if (Array.isArray(data) && data.length > 0 && Array.isArray(data[0])) {
          polygonCoords.value = data
          return true
        }
      } catch (err) {
        console.error('Error loading desa polygon:', err)
      }
      return false
    }

    const initMap = async () => {
      const mapContainer = document.getElementById('map-container')
      if (!mapContainer) return

      try {
        map = L.map(mapContainer, {
          zoomControl: true,
          attributionControl: true
        })

        // Build every base layer once; only the selected one is added.
        BASE_LAYERS.forEach((def) => {
          baseLayers[def.key] = L.tileLayer(def.url, {
            attribution: def.attribution,
            maxZoom: def.maxZoom,
            maxNativeZoom: def.maxNativeZoom ?? def.maxZoom
          })
        })
        const initialLayer = baseLayers[baseLayerKey.value] || baseLayers.osm
        initialLayer.addTo(map)
        map.setMaxZoom(baseLayerMaxZoom(baseLayerKey.value))

        let latLngCoords = []
        if (polygonCoords.value.length > 0) {
          latLngCoords = polygonCoords.value.map(coord => [coord[1], coord[0]])
          polygon = L.polygon(latLngCoords, {
            color: POLYGON_COLOR,
            fillColor: FILL_COLOR,
            fillOpacity: 0.1,
            weight: 2
          }).addTo(map)

          map.fitBounds(polygon.getBounds(), { padding: [30, 30], maxZoom: 16 })
        } else {
          // Fallback view centered on the village if the polygon is unavailable.
          map.setView([-6.7128, 107.6925], 14)
        }

        const minimapTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '© OpenStreetMap'
        })

        const minimapLayers = L.layerGroup([minimapTileLayer])
        if (latLngCoords.length > 0) {
          minimapLayers.addLayer(L.polygon(latLngCoords, {
            color: POLYGON_COLOR,
            fillColor: FILL_COLOR,
            fillOpacity: 0.15,
            weight: 2
          }))
        }

        minimap = new L.Control.MiniMap(minimapLayers, {
          toggleDisplay: true,
          minimized: false,
          position: 'bottomleft',
          width: 150,
          height: 150
        }).addTo(map)

        markerCluster = L.markerClusterGroup({
          showCoverageOnHover: false,
          maxClusterRadius: 50,
          iconCreateFunction: function(cluster) {
            const count = cluster.getChildCount()
            return L.divIcon({
              html: `<div style="
                background: ${POLYGON_COLOR};
                border-radius: 50%;
                width: 36px;
                height: 36px;
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-weight: bold;
                font-size: 12px;
                border: 3px solid white;
                box-shadow: 0 2px 6px rgba(0,0,0,0.3);
              ">${count}</div>`,
              className: 'custom-cluster-icon',
              iconSize: L.point(36, 36)
            })
          }
        })
        map.addLayer(markerCluster)

        mapError.value = false
        mapReady.value = true

        invalidateTimer = setTimeout(() => {
          if (map) map.invalidateSize()
        }, 300)
      } catch (err) {
        console.error('Error initializing map:', err)
        mapError.value = true
      }
    }

    const baseLayerMaxZoom = (key) =>
      BASE_LAYERS.find(l => l.key === key)?.maxZoom || 19

    const setBaseLayer = (key) => {
      if (!map || !baseLayers[key] || key === baseLayerKey.value) {
        showLayers.value = false
        return
      }
      if (baseLayers[baseLayerKey.value]) map.removeLayer(baseLayers[baseLayerKey.value])
      baseLayers[key].addTo(map)
      baseLayerKey.value = key
      map.setMaxZoom(baseLayerMaxZoom(key))
      localStorage.setItem(BASE_LAYER_STORAGE_KEY, key)
      showLayers.value = false
    }

    const locationImage = (location) => {
      const media = location.media
      if (!Array.isArray(media) || media.length === 0) return ''
      const first = media[0]
      if (!first) return ''
      const raw = typeof first === 'string' ? first : (first.url || first.thumbnail_url || '')
      return resolveGalleryAssetUrl(raw)
    }

    const updateMarkers = () => {
      if (!markerCluster) return

      markerCluster.clearLayers()
      locationMarkers = {}

      filteredLocations.value.forEach((location) => {
        if (!hasCoords(location)) return

        const marker = L.marker([Number(location.latitude), Number(location.longitude)], {
          icon: createCustomIcon(location.category, { active: selectedLocationId.value === location.id })
        })

        // focusOnMarker selects (opening the bottom sheet on mobile) and pans
        // the map to center the chosen pin.
        marker.on('click', () => focusOnMarker(location.id))
        marker.on('mouseover', () => { hoveredLocationId.value = location.id })
        marker.on('mouseout', () => { hoveredLocationId.value = null })

        markerCluster.addLayer(marker)
        locationMarkers[location.id] = marker
      })
    }

    // Swap only the affected markers' icons between default/hover/active, and
    // raise the active pin above its neighbors so it isn't hidden.
    const refreshMarkerIcon = (id, { active = false, hovered = false }) => {
      const loc = findLocation(id)
      const marker = locationMarkers[id]
      if (!loc || !marker) return
      marker.setIcon(createCustomIcon(loc.category, { active, hovered }))
      marker.setZIndexOffset(active ? 1000 : 0)
    }

    const findLocation = (id) => locations.value.find(l => l.id === id)

    // Scroll the sidebar/mobile list entry for id into view and highlight it.
    const scrollListTo = (id) => {
      nextTick(() => {
        const el = document.querySelector(`[data-location-id="${CSS.escape(String(id))}"]`)
        if (el && typeof el.scrollIntoView === 'function') {
          el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        }
      })
    }

    const clearSelection = () => {
      const prev = selectedLocationId.value
      selectedLocationId.value = null
      if (prev != null) refreshMarkerIcon(prev, { hovered: hoveredLocationId.value === prev })
    }

    const selectLocation = (id) => {
      const prev = selectedLocationId.value
      selectedLocationId.value = id
      if (prev != null && prev !== id) refreshMarkerIcon(prev, {})
      refreshMarkerIcon(id, { active: true })
      // The panel swaps to the detail view; on mobile open the bottom sheet.
      if (isMobile.value) {
        showSidebar.value = true
        scrollListTo(id)
      }
    }

    // Center the map on a location, using markercluster's zoomToShowLayer when
    // available so a clustered marker reliably declusters and becomes visible.
    const panToLocation = (location) => {
      if (!map || !hasCoords(location)) return
      const marker = locationMarkers[location.id]
      if (marker && markerCluster && typeof markerCluster.zoomToShowLayer === 'function') {
        markerCluster.zoomToShowLayer(marker, () => {
          map.setView(marker.getLatLng(), Math.max(map.getZoom(), 18), { animate: true })
        })
        return
      }
      const target = [Number(location.latitude), Number(location.longitude)]
      map.setView(target, Math.max(map.getZoom(), 18), { animate: true })
    }

    const focusOnMarker = (id) => {
      const location = findLocation(id)
      if (!location) return
      selectLocation(id)
      panToLocation(location)
    }

    const checkMobile = () => {
      isMobile.value = window.innerWidth < 768
    }

    const loadData = async () => {
      loading.value = true
      mapError.value = false
      try {
        const [polygonOk, result, cats] = await Promise.all([
          loadPolygon(),
          getPublicFasilitasAll(),
          getPublicFasilitasCategories()
        ])

        apiCategories.value = Array.isArray(cats) ? cats.filter(c => c && c.name) : []

        const fasilitas = Array.isArray(result?.items) ? result.items : []

        locations.value = fasilitas.map((item, index) => {
          const category = item.category || getCategoryFromName(item.name || '')
          return {
            id: item.id || index + 1,
            name: item.name || 'Tanpa nama',
            description: item.description || '',
            category,
            latitude: item.latitude,
            longitude: item.longitude,
            media: Array.isArray(item.media) ? item.media : []
          }
        })

        if (!polygonOk && polygonCoords.value.length === 0) {
          // Polygon is optional; the map still works centered on the village.
          console.warn('Desa polygon unavailable, using fallback view')
        }

        loading.value = false
        await nextTick()
        await new Promise(r => setTimeout(r, 100))

        await initMap()
        if (markerCluster) {
          updateMarkers()
        }
      } catch (error) {
        console.error('Error fetching peta data:', error)
        loading.value = false
        mapError.value = true
      }
    }

    const resetMapRefs = () => {
      map = null
      polygon = null
      markerCluster = null
      minimap = null
      locationMarkers = {}
      Object.keys(baseLayers).forEach((k) => delete baseLayers[k])
    }

    const reload = () => {
      if (map) map.remove()
      resetMapRefs()
      mapReady.value = false
      loadData()
    }

    onMounted(() => {
      checkMobile()
      window.addEventListener('resize', checkMobile)
      loadData()
    })

    watch(filteredLocations, (list) => {
      // Drop the selection if it was filtered out (its marker no longer exists).
      if (selectedLocationId.value && !list.some(l => l.id === selectedLocationId.value)) {
        selectedLocationId.value = null
      }
      if (markerCluster) {
        updateMarkers()
      }
    })

    watch(hoveredLocationId, (id, prev) => {
      if (prev != null && prev !== selectedLocationId.value) refreshMarkerIcon(prev, {})
      if (id != null && id !== selectedLocationId.value) refreshMarkerIcon(id, { hovered: true })
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', checkMobile)
      if (invalidateTimer) clearTimeout(invalidateTimer)
      if (map) map.remove()
      resetMapRefs()
    })

    return {
      locations,
      loading,
      mapReady,
      mapError,
      showLegend,
      showLayers,
      showSidebar,
      selectedCategory,
      searchText,
      isMobile,
      categories,
      legendCategories,
      baseLayersList,
      baseLayerKey,
      baseLayerLabel,
      setBaseLayer,
      filteredLocations,
      unmappedCount,
      selectedLocationId,
      hoveredLocationId,
      selectedLocation,
      locationImage,
      hasCoords,
      reload,
      getCategoryColor,
      focusOnMarker,
      scrollListTo,
      clearSelection
    }
  }
}
</script>

<style>
.custom-marker-icon,
.custom-cluster-icon {
  background: transparent;
}
.marker-halo {
  animation: marker-pulse 1.6s ease-out infinite;
  pointer-events: none;
}
@keyframes marker-pulse {
  0% {
    transform: translate(-50%, -30%) scale(0.7);
    opacity: 0.35;
  }
  100% {
    transform: translate(-50%, -30%) scale(1.6);
    opacity: 0;
  }
}
#map-container .leaflet-popup-content-wrapper {
  border-radius: 12px;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
#map-container .leaflet-top,
#map-container .leaflet-bottom,
#map-container .leaflet-control {
  z-index: 400 !important;
}
#map-container .leaflet-minimap {
  border: 3px solid #e5e7eb !important;
  border-radius: 8px !important;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15) !important;
  bottom: 10px !important;
  left: 10px !important;
}
#map-container .leaflet-minimap-container {
  z-index: 10 !important;
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease-in-out;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}
@media (max-width: 767px) {
  #map-container .leaflet-control-zoom {
    bottom: 80px !important;
    right: 10px !important;
    z-index: 500 !important;
  }
  #map-container .leaflet-minimap {
    bottom: 140px !important;
    left: 10px !important;
    z-index: 500 !important;
  }
  #map-container .leaflet-top,
  #map-container .leaflet-bottom,
  #map-container .leaflet-control {
    z-index: 400 !important;
  }
}
</style>
