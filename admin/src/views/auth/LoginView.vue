<template>
  <div class="min-h-screen flex items-center justify-center bg-secondary-50 dark:bg-secondary-900 px-4 transition-colors">
    <div class="w-full max-w-md bg-white dark:bg-secondary-800 rounded-lg shadow-md p-8 border border-secondary-200 dark:border-secondary-700">
      <h1 class="text-2xl font-bold text-center text-secondary-800 dark:text-secondary-100 mb-6">Desa Admin</h1>

      <form @submit.prevent="handleSubmit" novalidate>
        <div class="flex flex-col gap-4">
          <AppInput
            id="email"
            v-model="email"
            :label="$t('auth.email')"
            type="email"
            :placeholder="$t('auth.email')"
            :required="true"
            :error="errors.email"
          />

          <AppInput
            id="password"
            v-model="password"
            :label="$t('auth.password')"
            type="password"
            :placeholder="$t('auth.password')"
            :required="true"
            :error="errors.password"
          />

          <p v-if="apiError" class="text-red-500 text-sm text-center">{{ apiError }}</p>

          <AppButton type="submit" variant="primary" :loading="loading" loading-label="Memproses..." class="w-full mt-2">
            {{ $t('auth.login') }}
          </AppButton>
        </div>
      </form>

      <div class="mt-4 text-center">
        <router-link to="/forgot-password" class="text-sm text-blue-600 dark:text-blue-400 hover:underline">
          {{ $t('auth.forgotPassword') }}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import { authService } from '../../services/auth.service'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const apiError = ref('')
const errors = ref({ email: '', password: '' })

function validate() {
  errors.value = { email: '', password: '' }
  let valid = true

  if (!email.value.trim()) {
    errors.value.email = t('validation.required')
    valid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    errors.value.email = t('validation.email')
    valid = false
  }

  if (!password.value) {
    errors.value.password = t('validation.required')
    valid = false
  }

  return valid
}

async function handleSubmit() {
  apiError.value = ''
  if (!validate()) return

  loading.value = true
  try {
    const response = await authService.login(email.value.trim(), password.value)
    authStore.setTokens(response.data)
    router.push('/')
  } catch (err) {
    if (err.response?.status === 429) {
      apiError.value = t('auth.rateLimitExceeded')
    } else {
      apiError.value = err.response?.data?.error || t('auth.loginFailed')
    }
  } finally {
    loading.value = false
  }
}
</script>
