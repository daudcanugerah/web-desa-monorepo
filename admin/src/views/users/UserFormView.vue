<template>
  <div class="max-w-3xl mx-auto">
    <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 mb-6">{{ isEdit ? 'Edit User' : 'Tambah User' }}</h1>

    <!-- Main form -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6 mb-6">
      <form @submit.prevent="handleSubmit">
        <!-- Profile image -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Foto Profil</label>
          <ImageUpload
            :existing-url="existingImageUrl"
            @change="onImageChange"
          />
        </div>

        <div class="mb-4">
          <AppInput
            id="name"
            v-model="form.name"
            label="Name"
            required
            :error="errors.name"
            placeholder="Nama lengkap"
            @blur="touch('name')"
          />
        </div>

        <div class="mb-4">
          <AppInput
            id="email"
            v-model="form.email"
            label="Email"
            type="email"
            required
            :error="errors.email"
            placeholder="email@contoh.com"
            @blur="touch('email')"
          />
        </div>

        <!-- Password only in create mode -->
        <div v-if="!isEdit" class="mb-4">
          <AppInput
            id="password"
            v-model="form.password"
            label="Password"
            type="password"
            required
            :error="errors.password"
            placeholder="Minimal 8 karakter"
            hint="Minimal 8 karakter"
            @blur="touch('password')"
          />
        </div>

        <!-- Role assignment -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Role</label>
          <div v-if="rolesLoading" class="text-xs text-secondary-400 dark:text-secondary-500 flex items-center gap-2">
            <LoadingSpinner size="sm" /> Memuat roles...
          </div>
          <div v-else-if="availableRoles.length === 0" class="text-xs text-secondary-400 dark:text-secondary-500">
            Belum ada role tersedia. Buat role di halaman Roles terlebih dahulu.
          </div>
          <div v-else class="space-y-2">
            <label
              v-for="role in availableRoles"
              :key="role.name"
              class="flex items-center gap-2 p-2 rounded hover:bg-secondary-50 dark:hover:bg-secondary-700 cursor-pointer"
            >
              <input
                type="radio"
                :value="role.name"
                v-model="selectedRoles"
                class="border-secondary-300 dark:border-secondary-600 text-blue-600 focus:ring-blue-500"
              />
              <span class="text-sm text-secondary-700 dark:text-secondary-300">{{ role.name }}</span>
            </label>
          </div>
        </div>

        <div class="flex gap-3 mt-6">
          <AppButton type="submit" variant="primary" :loading="submitting" :loading-label="isEdit ? 'Memperbarui...' : 'Membuat...'">
            {{ isEdit ? 'Perbarui User' : 'Buat User' }}
            <span v-if="saveHint" class="ml-2 text-xs opacity-70 hidden sm:inline">({{ saveHint }})</span>
          </AppButton>
          <AppButton variant="secondary" @click="router.push('/users')">Batal</AppButton>
        </div>
      </form>
    </div>

    <!-- Password change section (edit mode only) -->
    <div v-if="isEdit" class="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6">
      <h2 class="text-lg font-semibold text-secondary-700 dark:text-secondary-300 mb-4">Ubah Password</h2>
      <form @submit.prevent="handlePasswordChange">
        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Old Password</label>
          <input
            v-model="pwForm.old_password"
            type="password"
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
          />
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">New Password <span class="text-red-500">*</span></label>
          <input
            v-model="pwForm.new_password"
            type="password"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            :class="pwErrors.new_password ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="pwErrors.new_password" class="text-red-500 dark:text-red-400 text-xs mt-1">{{ pwErrors.new_password }}</p>
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Confirm Password <span class="text-red-500">*</span></label>
          <input
            v-model="pwForm.confirm_password"
            type="password"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            :class="pwErrors.confirm_password ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="pwErrors.confirm_password" class="text-red-500 text-xs mt-1">{{ pwErrors.confirm_password }}</p>
        </div>

        <AppButton type="submit" variant="primary" :loading="pwSubmitting">Ubah Password</AppButton>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppButton from '../../components/common/AppButton.vue'
