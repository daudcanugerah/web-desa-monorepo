import api from './api'

export const profileService = {
  get: () => api.get('/desa'),
  update: (data) => api.put('/desa', data),
  uploadMedia: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post('/desa/upload-media', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}
