<template>
  <div>
    <div v-if="loading && !folder" class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <div class="h-8 w-64 bg-secondary-200 dark:bg-secondary-700 rounded animate-pulse" />
    </div>

    <template v-else-if="folder">
      <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div class="flex items-center gap-3 min-w-0">
          <AppBackButton />
          <div class="min-w-0">
            <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 truncate">{{ folder.name }}</h1>
            <p v-if="folder.description" class="text-sm text-secondary-500 dark:text-secondary-400 mt-0.5 truncate">
              {{ folder.description }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <AppButton size="sm" variant="secondary" @click="$router.push(`/gallery/folders/${folder.id}/edit`)">Edit</AppButton>
          <AppButton
            size="sm"
            :variant="folder.is_public ? 'success' : 'secondary'"
            :loading="togglingVisibility"
            @click="toggleFolderVisibility"
          >
            {{ folder.is_public ? 'Publik' : 'Privat' }}
          </AppButton>
          <AppButton size="sm" variant="danger" @click="handleDeleteFolder">Hapus Folder</AppButton>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <!-- Media uploader -->
        <div class="lg:col-span-1 bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6">
          <h3 class="text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-3">Unggah Media</h3>
          <label
            class="block border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors"
            :class="uploading
              ? 'opacity-50 cursor-not-allowed border-secondary-300 dark:border-secondary-600'
              : isUploadDragging
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : 'border-secondary-300 dark:border-secondary-600 hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/10'"
            @dragover.prevent="isUploadDragging = true"
            @dragleave.prevent="isUploadDragging = false"
            @drop.prevent="onDrop"
          >
            <svg class="w-10 h-10 mx-auto mb-2" :class="isUploadDragging ? 'text-primary-500' : 'text-secondary-400 dark:text-secondary-500'" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <span class="text-sm font-medium text-secondary-700 dark:text-secondary-300">
              {{ uploading ? 'Mengunggah...' : isUploadDragging ? 'Lepaskan file di sini' : 'Klik atau seret file' }}
            </span>
            <p class="text-xs text-secondary-500 dark:text-secondary-400 mt-1">JPEG, PNG, WebP, MP4</p>
            <input
              type="file"
              multiple
              accept="image/*,video/*"
              class="hidden"
              :disabled="uploading"
              @change="onUpload"
            />
          </label>
          <p v-if="uploadError" class="mt-2 text-xs text-red-600">{{ uploadError }}</p>
          <p v-else class="mt-2 text-xs text-secondary-500 dark:text-secondary-400">
            Beberapa file didukung. Ukuran dan tipe divalidasi per file.
          </p>

          <div class="mt-6 pt-6 border-t border-secondary-200 dark:border-secondary-700">
            <h4 class="text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-2">Statistik</h4>
            <dl class="space-y-1.5 text-xs">
              <div class="flex justify-between">
                <dt class="text-secondary-500 dark:text-secondary-400">Total media</dt>
                <dd class="font-medium text-secondary-800 dark:text-secondary-200">{{ folder.media_count ?? media.length }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-secondary-500 dark:text-secondary-400">Publik</dt>
                <dd class="font-medium text-secondary-800 dark:text-secondary-200">{{ folder.public_media_count ?? '' }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-secondary-500 dark:text-secondary-400">Dibuat</dt>
                <dd class="font-medium text-secondary-800 dark:text-secondary-200">{{ formatDate(folder.created_at) }}</dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Media grid -->
        <div class="lg:col-span-3">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-lg font-semibold text-secondary-800 dark:text-secondary-200">Media</h2>
            <span v-if="selectedIds.size > 0" class="text-sm text-secondary-500 dark:text-secondary-400">
              {{ selectedIds.size }} dipilih
            </span>
          </div>

          <div v-if="media.length === 0 && !loading" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-10">
            <AppEmptyState
              icon="📷"
              title="Belum ada media"
              hint="Unggah foto atau video untuk folder ini."
            />
          </div>

          <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-3">
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
                  :src="m.thumbnail_url"
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

                <span v-if="m.thumbnail_failed" class="absolute bottom-2 right-2 px-1.5 py-0.5 text-[10px] font-medium rounded bg-yellow-500/90 text-white">
                  Thumbnail gagal
                </span>

                <span v-if="folder.cover_media_id === m.id" class="absolute bottom-2 left-2 px-1.5 py-0.5 text-[10px] font-medium rounded bg-primary-500/90 text-white">
                  Cover
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

              <div class="border-t border-secondary-200 dark:border-secondary-700 p-1.5 flex items-center justify-between gap-1">
                <button
                  type="button"
                  class="text-xs px-2 py-1 rounded hover:bg-secondary-100 dark:hover:bg-secondary-700 text-secondary-700 dark:text-secondary-300"
                  :title="m.is_public ? 'Jadikan privat' : 'Jadikan publik'"
                  @click="toggleMediaVisibility(m)"
                >
                  {{ m.is_public ? '🔓' : '🔒' }}
                </button>
                <button
                  v-if="folder.cover_media_id !== m.id"
                  type="button"
                  class="text-xs px-2 py-1 rounded hover:bg-secondary-100 dark:hover:bg-secondary-700 text-secondary-700 dark:text-secondary-300"
                  title="Jadikan cover"
                  @click="setAsCover(m)"
                >
                  ⭐
                </button>
                <button
                  type="button"
                  class="text-xs px-2 py-1 rounded hover:bg-secondary-100 dark:hover:bg-secondary-700 text-secondary-700 dark:text-secondary-300"
                  title="Regenerate thumbnail"
                  :disabled="regeneratingId === m.id"
                  @click="regenerateThumb(m)"
                >
                  {{ regeneratingId === m.id ? '⏳' : '🔄' }}
                </button>
                <button
                  type="button"
                  class="text-xs px-2 py-1 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-red-600"
                  title="Hapus media"
                  @click="handleDeleteMedia(m)"
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
            class="mt-4"
            @change="loadPage"
          />
        </div>
      </div>

    </template>

    <!-- Lightbox preview -->
    <GalleryLightbox
      :show="showLightbox"
      :list="media"
      :start-index="lightboxIndex"
      :show-delete="true"
      @update:show="showLightbox = $event"
      @toggle-visibility="handleLightboxToggle"
      @delete="handleLightboxDelete"
    />

    <!-- Floating bulk action bar -->
    <AppBulkBar :selected="selectedIds" @update:selected="selectedIds = $event">
      <template #default="{ selected }">
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-green-500/20 text-green-300 hover:bg-green-500/30 transition-colors"
          :disabled="bulkUpdating"
          @click="applyBulkVisibility(true)"
        >
          Publik
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium bg-white/10 text-white/80 hover:bg-white/20 transition-colors"
          :disabled="bulkUpdating"
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
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppButton from '../../../components/common/AppButton.vue'
import AppBackButton from '../../../components/common/AppBackButton.vue'
import AppPagination from '../../../components/common/AppPagination.vue'
import AppEmptyState from '../../../components/common/AppEmptyState.vue'
import AppBulkBar from '../../../components/common/AppBulkBar.vue'
import GalleryLightbox from '../../../components/common/GalleryLightbox.vue'
import SafeImg from '../../../components/common/SafeImg.vue'
import { galleryService } from '../../../services/gallery.service'
import { useNotificationStore } from '../../../stores/notification'
import { useConfirm } from '../../../composables/useConfirm'
import { formatDate } from '../../../utils/dateFormat'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const folder = ref(null)
const media = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const uploading = ref(false)
const uploadError = ref('')
const isUploadDragging = ref(false)
const togglingVisibility = ref(false)
const regeneratingId = ref(null)
const selectedIds = ref(new Set())
const bulkUpdating = ref(false)
const showLightbox = ref(false)
const lightboxIndex = ref(0)

let pollHandle = null

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

async function fetchFolder() {
  loading.value = true
  try {
    const res = await galleryService.getFolder(route.params.id, {
      page: pagination.value.page,
      limit: 24,
    })
    folder.value = res.data.folder
    media.value = res.data.media
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat folder')
    router.push('/gallery')
  } finally {
    loading.value = false
  }
}

async function fetchMedia(page = 1) {
  try {
    const res = await galleryService.getFolder(route.params.id, {
      page,
      limit: 24,
    })
    media.value = res.data.media
    pagination.value = res.data.pagination
    folder.value = res.data.folder
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat media')
  }
}

function loadPage(page) {
  fetchMedia(page)
}

async function onUpload(e) {
  const files = Array.from(e.target.files || [])
  e.target.value = ''
  await uploadFiles(files)
}

async function onDrop(e) {
  isUploadDragging.value = false
  await uploadFiles(Array.from(e.dataTransfer.files || []))
}

async function uploadFiles(files) {
  if (files.length === 0) return
  uploading.value = true
  uploadError.value = ''
  try {
    const res = await galleryService.uploadMedia(route.params.id, files)
    const { uploaded = [], results = [] } = res.data || {}
    // BulkUploadResult carries an `error` field on failures — use it instead
    // of guessing status-string values.
    const failed = results.filter((r) => !!r.error)
    if (failed.length) {
      uploadError.value = failed.map((f) => `${f.filename}: ${f.error}`).join('; ')
    }
    if (uploaded.length) {
      notificationStore.success(`${uploaded.length} media berhasil diunggah`)
    }
    selectedIds.value = new Set()
    await fetchMedia(1)
  } catch (err) {
    uploadError.value = err.response?.data?.error || 'Gagal mengunggah media'
  } finally {
    uploading.value = false
  }
}

async function toggleFolderVisibility() {
  if (!folder.value) return
  togglingVisibility.value = true
  try {
    await galleryService.setFolderVisibility(folder.value.id, !folder.value.is_public)
    folder.value.is_public = !folder.value.is_public
    notificationStore.success(folder.value.is_public ? 'Folder dipublikasikan' : 'Folder dijadikan privat')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas folder')
  } finally {
    togglingVisibility.value = false
  }
}

async function toggleMediaVisibility(m) {
  try {
    const res = await galleryService.setMediaVisibility(m.id, !m.is_public)
    m.is_public = res.data.is_public
    notificationStore.success(m.is_public ? 'Media dipublikasikan' : 'Media dijadikan privat')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas')
  }
}

async function setAsCover(m) {
  try {
    await galleryService.setFolderCover(folder.value.id, m.id)
    folder.value.cover_media_id = m.id
    notificationStore.success('Cover folder diperbarui')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal mengatur cover')
  }
}

async function regenerateThumb(m) {
  regeneratingId.value = m.id
  try {
    const res = await galleryService.regenerateThumbnail(m.id)
    m.thumbnail_url = res.data.thumbnail_url
    m.thumbnail_failed = !!res.data.thumbnail_failed
    notificationStore.success('Thumbnail diperbarui')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal regenerate thumbnail')
  } finally {
    regeneratingId.value = null
  }
}

async function applyBulkVisibility(isPublic) {
  bulkUpdating.value = true
  try {
    await galleryService.setBulkMediaVisibility(
      folder.value.id,
      Array.from(selectedIds.value),
      isPublic,
    )
    notificationStore.success(`Visibilitas ${selectedIds.value.size} media diperbarui`)
    selectedIds.value = new Set()
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas')
  } finally {
    bulkUpdating.value = false
  }
}

function openLightbox(m) {
  lightboxIndex.value = media.value.findIndex((x) => x.id === m.id)
  if (lightboxIndex.value === -1) lightboxIndex.value = 0
  showLightbox.value = true
}

async function handleLightboxToggle(m) {
  await toggleMediaVisibility(m)
}

async function handleLightboxDelete(m) {
  const wasLast = media.value.length === 1
  await handleDeleteMedia(m)
  if (wasLast) showLightbox.value = false
}

async function handleBulkDelete(selected) {
  const confirmed = await confirmDelete(`${selected.length} Media`, '')
  if (!confirmed) return
  bulkUpdating.value = true
  try {
    for (const id of selected) {
      await galleryService.deleteMedia(id)
    }
    notificationStore.success(`${selected.length} media dihapus`)
    selectedIds.value = new Set()
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus media')
  } finally {
    bulkUpdating.value = false
  }
}

async function handleDeleteMedia(m) {
  const confirmed = await confirmDelete('Media', m.original_filename)
  if (!confirmed) return
  try {
    await galleryService.deleteMedia(m.id)
    notificationStore.success('Media dihapus')
    const next = new Set(selectedIds.value)
    next.delete(m.id)
    selectedIds.value = next
    await fetchMedia(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus media')
  }
}

async function handleDeleteFolder() {
  const confirmed = await confirmDelete('Folder', folder.value.name)
  if (!confirmed) return
  try {
    await galleryService.deleteFolder(folder.value.id)
    notificationStore.success('Folder dihapus')
    router.push('/gallery')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus folder')
  }
}

onMounted(() => {
  fetchFolder()
  pollHandle = setInterval(() => {
    if (media.value.some((m) => m.thumbnail_failed)) {
      fetchMedia(pagination.value.page)
    }
  }, 10000)
})

onBeforeUnmount(() => {
  if (pollHandle) clearInterval(pollHandle)
})
</script>