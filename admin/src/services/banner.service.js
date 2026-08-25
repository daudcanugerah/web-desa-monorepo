import api from './api'

export const bannerService = {
  list: (params) => api.get('/banners', { params }),
  get: (id) => api.get(`/banners/${id}`),
  getActive: () => api.get('/banners/active'),
  create: (formData) => api.post('/banners', formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  update: (id, formData) => api.put(`/banners/${id}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  delete: (id) => api.delete(`/banners/${id}`),
  updateStatus: (id, status) => api.patch(`/banners/${id}/status`, { status }),

  // Categories
  getCategories: (params) => api.get('/banners/categories', { params }),
  createCategory: (data) => api.post('/banners/categories', data),
  deleteCategory: (id) => api.delete(`/banners/categories/${id}`),
}
