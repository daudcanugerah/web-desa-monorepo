import api from './api'

export const umkmService = {
  list: (params) => api.get('/umkm', { params }),
  get: (id) => api.get(`/umkm/${id}`),
  create: (data) => api.post('/umkm', data),
  update: (id, data) => api.put(`/umkm/${id}`, data),
  delete: (id) => api.delete(`/umkm/${id}`),

  // Pre-upload one image, returning { media_id, url }. Attach the media_id
  // via `images_media_ids` on create/update.
  uploadMedia: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/umkm/upload-media', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // Categories
  getCategories: (params) => api.get('/public/umkm/categories', { params }),
  createCategory: (data) => api.post('/umkm/categories', data),
  deleteCategory: (id) => api.delete(`/umkm/categories/${id}`),
}
