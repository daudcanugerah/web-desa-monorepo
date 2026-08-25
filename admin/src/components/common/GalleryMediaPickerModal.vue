<template>
  <AppModal :show="show" title="Pilih Media dari Galeri" size="lg" @close="emit('close')">
    <div class="space-y-4">
      <!-- Filters -->
      <div class="flex flex-wrap gap-2">
        <AppSearchInput
          v-model="searchQuery"
          placeholder="Cari nama file..."
          class="flex-1 min-w-[200px]"
          @input="onSearchInput"
        />
        <div class="w-40">
          <AppSelect v-model="mediaTypeFilter" @change="fetchMedia(1)">
            <option value="">Semua Tipe</option>
            <option value="image">Gambar</option>
            <option value="video">Video</option>
          </AppSelect>
        </div>
      </div>

      <!-- Grid -->
      <div v-if="loading && media.length === 0" class="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-5 gap-2">
        <div v-for="i in 10" :key="`sk-${i}`" class="aspect-square bg-secondary-200 dark:bg-secondary-700 rounded animate-pulse" />
      </div>

      <div v-else-if="media.length === 0" class="text-center py-10 text-sm text-secondary-500 dark:text-secondary-400">
        Tidak ada media ditemukan.
      </div>

      <div v-else class="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-5 gap-2 max-h-96 overflow-y-auto">
        <button
          v-for="m in media"
          :key="m.id"
          type="button"
          class="relative aspect-square bg-secondary-100 dark:bg-secondary-700 rounded-lg overflow-hidden border-2 transition-colors hover:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
          :class="selectedId === m.id ? 'border-primary-500 ring-2 ring-primary-500' : 'border-transparent'"
          :title="m.original_filename"
          @click="selectMedia(m)"
        >
          <SafeImg
            v-if="m.thumbnail_url && m.media_type === 'image'"
            :src="m.thumbnail_url"
            :alt="m.original_filename"
            class="w-full h-full object-cover"
          />
          <video
            v-else-if="m.media_type === 'video' && m.thumbnail_url"
            :src="m.thumbnail_url"
            class="w-full h-full object-cover"
            muted
            preload="metadata"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-secondary-400">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <div class="absolute inset-x-0 bottom-0 px-1.5 py-1 bg-gradient-to-t from-black/70 to-transparent">
            <p class="text-[10px] text-white truncate">{{ m.original_filename }}</p>
          </div>
          <div v-if="selectedId === m.id" class="absolute top-1 right-1 w-5 h-5 rounded-full bg-primary-500 text-white flex items-center justify-center">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>

      <AppPagination
        v-if="pagination.total_pages > 1"
        :page="pagination.page"
        :total-pages="pagination.total_pages"
        class="mt-2"
        @change="loadPage"
      />
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <AppButton variant="secondary" @click="emit('close')">Batal</AppButton>
        <AppButton variant="primary" :disabled="!selectedMedia" @click="confirmSelection">
          Pilih
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup>
import { ref, watch } from 'vue'
import AppModal from './AppModal.vue'
import AppSearchInput from './AppSearchInput.vue'
import AppSelect from './AppSelect.vue'
import AppButton from './AppButton.vue'
import AppPagination from './AppPagination.vue'
import SafeImg from './SafeImg.vue'
import { galleryService } from '../../services/gallery.service'
import { useNotificationStore } from '../../stores/notification'

const props = defineProps({
  show: { type: Boolean, default: false },
})
const emit = defineEmits(['close', 'select'])

const notificationStore = useNotificationStore()

const media = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const searchQuery = ref('')
const mediaTypeFilter = ref('')
const selectedId = ref(null)
const selectedMedia = ref(null)

let searchTimer = null

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchMedia(1), 300)
}

function loadPage(page) {
  fetchMedia(page)
}

function selectMedia(m) {
  selectedId.value = m.id
  selectedMedia.value = m
}

function confirmSelection() {
  if (!selectedMedia.value) return
  emit('select', selectedMedia.value)
}

async function fetchMedia(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 24 }
    if (searchQuery.value.trim()) params.q = searchQuery.value.trim()
    if (mediaTypeFilter.value) params.media_type = mediaTypeFilter.value
    const res = await galleryService.listMedia(params)
    media.value = res.data.media
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat media galeri')
  } finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      selectedId.value = null
      selectedMedia.value = null
      searchQuery.value = ''
      mediaTypeFilter.value = ''
      fetchMedia(1)
    }
  },
)
</script>