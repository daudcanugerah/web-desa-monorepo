<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">{{ $t('navigation.roles') }}</h1>
      <AppButton variant="primary" @click="openCreateModal">{{ $t('common.add') }} {{ $t('navigation.roles') }}</AppButton>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700">
      <AppTable :columns="columns" :data="roles" :loading="loading">
        <template #created_at="{ row }">
          <span class="text-secondary-600 dark:text-secondary-400 text-sm">{{ formatDate(row.created_at) }}</span>
        </template>
        <template #actions="{ row }">
          <div class="flex items-center gap-2">
            <AppButton variant="secondary" @click="openEditModal(row)" class="text-xs px-3 py-1.5">
              {{ $t('common.edit') }}
            </AppButton>
            <AppButton 
              v-if="!isDefaultRole(row.name)"
              variant="danger" 
              @click="handleDelete(row)" 
              class="text-xs px-3 py-1.5"
            >
              {{ $t('common.delete') }}
            </AppButton>
            <span v-else class="text-xs text-secondary-400 dark:text-secondary-500">—</span>
          </div>
        </template>
        <template #empty>
          <AppEmptyState
            icon="🔑"
            title="Belum ada role"
            hint="Tambahkan role pertama untuk mulai mengelola izin akses pengguna."
          >
            <AppButton variant="primary" @click="openCreateModal">{{ $t('common.add') }} {{ $t('navigation.roles') }}</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <!-- Permissions Modal -->
    <AppModal
      :show="permissionsModal.show"
      :title="`${$t('common.actions')} — ${permissionsModal.roleName}`"
      size="md"
      @close="closePermissionsModal"
    >
      <div v-if="permissionsModal.loading" class="flex justify-center py-8">
        <LoadingSpinner />
      </div>
      <div v-else class="space-y-2">
        <div class="flex items-center justify-between text-xs">
          <span class="text-secondary-500 dark:text-secondary-400">
            <span v-if="permissionsLoaded">📡 Katalog izin dari server</span>
            <span v-else-if="permissionsError" class="text-danger-600">⚠ Gagal memuat dari server</span>
            <span v-else class="text-secondary-400">Memuat katalog izin...</span>
          </span>
          <button v-if="permissionsError" type="button" class="text-primary-600 hover:underline" @click="loadAvailablePermissions">Coba lagi</button>
        </div>
        <p v-if="KNOWN_PERMISSIONS.length === 0" class="text-sm text-secondary-500 dark:text-secondary-400 py-4 text-center">
          Belum ada izin yang tersedia. Hubungi administrator server.
        </p>
        <label
          v-for="perm in KNOWN_PERMISSIONS"
          :key="perm.resource + ':' + perm.action"
          class="flex items-center gap-3 p-2 rounded hover:bg-secondary-50 dark:bg-secondary-800/50 cursor-pointer select-none"
        >
          <input
            type="checkbox"
            :checked="permissionsModal.selected.has(`${perm.resource}:${perm.action}`)"
            class="rounded border-secondary-300 dark:border-secondary-600 text-blue-600 focus:ring-blue-500"
            @change="(e) => toggleModalPermission(perm.resource, perm.action, e.target.checked)"
          />
          <span class="text-sm text-secondary-700 dark:text-secondary-300">{{ perm.resource }}:{{ perm.action }}</span>
        </label>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <AppButton variant="secondary" @click="closePermissionsModal">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" :loading="permissionsModal.saving" @click="savePermissions">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </template>
    </AppModal>

    <!-- Create Role Modal -->
    <AppModal
      :show="createModal.show"
      :title="`${$t('common.add')} ${$t('navigation.roles')}`"
      size="md"
      @close="closeCreateModal"
    >
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">{{ $t('navigation.roles') }} <span class="text-red-500">*</span></label>
          <input
            v-model="createModal.name"
            type="text"
            :placeholder="`${$t('common.add')} ${$t('navigation.roles')}`"
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            @keyup.enter="handleCreateRole"
          />
          <p v-if="createModal.nameError" class="mt-1 text-xs text-red-500">{{ createModal.nameError }}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">{{ $t('common.actions') }}</p>
          <div class="space-y-1 max-h-64 overflow-y-auto border border-secondary-200 dark:border-secondary-700 rounded p-2">
            <label
              v-for="perm in KNOWN_PERMISSIONS"
              :key="perm.resource + ':' + perm.action"
              class="flex items-center gap-3 p-1.5 rounded hover:bg-secondary-50 dark:bg-secondary-800/50 cursor-pointer select-none"
            >
              <input
                type="checkbox"
                :checked="createModal.selected.has(`${perm.resource}:${perm.action}`)"
                class="rounded border-secondary-300 dark:border-secondary-600 text-blue-600 focus:ring-blue-500"
                @change="(e) => toggleCreatePermission(perm.resource, perm.action, e.target.checked)"
              />
              <span class="text-sm text-secondary-700 dark:text-secondary-300">{{ perm.resource }}:{{ perm.action }}</span>
            </label>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <AppButton variant="secondary" @click="closeCreateModal">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" :loading="createModal.saving" @click="handleCreateRole">
            {{ $t('common.add') }} {{ $t('navigation.roles') }}
          </AppButton>
        </div>
      </template>
    </AppModal>

    <!-- Edit Role Modal -->
    <AppModal
      :show="editModal.show"
      :title="`${$t('common.edit')} {{ $t('navigation.roles') }} — ${editModal.roleName}`"
      size="md"
      @close="closeEditModal"
    >
      <div v-if="editModal.loading" class="flex justify-center py-8">
        <LoadingSpinner />
      </div>
      <div v-else class="space-y-4">
        <div>
          <p class="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">{{ $t('navigation.roles') }}</p>
          <p class="px-3 py-2 bg-secondary-100 dark:bg-secondary-800 rounded text-sm text-secondary-600 dark:text-secondary-400">{{ editModal.roleName }}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">{{ $t('common.actions') }}</p>
          <div class="space-y-1 max-h-64 overflow-y-auto border border-secondary-200 dark:border-secondary-700 rounded p-2">
            <label
              v-for="perm in KNOWN_PERMISSIONS"
              :key="perm.resource + ':' + perm.action"
              class="flex items-center gap-3 p-1.5 rounded hover:bg-secondary-50 dark:bg-secondary-800/50 cursor-pointer select-none"
            >
              <input
                type="checkbox"
                :checked="editModal.selected.has(`${perm.resource}:${perm.action}`)"
                class="rounded border-secondary-300 dark:border-secondary-600 text-blue-600 focus:ring-blue-500"
                @change="(e) => toggleEditPermission(perm.resource, perm.action, e.target.checked)"
              />
              <span class="text-sm text-secondary-700 dark:text-secondary-300">{{ perm.resource }}:{{ perm.action }}</span>
            </label>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <AppButton variant="secondary" @click="closeEditModal">{{ $t('common.cancel') }}</AppButton>
          <AppButton variant="primary" :loading="editModal.saving" @click="saveEditPermissions">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppButton from '../../components/common/AppButton.vue'
