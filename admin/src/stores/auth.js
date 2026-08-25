import { defineStore } from 'pinia'
import { getAccessToken, getRefreshToken, setTokens as storageSetTokens, clearTokens as storageClearTokens } from '../utils/storage'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: getAccessToken(),
    refreshToken: getRefreshToken(),
    expiresAt: localStorage.getItem('expires_at'),
    currentUser: null,
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
  },

  actions: {
    setTokens(data) {
      storageSetTokens(data)
      this.accessToken = data.access_token
      if (data.refresh_token) this.refreshToken = data.refresh_token
      if (data.expires_at) this.expiresAt = data.expires_at
    },

    clearTokens() {
      storageClearTokens()
      this.accessToken = null
      this.refreshToken = null
      this.expiresAt = null
      this.currentUser = null
    },

    async fetchCurrentUser() {
      const { default: api } = await import('../services/api')
      const response = await api.get('/users/me')
      this.currentUser = response.data
    },
  },
})