import AppInput from '../../components/common/AppInput.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import { useForm, rules } from '../../composables/useForm'
import { useSaveShortcut } from '../../composables/useSaveShortcut'
import { useUnsavedChanges } from '../../composables/useUnsavedChanges'
import { userService } from '../../services/user.service'
import { roleService } from '../../services/role.service'
import { useNotificationStore } from '../../stores/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)

const { values: form, errors, touch, validateAll } = useForm(
  { name: '', email: '', password: '' },
  {
    name: [rules.required('Nama')],
    email: [rules.required('Email'), rules.email],
    password: isEdit.value ? [] : [rules.required('Password'), rules.minLength(8, 'Password')],
  },
)

const submitting = ref(false)

const selectedImage = ref(null)
const existingImageUrl = ref(null)

const availableRoles = ref([])
const selectedRoles = ref('')
const originalRoles = ref([])
const rolesLoading = ref(false)

useUnsavedChanges(form)
const saveHint = useSaveShortcut(() => {
  if (!submitting.value) handleSubmit()
}, { enabled: true })

const pwForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const pwErrors = ref({})
const pwSubmitting = ref(false)

function onImageChange(file) {
  selectedImage.value = file
}

async function loadRoles() {
  rolesLoading.value = true
  try {
    const res = await roleService.list()
    const rolesData = res.data
    availableRoles.value = Array.isArray(rolesData) ? rolesData : (rolesData.roles || [])
    availableRoles.value = availableRoles.value.map(role =>
      typeof role === 'string' ? { name: role } : role
    )
  } catch {
    notificationStore.error('Gagal memuat daftar role')
  } finally {
    rolesLoading.value = false
  }
}

async function syncRoles(userId) {
  const toAdd = selectedRoles.value ? [selectedRoles.value] : []
  const toRemove = originalRoles.value.filter(r => r !== selectedRoles.value)
  await Promise.all([
    ...toAdd.map(r => userService.assignRole(userId, r)),
    ...toRemove.map(r => userService.removeRole(userId, r)),
  ])
}

async function handleSubmit() {
  if (!validateAll()) return
  submitting.value = true
  try {
    let userId
    if (isEdit.value) {
      userId = route.params.id
      let payload
      if (selectedImage.value) {
        payload = new FormData()
        payload.append('name', form.name)
        payload.append('email', form.email)
        payload.append('profile_image', selectedImage.value)
        await userService.update(userId, payload)
      } else {
        await userService.update(userId, { name: form.name, email: form.email })
      }
      await syncRoles(userId)
      notificationStore.success('User berhasil diperbarui')
    } else {
      const res = await userService.create({ name: form.name, email: form.email, password: form.password })
      userId = res.data.id || res.data.user?.id
      if (selectedImage.value && userId) {
        const fd = new FormData()
        fd.append('name', form.name)
        fd.append('email', form.email)
        fd.append('password', form.password)
        fd.append('profile_image', selectedImage.value)
        await userService.update(userId, fd)
      }
      if (userId) {
        originalRoles.value = []
        await syncRoles(userId)
      }
      notificationStore.success('User berhasil dibuat')
    }
    router.push('/users')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan user')
  } finally {
    submitting.value = false
  }
}

function validatePasswordForm() {
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

async function handlePasswordChange() {
  if (!validatePasswordForm()) return
  pwSubmitting.value = true
  try {
    const payload = { new_password: pwForm.value.new_password }
    if (pwForm.value.old_password) payload.old_password = pwForm.value.old_password
    await userService.updatePassword(route.params.id, payload)
    notificationStore.success('Password berhasil diubah')
    pwForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal mengubah password')
  } finally {
    pwSubmitting.value = false
  }
}

watch(isEdit, (edit) => {
  if (edit) {
    form.password = ''
  }
}, { immediate: true })

onMounted(async () => {
  await loadRoles()
  if (isEdit.value) {
    try {
      const res = await userService.get(route.params.id)
      const user = res.data
      form.name = user.name || ''
      form.email = user.email || ''
      existingImageUrl.value = user.profile_image_url || null
      originalRoles.value = Array.isArray(user.roles) ? [...user.roles] : []
      selectedRoles.value = originalRoles.value.length > 0 ? originalRoles.value[0] : ''
    } catch (err) {
      notificationStore.error(err.response?.data?.error || 'Gagal memuat data user')
      router.push('/users')
    }
  }
})
</script>
