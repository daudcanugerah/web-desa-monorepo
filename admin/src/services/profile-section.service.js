import api from './api'

export const profileSectionService = {
  list: (params) => api.get('/profile', { params }),
  get: (id) => api.get(`/profile/${id}`),
  create: (data) => api.post('/profile', data),
  update: (id, data) => api.put(`/profile/${id}`, data),
  delete: (id) => api.delete(`/profile/${id}`),
  getSectionNames: () => api.get('/profile/sections/names'),
}