import AppTable from '../../components/common/AppTable.vue'
import AppModal from '../../components/common/AppModal.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { roleService } from '../../services/role.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { formatDate } from '../../utils/dateFormat'

const { t } = useI18n()
const notificationStore = useNotificationStore()
const { confirm } = useConfirm()

const DEFAULT_ROLES = ['admin', 'operator']

function isDefaultRole(roleName) {
  return DEFAULT_ROLES.includes(roleName)
}

const KNOWN_PERMISSIONS = ref([])
const permissionsLoaded = ref(false)
const permissionsError = ref(false)

const columns = computed(() => [
  { key: 'name', label: t('navigation.roles') },
  { key: 'created_at', label: t('common.createdAt'), width: '160px' },
  { key: 'actions', label: t('common.actions'), width: '220px' },
])

const roles = ref([])
const loading = ref(false)

// --- Permissions Modal ---
const permissionsModal = reactive({
  show: false,
  roleName: '',
  loading: false,
  saving: false,
  original: new Set(),
  selected: new Set(),
})

async function openPermissionsModal(role) {
  permissionsModal.roleName = role.name
  permissionsModal.show = true
  permissionsModal.loading = true
  permissionsModal.original = new Set()
  permissionsModal.selected = new Set()
  try {
    const res = await roleService.getPermissions(role.name)
    const perms = res.data
    const set = new Set(perms.map((p) => `${p.resource}:${p.action}`))
    permissionsModal.original = new Set(set)
    permissionsModal.selected = new Set(set)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.loadFailed'))
  } finally {
    permissionsModal.loading = false
  }
}

function closePermissionsModal() {
  permissionsModal.show = false
}

function toggleModalPermission(resource, action, checked) {
  const key = `${resource}:${action}`
  const updated = new Set(permissionsModal.selected)
  if (checked) updated.add(key)
  else updated.delete(key)
  permissionsModal.selected = updated
}

