<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div v-if="!flags.galleryPublic" class="bg-white rounded-xl shadow-sm border border-gray-200 p-8 text-center max-w-xl mx-auto">
        <Icon name="grid" class="w-12 h-12 mx-auto mb-4 text-emerald-400" />
        <h2 class="text-lg font-bold text-gray-900 mb-2">Galeri sedang dalam pemeliharaan</h2>
        <p class="text-sm text-gray-600 mb-4">Fitur galeri publik belum tersedia pada backend.</p>
        <RouterLink to="/" class="text-sm text-emerald-600 hover:text-emerald-700 font-medium">Kembali ke Beranda</RouterLink>
      </div>

      <template v-else>
        <nav class="mb-6 text-sm">
          <RouterLink to="/galeri" class="text-emerald-600 hover:text-emerald-700 font-medium inline-flex items-center gap-1.5">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            Kembali ke Galeri
          </RouterLink>
        </nav>

      <LoadingSpinner v-if="loading" />

      <EmptyState
        v-else-if="!folder"
        message="Album tidak ditemukan"
        description="Album ini tidak tersedia, privat, atau belum memiliki media publik."
        action-label="Kembali ke Galeri"
        @action="goBack"
      />

      <template v-else>
        <PageHeader :title="folder.name" :subtitle="folder.description || `Album publik ${desaName}`" />

        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6 flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-4 text-xs text-gray-600">
            <span class="inline-flex items-center gap-1.5">
              <svg class="w-4 h-4 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              {{ imageCount }} foto
            </span>
            <span class="inline-flex items-center gap-1.5">
              <svg class="w-4 h-4 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              {{ videoCount }} video
            </span>
            <span class="text-gray-400">·</span>
            <span>{{ formatDate(folder.created_at) }}</span>
          </div>
          <div class="flex gap-2">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              @click="activeType = opt.value"
              :class="[
                'px-3 py-1.5 text-xs font-medium rounded-lg transition-colors',
                activeType === opt.value
                  ? 'bg-emerald-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              ]"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <EmptyState
          v-if="filteredMedia.length === 0"
          message="Belum ada media"
          description="Album ini belum memiliki media publik pada filter yang dipilih."
        />

        <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
          <button
            v-for="(item, idx) in filteredMedia"
            :key="item.id"
            @click="openPreview(idx)"
            class="group relative aspect-square bg-gradient-to-br from-emerald-100 to-teal-100 rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all border border-gray-200 focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <img
              v-if="mediaThumbUrl(item)"
              :src="mediaThumbUrl(item)"
              :alt="item.original_filename || 'Media'"
              loading="lazy"
              class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
              @error="handleImageError($event, idx)"
            />
            <div v-else class="absolute inset-0 flex items-center justify-center text-emerald-300">
              <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>

            <div v-if="isVideo(item)" class="absolute inset-0 flex items-center justify-center pointer-events-none">
              <div class="w-12 h-12 rounded-full bg-black/50 backdrop-blur-sm flex items-center justify-center">
                <svg class="w-6 h-6 text-white ml-1" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M8 5v14l11-7z" />
                </svg>
              </div>
            </div>

            <div class="absolute inset-x-0 bottom-0 p-2 bg-gradient-to-t from-black/70 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
              <p class="text-xs text-white truncate">
                {{ item.original_filename || (isVideo(item) ? 'Video' : 'Foto') }}
              </p>
            </div>
          </button>
        </div>
      </template>

    <Teleport to="body">
      <div
        v-if="preview"
        @click="closePreview"
        class="fixed inset-0 z-[100] bg-black/90 backdrop-blur-sm flex items-center justify-center p-4"
      >
        <button
          @click.stop="closePreview"
          class="absolute top-4 right-4 w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition-colors"
          aria-label="Tutup"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        <button
          v-if="filteredMedia.length > 1"
          @click.stop="prev"
          class="absolute left-4 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition-colors"
          aria-label="Sebelumnya"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <button
          v-if="filteredMedia.length > 1"
          @click.stop="next"
          class="absolute right-4 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition-colors"
          aria-label="Berikutnya"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <div @click.stop class="max-w-5xl w-full max-h-[90vh] flex flex-col items-center">
          <img
            v-if="!isVideo(preview)"
            :src="mediaContentUrl(preview)"
            :alt="preview.original_filename || 'Media'"
            class="max-w-full max-h-[85vh] object-contain rounded-lg shadow-2xl"
          />
          <video
            v-else
            :src="mediaContentUrl(preview)"
            controls
            autoplay
            class="max-w-full max-h-[85vh] rounded-lg shadow-2xl bg-black"
          />
          <div class="mt-3 text-center">
            <p class="text-sm text-white font-medium truncate">
              {{ preview.original_filename || (isVideo(preview) ? 'Video' : 'Foto') }}
            </p>
            <p class="text-xs text-white/60 mt-1">
              {{ currentIndex + 1 }} / {{ filteredMedia.length }}
            </p>
          </div>
        </div>
      </div>
    </Teleport>
      </template>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getPublicGalleryFolderById,
  resolveMediaThumbnailUrl,
  resolveMediaContentUrl,
  formatMediaDate
} from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import PageHeader from '../components/common/PageHeader.vue'
import Icon from '../components/common/Icon.vue'
import { useDesaInfo } from '../composables/useDesaInfo'
import { useFeatureFlags } from '../composables/useFeatureFlags'

