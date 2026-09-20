<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Semua Media</h1>
      <span v-if="selectedIds.size > 0" class="text-sm text-secondary-500 dark:text-secondary-400">
        {{ selectedIds.size }} dipilih
      </span>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <AppSearchInput
        v-model="searchQuery"
        placeholder="Cari nama file..."
        class="flex-1 min-w-[200px] max-w-md"
        @input="onSearchInput"
      />
      <div class="w-44">
        <AppSelect v-model="folderFilter" @change="fetchMedia(1)">
          <option value="">Semua Folder</option>
          <option v-for="f in folders" :key="f.id" :value="f.id">
            {{ f.name }}
          </option>
        </AppSelect>
      </div>
      <div class="w-40">
        <AppSelect v-model="mediaTypeFilter" @change="fetchMedia(1)">
          <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </AppSelect>
      </div>
      <div class="inline-flex rounded-lg border border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800 p-1" role="tablist" aria-label="Filter visibilitas media">
        <button
          v-for="opt in visibilityOptions"
          :key="opt.value"
          type="button"
          role="tab"
          :aria-selected="visibilityFilter === opt.value"
          class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
          :class="visibilityFilter === opt.value
            ? 'bg-primary-600 text-white'
            : 'text-secondary-600 dark:text-secondary-300 hover:bg-secondary-100 dark:hover:bg-secondary-700'"
          @click="setVisibilityFilter(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <div v-if="loading && media.length === 0" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
      <div v-for="i in 10" :key="`sk-${i}`" class="aspect-square bg-secondary-200 dark:bg-secondary-700 rounded-lg animate-pulse" />
    </div>

    <div v-else-if="media.length === 0" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-10">
      <AppEmptyState
        icon="📷"
        title="Tidak ada media ditemukan"
        hint="Coba ubah filter atau cari nama file lain."
      />
    </div>

    <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
      <div
        v-for="m in media"
        :key="m.id"
        class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden group relative"
        :class="{ 'ring-2 ring-primary-500': selectedIds.has(m.id) }"
      >
        <label class="absolute top-2 left-2 z-10 bg-white/90 dark:bg-secondary-900/90 rounded p-1 cursor-pointer">
          <input
            type="checkbox"
            class="w-4 h-4 rounded border-secondary-300 dark:border-secondary-600 text-primary-600 focus:ring-primary-500"
            :checked="selectedIds.has(m.id)"
            :aria-label="`Pilih ${m.original_filename}`"
            @change="toggleSelect(m.id)"
          />
        </label>

        <button
          type="button"
          class="aspect-square w-full bg-secondary-100 dark:bg-secondary-700 relative overflow-hidden block focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500"
          :aria-label="`Lihat ${m.original_filename}`"
          @click="openLightbox(m)"
        >
          <SafeImg
            v-if="m.thumbnail_url && m.media_type === 'image'"
            :src="m.thumbnail_url"
            :alt="m.original_filename"
            class="w-full h-full object-cover"
          />
          <video
            v-else-if="m.media_type === 'video' && m.thumbnail_url"
            :src="resolveMediaUrl(m.thumbnail_url)"
            class="w-full h-full object-cover"
            muted
            preload="metadata"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-secondary-400 dark:text-secondary-500">
            <svg class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>

          <span
            class="absolute top-2 right-2 px-1.5 py-0.5 text-[10px] font-medium rounded uppercase"
            :class="m.is_public
              ? 'bg-green-500/90 text-white'
              : 'bg-secondary-700/90 text-white'"
          >
            {{ m.is_public ? 'Publik' : 'Privat' }}
          </span>
        </button>

        <div class="p-2">
          <p class="text-xs font-medium text-secondary-800 dark:text-secondary-200 truncate" :title="m.original_filename">
            {{ m.original_filename }}
          </p>
          <p class="text-[10px] text-secondary-500 dark:text-secondary-400">
            {{ m.media_type }} · {{ formatBytes(m.file_size) }}
          </p>
        </div>

        <div class="border-t border-secondary-200 dark:border-secondary-700 p-1.5 flex items-center justify-end gap-1">
          <button
            type="button"
            class="text-xs px-2 py-1 rounded hover:bg-secondary-100 dark:hover:bg-secondary-700 text-secondary-700 dark:text-secondary-300"
            :title="m.is_public ? 'Jadikan privat' : 'Jadikan publik'"
            @click="toggleVisibility(m)"
          >
            {{ m.is_public ? '🔓' : '🔒' }}
          </button>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded hover:bg-secondary-100 dark:hover:bg-secondary-700 text-secondary-700 dark:text-secondary-300"
            title="Lihat di folder"
            @click="goToFolder(m)"
          >
            📁
          </button>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-red-600"
            title="Hapus media"
            @click="handleDelete(m)"
          >
            🗑️
          </button>
        </div>
      </div>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-6"
      @change="loadPage"
    />

    <!-- Lightbox preview -->
    <GalleryLightbox
      :show="showLightbox"
      :list="media"
      :start-index="lightboxIndex"
      :show-delete="true"
      @update:show="showLightbox = $event"
      @toggle-visibility="toggleVisibility"
      @delete="handleLightboxDelete"
    />

    <!-- Floating bulk action bar -->
    <AppBulkBar :selected="selectedIds" @update:selected="selectedIds = $event">
      <template #default="{ selected }">
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-green-500/20 text-green-300 hover:bg-green-500/30 transition-colors"
          @click="applyBulkVisibility(true)"
        >
          Publik
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-white/10 text-white/80 hover:bg-white/20 transition-colors"
          @click="applyBulkVisibility(false)"
        >
          Privat
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-red-500/20 text-red-300 hover:bg-red-500/30 transition-colors"
          @click="handleBulkDelete(selected)"
        >
          Hapus
        </button>
      </template>
    </AppBulkBar>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { resolveMediaUrl } from '../../../utils/imageUrl'
