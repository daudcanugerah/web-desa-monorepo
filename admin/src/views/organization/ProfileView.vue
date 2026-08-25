<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Profil Desa</h1>
      <AppButton v-if="!editing" variant="primary" @click="startEdit">Edit</AppButton>
    </div>

    <LoadingSpinner v-if="loading" class="py-12" />

    <div v-else-if="!editing" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Nama Desa</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.name || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Alamat</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.address || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Telepon</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.phone || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Email</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.email || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Website</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.website || '—' }}</p>
        </div>
      </div>
      <div>
        <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Visi &amp; Misi</p>
        <p class="text-secondary-800 dark:text-secondary-200 whitespace-pre-line">{{ profile.vision_mission || '—' }}</p>
      </div>
    </div>

    <form v-else class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5" @submit.prevent="handleSave">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <AppInput
          v-model="form.name"
          label="Nama Desa"
          required
          :error="formErrors.name"
          placeholder="Nama desa"
        />
        <AppInput
          v-model="form.address"
          label="Alamat"
          placeholder="Alamat desa"
        />
        <AppInput
          v-model="form.phone"
          label="Telepon"
          placeholder="08xxxxxxxxxx"
        />
        <AppInput
          v-model="form.email"
          label="Email"
          type="email"
          placeholder="email@desa.go.id"
        />
        <AppInput
          v-model="form.website"
          label="Website"
          placeholder="https://desa.go.id"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Visi &amp; Misi</label>
        <textarea
          v-model="form.vision_mission"
          rows="5"
          placeholder="Tuliskan visi dan misi desa..."
          class="w-full border border-secondary-300 dark:border-secondary-600 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
        />
      </div>
      <div class="flex justify-end gap-3 pt-2">
        <AppButton variant="secondary" type="button" @click="cancelEdit">Batal</AppButton>
        <AppButton variant="primary" type="submit" :loading="saving">Simpan</AppButton>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AppButton from '../../components/common/AppButton.vue'
import AppInput from '../../components/common/AppInput.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import { profileService } from '../../services/profile.service'
import { useNotificationStore } from '../../stores/notification'

const notificationStore = useNotificationStore()

const loading = ref(false)
const saving = ref(false)
const editing = ref(false)
const profile = ref({})
const form = ref({})
const formErrors = ref({})

async function fetchProfile() {
  loading.value = true
  try {
    const res = await profileService.get()
    profile.value = res.data
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat profil desa')
  } finally {
    loading.value = false
  }
}

function startEdit() {
  form.value = {
    name: profile.value.name || '',
    address: profile.value.address || '',
    phone: profile.value.phone || '',
    email: profile.value.email || '',
    website: profile.value.website || '',
    vision_mission: profile.value.vision_mission || '',
  }
  formErrors.value = {}
  editing.value = true
}

function cancelEdit() {
  editing.value = false
  formErrors.value = {}
}

function validate() {
  formErrors.value = {}
  if (!form.value.name?.trim()) {
    formErrors.value.name = 'Nama desa wajib diisi'
  }
  return Object.keys(formErrors.value).length === 0
}

async function handleSave() {
  if (!validate()) return
  saving.value = true
  try {
    const res = await profileService.update(form.value)
    profile.value = res.data
    notificationStore.success('Profil desa berhasil diperbarui')
    editing.value = false
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan profil desa')
  } finally {
    saving.value = false
  }
}

onMounted(fetchProfile)
</script>
