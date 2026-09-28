import api from './api'

export const infographicService = {
  list: (params) => api.get('/infographic', { params }),
  get: (id) => api.get(`/infographic/${id}`),
  create: (data) => api.post('/infographic', data),
  update: (id, data) => api.put(`/infographic/${id}`, data),
  delete: (id) => api.delete(`/infographic/${id}`),
  generatePreviewToken: (data) => api.post('/infographic/preview/token', data),
  getSectionNames: () => api.get('/infographic/sections/names'),

  // Categories — admin has no GET for /infographic/categories,
  // so we use the public endpoint for listing (same pattern as
  // news/umkm/fasilitas/ppid in this codebase).
  getCategories: (params) => api.get('/public/infographic/categories', { params }),
  createCategory: (data) => api.post('/infographic/categories', data),
  deleteCategory: (id) => api.delete(`/infographic/categories/${id}`),
}