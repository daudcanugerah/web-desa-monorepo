import api from './api'

export const newsService = {
  list: (params) => api.get('/berita', { params }),
  get: (id) => api.get(`/berita/${id}`),
  create: (formData) => api.post('/berita', formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  update: (id, formData) => api.put(`/berita/${id}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  updateStatus: (id, status) => api.patch(`/berita/${id}/status`, { status }),
  delete: (id) => api.delete(`/berita/${id}`),
  uploadMedia: (file, type) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('type', type)
    return api.post('/berita/upload-media', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
  },

  // Categories
  getCategories: (params) => api.get('/public/berita/categories', { params }),
  createCategory: (data) => api.post('/berita/categories', data),
  deleteCategory: (id) => api.delete(`/berita/categories/${id}`),
}
