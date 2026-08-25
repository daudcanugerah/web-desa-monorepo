<template>
  <div class="h-[calc(100vh-80px)] flex overflow-hidden relative">
    <div class="flex-1 relative">
      <div id="map-container" class="absolute inset-0"></div>
      
      <LoadingSpinner v-if="loading" class="absolute top-4 left-1/2 -translate-x-1/2 z-[1000]" />
      
      <div v-if="!mapReady && !loading" class="absolute inset-0 flex items-center justify-center bg-gray-100 z-[1]">
        <p class="text-gray-500">Memuat peta...</p>
      </div>

      <div class="absolute top-4 left-1/2 -translate-x-1/2 z-[500] flex items-center gap-2 md:hidden">
        <button
          @click="showLegend = !showLegend"
          class="bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors"
          title="Legenda"
        >
          <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
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

      <button
        @click="showLegend = !showLegend"
        class="absolute top-4 right-4 z-[500] bg-white rounded-lg shadow-md border border-gray-200 p-2.5 hover:bg-gray-50 transition-colors hidden md:block"
        title="Legenda"
      >
        <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      <Transition name="fade">
        <div v-if="showLegend && !isMobile" class="absolute top-16 right-4 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-4 w-48">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-medium text-gray-900">Legenda</h3>
            <button @click="showLegend = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="space-y-2">
            <div v-for="cat in legendCategories" :key="cat.name" class="flex items-center gap-2">
              <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: cat.color }"></div>
              <span class="text-xs text-gray-600">{{ cat.name }}</span>
            </div>
          </div>
        </div>
      </Transition>

      <Transition name="fade">
        <div v-if="showLegend && isMobile" class="absolute top-20 left-1/2 -translate-x-1/2 z-[500] bg-white rounded-lg shadow-lg border border-gray-200 p-4 w-48">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-medium text-gray-900">Legenda</h3>
            <button @click="showLegend = false" class="text-gray-400 hover:text-gray-600">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="space-y-2">
            <div v-for="cat in legendCategories" :key="cat.name" class="flex items-center gap-2">
              <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: cat.color }"></div>
              <span class="text-xs text-gray-600">{{ cat.name }}</span>
            </div>
          </div>
        </div>
      </Transition>
    </div>

    <div class="hidden md:block w-96 bg-white border-l border-gray-200 flex flex-col overflow-hidden z-[50]">
      <div class="p-5 border-b border-gray-200">
        <h2 class="text-base font-semibold text-gray-900 mb-4">Daftar Lokasi</h2>
        <div class="space-y-3">
          <div class="relative">
            <input
              v-model="searchText"
              type="text"
              placeholder="Cari lokasi..."
              class="w-full pl-10 pr-4 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            />
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <select
            v-model="selectedCategory"
            class="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
          >
            <option value="Semua">Semua Kategori</option>
            <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
        </div>
      </div>
      
      <div class="flex-1 overflow-y-auto">
        <div v-if="filteredLocations.length === 0" class="p-8 text-center text-gray-500">
          Tidak ada lokasi
        </div>
        <div v-else class="divide-y divide-gray-100">
          <div
            v-for="location in filteredLocations"
            :key="location.id"
            @click="focusOnMarker(location.id)"
            class="p-4 hover:bg-emerald-50 cursor-pointer transition-colors border-b border-gray-50 last:border-b-0"
          >
            <div class="flex items-start gap-3">
              <div
                class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0"
                :style="{ backgroundColor: getCategoryColor(location.category) + '20' }"
              >
                <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: getCategoryColor(location.category) }"></div>
              </div>
              <div class="flex-1 min-w-0">
                <h4 class="text-sm font-medium text-gray-900">{{ location.name }}</h4>
                <p class="text-xs text-gray-500 mt-0.5">{{ location.category }}</p>
                <p v-if="location.location" class="text-xs text-emerald-600 mt-1">{{ location.location }}</p>
              </div>
              <svg class="w-5 h-5 text-gray-300 flex-shrink-0 mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <div class="p-3 border-t border-gray-100 bg-gray-50">
        <p class="text-xs text-gray-500 text-center">{{ filteredLocations.length }} lokasi</p>
      </div>
    </div>

    <Transition name="slide-up">
      <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="p-4 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-base font-semibold text-gray-900">Daftar Lokasi</h2>
          <button @click="showSidebar = false" class="p-2 hover:bg-gray-100 rounded-full">
            <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="p-4 border-b border-gray-200">
          <div class="space-y-3">
            <div class="relative">
              <input
                v-model="searchText"
                type="text"
                placeholder="Cari lokasi..."
                class="w-full pl-10 pr-4 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
              />
              <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select
              v-model="selectedCategory"
              class="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            >
              <option value="Semua">Semua Kategori</option>
              <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
            </select>
          </div>
        </div>
        <div class="flex-1 overflow-y-auto">
          <div v-if="filteredLocations.length === 0" class="p-8 text-center text-gray-500">
            Tidak ada lokasi
          </div>
          <div v-else class="divide-y divide-gray-100">
            <div
              v-for="location in filteredLocations"
              :key="location.id"
              @click="focusOnMarker(location.id); showSidebar = false"
              class="p-4 hover:bg-emerald-50 cursor-pointer transition-colors border-b border-gray-50 last:border-b-0"
            >
              <div class="flex items-start gap-3">
                <div
                  class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0"
                  :style="{ backgroundColor: getCategoryColor(location.category) + '20' }"
                >
                  <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: getCategoryColor(location.category) }"></div>
                </div>
                <div class="flex-1 min-w-0">
                  <h4 class="text-sm font-medium text-gray-900">{{ location.name }}</h4>
                  <p class="text-xs text-gray-500 mt-0.5">{{ location.category }}</p>
                  <p v-if="location.location" class="text-xs text-emerald-600 mt-1">{{ location.location }}</p>
                </div>
                <svg class="w-5 h-5 text-gray-300 flex-shrink-0 mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </div>
          </div>
        </div>
        <div class="p-3 border-t border-gray-100 bg-gray-50">
          <p class="text-xs text-gray-500 text-center">{{ filteredLocations.length }} lokasi</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { getPublicFasilitasList, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
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

const polygonCoords = [[107.6925,-6.7128],[107.693,-6.7125],[107.6931,-6.7122],[107.693,-6.7117],[107.6929,-6.7113],[107.693,-6.711],[107.6928,-6.7108],[107.6926,-6.7105],[107.6926,-6.71],[107.6928,-6.7078],[107.6932,-6.7069],[107.6936,-6.7065],[107.6941,-6.7061],[107.6942,-6.7057],[107.694,-6.7056],[107.6938,-6.7055],[107.6934,-6.7056],[107.6926,-6.7059],[107.6923,-6.706],[107.692,-6.706],[107.6918,-6.706],[107.6916,-6.706],[107.6913,-6.7061],[107.6912,-6.7063],[107.6909,-6.707],[107.6906,-6.7072],[107.6902,-6.7075],[107.6898,-6.7074],[107.6894,-6.7072],[107.689,-6.7072],[107.6894,-6.7064],[107.6898,-6.7056],[107.6899,-6.7051],[107.6898,-6.7047],[107.6896,-6.7043],[107.6893,-6.7041],[107.6891,-6.704],[107.689,-6.7039],[107.6888,-6.7038],[107.6875,-6.7046],[107.6867,-6.7053],[107.6868,-6.7055],[107.6864,-6.7063],[107.6863,-6.707],[107.6859,-6.7079],[107.6856,-6.7082],[107.6848,-6.7089],[107.6844,-6.7092],[107.6836,-6.7099],[107.6832,-6.7103],[107.6828,-6.7109],[107.6827,-6.7111],[107.6825,-6.7113],[107.6824,-6.7116],[107.6822,-6.7118],[107.6818,-6.7125],[107.6805,-6.7124],[107.6799,-6.7123],[107.6797,-6.7122],[107.6793,-6.7121],[107.6791,-6.712],[107.6788,-6.7119],[107.6784,-6.7113],[107.6783,-6.7114],[107.6775,-6.7113],[107.6774,-6.7113],[107.677,-6.7113],[107.6765,-6.7114],[107.6762,-6.7115],[107.6758,-6.7118],[107.6754,-6.712],[107.6747,-6.7121],[107.6739,-6.7119],[107.6731,-6.7118],[107.6724,-6.7117],[107.6721,-6.7115],[107.6717,-6.712],[107.6712,-6.7126],[107.6709,-6.7128],[107.6706,-6.7132],[107.6703,-6.7134],[107.6698,-6.7134],[107.6689,-6.7133],[107.6676,-6.7132],[107.6665,-6.7138],[107.6658,-6.7146],[107.6657,-6.7146],[107.6652,-6.7156],[107.665,-6.7159],[107.665,-6.7161],[107.6648,-6.7172],[107.6651,-6.7182],[107.6653,-6.7191],[107.6653,-6.7193],[107.6651,-6.72],[107.665,-6.7203],[107.664,-6.7207],[107.6626,-6.7218],[107.6619,-6.7226],[107.6616,-6.723],[107.6601,-6.7238],[107.6595,-6.7242],[107.6593,-6.7243],[107.658,-6.7255],[107.6571,-6.7266],[107.6567,-6.7274],[107.6563,-6.7279],[107.6562,-6.7281],[107.6561,-6.7284],[107.6556,-6.729],[107.6556,-6.7291],[107.6555,-6.7292],[107.6554,-6.7293],[107.6552,-6.7296],[107.6546,-6.7301],[107.6536,-6.7306],[107.6535,-6.7307],[107.6526,-6.7312],[107.6526,-6.7323],[107.653,-6.7324],[107.6533,-6.7324],[107.6534,-6.7326],[107.6537,-6.7328],[107.654,-6.7334],[107.654,-6.7335],[107.6543,-6.734],[107.6543,-6.7342],[107.6545,-6.7349],[107.6547,-6.7354],[107.655,-6.7356],[107.6556,-6.7356],[107.6559,-6.7358],[107.6563,-6.7359],[107.6572,-6.7359],[107.6574,-6.7359],[107.6579,-6.7358],[107.6581,-6.7357],[107.6584,-6.7356],[107.6589,-6.7351],[107.6595,-6.7349],[107.6602,-6.7347],[107.6604,-6.7347],[107.6606,-6.7347],[107.6619,-6.7339],[107.6622,-6.7336],[107.6633,-6.7327],[107.6636,-6.7323],[107.664,-6.7317],[107.6641,-6.7315],[107.6648,-6.7314],[107.6657,-6.7312],[107.6664,-6.7308],[107.6668,-6.7306],[107.6669,-6.7305],[107.6678,-6.7301],[107.6691,-6.7298],[107.6695,-6.7298],[107.6697,-6.7298],[107.6702,-6.7299],[107.6707,-6.7299],[107.671,-6.7299],[107.6713,-6.73],[107.6716,-6.73],[107.6718,-6.73],[107.6725,-6.7301],[107.6733,-6.73],[107.6738,-6.7297],[107.6744,-6.7293],[107.6752,-6.7288],[107.6756,-6.7284],[107.676,-6.7278],[107.6761,-6.7275],[107.676,-6.7272],[107.6763,-6.7271],[107.6775,-6.7271],[107.6782,-6.7272],[107.6791,-6.7269],[107.6798,-6.7267],[107.6798,-6.7265],[107.6797,-6.7261],[107.6799,-6.7258],[107.6804,-6.7256],[107.6811,-6.7254],[107.6822,-6.7249],[107.6835,-6.7247],[107.6841,-6.7243],[107.6842,-6.7243],[107.6844,-6.7243],[107.6851,-6.7242],[107.6854,-6.724],[107.6858,-6.7238],[107.6858,-6.7237],[107.6861,-6.7233],[107.6865,-6.7232],[107.6878,-6.7229],[107.6879,-6.7229],[107.6885,-6.7229],[107.6891,-6.7225],[107.6898,-6.7219],[107.6903,-6.7213],[107.6908,-6.7208],[107.6917,-6.7204],[107.6917,-6.7202],[107.6918,-6.7196],[107.6918,-6.7194],[107.6919,-6.7192],[107.6922,-6.7187],[107.6923,-6.7184],[107.6924,-6.718],[107.6927,-6.7174],[107.6928,-6.7171],[107.6927,-6.7167],[107.6926,-6.7164],[107.6922,-6.716],[107.6921,-6.7158],[107.6919,-6.7156],[107.6916,-6.7155],[107.6912,-6.7155],[107.6913,-6.7152],[107.6913,-6.7151],[107.6915,-6.7145],[107.6919,-6.7136],[107.6917,-6.7132],[107.6911,-6.7129],[107.6914,-6.7126],[107.6915,-6.7123],[107.6915,-6.7123],[107.6918,-6.7127],[107.692,-6.7128],[107.6925,-6.7128]]

export default {
  name: 'Peta',
  components: {
    LoadingSpinner
  },
  setup() {
    const locations = ref([])
    const loading = ref(true)
    const mapReady = ref(false)
    const showLegend = ref(true)
    const showSidebar = ref(false)
    const selectedCategory = ref('Semua')
    const searchText = ref('')
    const isMobile = ref(false)

    let map = null
    let polygon = null
    let markerCluster = null
    let locationMarkers = {}

    const legendCategories = [
      { name: 'Pemerintahan', color: '#10B981' },
      { name: 'Pendidikan', color: '#3B82F6' },
      { name: 'Kesehatan', color: '#EF4444' },
      { name: 'Ibadah', color: '#8B5CF6' },
      { name: 'Ekonomi', color: '#F59E0B' },
      { name: 'Umum', color: '#06B6D4' }
    ]

    const getCategoryColor = (category) => {
      const colors = {
        'Pemerintahan': '#10B981',
        'Pendidikan': '#3B82F6',
        'Kesehatan': '#EF4444',
        'Ibadah': '#8B5CF6',
        'Ekonomi': '#F59E0B',
        'Umum': '#06B6D4'
      }
      return colors[category] || '#6B7280'
    }

    const getCategoryFromName = (name) => {
      const lower = name.toLowerCase()
      if (lower.includes('kantor') || lower.includes('desa') || lower.includes('pemerintah')) return 'Pemerintahan'
      if (lower.includes('sekolah') || lower.includes('sd') || lower.includes('smp') || lower.includes('smk')) return 'Pendidikan'
      if (lower.includes('puskesmas') || lower.includes('klinik') || lower.includes('rumah sakit')) return 'Kesehatan'
      if (lower.includes('masjid') || lower.includes('muslim') || lower.includes('gereja') || lower.includes(' pura')) return 'Ibadah'
      if (lower.includes('pasar') || lower.includes('toko') || lower.includes('warung')) return 'Ekonomi'
      return 'Umum'
    }

    const categories = computed(() => {
      const cats = new Set(locations.value.map(l => l.category).filter(Boolean))
      return [...cats]
    })

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

    const createCustomIcon = (category) => {
      const color = getCategoryColor(category)
      return L.divIcon({
        html: `<div style="
          width: 28px;
          height: 28px;
          background: ${color};
          border: 3px solid white;
          border-radius: 50%;
          box-shadow: 0 2px 6px rgba(0,0,0,0.3);
        "></div>`,
        className: 'custom-marker-icon',
        iconSize: [28, 28],
        iconAnchor: [14, 14],
        popupAnchor: [0, -14]
      })
    }

    const initMap = async () => {
      const mapContainer = document.getElementById('map-container')
      if (!mapContainer) return

      try {
        map = L.map(mapContainer, {
          zoomControl: true,
          attributionControl: true
        })

        const mainTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '© OpenStreetMap contributors'
        }).addTo(map)

        const latLngCoords = polygonCoords.map(coord => [coord[1], coord[0]])

        polygon = L.polygon(latLngCoords, {
          color: POLYGON_COLOR,
          fillColor: FILL_COLOR,
          fillOpacity: 0.1,
          weight: 2
        }).addTo(map)

        const bounds = polygon.getBounds()
        map.fitBounds(bounds, { padding: [30, 30], maxZoom: 16 })

        const minimapTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '© OpenStreetMap'
        })
        
        const minimapPolygon = L.polygon(latLngCoords, {
          color: POLYGON_COLOR,
          fillColor: FILL_COLOR,
          fillOpacity: 0.15,
          weight: 2
        })
        
        const minimapLayers = L.layerGroup([minimapTileLayer, minimapPolygon])
        
        const minimap = new L.Control.MiniMap(minimapLayers, {
          toggleDisplay: true,
          minimizedOnStart: false,
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

        mapReady.value = true

        setTimeout(() => {
          if (map) map.invalidateSize()
        }, 300)
      } catch (err) {
        console.error('Error initializing map:', err)
      }
    }

    const updateMarkers = () => {
      if (!markerCluster) return

      markerCluster.clearLayers()
      locationMarkers = {}

      filteredLocations.value.forEach((location) => {
        const hasCoords = typeof location.latitude === 'number' && typeof location.longitude === 'number'

        let lat, lng
        if (hasCoords) {
          lat = location.latitude
          lng = location.longitude
        } else {
          const bounds = polygon.getBounds()
          const center = bounds.getCenter()
          lat = center.lat + (Math.random() - 0.5) * 0.005
          lng = center.lng + (Math.random() - 0.5) * 0.008
        }

        const marker = L.marker([lat, lng], {
          icon: createCustomIcon(location.category)
        })

        marker.bindPopup(`
          <div style="min-width: 180px;">
            <h4 style="font-weight: 600; margin-bottom: 4px;">${location.name || 'Unknown'}</h4>
            <p style="font-size: 12px; color: #666; margin-bottom: 4px;">${location.category || 'Uncategorized'}</p>
            ${location.description ? `<p style="font-size: 11px;">${location.description}</p>` : ''}
          </div>
        `)

        markerCluster.addLayer(marker)
        locationMarkers[location.id] = marker
      })
    }

    const focusOnMarker = (id) => {
      const location = filteredLocations.value.find(l => l.id === id)
      if (!location) return

      const marker = locationMarkers[id]
      if (marker) {
        const latlng = marker.getLatLng()
        map.setView(latlng, 18, { animate: true })
        setTimeout(() => {
          if (!marker.isPopupOpen()) {
            marker.openPopup()
          }
        }, 300)
      } else if (typeof location.latitude === 'number' && typeof location.longitude === 'number') {
        map.setView([location.latitude, location.longitude], 18, { animate: true })
      }
    }

    const checkMobile = () => {
      isMobile.value = window.innerWidth < 768
    }

    onMounted(async () => {
      checkMobile()
      window.addEventListener('resize', checkMobile)

      try {
        const result = await getPublicFasilitasList()
        
        let fasilitas = []
        if (result === null || result === undefined) {
          fasilitas = []
        } else if (Array.isArray(result)) {
          fasilitas = result
        } else if (typeof result === 'object') {
          if (Array.isArray(result.fasilitas)) {
            fasilitas = result.fasilitas
          } else if (result.data?.fasilitas) {
            fasilitas = result.data.fasilitas
          }
        }

        const transformedLocations = fasilitas.map((item, index) => ({
          id: item.id || index + 1,
          name: item.name || 'Unknown',
          description: item.description || '',
          category: item.category || getCategoryFromName(item.name || ''),
          latitude: item.latitude,
          longitude: item.longitude,
          media: Array.isArray(item.media) ? item.media : []
        }))

        locations.value = transformedLocations
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
      }
    })

    watch(filteredLocations, () => {
      if (markerCluster) {
        updateMarkers()
      }
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', checkMobile)
      if (map) {
        map.remove()
        map = null
      }
    })

    return {
      locations,
      loading,
      mapReady,
      showLegend,
      showSidebar,
      selectedCategory,
      searchText,
      isMobile,
      categories,
      legendCategories,
      filteredLocations,
      getCategoryColor,
      focusOnMarker
    }
  }
}
</script>

<style>
.custom-marker-icon,
.custom-cluster-icon {
  background: transparent;
}
.leaflet-popup-content-wrapper {
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
.leaflet-top,
.leaflet-bottom {
  z-index: 400 !important;
}
.leaflet-control {
  z-index: 400 !important;
}
.leaflet-minimap {
  border: 3px solid #e5e7eb !important;
  border-radius: 8px !important;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15) !important;
}
.leaflet-minimap-container {
  z-index: 10 !important;
}
.leaflet-minimap {
  bottom: 10px !important;
  left: 10px !important;
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
  .leaflet-control-zoom {
    bottom: 80px !important;
    right: 10px !important;
    z-index: 500 !important;
  }
  .leaflet-minimap {
    bottom: 140px !important;
    left: 10px !important;
    z-index: 500 !important;
  }
  .leaflet-top,
  .leaflet-bottom,
  .leaflet-control {
    z-index: 400 !important;
  }
}
</style>
