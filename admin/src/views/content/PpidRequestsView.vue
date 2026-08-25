<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Permintaan Informasi PPID</h1>
    </div>

    <div class="flex flex-wrap gap-3 mb-4">
      <AppSelect v-model="statusFilter" @update:model-value="fetchRequests(1)">
        <option value="pending">Menunggu</option>
        <option value="approved">Disetujui</option>
        <option value="revoked">Dibatalkan</option>
      </AppSelect>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <LoadingSpinner v-if="loading" class="py-12" />
      <AppTable v-else :columns="columns" :data="requests" :loading="false">
        <template #ppid_id="{ row }">
          <span class="text-sm font-medium text-secondary-800 dark:text-secondary-200">{{ ppidTitles[row.ppid_id] || row.ppid_id || '—' }}</span>
        </template>

        <template #status="{ row }">
          <AppBadge :variant="getStatusVariant(row.status)">
            {{ formatStatus(row.status) }}
          </AppBadge>
        </template>

        <template #approved_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ row.approved_at ? formatDateTime(row.approved_at) : '—' }}</span>
        </template>

        <template #revoked_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ row.revoked_at ? formatDateTime(row.revoked_at) : '—' }}</span>
        </template>

        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDateTime(row.created_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton
              v-if="row.status === 'pending'"
              size="sm"
              variant="primary"
              @click="handleApprove(row)"
            >
              Setujui
            </AppButton>
            <AppButton
              v-if="row.status === 'pending'"
              size="sm"
              variant="danger"
              @click="handleRevoke(row)"
            >
              Batalkan
            </AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="statusFilter !== 'pending'"
            icon="📋"
            title="Tidak ada permintaan pada status ini"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="statusFilter = 'pending'; fetchRequests(1)">Lihat permintaan menunggu</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="📋"
            title="Tidak ada permintaan menunggu"
            hint="Semua permintaan informasi publik telah diproses."
          />
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="fetchRequests"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import AppTable from '../../components/common/AppTable.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { ppidService } from '../../services/ppid.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDateTime } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()
const { confirm } = useConfirm()

const columns = [
  { key: 'ppid_id', label: 'Dokumen' },
  { key: 'requester_name', label: 'Nama Pemohon' },
  { key: 'requester_email', label: 'Email', width: '200px' },
  { key: 'purpose', label: 'Tujuan' },
  { key: 'status', label: 'Status', width: '110px' },
  { key: 'approved_at', label: 'Disetujui', width: '160px' },
  { key: 'revoked_at', label: 'Dibatalkan', width: '160px' },
  { key: 'created_at', label: 'Tanggal Permintaan', width: '160px' },
  { key: 'actions', label: '', width: '120px' },
]

const requests = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const ppidTitles = ref({})

const { refs: urlRefs, syncToUrl } = useUrlFilters(['status'])
const statusFilter = urlRefs.status
if (!statusFilter.value) statusFilter.value = 'pending'

watch(statusFilter, () => syncToUrl())

function formatStatus(status) {
  const map = { pending: 'Menunggu', approved: 'Disetujui', revoked: 'Dibatalkan' }
  return map[status] || status
}

const emptyHint = computed(() => `Tidak ada permintaan dengan status "${formatStatus(statusFilter.value)}".`)

function getStatusVariant(status) {
  const map = { pending: 'warning', approved: 'success', revoked: 'danger' }
  return map[status] || 'secondary'
}

async function fetchPpidTitles() {
  try {
    const res = await ppidService.adminList({ limit: 100 })
    const items = res.data.ppid || res.data || []
    items.forEach((item) => {
      ppidTitles.value[item.id] = item.title
    })
  } catch {
    // Non-critical, leave ppidTitles empty
  }
}

async function fetchRequests(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (statusFilter.value) params.status = statusFilter.value
    const res = await ppidService.listRequests(params)
    const payload = res.data
    requests.value = payload.requests || []
    pagination.value = payload.pagination || { page: 1, total_pages: 1 }
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat permintaan')
  } finally {
    loading.value = false
  }
}

async function handleApprove(request) {
  const confirmed = await confirm(
    'Setujui Permintaan',
    `Setujui permintaan dari "${request.requester_name}"?`,
    request.requester_name
  )
  if (!confirmed) return
  try {
    await ppidService.approveRequest(request.id)
    notificationStore.success('Permintaan disetujui')
    fetchRequests(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyetujui permintaan')
  }
}

async function handleRevoke(request) {
  const confirmed = await confirm(
    'Batalkan Permintaan',
    `Batalkan permintaan dari "${request.requester_name}"?`,
    request.requester_name
  )
  if (!confirmed) return
  try {
    await ppidService.revokeRequest(request.id)
    notificationStore.success('Permintaan dibatalkan')
    fetchRequests(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal membatalkan permintaan')
  }
}

onMounted(() => {
  fetchPpidTitles()
  fetchRequests()
})
</script>
