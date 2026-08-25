<template>
  <div class="min-h-screen flex items-center justify-center bg-secondary-50 dark:bg-secondary-900 px-4 transition-colors">
    <div class="w-full max-w-md bg-white dark:bg-secondary-800 rounded-lg shadow-md p-8 border border-secondary-200 dark:border-secondary-700">
      <h1 class="text-2xl font-bold text-center text-secondary-800 dark:text-secondary-100 mb-2">Lupa Password</h1>
      <p class="text-sm text-secondary-500 dark:text-secondary-400 text-center mb-6">
        Masukkan email Anda untuk menerima link reset password.
      </p>

      <template v-if="!submitted">
        <form @submit.prevent="handleSubmit" novalidate>
          <div class="flex flex-col gap-4">
            <AppInput
              id="email"
              v-model="email"
              label="Email"
              type="email"
              placeholder="admin@example.com"
              :required="true"
              :error="errors.email"
            />

            <p v-if="apiError" class="text-red-500 text-sm">{{ apiError }}</p>

            <AppButton type="submit" variant="primary" :loading="loading" loading-label="Mengirim..." class="w-full mt-2">
              Kirim Link Reset
            </AppButton>
          </div>
        </form>
      </template>

      <template v-else>
        <div class="text-center py-4">
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
            Jika email terdaftar, link reset password telah dikirim.
          </p>
        </div>
      </template>

      <div class="mt-6 text-center">
        <router-link to="/login" class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
          Kembali ke Login
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import { authService } from '../../services/auth.service'

const email = ref('')
const loading = ref(false)
const submitted = ref(false)
const apiError = ref('')
const errors = ref({ email: '' })

function validate() {
  errors.value = { email: '' }
  if (!email.value.trim()) {
    errors.value.email = 'Email wajib diisi.'
    return false
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    errors.value.email = 'Format email tidak valid.'
    return false
  }
  return true
}

async function handleSubmit() {
  apiError.value = ''
  if (!validate()) return

  loading.value = true
  try {
    await authService.requestPasswordReset(email.value.trim())
    submitted.value = true
  } catch (err) {
    if (err.response?.status === 429) {
      apiError.value = 'Terlalu banyak percobaan. Silakan coba lagi nanti.'
    } else {
      apiError.value =
        err.response?.data?.error || 'Terjadi kesalahan. Silakan coba lagi.'
    }
  } finally {
    loading.value = false
  }
}
</script>
