<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Users</h1>
      <AppButton variant="primary" @click="router.push('/users/create')">Tambah User</AppButton>
    </div>

    <!-- Search -->
    <div class="mb-4">
      <AppSearchInput
        v-model="searchInput"
        placeholder="Cari user..."
        width="sm"
        @search="onSearch"
      />
    </div>

    <!-- Table -->
    <AppTable
      :columns="columns"
      :data="users"
      :loading="loading"
      selectable
      :selected="bulk.selected.value"
      @toggle="bulk.toggle"
      @toggle-all="bulk.toggleAll"
    >
      <template #roles="{ row }">
        <div v-if="row.id === authStore.currentUser?.id" class="flex gap-1">
          <AppBadge v-for="role in row.roles" :key="role" variant="primary" size="sm">
            {{ role }}
          </AppBadge>
        </div>
        <select
          v-else
          :value="row.roles?.[0] || ''"
          :disabled="updatingRoleId === row.id"
          class="text-xs px-2 py-1 border border-secondary-300 dark:border-secondary-600 rounded bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
          :aria-label="`Ubah role untuk ${row.name}`"
          @change="(e) => updateRoleInline(row, e.target.value)"
        >
          <option value="" disabled>Pilih role</option>
          <option v-for="role in availableRoles" :key="role.name" :value="role.name">{{ role.name }}</option>
        </select>
      </template>

      <template #created_at="{ row }">
        <span class="text-secondary-600 dark:text-secondary-400 text-sm">{{ formatDate(row.created_at) }}</span>
      </template>

      <template #actions="{ row }">
        <div class="flex gap-2">
          <AppButton size="sm" variant="secondary" @click="router.push('/users/' + row.id + '/edit')">Edit</AppButton>
          <AppButton
            v-if="row.id !== authStore.currentUser?.id"
            size="sm"
            variant="danger"
            @click="handleDelete(row)"
          >Hapus</AppButton>
          <span v-else class="px-3 py-1 text-xs text-secondary-400 dark:text-secondary-500">—</span>
        </div>
      </template>

      <template #empty>
        <AppEmptyState
          v-if="searchInput"
          icon="search"
          title="Tidak ada user yang cocok"
          :hint="emptyHint"
        >
          <AppButton variant="ghost" @click="resetFilters">Reset pencarian</AppButton>
        </AppEmptyState>
        <AppEmptyState
          v-else
          icon="users"
          title="Belum ada user"
          hint="Tambahkan user pertama untuk mulai mengelola akses admin."
        >
          <AppButton variant="primary" @click="router.push('/users/create')">Tambah User</AppButton>
        </AppEmptyState>
      </template>
    </AppTable>

    <!-- Bulk action bar -->
    <AppBulkBar :selected="bulk.selected.value" @update:selected="(s) => bulk.selected.value = s">
      <template #default="{ selected: ids, clear }">
        <AppButton
          variant="danger"
          size="sm"
          :loading="bulkDeleting"
          @click="handleBulkDelete(ids, clear)"
        >
          Hapus {{ ids.length }} user
        </AppButton>
      </template>
    </AppBulkBar>

    <!-- Pagination -->
    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      @change="onPageChange"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppButton from '../../components/common/AppButton.vue'
import AppTable from '../../components/common/AppTable.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import AppBulkBar from '../../components/common/AppBulkBar.vue'
import { userService } from '../../services/user.service'
import { useNotificationStore } from '../../stores/notification'
import { useAuthStore } from '../../stores/auth'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { useBulkSelect } from '../../composables/useBulkSelect'
import { formatDate } from '../../utils/dateFormat'

const router = useRouter()
const notificationStore = useNotificationStore()
const authStore = useAuthStore()
const { confirmDelete } = useConfirm()

const users = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, limit: 10, total: 0, total_pages: 1 })

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q'])
const searchInput = urlRefs.q

watch(searchInput, () => syncToUrl())

const bulk = useBulkSelect()
const bulkDeleting = ref(false)

const availableRoles = ref([])
const updatingRoleId = ref(null)

async function loadAvailableRoles() {
  try {
    const res = await fetch(import.meta.env.VITE_API_BASE_URL + '/roles', {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!res.ok) return
    const data = await res.json()
    const list = Array.isArray(data) ? data : (data.roles || [])
    availableRoles.value = list.map((r) => (typeof r === 'string' ? { name: r } : r))
  } catch {
    // non-blocking
  }
}

async function updateRoleInline(row, newRole) {
  if (!newRole || row.roles?.includes(newRole)) return
  updatingRoleId.value = row.id
  const previous = [...(row.roles || [])]
  try {
    await userService.assignRole(row.id, newRole)
    row.roles = [newRole]
    notificationStore.success(`Role ${row.name} diubah ke ${newRole}`)
  } catch (err) {
    row.roles = previous
    notificationStore.error(err.response?.data?.error || 'Gagal mengubah role')
  } finally {
    updatingRoleId.value = null
  }
}

const emptyHint = computed(() => `Pencarian "${searchInput.value}" tidak menemukan hasil.`)

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'roles', label: 'Roles' },
  { key: 'created_at', label: 'Dibuat', width: '140px' },
  { key: 'actions', label: 'Actions', width: '160px' },
]

function onSearch() {
  pagination.value.page = 1
  fetchUsers()
}

function resetFilters() {
  resetUrlFilters()
  pagination.value.page = 1
  fetchUsers()
}

async function fetchUsers() {
  loading.value = true
  try {
    const res = await userService.list(pagination.value.page, pagination.value.limit, searchInput.value)
    users.value = res.data.users
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Failed to load users')
  } finally {
    loading.value = false
  }
}

function onPageChange(page) {
  pagination.value.page = page
  fetchUsers()
}

async function handleDelete(row) {
  const confirmed = await confirmDelete('User', row.name)
  if (!confirmed) return
  const snapshot = JSON.parse(JSON.stringify(row))
  try {
    await userService.delete(row.id)
    notificationStore.undo(
      `User "${snapshot.name}" dihapus`,
      async () => {
        try {
          await userService.create({
            name: snapshot.name,
            email: snapshot.email,
            password: snapshot.password || 'Webdesa2026!',
            roles: snapshot.roles || [],
          })
          notificationStore.success('User dipulihkan')
          fetchUsers()
        } catch (err) {
          notificationStore.error(err.response?.data?.error || 'Gagal memulihkan user')
        }
      },
      { actionLabel: 'Batalkan' },
    )
    fetchUsers()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus user')
  }
}

async function handleBulkDelete(ids, clear) {
  const confirmed = await confirmDelete(`${ids.length} User`, `${ids.length} user akan dihapus sekaligus.`)
  if (!confirmed) return
  bulkDeleting.value = true
  try {
    const results = await Promise.allSettled(ids.map((id) => userService.delete(id)))
    const succeeded = results.filter((r) => r.status === 'fulfilled').length
    const failed = results.length - succeeded
    clear()
    if (failed === 0) {
      notificationStore.success(`${succeeded} user dihapus`)
    } else {
      notificationStore.error(`${succeeded} berhasil, ${failed} gagal dihapus`)
    }
    fetchUsers()
  } finally {
    bulkDeleting.value = false
  }
}

onMounted(async () => {
  // Ensure currentUser is loaded for admin check
  if (!authStore.currentUser) {
    try {
      await authStore.fetchCurrentUser()
    } catch {
      // ignore
    }
  }
  if (!authStore.currentUser?.roles?.includes('admin')) {
    notificationStore.error('Akses ditolak')
    router.push('/')
    return
  }
  await Promise.all([fetchUsers(), loadAvailableRoles()])
})
</script>