import AppSearchInput from '../../../components/common/AppSearchInput.vue'
import AppSelect from '../../../components/common/AppSelect.vue'
import AppPagination from '../../../components/common/AppPagination.vue'
import AppEmptyState from '../../../components/common/AppEmptyState.vue'
import AppBulkBar from '../../../components/common/AppBulkBar.vue'
import GalleryLightbox from '../../../components/common/GalleryLightbox.vue'
import SafeImg from '../../../components/common/SafeImg.vue'
import { galleryService } from '../../../services/gallery.service'
import { useNotificationStore } from '../../../stores/notification'
import { useConfirm } from '../../../composables/useConfirm'

const router = useRouter()
const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const typeOptions = [
  { label: 'Semua Tipe', value: '' },
  { label: 'Gambar', value: 'image' },
  { label: 'Video', value: 'video' },
]
const visibilityOptions = [
  { label: 'Semua', value: '' },
  { label: 'Publik', value: 'true' },
  { label: 'Privat', value: 'false' },
]

const media = ref([])
const folders = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const searchQuery = ref('')
const mediaTypeFilter = ref('')
const visibilityFilter = ref('')
const folderFilter = ref('')
const selectedIds = ref(new Set())
const showLightbox = ref(false)
const lightboxIndex = ref(0)

let searchTimer = null

function formatBytes(bytes) {
  if (!bytes) return '—'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`
}

function toggleSelect(id) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchMedia(1), 300)
}

function setVisibilityFilter(val) {
  visibilityFilter.value = val
  fetchMedia(1)
}

function loadPage(page) {
  fetchMedia(page)
}

function openLightbox(m) {
  lightboxIndex.value = media.value.findIndex((x) => x.id === m.id)
  if (lightboxIndex.value === -1) lightboxIndex.value = 0
  showLightbox.value = true
}

async function fetchMedia(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 24 }
    if (searchQuery.value.trim()) params.q = searchQuery.value.trim()
    if (mediaTypeFilter.value) params.media_type = mediaTypeFilter.value
    if (visibilityFilter.value !== '') params.is_public = visibilityFilter.value
    if (folderFilter.value) params.folder_id = folderFilter.value
    const res = await galleryService.listMedia(params)
    media.value = res.data.media
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat media')
  } finally {
    loading.value = false
  }
}

async function fetchFolders() {
  try {
    const res = await galleryService.listFolders({ page: 1, limit: 100 })
    folders.value = res.data.folders || []
  } catch {
    folders.value = []
  }
}

async function toggleVisibility(m) {
  try {
    const res = await galleryService.setMediaVisibility(m.id, !m.is_public)
    m.is_public = res.data.is_public
    notificationStore.success(m.is_public ? 'Media dipublikasikan' : 'Media dijadikan privat')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas')
  }
}

function goToFolder(m) {
  if (m.folder_id) {
    router.push(`/gallery/folders/${m.folder_id}`)
  }
}

async function handleDelete(m) {
  const confirmed = await confirmDelete('Media', m.original_filename)
  if (!confirmed) return
  try {
    await galleryService.deleteMedia(m.id)
    notificationStore.success('Media dihapus')
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus media')
  }
}

async function handleLightboxDelete(m) {
  const wasLast = media.value.length === 1
  await handleDelete(m)
  if (wasLast) showLightbox.value = false
}

async function applyBulkVisibility(isPublic) {
  const ids = Array.from(selectedIds.value)
  if (ids.length === 0) return
  try {
    // Group by folder: bulk endpoint is scoped to a folder, so update per folder.
    const byFolder = {}
    for (const id of ids) {
      const m = media.value.find((x) => x.id === id)
      const fid = m?.folder_id || 'orphan'
      ;(byFolder[fid] ||= []).push(id)
    }
    for (const [fid, mids] of Object.entries(byFolder)) {
      if (fid === 'orphan') {
        for (const id of mids) await galleryService.setMediaVisibility(id, isPublic)
      } else {
        await galleryService.setBulkMediaVisibility(fid, mids, isPublic)
      }
    }
    notificationStore.success(`Visibilitas ${ids.length} media diperbarui`)
    selectedIds.value = new Set()
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas')
  }
}

async function handleBulkDelete(selected) {
  const confirmed = await confirmDelete(`${selected.length} Media`, '')
  if (!confirmed) return
  try {
    for (const id of selected) {
      await galleryService.deleteMedia(id)
    }
    notificationStore.success(`${selected.length} media dihapus`)
    selectedIds.value = new Set()
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus media')
  }
}

onMounted(() => {
  fetchMedia()
  fetchFolders()
})
</script>