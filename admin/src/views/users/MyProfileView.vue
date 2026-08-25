<template>
  <div class="max-w-6xl mx-auto">
    <!-- Profile View Mode -->
    <div v-if="!editMode" class="space-y-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 mb-6">Profil Saya</h1>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Profile Card (left, 2/3) -->
        <div class="lg:col-span-2 bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-8">
          <div class="flex items-start gap-8">
            <!-- Avatar -->
            <div class="flex-shrink-0">
              <SafeImg
                v-if="authStore.currentUser?.profile_image_url"
                :src="authStore.currentUser.profile_image_url"
                :alt="authStore.currentUser.name"
                circle
                class="w-32 h-32 border-4 border-secondary-200 dark:border-secondary-700"
              />
              <div
                v-else
                class="w-32 h-32 rounded-full bg-blue-600 text-white text-5xl font-semibold flex items-center justify-center border-4 border-secondary-200 dark:border-secondary-700"
              >
                {{ userInitial }}
              </div>
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <h2 class="text-3xl font-semibold text-secondary-800 dark:text-secondary-200 mb-2">{{ authStore.currentUser?.name }}</h2>
              <p class="text-secondary-600 dark:text-secondary-400 text-lg mb-6">{{ authStore.currentUser?.email }}</p>
              <div class="space-y-3">
                <div class="flex items-center gap-3">
                  <span class="text-sm font-medium text-secondary-600 dark:text-secondary-400 w-24">Roles:</span>
                  <div class="flex gap-2 flex-wrap">
                    <AppBadge v-for="role in authStore.currentUser?.roles" :key="role" variant="primary">
                      {{ role }}
                    </AppBadge>
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <span class="text-sm font-medium text-secondary-600 dark:text-secondary-400 w-24">Bergabung:</span>
                  <span class="text-sm text-secondary-700 dark:text-secondary-300">{{ formatLongDate(authStore.currentUser?.created_at) }}</span>
                </div>
              </div>

              <div class="flex gap-3 mt-6">
                <AppButton variant="primary" @click="editMode = true" class="px-6">Edit Profil</AppButton>
                <AppButton variant="secondary" @click="showPasswordModal = true" class="px-6">Ubah Password</AppButton>
              </div>
            </div>
          </div>
        </div>

        <!-- Sidebar (right, 1/3) -->
        <aside class="space-y-4">
          <div class="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6">
            <h3 class="text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-3">Akses Cepat</h3>
            <ul class="space-y-2">
              <li>
                <router-link to="/users" class="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 hover:underline">
                  Kelola User
                </router-link>
              </li>
              <li>
                <router-link to="/roles" class="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 hover:underline">
                  Kelola Role &amp; Izin
                </router-link>
              </li>
              <li>
                <router-link to="/" class="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 hover:underline">
                  Kembali ke Dashboard
                </router-link>
              </li>
            </ul>
          </div>
          <div class="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6">
            <h3 class="text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-2">Tips</h3>
            <p class="text-xs text-secondary-600 dark:text-secondary-400 leading-relaxed">
              Gunakan foto profil dengan rasio 1:1 minimal 256×256 piksel untuk hasil terbaik. Format yang didukung: JPEG, PNG, WebP.
            </p>
          </div>
        </aside>
      </div>
    </div>

    <!-- Profile Edit Mode -->
    <div v-else class="space-y-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Edit Profil</h1>
        <AppIconButton
          variant="ghost"
          size="lg"
          title="Close"
          @click="editMode = false"
        >
          ✕
        </AppIconButton>
      </div>

      <div class="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-8">
        <form @submit.prevent="handleSave">
          <div class="mb-6">
            <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Foto Profil</label>
            <ImageUpload
              :existing-url="authStore.currentUser?.profile_image_url"
              @change="onImageChange"
            />
          </div>

          <div class="mb-6">
            <AppInput
              id="profile-name"
              v-model="form.name"
              label="Nama"
              required
              :error="errors.name"
              placeholder="Nama lengkap"
            />
          </div>

          <div class="mb-6">
            <p class="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Email</p>
            <p class="px-4 py-2 bg-secondary-100 dark:bg-secondary-800 rounded text-sm text-secondary-600 dark:text-secondary-400">
              {{ authStore.currentUser?.email || '—' }}
            </p>
            <p class="text-xs text-secondary-500 dark:text-secondary-400 mt-1">Email tidak dapat diubah. Hubungi administrator jika perlu memperbarui.</p>
          </div>

          <div class="flex gap-3">
            <AppButton type="submit" variant="primary" :loading="saving" loading-label="Menyimpan..." class="px-6">Simpan Perubahan</AppButton>
            <AppButton type="button" variant="secondary" @click="editMode = false" class="px-6">Batal</AppButton>
          </div>
        </form>
      </div>
    </div>

    <!-- Change Password Modal -->
    <AppModal
      :show="showPasswordModal"
      title="Ubah Password"
      size="md"
      @close="showPasswordModal = false"
    >
      <form @submit.prevent="handlePasswordChange">
        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Password Lama</label>
          <input
            v-model="pwForm.old_password"
            type="password"
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Password Baru <span class="text-red-500">*</span></label>
          <input
            v-model="pwForm.new_password"
            type="password"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            :class="pwErrors.new_password ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="pwErrors.new_password" class="text-red-500 text-xs mt-1">{{ pwErrors.new_password }}</p>
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Konfirmasi Password <span class="text-red-500">*</span></label>
          <input
            v-model="pwForm.confirm_password"
            type="password"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            :class="pwErrors.confirm_password ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="pwErrors.confirm_password" class="text-red-500 text-xs mt-1">{{ pwErrors.confirm_password }}</p>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <AppButton variant="secondary" @click="showPasswordModal = false">Batal</AppButton>
          <AppButton variant="primary" :loading="pwSaving" @click="handlePasswordChange">Ubah Password</AppButton>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppIconButton from '../../components/common/AppIconButton.vue'
