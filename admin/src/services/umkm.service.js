import api from './api'

export const umkmService = {
  list: (params) => api.get('/umkm', { params }),
  get: (id) => api.get(`/umkm/${id}`),
  create: (data) => api.post('/umkm', data),
  update: (id, data) => api.put(`/umkm/${id}`, data),
  delete: (id) => api.delete(`/umkm/${id}`),

  // Categories
  getCategories: (params) => api.get('/public/umkm/categories', { params }),
  createCategory: (data) => api.post('/umkm/categories', data),
  deleteCategory: (id) => api.delete(`/umkm/categories/${id}`),
}
