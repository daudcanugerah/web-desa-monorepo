<template>
  <div class="flex flex-col h-full">
    <div 
      ref="mapEl" 
      class="flex-1 w-full rounded-lg border border-secondary-300 dark:border-secondary-600 min-h-[400px]"
    />
    <p class="text-xs text-secondary-500 dark:text-secondary-400 mt-1">Klik pada peta untuk menentukan lokasi ()</p>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import '../../../node_modules/leaflet-minimap/dist/Control.MiniMap.min.css'
import L_Minimap from 'leaflet-minimap'
import areaDesaPoly from '../../assets/maps/area-desa-poly.json'
import { useNotificationStore } from '../../stores/notification'

const notificationStore = useNotificationStore()

delete L.Icon.Default.prototype._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
  iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
  shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
})

const props = defineProps({
  lat: { type: [Number, String], default: -6.2 },
  lng: { type: [Number, String], default: 106.816 },
})

const emit = defineEmits(['update:lat', 'update:lng'])

const mapEl = ref(null)
let map = null
let marker = null
let polygon = null
let minimap = null

onMounted(async () => {
  await nextTick()

  try {
    const polygonCoords = areaDesaPoly.map(coord => [coord[1], coord[0]])

    if (polygonCoords.length === 0) {
      console.error('Polygon coordinates are empty')
      return
    }

    map = L.map(mapEl.value, {
      zoomControl: true,
      attributionControl: true,
    })

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
    }).addTo(map)

    polygon = L.polygon(polygonCoords, {
      color: '#3B82F6',
      fillColor: '#3B82F6',
      fillOpacity: 0.1,
      weight: 2,
    }).addTo(map)

    const bounds = polygon.getBounds()
    if (!bounds.isValid()) {
      console.error('Invalid polygon bounds')
      return
    }

    const center = bounds.getCenter()

    const initLat = parseFloat(props.lat) || center.lat
    const initLng = parseFloat(props.lng) || center.lng

    map.fitBounds(bounds, { padding: [10, 10], maxZoom: 16 })

    marker = L.marker([initLat, initLng], { draggable: true }).addTo(map)

    if (!props.lat || !props.lng) {
      emit('update:lat', parseFloat(center.lat.toFixed(7)))
      emit('update:lng', parseFloat(center.lng.toFixed(7)))
    }

    const mainTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
    })

    const minimapTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors',
    })

    const minimapPolygon = L.polygon(polygonCoords, {
      color: '#3B82F6',
      fillColor: '#3B82F6',
      fillOpacity: 0.1,
      weight: 2,
    })

    const minimapLayerGroup = L.layerGroup([minimapTileLayer, minimapPolygon])

    minimap = L.control.minimap(minimapLayerGroup, {
      toggleDisplay: true,
      minimizedOnStart: false,
      position: 'bottomleft',
      zoomLevelFixed: 12,
      centerFixed: bounds.getCenter(),
    }).addTo(map)

    marker.on('dragend', () => {
      const { lat, lng } = marker.getLatLng()
      if (isInsidePolygon(lat, lng)) {
        emit('update:lat', parseFloat(lat.toFixed(7)))
        emit('update:lng', parseFloat(lng.toFixed(7)))
      } else {
        const centerPt = polygon.getBounds().getCenter()
        marker.setLatLng(centerPt)
        emit('update:lat', parseFloat(centerPt.lat.toFixed(7)))
        emit('update:lng', parseFloat(centerPt.lng.toFixed(7)))
      }
    })

    map.on('click', (e) => {
      const { lat, lng } = e.latlng
      if (isInsidePolygon(lat, lng)) {
        marker.setLatLng([lat, lng])
        emit('update:lat', parseFloat(lat.toFixed(7)))
        emit('update:lng', parseFloat(lng.toFixed(7)))
      } else {
        notificationStore.error('Lokasi ')
      }
    })

    setTimeout(() => {
      map.invalidateSize()
    }, 100)
  } catch (err) {
    console.error('Error initializing map:', err)
  }
})

function isInsidePolygon(lat, lng) {
  if (!polygon) return true
  return polygon.getBounds().contains(L.latLng(lat, lng))
}

watch([() => props.lat, () => props.lng], ([lat, lng]) => {
  const la = parseFloat(lat)
  const lo = parseFloat(lng)
  if (!isNaN(la) && !isNaN(lo) && marker) {
    marker.setLatLng([la, lo])
    if (!map.getBounds().contains([la, lo])) {
      map.setView([la, lo], map.getZoom())
    }
    map.invalidateSize()
  }
})

onBeforeUnmount(() => {
  if (map) {
    map.remove()
    map = null
  }
})
</script>
