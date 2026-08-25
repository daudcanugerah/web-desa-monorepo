import api from './api'

export const facilityService = {
  list: (params) => api.get('/fasilitas', { params }),
  get: (id) => api.get(`/fasilitas/${id}`),
  create: (data) => api.post('/fasilitas', data),
  update: (id, data) => api.put(`/fasilitas/${id}`, data),
  delete: (id) => api.delete(`/fasilitas/${id}`),
  deleteImage: (id, imageIndex) => api.delete(`/fasilitas/${id}/images/${imageIndex}`),

  // Categories
  getCategories: (params) => api.get('/public/fasilitas/categories', { params }),
  createCategory: (data) => api.post('/fasilitas/categories', data),
  deleteCategory: (id) => api.delete(`/fasilitas/categories/${id}`),
}
