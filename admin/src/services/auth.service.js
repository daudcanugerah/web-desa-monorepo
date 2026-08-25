import api from './api'

export const authService = {
  login: (email, password) => api.post('/auth/login', { email, password }),
  refresh: (refresh_token) => api.post('/auth/refresh', { refresh_token }),
  requestPasswordReset: (email) => api.post('/auth/password-reset/request', { email }),
  checkResetToken: (token) => api.get('/auth/password-reset/check', { params: { token } }),
  confirmPasswordReset: (token, new_password) =>
    api.post('/auth/password-reset/confirm', { token, new_password }),
}
