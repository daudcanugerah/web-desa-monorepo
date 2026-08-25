<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Berita</h1>
      <AppButton variant="primary" @click="$router.push('/berita/create')">Tambah Berita</AppButton>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4 p-4 bg-secondary-100 dark:bg-secondary-800/50 rounded-lg">
      <AppSearchInput
        v-model="search"
        placeholder="Cari judul..."
        width="xs"
        @search="onSearch"
      />
      <AppSelect v-model="categoryFilter" @update:model-value="fetchNews(1)">
        <option value="">{{ t('filter.allCategories') }}</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
      </AppSelect>
      <div class="flex gap-2">
        <input
          v-model="dateFrom"
          type="date"
          class="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
          @change="fetchNews(1)"
        />
        <span class="text-sm text-secondary-500 dark:text-secondary-400 flex items-center">s/d</span>
        <input
          v-model="dateTo"
          type="date"
          class="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
          @change="fetchNews(1)"
        />
      </div>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="articles" :loading="loading">
        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDateTime(row.created_at) }}</span>
        </template>

        <template #category="{ row }">
          <AppBadge v-if="row.category?.name" variant="primary">{{ row.category.name }}</AppBadge>
          <span v-else-if="row.category_name" class="text-sm text-secondary-700 dark:text-secondary-200">{{ row.category_name }}</span>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-xs">—</span>
        </template>

        <template #updated_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDateTime(row.updated_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="secondary" @click="$router.push(`/berita/${row.id}/edit`)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="hasFilters"
            icon="search"
            title="Tidak ada berita yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset filter</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="📰"
            title="Belum ada berita"
            hint="Publikasikan berita pertama untuk mulai mengelola konten."
          >
            <AppButton variant="primary" @click="$router.push('/berita/create')">Publikasikan Berita</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="fetchNews"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppTable from '../../components/common/AppTable.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { newsService } from '../../services/news.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDateTime } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()
const { t } = useI18n()
const { confirmDelete } = useConfirm()

const columns = [
  { key: 'title', label: 'Judul' },
  { key: 'category', label: 'Kategori', width: '120px' },
  { key: 'created_at', label: 'Dibuat', width: '130px' },
  { key: 'updated_at', label: 'Diperbarui', width: '130px' },
  { key: 'actions', label: '', width: '140px' },
]

const articles = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const categories = ref([])

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q', 'since', 'until', 'category'])
const search = urlRefs.q
const dateFrom = urlRefs.since
const dateTo = urlRefs.until
const categoryFilter = urlRefs.category

watch([search, dateFrom, dateTo, categoryFilter], () => syncToUrl())

const hasFilters = computed(() => !!(search.value || dateFrom.value || dateTo.value || categoryFilter.value))
const emptyHint = computed(() => {
  const parts = []
  if (search.value) parts.push(`pencarian "${search.value}"`)
  if (categoryFilter.value) parts.push('kategori ini')
  if (dateFrom.value || dateTo.value) parts.push('rentang tanggal ini')
  return `Tidak ada hasil untuk ${parts.join(', ')}.`
})

function onSearch() {
  fetchNews(1)
}

function resetFilters() {
  resetUrlFilters()
  fetchNews(1)
}

async function fetchNews(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (search.value) params.q = search.value
    if (dateFrom.value) params.since = dateFrom.value
    if (dateTo.value) params.until = dateTo.value
    if (categoryFilter.value) params.category = categoryFilter.value
    const res = await newsService.list(params)
    articles.value = res.data.berita
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat berita')
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await newsService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categories.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
      .sort((a, b) => a.name.localeCompare(b.name))
  } catch {
    categories.value = []
  }
}

async function handleDelete(article) {
  const confirmed = await confirmDelete('Berita', article.title)
  if (!confirmed) return
  try {
    await newsService.delete(article.id)
    notificationStore.success('Berita dihapus')
    fetchNews(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus berita')
  }
}

onMounted(() => {
  loadCategories()
  fetchNews()
})
</script>