import AppModal from '../../components/common/AppModal.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import { useAuthStore } from '../../stores/auth'
import { userService } from '../../services/user.service'
import { useNotificationStore } from '../../stores/notification'
import api from '../../services/api'
import { formatLongDate } from '../../utils/dateFormat'

const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const editMode = ref(false)
const showPasswordModal = ref(false)
const form = ref({ name: '', email: '' })
const errors = ref({ name: '' })
const saving = ref(false)
const selectedImage = ref(null)

const pwForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const pwErrors = ref({})
const pwSaving = ref(false)

const userInitial = computed(() => {
  const name = authStore.currentUser?.name ?? 'A'
  return name.charAt(0).toUpperCase()
})

function onImageChange(file) {
  selectedImage.value = file
}

function validateProfile() {
  errors.value = { name: '' }
  let valid = true
  if (!form.value.name.trim()) {
    errors.value.name = 'Nama wajib diisi'
    valid = false
  }
  return valid
}

function validatePassword() {
  const e = {}
  if (!pwForm.value.new_password) {
    e.new_password = 'Password baru wajib diisi'
  } else if (pwForm.value.new_password.length < 8) {
    e.new_password = 'Password minimal 8 karakter'
  }
  if (pwForm.value.confirm_password !== pwForm.value.new_password) {
    e.confirm_password = 'Password tidak cocok'
  }
  pwErrors.value = e
  return Object.keys(e).length === 0
}

async function handleSave() {
  if (!validateProfile()) return
  saving.value = true
  try {
    if (selectedImage.value) {
      const fd = new FormData()
      fd.append('name', form.value.name)
      fd.append('profile_image', selectedImage.value)
      await userService.updateMe(fd)
    } else {
      await userService.updateMe({ name: form.value.name })
    }
    await authStore.fetchCurrentUser()
    notificationStore.success('Profil berhasil disimpan')
    editMode.value = false
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan profil')
  } finally {
    saving.value = false
  }
}

async function handlePasswordChange() {
  if (!validatePassword()) return
  pwSaving.value = true
  try {
    const payload = { new_password: pwForm.value.new_password }
    if (pwForm.value.old_password) payload.old_password = pwForm.value.old_password
    await userService.updatePassword(authStore.currentUser.id, payload)
    notificationStore.success('Password berhasil diubah')
    pwForm.value = { old_password: '', new_password: '', confirm_password: '' }
    showPasswordModal.value = false
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal mengubah password')
  } finally {
    pwSaving.value = false
  }
}

onMounted(async () => {
  if (!authStore.currentUser) {
    try {
      await authStore.fetchCurrentUser()
    } catch {
      // ignore
    }
  }
  form.value.name = authStore.currentUser?.name ?? ''
  form.value.email = authStore.currentUser?.email ?? ''
})
</script>
