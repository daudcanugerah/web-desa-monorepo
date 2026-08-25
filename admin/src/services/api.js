import axios from 'axios'
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '../utils/storage'
import { API_V1_PREFIX } from '../utils/imageUrl'

const RAW_HOST = import.meta.env.VITE_API_BASE_URL

if (!RAW_HOST) {
  console.error(
    '[api] VITE_API_BASE_URL is not defined. Set it in .env (e.g. VITE_API_BASE_URL=http://localhost:8080). API requests will fail until fixed.'
  )
}

const HOST = (RAW_HOST || '').replace(/\/+$/, '')
const API_BASE_URL = `${HOST}${API_V1_PREFIX}`

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
})

// Request interceptor — attach Bearer token from storage
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor — normalize envelope + handle 401 with token refresh + 429 toast
api.interceptors.response.use(
  (response) => {
    // Unwrap a `{data: ...}` envelope if the backend still returns one.
    // Swagger v3 defines flat responses (e.g. `user.UserListResponse` =
    // `{users, pagination}`), but this normalizer keeps the frontend
    // compatible with both shapes — wrapped and flat. No swagger response
    // type uses a top-level `data` field, so the check is safe.
    const body = response.data
    if (
      body &&
      typeof body === 'object' &&
      !Array.isArray(body) &&
      'data' in body &&
      body.data !== null &&
      typeof body.data === 'object'
    ) {
      response.data = body.data
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config
    const status = error.response?.status

    if (status === 429 && !originalRequest._rateLimitNotified) {
      originalRequest._rateLimitNotified = true
      try {
        const { useNotificationStore } = await import('../stores/notification')
        const notificationStore = useNotificationStore()
        notificationStore.error('Terlalu banyak permintaan. Silakan coba lagi nanti.')
      } catch {
        // Pinia not ready yet — fail silently; caller can handle the error
      }
      setTimeout(() => {
        originalRequest._rateLimitNotified = false
      }, 5000)
    }

    if (status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = getRefreshToken()
        const { data } = await axios.post(
          `${API_BASE_URL}/auth/refresh`,
          { refresh_token: refreshToken },
        )

        // Raw axios call bypasses the response interceptor, so apply the
        // same envelope-unwrapping here.
        const unwrapped =
          data && typeof data === 'object' && 'data' in data && data.data && typeof data.data === 'object'
            ? data.data
            : data

        // Update storage directly to avoid circular dep with auth store
        setTokens(unwrapped)

        // Also sync the Pinia store if it's already initialised
        try {
          const { useAuthStore } = await import('../stores/auth')
          const authStore = useAuthStore()
          authStore.setTokens(unwrapped)
        } catch {
          // Pinia may not be initialised yet — storage update above is sufficient
        }

        originalRequest.headers.Authorization = `Bearer ${unwrapped.access_token}`
        return api(originalRequest)
      } catch {
        clearTokens()

        // Sync Pinia store if available
        try {
          const { useAuthStore } = await import('../stores/auth')
          const authStore = useAuthStore()
          authStore.clearTokens()
        } catch {
          // ignore
        }

        const { default: router } = await import('../router')
        router.push('/login')
      }
    }

    return Promise.reject(error)
  },
)

export default api
export { API_BASE_URL }