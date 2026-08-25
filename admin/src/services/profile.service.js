import api from './api'

export const profileService = {
  get: () => api.get('/desa'),
  update: (data) => api.put('/desa', data),
}
