<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">UMKM</h1>
      <AppButton variant="primary" @click="$router.push('/umkm/create')">Tambah UMKM</AppButton>
    </div>

    <div class="mb-4">
      <AppSearchInput
        v-model="search"
        placeholder="Cari nama usaha..."
        width="sm"
        @search="onSearch"
      />
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="items" :loading="loading">
        <template #category="{ row }">
          <AppBadge v-if="row.category?.name" variant="primary">{{ row.category.name }}</AppBadge>
          <span v-else-if="row.category_name" class="text-sm text-secondary-700 dark:text-secondary-200">{{ row.category_name }}</span>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-xs">—</span>
        </template>

        <template #created_at="{ row }">
          <div class="text-sm text-secondary-600 dark:text-secondary-400">
            <div>{{ formatDate(row.created_at) }}</div>
            <div v-if="row.updated_at && row.updated_at !== row.created_at" class="text-xs text-secondary-500 dark:text-secondary-400">{{ formatDate(row.updated_at) }}</div>
          </div>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="secondary" @click="$router.push(`/umkm/${row.id}/edit`)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="search"
            icon="search"
            title="Tidak ada UMKM yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset pencarian</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="🏪"
            title="Belum ada UMKM"
            hint="Tambahkan UMKM pertama untuk mulai mengelola direktori usaha desa."
          >
            <AppButton variant="primary" @click="$router.push('/umkm/create')">Tambah UMKM</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="fetchUmkm"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import AppTable from '../../components/common/AppTable.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { umkmService } from '../../services/umkm.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDate } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const columns = [
  { key: 'name', label: 'Nama Usaha' },
  { key: 'owner', label: 'Pemilik' },
  { key: 'category', label: 'Kategori', width: '120px' },
  { key: 'phone', label: 'Telepon', width: '140px' },
  { key: 'created_at', label: 'Dibuat / Diperbarui', width: '180px' },
  { key: 'actions', label: '', width: '140px' },
]

const items = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q'])
const search = urlRefs.q

watch(search, () => syncToUrl())

const emptyHint = computed(() => `Pencarian "${search.value}" tidak menemukan hasil.`)

function onSearch() {
  fetchUmkm(1)
}

function resetFilters() {
  resetUrlFilters()
  fetchUmkm(1)
}

async function fetchUmkm(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (search.value) params.search = search.value
    const res = await umkmService.list(params)
    items.value = res.data.umkm
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat data UMKM')
  } finally {
    loading.value = false
  }
}

async function handleDelete(item) {
  const confirmed = await confirmDelete('UMKM', item.name)
  if (!confirmed) return
  try {
    await umkmService.delete(item.id)
    notificationStore.success('UMKM dihapus')
    fetchUmkm(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus UMKM')
  }
}

onMounted(() => fetchUmkm())
</script>
