import api from './api'

export const userService = {
  list: (page = 1, limit = 10, search = '') =>
    api.get('/users', { params: { page, limit, ...(search ? { q: search } : {}) } }),
  get: (id) => api.get(`/users/${id}`),
  getMe: () => api.get('/users/me'),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  updateMe: (payload) => {
    if (payload instanceof FormData) {
      return api.put('/users/me', payload, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
    }
    return api.put('/users/me', payload)
  },
  delete: (id) => api.delete(`/users/${id}`),
  updatePassword: (id, data) => api.put(`/users/${id}/password`, data),
  assignRole: (id, role) => api.post(`/users/${id}/roles`, { role }),
  removeRole: (id, role) => api.delete(`/users/${id}/roles/${role}`),
}
