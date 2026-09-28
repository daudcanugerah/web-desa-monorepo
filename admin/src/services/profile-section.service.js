import api from './api'

export const profileSectionService = {
  list: (params) => api.get('/profile', { params }),
  get: (id) => api.get(`/profile/${id}`),
  create: (data) => api.post('/profile', data),
  update: (id, data) => api.put(`/profile/${id}`, data),
  delete: (id) => api.delete(`/profile/${id}`),
  getSectionNames: () => api.get('/profile/sections/names'),

  // Categories (grouping). Admin has no GET for /profile/categories, so list
  // via the public endpoint (same pattern as other features in this codebase).
  getCategories: (params) => api.get('/public/profile/categories', { params }),
  createCategory: (data) => api.post('/profile/categories', data),
  deleteCategory: (id) => api.delete(`/profile/categories/${id}`),
}