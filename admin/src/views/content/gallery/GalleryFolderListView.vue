<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Galeri</h1>
      <AppButton variant="primary" @click="$router.push('/gallery/folders/create')">Tambah Folder</AppButton>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <AppSearchInput
        v-model="searchQuery"
        placeholder="Cari folder..."
        class="flex-1 min-w-[200px] max-w-md"
        @input="onSearchInput"
      />
      <div class="inline-flex rounded-lg border border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800 p-1" role="tablist" aria-label="Filter visibilitas folder">
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

    <div v-if="loading && folders.length === 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div v-for="i in 8" :key="`sk-${i}`" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden animate-pulse">
        <div class="aspect-video bg-secondary-200 dark:bg-secondary-700" />
        <div class="p-4 space-y-2">
          <div class="h-4 bg-secondary-200 dark:bg-secondary-700 rounded w-3/4" />
          <div class="h-3 bg-secondary-200 dark:bg-secondary-700 rounded w-1/2" />
        </div>
      </div>
    </div>

    <div v-else-if="folders.length === 0" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-10">
      <AppEmptyState
        icon="🖼️"
        title="Belum ada folder galeri"
        hint="Buat folder untuk mengelompokkan foto dan video."
      >
        <AppButton variant="primary" @click="$router.push('/gallery/folders/create')">Tambah Folder</AppButton>
      </AppEmptyState>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <article
        v-for="folder in folders"
        :key="folder.id"
        class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden hover:shadow-md transition-shadow cursor-pointer group"
        @click="$router.push(`/gallery/folders/${folder.id}`)"
      >
        <div class="aspect-video bg-secondary-100 dark:bg-secondary-700 relative overflow-hidden">
          <SafeImg
            v-if="folder.cover_thumbnail_url"
            :src="folder.cover_thumbnail_url"
            :alt="folder.name"
            class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-secondary-400 dark:text-secondary-500">
            <svg class="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>

          <!-- Media count badge -->
          <span
            class="absolute bottom-2 left-2 px-2 py-0.5 text-xs font-medium rounded bg-black/50 text-white backdrop-blur-sm"
          >
            {{ folder.media_count ?? 0 }} 📷
          </span>

          <!-- Public/private + quick actions overlay -->
          <div class="absolute inset-0 flex flex-col justify-between p-2 bg-gradient-to-t from-black/60 via-transparent to-black/30 opacity-0 group-hover:opacity-100 transition-opacity">
            <div class="flex justify-end gap-1.5">
              <button
                type="button"
                class="p-1.5 rounded-lg bg-white/90 text-secondary-800 shadow hover:bg-white"
                :title="folder.is_public ? 'Jadikan privat' : 'Jadikan publik'"
                :aria-label="folder.is_public ? 'Jadikan privat' : 'Jadikan publik'"
                @click.stop="toggleVisibility(folder)"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    v-if="folder.is_public"
                    stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z"
                  />
                  <path
                    v-else
                    stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                  />
                </svg>
              </button>
              <button
                type="button"
                class="p-1.5 rounded-lg bg-white/90 text-secondary-800 shadow hover:bg-white"
                :title="`Edit ${folder.name}`"
                :aria-label="`Edit ${folder.name}`"
                @click.stop="$router.push(`/gallery/folders/${folder.id}/edit`)"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                type="button"
                class="p-1.5 rounded-lg bg-red-600/90 text-white shadow hover:bg-red-600"
                :title="`Hapus ${folder.name}`"
                :aria-label="`Hapus ${folder.name}`"
                @click.stop="handleDelete(folder)"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>

            <span
              class="self-start px-2 py-0.5 text-xs font-medium rounded"
              :class="folder.is_public
                ? 'bg-green-500/90 text-white'
                : 'bg-secondary-700/90 text-white'"
            >
              {{ folder.is_public ? 'Publik' : 'Privat' }}
            </span>
          </div>
        </div>
        <div class="p-4">
          <h3 class="font-semibold text-secondary-800 dark:text-secondary-200 truncate">{{ folder.name }}</h3>
          <p v-if="folder.description" class="mt-1 text-xs text-secondary-500 dark:text-secondary-400 line-clamp-2">
            {{ folder.description }}
          </p>
          <div class="mt-3 flex items-center justify-between text-xs text-secondary-500 dark:text-secondary-400">
            <span>{{ folder.media_count ?? 0 }} media</span>
            <span v-if="folder.public_media_count != null && folder.public_media_count < folder.media_count">
              ({{ folder.public_media_count }} publik)
            </span>
          </div>
        </div>
      </article>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-6"
      @change="loadPage"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AppButton from '../../../components/common/AppButton.vue'
import AppSearchInput from '../../../components/common/AppSearchInput.vue'
import AppPagination from '../../../components/common/AppPagination.vue'
import AppEmptyState from '../../../components/common/AppEmptyState.vue'
import SafeImg from '../../../components/common/SafeImg.vue'
import { galleryService } from '../../../services/gallery.service'
import { useNotificationStore } from '../../../stores/notification'
import { useConfirm } from '../../../composables/useConfirm'

const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const visibilityOptions = [
  { label: 'Semua', value: '' },
  { label: 'Publik', value: 'true' },
  { label: 'Privat', value: 'false' },
]

const folders = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const searchQuery = ref('')
const visibilityFilter = ref('')

let searchTimer = null

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchFolders(1), 300)
}

function setVisibilityFilter(val) {
  visibilityFilter.value = val
  fetchFolders(1)
}

function loadPage(page) {
  fetchFolders(page)
}

async function toggleVisibility(folder) {
  try {
    const res = await galleryService.setFolderVisibility(folder.id, !folder.is_public)
    folder.is_public = res.data.is_public
    notificationStore.success(folder.is_public ? 'Folder dipublikasikan' : 'Folder dijadikan privat')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui visibilitas folder')
  }
}

async function handleDelete(folder) {
  const confirmed = await confirmDelete('Folder', folder.name)
  if (!confirmed) return
  try {
    await galleryService.deleteFolder(folder.id)
    notificationStore.success('Folder dihapus')
    fetchFolders(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus folder')
  }
}

async function fetchFolders(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 12 }
    if (searchQuery.value.trim()) params.q = searchQuery.value.trim()
    if (visibilityFilter.value !== '') params.is_public = visibilityFilter.value
    const res = await galleryService.listFolders(params)
    folders.value = res.data.folders
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat folder galeri')
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchFolders())
</script>