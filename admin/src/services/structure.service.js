import api from './api'

export const structureService = {
  list: (params) => api.get('/struktur', { params }),
  get: (id) => api.get(`/struktur/${id}`),
  create: (formData) => api.post('/struktur', formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  update: (id, formData) => api.put(`/struktur/${id}`, formData, { headers: { 'Content-Type': 'multipart/form-data' } }),
  delete: (id) => api.delete(`/struktur/${id}`),
}
