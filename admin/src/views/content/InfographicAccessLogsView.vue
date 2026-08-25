<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3 min-w-0">
        <AppBackButton />
        <div class="min-w-0">
          <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 truncate">Access Log Infographic</h1>
          <p class="text-sm text-secondary-500 dark:text-secondary-400 mt-0.5">
            Audit trail untuk setiap generate token Metabase.
          </p>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-4 mb-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <AppInput
          v-model="infographicIdFilter"
          label="Filter Infographic ID"
          placeholder="UUID infographic"
        />
        <AppInput
          v-model="ipFilter"
          label="Filter IP Address"
          placeholder="192.168.1.1"
        />
      </div>
      <div class="flex gap-2 mt-4">
        <AppButton variant="primary" size="sm" :loading="loading" @click="fetchLogs(1)">Terapkan</AppButton>
        <AppButton variant="ghost" size="sm" @click="resetFilters">Reset</AppButton>
      </div>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="logs" :loading="loading">
        <template #infographic_id="{ row }">
          <span class="font-mono text-xs text-secondary-700 dark:text-secondary-300">{{ row.infographic_id }}</span>
        </template>

        <template #component="{ row }">
          <span class="text-xs">
            <AppBadge :variant="row.component_type === 'dashboard' ? 'primary' : 'warning'" class="mr-1">
              {{ row.component_type }}
            </AppBadge>
            <span class="font-mono text-secondary-600 dark:text-secondary-400">#{{ row.component_id }}</span>
          </span>
        </template>

        <template #ip="{ row }">
          <span class="font-mono text-xs">{{ row.ip_address || '—' }}</span>
        </template>

        <template #endpoint="{ row }">
          <span class="text-xs text-secondary-600 dark:text-secondary-400 font-mono">{{ row.endpoint || '—' }}</span>
        </template>

        <template #token_window="{ row }">
          <span class="text-xs text-secondary-600 dark:text-secondary-400">
            {{ formatDateTime(row.token_issued_at) }}
            <span class="text-secondary-400 dark:text-secondary-500">→</span>
            {{ formatDateTime(row.token_expires_at) }}
          </span>
        </template>

        <template #created_at="{ row }">
          <span class="text-xs text-secondary-600 dark:text-secondary-400">{{ formatDateTime(row.created_at) }}</span>
        </template>

        <template #empty>
          <AppEmptyState
            icon="📜"
            title="Belum ada access log"
            hint="Log akan muncul setelah ada yang membuka preview infographic."
          />
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBackButton from '../../components/common/AppBackButton.vue'
import AppInput from '../../components/common/AppInput.vue'
import AppTable from '../../components/common/AppTable.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import { infographicService } from '../../services/infographic.service'
import { useNotificationStore } from '../../stores/notification'
import { formatDateTime } from '../../utils/dateFormat'

const notificationStore = useNotificationStore()

const logs = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const infographicIdFilter = ref('')
const ipFilter = ref('')

const columns = [
  { key: 'created_at', label: 'Waktu', width: '180px' },
  { key: 'infographic_id', label: 'Infographic', width: '180px' },
  { key: 'component', label: 'Component', width: '160px' },
  { key: 'ip', label: 'IP', width: '140px' },
  { key: 'endpoint', label: 'Endpoint', width: '180px' },
  { key: 'token_window', label: 'Token Window', width: '220px' },
]

function resetFilters() {
  infographicIdFilter.value = ''
  ipFilter.value = ''
  fetchLogs(1)
}

function loadPage(page) {
  fetchLogs(page)
}

async function fetchLogs(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 20 }
    if (infographicIdFilter.value.trim()) params.infographic_id = infographicIdFilter.value.trim()
    if (ipFilter.value.trim()) params.ip = ipFilter.value.trim()
    const res = await infographicService.getAccessLogs(params)
    logs.value = res.data.logs
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat access log')
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchLogs())
</script>