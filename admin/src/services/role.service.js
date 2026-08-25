import api from './api'

export const roleService = {
  list: () => api.get('/roles'),
  create: (name) => api.post('/roles', { name }),
  delete: (role) => api.delete(`/roles/${role}`),
  getPermissions: (role) => api.get(`/roles/${role}/permissions`),
  addPermission: (role, resource, action) => api.post(`/roles/${role}/permissions`, { resource, action }),
  removePermission: (role, resource, action) => api.delete(`/roles/${role}/permissions`, { data: { resource, action } }),
  getAvailablePermissions: () => api.get('/permissions'),
}
