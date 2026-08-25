<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">PPID</h1>
      <AppButton variant="primary" @click="$router.push('/ppid/create')">Tambah Dokumen</AppButton>
    </div>

    <div class="flex flex-wrap gap-3 mb-4">
      <AppSearchInput
        v-model="search"
        placeholder="Cari judul..."
        width="xs"
        @search="onSearch"
      />
      <AppSelect v-model="categoryFilter" @update:model-value="fetchPpid(1)">
        <option value="">{{ t('filter.allCategories') }}</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
      </AppSelect>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <LoadingSpinner v-if="loading" class="py-12" />
      <AppTable v-else :columns="columns" :data="items" :loading="false">
        <template #thumbnail="{ row }">
          <SafeImg
            v-if="ppidThumbnail(row)"
            :src="ppidThumbnail(row)"
            :alt="row.title"
            class="w-12 h-12 rounded"
          />
          <div v-else class="w-12 h-12 rounded bg-secondary-200 dark:bg-secondary-700 flex items-center justify-center">
            <svg class="w-6 h-6 text-secondary-400 dark:text-secondary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        </template>

        <template #category="{ row }">
          <AppBadge v-if="row.category?.name" variant="primary">
            {{ row.category.name }}
          </AppBadge>
          <span v-else-if="row.category_name" class="text-sm text-secondary-700 dark:text-secondary-200">{{ row.category_name }}</span>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-xs">—</span>
        </template>

        <template #publication_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDate(row.publication_at) }}</span>
        </template>

        <template #created_at="{ row }">
          <div class="text-sm text-secondary-600 dark:text-secondary-400">
            <div>{{ formatDateTime(row.created_at) }}</div>
            <div v-if="row.updated_at && row.updated_at !== row.created_at" class="text-xs text-secondary-500 dark:text-secondary-400">{{ formatDateTime(row.updated_at) }}</div>
          </div>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton
              v-if="ppidDocumentUrl(row)"
              size="sm"
              variant="primary"
              :loading="downloadingId === row.id"
              @click="handleDownload(row)"
            >
              Download
            </AppButton>
            <AppButton size="sm" variant="secondary" @click="$router.push(`/ppid/${row.id}/edit`)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="hasFilters"
            icon="search"
            title="Tidak ada dokumen PPID yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset filter</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="document"
            title="Belum ada dokumen PPID"
            hint="Unggah dokumen publik pertama untuk mulai mengelola informasi PPID."
          >
            <AppButton variant="primary" @click="$router.push('/ppid/create')">Unggah Dokumen</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="fetchPpid"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import AppTable from '../../components/common/AppTable.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { getImageUrl } from '../../utils/imageUrl'
import { useI18n } from 'vue-i18n'
import { ppidService } from '../../services/ppid.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDate, formatDateTime } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const columns = [
  { key: 'thumbnail', label: '', width: '60px' },
  { key: 'title', label: 'Judul' },
  { key: 'category', label: 'Kategori', width: '160px' },
  { key: 'publication_at', label: 'Publikasi', width: '180px' },
  { key: 'created_at', label: 'Dibuat / Diperbarui', width: '200px' },
  { key: 'actions', label: '', width: '150px' },
]

const categories = ref([])
const { t } = useI18n()

const items = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const downloadingId = ref(null)

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q', 'category'])
const search = urlRefs.q
const categoryFilter = urlRefs.category

watch([search, categoryFilter], () => syncToUrl())

const hasFilters = computed(() => !!(search.value || categoryFilter.value))
const emptyHint = computed(() => {
  const parts = []
  if (search.value) parts.push(`pencarian "${search.value}"`)
  if (categoryFilter.value) parts.push('kategori ini')
  return `Tidak ada hasil untuk ${parts.join(', ')}.`
})

function onSearch() {
  fetchPpid(1)
}

function resetFilters() {
  resetUrlFilters()
  fetchPpid(1)
}

async function fetchPpid(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (search.value) params.q = search.value
    if (categoryFilter.value) params.category = categoryFilter.value
    const res = await ppidService.adminList(params)
    const payload = res.data
    items.value = payload.ppid || []
    pagination.value = payload.pagination || { page: 1, total_pages: 1 }
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat dokumen PPID')
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await ppidService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categories.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
      .sort((a, b) => a.name.localeCompare(b.name))
  } catch {
    categories.value = []
  }
}

// New schema: document/thumbnail are MediaInfo {media_id, url, thumbnail_url}.
// getImageUrl prefixes the API host so relative signed URLs (/api/v1/media/...)
// resolve against the backend, not the SPA origin.
function ppidDocumentUrl(row) {
  return getImageUrl(row.document?.url || row.document?.thumbnail_url || row.document_url || '')
}

function ppidThumbnail(row) {
  return row.thumbnail?.thumbnail_url || row.thumbnail?.url || row.thumbnail_url || ''
}

async function handleDownload(row) {
  const url = ppidDocumentUrl(row)
  if (!url) return
  downloadingId.value = row.id
  try {
    // Fetch the signed content URL as a blob so we can force a download.
    // The token rides in the query string, so no auth header is needed.
    const res = await fetch(url)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const objUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = objUrl
    a.download = row.document?.filename || row.document_name || `${row.title || 'dokumen'}.pdf`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(objUrl)
  } catch {
    // Fallback: open the signed URL directly (browser shows PDF inline).
    window.open(url, '_blank', 'noopener,noreferrer')
  } finally {
    downloadingId.value = null
  }
}

async function handleDelete(item) {
  const confirmed = await confirmDelete('Dokumen', item.title)
  if (!confirmed) return
  try {
    await ppidService.delete(item.id)
    notificationStore.success('Dokumen dihapus')
    fetchPpid(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus dokumen')
  }
}

onMounted(() => {
  loadCategories()
  fetchPpid()
})
</script>