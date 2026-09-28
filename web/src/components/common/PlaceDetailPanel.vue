<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="relative flex-shrink-0">
      <div class="h-36 bg-gray-100">
        <img
          v-if="images.length"
          :src="images[0]"
          :alt="location.name"
          class="w-full h-full object-cover"
          loading="lazy"
        />
        <div v-else class="w-full h-full flex items-center justify-center text-gray-300">
          <svg class="w-9 h-9" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
      </div>

      <button
        type="button"
        class="absolute top-2 left-2 w-8 h-8 flex items-center justify-center rounded-full bg-white/95 text-gray-700 shadow-md hover:bg-white transition-colors"
        aria-label="Kembali ke daftar"
        @click="$emit('back')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <span
        v-if="images.length > 1"
        class="absolute bottom-2 right-2 px-2 py-0.5 text-[11px] font-medium text-white bg-black/50 rounded-full"
      >
        {{ images.length }} foto
      </span>
    </div>

    <!-- Tabs -->
    <div class="flex flex-shrink-0 border-b border-gray-200">
      <button
        type="button"
        class="flex-1 py-2.5 text-xs font-medium transition-colors border-b-2 -mb-px"
        :class="tab === 'description' ? 'text-emerald-600 border-emerald-600' : 'text-gray-500 border-transparent hover:text-gray-700'"
        @click="tab = 'description'"
      >
        Deskripsi
      </button>
      <button
        type="button"
        class="flex-1 py-2.5 text-xs font-medium transition-colors border-b-2 -mb-px"
        :class="tab === 'photos' ? 'text-emerald-600 border-emerald-600' : 'text-gray-500 border-transparent hover:text-gray-700'"
        @click="tab = 'photos'"
      >
        Foto <span v-if="images.length" class="text-[11px]">({{ images.length }})</span>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto">
      <!-- Description tab -->
      <div v-show="tab === 'description'">
        <div class="p-4">
          <h2 class="text-base font-bold text-gray-900 leading-snug">{{ location.name }}</h2>
          <span
            v-if="location.category"
            class="inline-block mt-1.5 px-2 py-0.5 text-[11px] font-medium rounded-full"
            :style="{ backgroundColor: categoryColor + '20', color: categoryColor }"
          >
            {{ location.category }}
          </span>

          <p v-if="location.description" class="mt-3 text-xs text-gray-600 leading-relaxed whitespace-pre-line">
            {{ location.description }}
          </p>
          <p v-else class="mt-3 text-xs text-gray-400 italic">Tidak ada deskripsi.</p>
        </div>

        <div class="px-4 pb-4 border-t border-gray-100 pt-3">
          <div class="flex items-start gap-2.5">
            <svg class="w-4 h-4 text-gray-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <div class="text-xs">
              <p class="text-gray-700">Koordinat</p>
              <p v-if="hasCoords" class="text-[11px] text-gray-500 mt-0.5">
                {{ Number(location.latitude).toFixed(6) }}, {{ Number(location.longitude).toFixed(6) }}
              </p>
              <p v-else class="text-[11px] text-amber-600 mt-0.5">Belum tersedia</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Photos tab -->
      <div v-show="tab === 'photos'" class="p-4">
        <div v-if="images.length" class="grid grid-cols-2 gap-2">
          <button
            v-for="(src, idx) in images"
            :key="idx"
            type="button"
            class="relative aspect-square rounded-lg overflow-hidden group focus:outline-none focus:ring-2 focus:ring-emerald-500"
            @click="openLightbox(idx)"
          >
            <img :src="src" :alt="`${location.name} ${idx + 1}`" class="w-full h-full object-cover group-hover:scale-105 transition-transform" loading="lazy" />
          </button>
        </div>
        <p v-else class="text-xs text-gray-400 italic">Belum ada foto.</p>
      </div>
    </div>

    <div class="p-3 border-t border-gray-100 bg-gray-50 flex gap-2">
      <button
        type="button"
        class="flex-1 px-3 py-2 text-xs font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700 transition-colors"
        @click="$emit('back')"
      >
        Kembali ke daftar
      </button>
      <a
        v-if="hasCoords"
        :href="directionsUrl"
        target="_blank"
        rel="noopener"
        class="px-3 py-2 text-xs font-medium text-emerald-700 bg-emerald-50 rounded-lg hover:bg-emerald-100 transition-colors"
      >
        Rute
      </a>
    </div>

    <!-- Lightbox -->
    <Teleport to="body">
      <div
        v-if="lightboxIndex !== null"
        class="fixed inset-0 z-[1000] bg-black/85 flex items-center justify-center p-4"
        @click.self="closeLightbox"
      >
        <button
          type="button"
          class="absolute top-4 right-4 w-9 h-9 flex items-center justify-center rounded-full bg-white/15 text-white hover:bg-white/25 transition-colors"
          aria-label="Tutup"
          @click="closeLightbox"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <img :src="images[lightboxIndex]" :alt="location.name" class="max-w-full max-h-[85vh] object-contain rounded-lg" />
        <template v-if="images.length > 1">
          <button
            type="button"
            class="absolute left-4 w-10 h-10 flex items-center justify-center rounded-full bg-white/15 text-white hover:bg-white/25 transition-colors"
            aria-label="Sebelumnya"
            @click.stop="prevLightbox"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <button
            type="button"
            class="absolute right-4 w-10 h-10 flex items-center justify-center rounded-full bg-white/15 text-white hover:bg-white/25 transition-colors"
            aria-label="Berikutnya"
            @click.stop="nextLightbox"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <span class="absolute bottom-4 left-1/2 -translate-x-1/2 px-2.5 py-0.5 text-xs text-white bg-black/50 rounded-full">
            {{ lightboxIndex + 1 }} / {{ images.length }}
          </span>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { resolveGalleryAssetUrl } from '../../services/desaService'