export default {
  name: 'GaleriDetail',
  components: { LoadingSpinner, EmptyState, PageHeader, Icon },
  setup() {
    const route = useRoute()
    const router = useRouter()

    const folder = ref(null)
    const media = ref([])
    const loading = ref(true)
    const activeType = ref('all')
    const preview = ref(null)
    const currentIndex = ref(0)

    const { desaInfo } = useDesaInfo()
    const flags = useFeatureFlags()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const formatDate = formatMediaDate

    const typeOptions = [
      { value: 'all', label: 'Semua' },
      { value: 'image', label: 'Foto' },
      { value: 'video', label: 'Video' }
    ]

    const isVideo = (item) => {
      if (!item) return false
      const t = (item.media_type || '').toLowerCase()
      if (t === 'video') return true
      const m = (item.mime_type || '').toLowerCase()
      return m.startsWith('video/')
    }

    const imageCount = computed(() => media.value.filter(m => !isVideo(m)).length)
    const videoCount = computed(() => media.value.filter(isVideo).length)

    const filteredMedia = computed(() => {
      if (activeType.value === 'all') return media.value
      if (activeType.value === 'video') return media.value.filter(isVideo)
      return media.value.filter(m => !isVideo(m))
    })

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const fetchFolder = async () => {
      loading.value = true
      folder.value = null
      media.value = []
      try {
        const result = await getPublicGalleryFolderById(route.params.id)
        folder.value = result.folder
        media.value = result.media || []
      } catch (error) {
        console.error('Error fetching gallery folder:', error)
      } finally {
        loading.value = false
      }
    }

    const mediaThumbUrl = (m) => resolveMediaThumbnailUrl(m)
    const mediaContentUrl = (m) => resolveMediaContentUrl(m)

    const openPreview = (idx) => {
      currentIndex.value = idx
      preview.value = filteredMedia.value[idx]
    }

    const closePreview = () => {
      preview.value = null
    }

    const next = () => {
      const len = filteredMedia.value.length
      currentIndex.value = (currentIndex.value + 1) % len
      preview.value = filteredMedia.value[currentIndex.value]
    }

    const prev = () => {
      const len = filteredMedia.value.length
      currentIndex.value = (currentIndex.value - 1 + len) % len
      preview.value = filteredMedia.value[currentIndex.value]
    }

    const onKeydown = (e) => {
      if (!preview.value) return
      if (e.key === 'Escape') closePreview()
      else if (e.key === 'ArrowRight') next()
      else if (e.key === 'ArrowLeft') prev()
    }

    const goBack = () => {
      router.push('/galeri')
    }

    onMounted(() => {
      fetchFolder()
      window.addEventListener('keydown', onKeydown)
    })

    onUnmounted(() => {
      window.removeEventListener('keydown', onKeydown)
    })

    watch(() => route.params.id, () => {
      closePreview()
      fetchFolder()
    })

    return {
      folder,
      media,
      loading,
      activeType,
      typeOptions,
      preview,
      currentIndex,
      desaName,
      imageCount,
      videoCount,
      filteredMedia,
      formatDate,
      isVideo,
      handleImageError,
      openPreview,
      closePreview,
      next,
      prev,
      goBack,
      mediaThumbUrl,
      mediaContentUrl,
      flags
    }
  }
}
</script>