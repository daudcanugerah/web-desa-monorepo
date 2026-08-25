<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">{{ isEdit ? 'Edit Informasi Detail' : 'Tambah Informasi Detail' }}</h1>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Nama Bagian <span class="text-red-500">*</span></label>
          <input
            v-model="form.section_name"
            type="text"
            placeholder="Contoh: keuangan, lokasi-desa, potensi"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            :class="errors.section_name ? 'border-danger-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="errors.section_name" class="text-red-500 dark:text-red-400 text-xs mt-1">{{ errors.section_name }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Endpoint <span class="text-red-500">*</span></label>
          <input
            :value="form.section_endpoint"
            type="text"
            disabled
            class="w-full px-3 py-2 border rounded text-sm bg-secondary-100 dark:bg-secondary-800 text-secondary-600 dark:text-secondary-400 cursor-not-allowed focus:outline-none"
            :class="errors.section_endpoint ? 'border-danger-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p class="text-xs text-secondary-500 dark:text-secondary-400 mt-1">Endpoint otomatis di-generate</p>
        </div>
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Status</label>
        <div class="flex items-center">
          <input
            v-model="form.state"
            type="checkbox"
            class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-secondary-300 dark:border-secondary-600 rounded"
          />
          <label class="ml-2 text-sm text-secondary-700 dark:text-secondary-300">Aktif</label>
        </div>
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Konten <span class="text-red-500">*</span></label>
        <RichTextEditor
          v-model="form.content"
          :class="errors.content ? 'border-red-500' : ''"
        />
        <p v-if="errors.content" class="text-red-500 text-xs mt-1">{{ errors.content }}</p>
      </div>

      <div class="flex gap-3 pt-4">
        <AppButton
          type="button"
          variant="primary"
          :loading="saving"
          @click="handleSubmit"
        >
          {{ isEdit ? 'Simpan Perubahan' : 'Tambah Informasi Detail' }}
        </AppButton>
        <AppButton
          type="button"
          variant="secondary"
          @click="$router.back()"
        >
          Batal
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppButton from '../../components/common/AppButton.vue'
import AppBackButton from '../../components/common/AppBackButton.vue'
import RichTextEditor from '../../components/forms/RichTextEditor.vue'
import { profileSectionService } from '../../services/profile-section.service'
import { useNotificationStore } from '../../stores/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const form = ref({
  section_name: '',
  section_endpoint: '',
  content: '',
  state: true,
})
const errors = ref({})
const saving = ref(false)

function generateEndpoint() {
  // Generate random short string (8 chars)
  const randomStr = Math.random().toString(36).substring(2, 10)
  form.value.section_endpoint = `/profile/page-${randomStr}`
}

function validate() {
  errors.value = {}
  
  if (!form.value.section_name.trim()) {
    errors.value.section_name = 'Nama bagian wajib diisi'
  }
  
  if (!form.value.content.trim()) {
    errors.value.content = 'Konten wajib diisi'
  }
  
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  
  saving.value = true
  try {
    const data = {
      section_name: form.value.section_name,
      section_endpoint: form.value.section_endpoint,
      content: form.value.content,
      state: form.value.state,
    }
    
    if (isEdit.value) {
      await profileSectionService.update(route.params.id, data)
      notificationStore.success('Informasi detail berhasil diperbarui')
    } else {
      await profileSectionService.create(data)
      notificationStore.success('Informasi detail berhasil ditambahkan')
    }
    
    router.push('/profile/sections')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan informasi detail')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  if (isEdit.value) {
    try {
      const res = await profileSectionService.get(route.params.id)
      const data = res.data
      form.value = {
        section_name: data.section_name,
        section_endpoint: data.section_endpoint,
        content: data.content,
        state: data.state,
      }
    } catch (err) {
      notificationStore.error('Gagal memuat data informasi detail')
      router.push('/profile/sections')
    }
  } else {
    // Generate endpoint for new section
    generateEndpoint()
  }
})
</script>