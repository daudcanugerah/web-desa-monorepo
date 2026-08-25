import api from './api'

export const healthService = {
  ping: () => api.get('/health'),
}