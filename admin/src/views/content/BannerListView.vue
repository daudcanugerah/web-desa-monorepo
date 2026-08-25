<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Banner</h1>
      <div class="flex gap-2">
        <AppButton variant="secondary" @click="openActivePreview">Lihat Aktif</AppButton>
        <AppButton variant="primary" @click="$router.push('/banners/create')">Tambah Banner</AppButton>
      </div>
    </div>

<!-- Status filter -->
    <div class="mb-4 inline-flex rounded-lg border border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800 p-1" role="tablist" aria-label="Filter status banner">
      <button
        v-for="opt in statusOptions"
        :key="opt.value"
        type="button"
        role="tab"
        :aria-selected="statusFilter === opt.value"
        class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
        :class="statusFilter === opt.value
          ? 'bg-primary-600 text-white'
          : 'text-secondary-600 dark:text-secondary-300 hover:bg-secondary-100 dark:hover:bg-secondary-700'"
        @click="setStatusFilter(opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="banners" :loading="loading">
        <template #link="{ row }">
          <a
            v-if="row.link"
            :href="row.link"
            target="_blank"
            rel="noopener noreferrer"
            class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline text-sm truncate max-w-xs block"
          >
            {{ row.link }}
          </a>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-sm">—</span>
        </template>

<template #category="{ row }">
          <AppBadge v-if="row.category?.name" variant="primary">{{ row.category.name }}</AppBadge>
          <span v-else-if="row.category_name" class="text-sm text-secondary-700 dark:text-secondary-200">{{ row.category_name }}</span>
          <span v-else class="text-secondary-400 dark:text-secondary-500 text-xs">—</span>
        </template>

        <template #status="{ row }">
          <AppButton
            size="sm"
            :variant="row.status === 'active' ? 'success' : 'secondary'"
            :loading="togglingId === row.id"
            @click="toggleStatus(row)"
          >
            {{ row.status === 'active' ? 'Aktif' : 'Nonaktif' }}
          </AppButton>
        </template>

        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDate(row.created_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="secondary" @click="$router.push(`/banners/${row.id}/edit`)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="statusFilter"
            icon="search"
            title="Tidak ada banner dengan filter ini"
            :hint="`Coba pilih status lain atau tampilkan semua.`"
          >
            <AppButton variant="ghost" @click="setStatusFilter('')">Tampilkan semua</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="🖼️"
            title="Belum ada banner"
            hint="Tambahkan banner pertama untuk mulai mengelola tampilan beranda."
          >
            <AppButton variant="primary" @click="$router.push('/banners/create')">Tambah Banner</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="loadPage"
    />

    <AppModal :show="showActivePreview" title="Banner Aktif (Tampilan Publik)" size="lg" @close="showActivePreview = false">
      <div v-if="activeLoading" class="flex items-center justify-center h-48">
        <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-primary-600" />
      </div>
      <div v-else-if="activeBanners.length === 0" class="text-center text-sm text-secondary-500 dark:text-secondary-400 py-6">
        Tidak ada banner aktif.
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="b in activeBanners"
          :key="b.id"
          class="border border-secondary-200 dark:border-secondary-700 rounded-lg overflow-hidden"
        >
          <SafeImg :src="bannerImageUrl(b)" :alt="b.title" class="w-full aspect-video object-cover" />
          <div class="p-3">
            <p class="text-sm font-medium text-secondary-800 dark:text-secondary-200 truncate">{{ b.title }}</p>
            <a
              v-if="b.link"
              :href="b.link"
              target="_blank"
              rel="noopener noreferrer"
              class="text-xs text-blue-600 dark:text-blue-400 hover:underline truncate block"
            >
              {{ b.link }}
            </a>
          </div>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import AppTable from '../../components/common/AppTable.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import AppModal from '../../components/common/AppModal.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import { bannerService } from '../../services/banner.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDate } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()

const columns = [
  { key: 'title', label: 'Judul' },
  { key: 'category', label: 'Kategori', width: '140px' },
  { key: 'link', label: 'Link', width: '200px' },
  { key: 'status', label: 'Status', width: '110px' },
  { key: 'created_at', label: 'Dibuat', width: '130px' },
  { key: 'actions', label: '', width: '140px' },
]

const statusOptions = [
  { label: 'Semua', value: '' },
  { label: 'Aktif', value: 'active' },
  { label: 'Nonaktif', value: 'inactive' },
]

const banners = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const togglingId = ref(null)
const showActivePreview = ref(false)
const activeLoading = ref(false)
const activeBanners = ref([])

const { refs: urlRefs, syncToUrl } = useUrlFilters(['status'])
const statusFilter = urlRefs.status

watch(statusFilter, () => syncToUrl())

async function fetchBanners(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (statusFilter.value) params.status = statusFilter.value
    const res = await bannerService.list(params)
    banners.value = res.data.banners
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat banners')
  } finally {
    loading.value = false
  }
}

function setStatusFilter(val) {
  statusFilter.value = val
  fetchBanners(1)
}

function loadPage(page) {
  fetchBanners(page)
}

async function toggleStatus(banner) {
  togglingId.value = banner.id
  const newStatus = banner.status === 'active' ? 'inactive' : 'active'
  try {
    await bannerService.updateStatus(banner.id, newStatus)
    banner.status = newStatus
    notificationStore.success(`Banner diatur ke ${newStatus}`)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui status')
  } finally {
    togglingId.value = null
  }
}

async function handleDelete(banner) {
  const confirmed = await confirmDelete('Banner', banner.title)
  if (!confirmed) return
  try {
    await bannerService.delete(banner.id)
    notificationStore.success('Banner dihapus')
    fetchBanners(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus banner')
  }
}

async function openActivePreview() {
  showActivePreview.value = true
  if (activeBanners.value.length > 0) return
  activeLoading.value = true
  try {
    const res = await bannerService.getActive()
    activeBanners.value = Array.isArray(res.data) ? res.data : []
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat banner aktif')
  } finally {
    activeLoading.value = false
  }
}

// Resolve image URL from either the new schema (media.url / media.thumbnail_url)
// or the legacy top-level image_url field.
function bannerImageUrl(banner) {
  return banner?.media?.thumbnail_url || banner?.media?.url || banner?.image_url || ''
}

onMounted(() => fetchBanners())
</script>
