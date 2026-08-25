<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Fasilitas</h1>
      <AppButton variant="primary" @click="$router.push('/fasilitas/create')">Tambah Fasilitas</AppButton>
    </div>

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <AppSearchInput
        v-model="search"
        placeholder="Cari fasilitas..."
        width="md"
        @search="onSearch"
      />
      <div class="w-48">
        <AppSelect v-model="categoryFilter" @update:model-value="fetchFacilities(1)">
          <option value="">{{ t('filter.allCategories') }}</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </AppSelect>
      </div>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="items" :loading="loading">
        <template #thumbnail="{ row }">
          <SafeImg
            v-if="facilityThumbnail(row)"
            :src="facilityThumbnail(row)"
            :alt="row.name"
            class="w-12 h-12 rounded object-cover"
          />
          <div v-else class="w-12 h-12 rounded bg-secondary-100 dark:bg-secondary-700 flex items-center justify-center">
            <svg class="w-6 h-6 text-secondary-400 dark:text-secondary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h3m-3 0H2m9-16h2m-2 4h2" />
            </svg>
          </div>
        </template>

        <template #name="{ row }">
          <div class="min-w-0">
            <p class="text-sm font-medium text-secondary-800 dark:text-secondary-100 truncate">{{ row.name }}</p>
            <p v-if="row.description" class="text-xs text-secondary-500 dark:text-secondary-400 truncate max-w-xs">{{ row.description }}</p>
          </div>
        </template>

        <template #category="{ row }">
          <AppBadge v-if="row.category?.name" variant="primary">{{ row.category.name }}</AppBadge>
          <span v-else-if="row.category_name" class="text-sm text-secondary-700 dark:text-secondary-200">{{ row.category_name }}</span>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-xs">—</span>
        </template>

        <template #coordinates="{ row }">
          <span class="text-xs text-secondary-500 dark:text-secondary-400 font-mono">{{ row.latitude }}, {{ row.longitude }}</span>
        </template>

        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDate(row.created_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="secondary" @click="$router.push(`/fasilitas/${row.id}/edit`)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="hasFilters"
            icon="search"
            title="Tidak ada fasilitas yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset filter</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="🏛️"
            title="Belum ada fasilitas"
            hint="Tambahkan fasilitas pertama untuk mulai menampilkan lokasi penting di desa."
          >
            <AppButton variant="primary" @click="$router.push('/fasilitas/create')">Tambah Fasilitas</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="fetchFacilities"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppTable from '../../components/common/AppTable.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import { getImageUrl } from '../../utils/imageUrl'
import { formatDate } from '../../utils/dateFormat'
import { facilityService } from '../../services/facility.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'

const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()
const { t } = useI18n()

const columns = [
  { key: 'thumbnail', label: '', width: '60px' },
  { key: 'name', label: 'Nama' },
  { key: 'category', label: 'Kategori', width: '140px' },
  { key: 'coordinates', label: 'Koordinat', width: '180px' },
  { key: 'created_at', label: 'Dibuat', width: '120px' },
  { key: 'actions', label: '', width: '140px' },
]

const items = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const search = ref('')
const categoryFilter = ref('')
const categories = ref([])

const hasFilters = computed(() => !!(search.value || categoryFilter.value))
const emptyHint = computed(() => {
  const parts = []
  if (search.value) parts.push(`pencarian "${search.value}"`)
  if (categoryFilter.value) parts.push('kategori terpilih')
  return `Tidak ada hasil untuk ${parts.join(', ')}.`
})

let searchTimer = null

function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchFacilities(1), 300)
}

function resetFilters() {
  search.value = ''
  categoryFilter.value = ''
  fetchFacilities(1)
}

function facilityThumbnail(row) {
  const first = row.media?.[0]
  return first ? getImageUrl(first.thumbnail_url || first.url || '') : ''
}

async function fetchFacilities(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (search.value.trim()) params.q = search.value.trim()
    if (categoryFilter.value) params.category = categoryFilter.value
    const res = await facilityService.list(params)
    items.value = res.data.fasilitas
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat fasilitas')
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await facilityService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categories.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
      .sort((a, b) => a.name.localeCompare(b.name))
  } catch {
    categories.value = []
  }
}

async function handleDelete(item) {
  const confirmed = await confirmDelete('Fasilitas', item.name)
  if (!confirmed) return
  try {
    await facilityService.delete(item.id)
    notificationStore.success('Fasilitas dihapus')
    fetchFacilities(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus fasilitas')
  }
}

watch([search, categoryFilter], () => {
  if (search.value || categoryFilter.value) {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(() => fetchFacilities(1), 300)
  }
})

onMounted(() => {
  loadCategories()
  fetchFacilities()
})
</script>