export default {
  name: 'PlaceDetailPanel',
  props: {
    location: { type: Object, required: true },
    hasCoords: { type: Boolean, default: false },
  },
  emits: ['back'],
  setup(props) {
    const tab = ref('description')
    const lightboxIndex = ref(null)

    const images = computed(() => {
      const media = props.location?.media
      if (!Array.isArray(media)) return []
      return media
        .map((m) => {
          const raw = typeof m === 'string' ? m : (m?.url || m?.thumbnail_url || '')
          return resolveGalleryAssetUrl(raw)
        })
        .filter(Boolean)
    })

    watch(() => props.location?.id, () => {
      tab.value = 'description'
      lightboxIndex.value = null
    })

    const clamp = (i, len) => ((i % len) + len) % len
    const openLightbox = (i) => { lightboxIndex.value = i }
    const closeLightbox = () => { lightboxIndex.value = null }
    const prevLightbox = () => { lightboxIndex.value = clamp(lightboxIndex.value - 1, images.value.length) }
    const nextLightbox = () => { lightboxIndex.value = clamp(lightboxIndex.value + 1, images.value.length) }

    return {
      tab,
      images,
      lightboxIndex,
      openLightbox,
      closeLightbox,
      prevLightbox,
      nextLightbox,
      categoryColor: computed(() => {
        const colors = {
          Pemerintahan: '#10B981',
          Pendidikan: '#3B82F6',
          Kesehatan: '#EF4444',
          Ibadah: '#8B5CF6',
          Olahraga: '#2390AE',
          Pasar: '#F59E0B',
          Ekonomi: '#F59E0B',
          Umum: '#06B6D4',
        }
        return colors[props.location?.category] || '#6B7280'
      }),
      directionsUrl: computed(() => {
        const { latitude, longitude } = props.location
        return `https://www.google.com/maps/dir/?api=1&destination=${latitude},${longitude}`
      }),
    }
  },
}
</script>
