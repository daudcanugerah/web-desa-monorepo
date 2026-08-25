<template>
  <div class="min-h-screen flex items-center justify-center bg-secondary-50 dark:bg-secondary-900 px-4 transition-colors">
    <div class="w-full max-w-md bg-white dark:bg-secondary-800 rounded-lg shadow-md p-8 border border-secondary-200 dark:border-secondary-700">
      <h1 class="text-2xl font-bold text-center text-secondary-800 dark:text-secondary-100 mb-6">Reset Password</h1>

      <!-- Token checking -->
      <div v-if="checkingToken" class="flex justify-center py-8">
        <svg
          class="animate-spin h-8 w-8 text-blue-600"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-label="Memvalidasi token..."
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <!-- Token invalid -->
      <div v-else-if="tokenError" class="text-center py-4">
        <svg
          class="mx-auto mb-4 h-12 w-12 text-red-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
          />
        </svg>
        <p class="text-red-600 dark:text-red-400 text-sm mb-4">{{ tokenError }}</p>
        <router-link to="/forgot-password" class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
          Minta link reset baru
        </router-link>
      </div>

      <!-- Success -->
      <div v-else-if="success" class="text-center py-4">
        <svg
          class="mx-auto mb-4 h-12 w-12 text-green-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <p class="text-secondary-700 dark:text-secondary-300 text-sm">
          Password berhasil direset. Mengalihkan ke halaman login...
        </p>
      </div>

      <!-- Reset form -->
      <template v-else>
        <form @submit.prevent="handleSubmit" novalidate>
          <div class="flex flex-col gap-4">
            <AppInput
              id="new_password"
              v-model="newPassword"
              label="Password Baru"
              type="password"
              placeholder="Minimal 8 karakter"
              :required="true"
              :error="errors.newPassword"
            />

            <AppInput
              id="confirm_password"
              v-model="confirmPassword"
              label="Konfirmasi Password"
              type="password"
              placeholder="Ulangi password baru"
              :required="true"
              :error="errors.confirmPassword"
            />

            <p v-if="apiError" class="text-red-500 text-sm">{{ apiError }}</p>

            <AppButton type="submit" variant="primary" :loading="loading" class="w-full mt-2">
              Reset Password
            </AppButton>
          </div>
        </form>
      </template>

      <div v-if="!success" class="mt-6 text-center">
        <router-link to="/login" class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
          Kembali ke Login
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import { authService } from '../../services/auth.service'

const route = useRoute()
const router = useRouter()

const token = ref('')
const checkingToken = ref(true)
const tokenError = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const success = ref(false)
const apiError = ref('')
const errors = ref({ newPassword: '', confirmPassword: '' })

onMounted(async () => {
  const queryToken = route.query.token
  if (!queryToken) {
    router.replace('/login')
    return
  }
  token.value = queryToken

  try {
    const res = await authService.checkResetToken(queryToken)
    if (!res.data?.valid) {
      tokenError.value = 'Link reset password tidak valid atau sudah kedaluwarsa.'
    }
  } catch {
    tokenError.value = 'Link reset password tidak valid atau sudah kedaluwarsa.'
  } finally {
    checkingToken.value = false
  }
})

function validate() {
  errors.value = { newPassword: '', confirmPassword: '' }
  let valid = true

  if (!newPassword.value) {
    errors.value.newPassword = 'Password baru wajib diisi.'
    valid = false
  } else if (newPassword.value.length < 8) {
    errors.value.newPassword = 'Password minimal 8 karakter.'
    valid = false
  }

  if (!confirmPassword.value) {
    errors.value.confirmPassword = 'Konfirmasi password wajib diisi.'
    valid = false
  } else if (newPassword.value !== confirmPassword.value) {
    errors.value.confirmPassword = 'Password tidak cocok.'
    valid = false
  }

  return valid
}

async function handleSubmit() {
  apiError.value = ''
  if (!validate()) return

  loading.value = true
  try {
    await authService.confirmPasswordReset(token.value, newPassword.value)
    success.value = true
    setTimeout(() => router.push('/login'), 2000)
  } catch (err) {
    apiError.value =
      err.response?.data?.error || 'Terjadi kesalahan. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}
</script>
