import api from './api'

export const ppidService = {
  // Admin endpoints
  adminList: (params) => api.get('/ppid', { params }),
  adminGet: (id) => api.get(`/ppid/${id}`),
  create: (payload) =>
    api.post('/ppid', payload, {
      headers: payload instanceof FormData
        ? { 'Content-Type': 'multipart/form-data' }
        : { 'Content-Type': 'application/json' },
    }),
  update: (id, payload) => api.put(`/ppid/${id}`, payload),
  delete: (id) => api.delete(`/ppid/${id}`),
  getCategories: () => api.get('/public/ppid/categories'),
  createCategory: (data) => api.post('/ppid/categories', data),
  deleteCategory: (id) => api.delete(`/ppid/categories/${id}`),

  // Combined media upload (document + thumbnail in one multipart request)
  upload: (documentFile, thumbnailFile) => {
    const fd = new FormData()
    if (documentFile) fd.append('document', documentFile)
    if (thumbnailFile) fd.append('thumbnail', thumbnailFile)
    return api.post('/ppid/upload', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // Request management
  listRequests: (params) => api.get('/ppid/requests', { params }),
  approveRequest: (id) => api.post(`/ppid/requests/${id}/approve`),
  revokeRequest: (id) => api.post(`/ppid/requests/${id}/revoke`),
}