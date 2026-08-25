<template>
  <AppModal :show="show" :title="title" size="md" @close="emit('close')">
    <div class="space-y-4">
      <div class="flex gap-2">
        <AppInput
          v-model="newName"
          label="Nama Kategori Baru"
          placeholder="Contoh: Pengumuman, Layanan"
          :error="error"
          @keydown.enter.prevent="handleAdd"
        />
        <div class="flex items-end">
          <AppButton variant="primary" :loading="adding" :disabled="!newName.trim()" @click="handleAdd">
            Tambah
          </AppButton>
        </div>
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
          Daftar Kategori ({{ categories.length }})
        </label>
        <LoadingSpinner v-if="loading" class="py-6" />
        <div v-else-if="categories.length === 0" class="text-sm text-secondary-500 dark:text-secondary-400 py-6 text-center">
          Belum ada kategori. Tambahkan kategori pertama di atas.
        </div>
        <ul v-else class="divide-y divide-secondary-200 dark:divide-secondary-700 border border-secondary-200 dark:border-secondary-700 rounded-lg overflow-hidden">
          <li
            v-for="cat in categories"
            :key="cat.id || cat.name"
            class="flex items-center justify-between px-3 py-2 bg-white dark:bg-secondary-800"
          >
            <button
              type="button"
              class="flex-1 text-left text-sm text-secondary-700 dark:text-secondary-200 hover:text-blue-600 dark:hover:text-blue-400"
              :title="allowSelect ? 'Pilih kategori ini' : ''"
              @click="handleSelect(cat)"
            >
              <span class="font-mono text-xs text-secondary-400 dark:text-secondary-500 mr-2">{{ cat.id }}</span>
              {{ cat.name }}
            </button>
            <AppIconButton
              v-if="cat.id"
              type="button"
              variant="danger"
              size="sm"
              :disabled="deletingId === cat.id"
              title="Hapus kategori"
              @click="handleDelete(cat)"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6M1 7h22M9 7V4a1 1 0 011-1h4a1 1 0 011 1v3" />
              </svg>
            </AppIconButton>
          </li>
        </ul>
      </div>
    </div>

    <template #footer>
      <AppButton variant="secondary" @click="emit('close')">Tutup</AppButton>
    </template>
  </AppModal>
</template>

<script setup>
import { ref, watch } from 'vue'
import AppModal from '../common/AppModal.vue'
import AppButton from '../common/AppButton.vue'
import AppInput from '../common/AppInput.vue'
import AppIconButton from '../common/AppIconButton.vue'
import LoadingSpinner from '../common/LoadingSpinner.vue'
import { useNotificationStore } from '../../stores/notification'

const props = defineProps({
  show: { type: Boolean, default: false },
  service: { type: Object, required: true },
  title: { type: String, default: 'Kelola Kategori' },
  allowSelect: { type: Boolean, default: true },
})

const emit = defineEmits(['close', 'selected'])

const notificationStore = useNotificationStore()

const categories = ref([])
const loading = ref(false)
const adding = ref(false)
const deletingId = ref(null)
const newName = ref('')
const error = ref('')

function extractCategories(payload) {
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload?.data)) return payload.data
  if (Array.isArray(payload?.categories)) return payload.categories
  if (Array.isArray(payload?.data?.categories)) return payload.data.categories
  // banner.BannerCategoryListResponse uses plural-prefixed `banner_categories`
  if (Array.isArray(payload?.banner_categories)) return payload.banner_categories
  if (Array.isArray(payload?.data?.banner_categories)) return payload.data.banner_categories
  return []
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await props.service.getCategories({ limit: 100 })
    categories.value = extractCategories(res.data)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat kategori')
    categories.value = []
  } finally {
    loading.value = false
  }
}

async function handleAdd() {
  const name = newName.value.trim()
  if (!name) {
    error.value = 'Nama wajib diisi'
    return
  }
  if (categories.value.some((c) => c.name?.toLowerCase() === name.toLowerCase())) {
    error.value = 'Kategori sudah ada'
    return
  }
  error.value = ''
  adding.value = true
  try {
    const res = await props.service.createCategory({ name })
    const created = res?.data || { name }
    if (!created.id && created.name) {
      created.id = `local-${Date.now()}`
    }
    categories.value.unshift(created)
    newName.value = ''
    notificationStore.success('Kategori ditambahkan')
    emit('selected', { id: created.id, name: created.name })
  } catch (err) {
    if (err.response?.status === 409) {
      error.value = 'Kategori sudah ada'
    } else {
      notificationStore.error(err.response?.data?.error || 'Gagal menambah kategori')
    }
  } finally {
    adding.value = false
  }
}

async function handleDelete(cat) {
  if (!cat.id || String(cat.id).startsWith('local-')) {
    categories.value = categories.value.filter((c) => c !== cat)
    return
  }
  deletingId.value = cat.id
  try {
    await props.service.deleteCategory(cat.id)
    categories.value = categories.value.filter((c) => c.id !== cat.id)
    notificationStore.success('Kategori dihapus')
  } catch (err) {
    if (err.response?.status === 409) {
      notificationStore.error('Kategori masih digunakan dan tidak dapat dihapus')
    } else {
      notificationStore.error(err.response?.data?.error || 'Gagal menghapus kategori')
    }
  } finally {
    deletingId.value = null
  }
}

function handleSelect(cat) {
  if (!props.allowSelect) return
  emit('selected', { id: cat.id, name: cat.name })
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      newName.value = ''
      error.value = ''
      load()
    }
  }
)
</script>