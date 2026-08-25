<template>
  <div class="max-w-3xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">
        {{ isEdit ? 'Edit Folder Galeri' : 'Tambah Folder Galeri' }}
      </h1>
    </div>

    <form
      class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5"
      @submit.prevent="handleSubmit"
    >
      <AppInput
        v-model="form.name"
        label="Nama Folder"
        :error="errors.name"
        placeholder="Contoh: Panen Raya 2026"
        required
        :maxlength="255"
      />

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1.5">
          Deskripsi
        </label>
        <textarea
          v-model="form.description"
          rows="4"
          maxlength="1000"
          class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg bg-white dark:bg-secondary-900 text-secondary-800 dark:text-secondary-200 focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          placeholder="Deskripsi singkat (opsional)"
        />
        <p v-if="errors.description" class="mt-1 text-xs text-red-600">{{ errors.description }}</p>
        <p v-else class="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
          Maksimal 1000 karakter
        </p>
      </div>

      <div class="flex items-center gap-3 pt-2">
        <input
          id="isPublic"
          v-model="form.is_public"
          type="checkbox"
          class="w-4 h-4 rounded border-secondary-300 dark:border-secondary-600 text-primary-600 focus:ring-primary-500"
        />
        <label for="isPublic" class="text-sm font-medium text-secondary-700 dark:text-secondary-300 select-none">
          Publik (dapat dilihat oleh semua pengguna)
        </label>
      </div>

      <div class="flex gap-3 pt-2 border-t border-secondary-200 dark:border-secondary-700">
        <AppButton type="submit" variant="primary" :loading="saving" :loading-label="isEdit ? 'Menyimpan...' : 'Membuat...'">
          {{ isEdit ? 'Simpan Perubahan' : 'Buat Folder' }}
        </AppButton>
        <AppButton type="button" variant="secondary" @click="$router.back()">Batal</AppButton>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppInput from '../../../components/common/AppInput.vue'
import AppButton from '../../../components/common/AppButton.vue'
import AppBackButton from '../../../components/common/AppBackButton.vue'
import { galleryService } from '../../../services/gallery.service'
import { useNotificationStore } from '../../../stores/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)

const form = ref({ name: '', description: '', is_public: false })
const errors = ref({})
const saving = ref(false)

function validate() {
  errors.value = {}
  const name = form.value.name.trim()
  if (!name) errors.value.name = 'Nama folder wajib diisi'
  else if (name.length > 255) errors.value.name = 'Maksimal 255 karakter'
  if (form.value.description && form.value.description.length > 1000) {
    errors.value.description = 'Maksimal 1000 karakter'
  }
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    if (isEdit.value) {
      // UpdateFolderRequest only accepts name + description; is_public is
      // changed via PATCH /gallery/folders/{id}/visibility.
      await galleryService.updateFolder(route.params.id, {
        name: form.value.name.trim(),
        description: form.value.description?.trim() || undefined,
      })
      notificationStore.success('Folder berhasil diperbarui')
    } else {
      await galleryService.createFolder({
        name: form.value.name.trim(),
        description: form.value.description?.trim() || undefined,
        is_public: form.value.is_public,
      })
      notificationStore.success('Folder berhasil dibuat')
    }
    router.push('/gallery')
  } catch (err) {
    const msg = err.response?.data?.error || 'Gagal menyimpan folder'
    if (err.response?.status === 409) {
      errors.value.name = 'Nama folder sudah digunakan'
    } else {
      notificationStore.error(msg)
    }
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  if (!isEdit.value) return
  try {
    const res = await galleryService.getFolder(route.params.id)
    const f = res.data.folder
    form.value.name = f.name
    form.value.description = f.description || ''
    form.value.is_public = !!f.is_public
  } catch (err) {
    notificationStore.error('Gagal memuat folder')
    router.push('/gallery')
  }
})
</script>