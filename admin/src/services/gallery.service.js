import api from './api'

export const galleryService = {
  // Folders
  listFolders: (params) => api.get('/gallery/folders', { params }),
  getFolder: (id, params) => api.get(`/gallery/folders/${id}`, { params }),
  createFolder: (data) => api.post('/gallery/folders', data),
  updateFolder: (id, data) => api.put(`/gallery/folders/${id}`, data),
  deleteFolder: (id) => api.delete(`/gallery/folders/${id}`),
  setFolderVisibility: (id, isPublic) =>
    api.patch(`/gallery/folders/${id}/visibility`, { is_public: isPublic }),
  setFolderCover: (id, coverMediaId) =>
    api.patch(`/gallery/folders/${id}/cover`, { cover_media_id: coverMediaId }),

  // Folder media
  uploadMedia: (id, files) => {
    const fd = new FormData()
    for (const file of files) fd.append('media', file)
    return api.post(`/gallery/folders/${id}/media`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  setBulkMediaVisibility: (id, mediaIds, isPublic) =>
    api.post(`/gallery/folders/${id}/media/visibility`, {
      media_ids: mediaIds,
      is_public: isPublic,
    }),

  // Media (flat)
  listMedia: (params) => api.get('/gallery/media', { params }),
  getMedia: (id) => api.get(`/gallery/media/${id}`),
  deleteMedia: (id) => api.delete(`/gallery/media/${id}`),
  setMediaVisibility: (id, isPublic) =>
    api.patch(`/gallery/media/${id}/visibility`, { is_public: isPublic }),
  regenerateThumbnail: (id) =>
    api.post(`/gallery/media/${id}/regenerate-thumbnail`),
}