async function savePermissions() {
  permissionsModal.saving = true
  try {
    const toAdd = [...permissionsModal.selected].filter((k) => !permissionsModal.original.has(k))
    const toRemove = [...permissionsModal.original].filter((k) => !permissionsModal.selected.has(k))

    await Promise.all([
      ...toAdd.map((k) => {
        const [resource, action] = k.split(':')
        return roleService.addPermission(permissionsModal.roleName, resource, action)
      }),
      ...toRemove.map((k) => {
        const [resource, action] = k.split(':')
        return roleService.removePermission(permissionsModal.roleName, resource, action)
      }),
    ])

    notificationStore.success(t('messages.saveSuccess'))
    closePermissionsModal()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.saveFailed'))
  } finally {
    permissionsModal.saving = false
  }
}

// --- Create Modal ---
const createModal = reactive({
  show: false,
  name: '',
  nameError: '',
  saving: false,
  selected: new Set(),
})

function openCreateModal() {
  createModal.show = true
  createModal.name = ''
  createModal.nameError = ''
  createModal.saving = false
  createModal.selected = new Set()
  if (KNOWN_PERMISSIONS.value.length === 0) loadAvailablePermissions()
}

function closeCreateModal() {
  createModal.show = false
}

function toggleCreatePermission(resource, action, checked) {
  const key = `${resource}:${action}`
  const updated = new Set(createModal.selected)
  if (checked) updated.add(key)
  else updated.delete(key)
  createModal.selected = updated
}

async function handleCreateRole() {
  const name = createModal.name.trim()
  if (!name) {
    createModal.nameError = t('validation.required')
    return
  }
  createModal.nameError = ''
  createModal.saving = true
  try {
    await roleService.create(name)
    // Assign all selected permissions in batch
    await Promise.all(
      [...createModal.selected].map((k) => {
        const [resource, action] = k.split(':')
        return roleService.addPermission(name, resource, action)
      })
    )
    notificationStore.success(t('messages.saveSuccess'))
    closeCreateModal()
    await fetchRoles()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.saveFailed'))
  } finally {
    createModal.saving = false
  }
}

// --- Edit Modal ---
const editModal = reactive({
  show: false,
  roleName: '',
  loading: false,
  saving: false,
  original: new Set(),
  selected: new Set(),
})

async function openEditModal(role) {
  editModal.roleName = role.name
  editModal.show = true
  editModal.loading = true
  editModal.original = new Set()
  editModal.selected = new Set()
  try {
    const res = await roleService.getPermissions(role.name)
    const perms = res.data
    const set = new Set(perms.map((p) => `${p.resource}:${p.action}`))
    editModal.original = new Set(set)
    editModal.selected = new Set(set)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.loadFailed'))
  } finally {
    editModal.loading = false
  }
}

function closeEditModal() {
  editModal.show = false
}

function toggleEditPermission(resource, action, checked) {
  const key = `${resource}:${action}`
  const updated = new Set(editModal.selected)
  if (checked) updated.add(key)
  else updated.delete(key)
  editModal.selected = updated
}

async function saveEditPermissions() {
  editModal.saving = true
  try {
    const toAdd = [...editModal.selected].filter((k) => !editModal.original.has(k))
    const toRemove = [...editModal.original].filter((k) => !editModal.selected.has(k))

    await Promise.all([
      ...toAdd.map((k) => {
        const [resource, action] = k.split(':')
        return roleService.addPermission(editModal.roleName, resource, action)
      }),
      ...toRemove.map((k) => {
        const [resource, action] = k.split(':')
        return roleService.removePermission(editModal.roleName, resource, action)
      }),
    ])

    notificationStore.success(t('messages.saveSuccess'))
    closeEditModal()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.saveFailed'))
  } finally {
    editModal.saving = false
  }
}

// --- Delete ---
async function handleDelete(role) {
  const confirmed = await confirm(
    t('common.confirm'),
    t('messages.confirmDelete'),
    role.name
  )
  if (!confirmed) return
  try {
    await roleService.delete(role.name)
    notificationStore.success(t('messages.deleteSuccess'))
    await fetchRoles()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.deleteFailed'))
  }
}

// --- Fetch ---
async function fetchRoles() {
  loading.value = true
  try {
    const res = await roleService.list()
    roles.value = res.data
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadAvailablePermissions() {
  try {
    const res = await roleService.getAvailablePermissions()
    const list = Array.isArray(res.data) ? res.data : (res.data?.permissions || [])
    if (Array.isArray(list) && list.length) {
      KNOWN_PERMISSIONS.value = list
        .filter((p) => p && p.resource && p.action)
        .map((p) => ({ resource: p.resource, action: p.action }))
      permissionsLoaded.value = true
    }
  } catch {
    permissionsError.value = true
  }
}

onMounted(() => {
  loadAvailablePermissions()
  fetchRoles()
})
</script>
