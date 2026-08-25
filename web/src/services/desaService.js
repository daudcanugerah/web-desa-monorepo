import { apiClient, API_BASE_URL } from './apiClient.js'

const unwrapData = (response) => {
  if (response == null) return null
  return response.data !== undefined ? response.data : response
}

const unwrapList = (response, key) => {
  const data = unwrapData(response)
  if (data == null) return []
  if (Array.isArray(data)) return data
  if (Array.isArray(data[key])) return data[key]
  if (Array.isArray(data?.data?.[key])) return data.data[key]
  return []
}

const unwrapPaginated = (response, key) => {
  const data = unwrapData(response) || {}
  return {
    items: Array.isArray(data[key]) ? data[key] : [],
    pagination: data.pagination || null
  }
}

const normalize = (item) => {
  if (!item || typeof item !== 'object') return item
  const categoryName = typeof item.category === 'object' && item.category !== null
    ? item.category.name
    : item.category
  return { ...item, category: categoryName }
}

const normalizeList = (items) => Array.isArray(items) ? items.map(normalize) : []

export const mediaUrl = (item, variant = 'single') => {
  if (!item) return ''
  const m = item.media
  if (variant === 'array') {
    if (Array.isArray(m) && m.length > 0) return m[0]?.url || m[0]?.thumbnail_url || ''
    if (Array.isArray(item.images) && item.images.length > 0) return item.images[0]
    return ''
  }
  if (m && !Array.isArray(m)) return m.url || m.thumbnail_url || ''
  return item.image_url || item.profile_image_url || ''
}

export const thumbnailUrl = (item) => {
  if (!item) return ''
  const t = item.thumbnail
  if (t && typeof t === 'object') return t.url || t.thumbnail_url || ''
  return item.thumbnail_url || ''
}

export const documentUrl = (item) => {
  if (!item) return ''
  const d = item.document
  if (d && typeof d === 'object') return d.url || ''
  return item.document_url || ''
}

export const decodeJwtExp = (token) => {
  if (!token || typeof token !== 'string') return null
  const parts = token.split('.')
  if (parts.length !== 3) return null
  try {
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = payload + '==='.slice((payload.length + 3) % 4)
    const decoded = atob(padded)
    const json = JSON.parse(decoded)
    return typeof json.exp === 'number' ? json.exp * 1000 : null
  } catch {
    return null
  }
}

const handleError = (scope, err, fallback) => {
  console.error(`[${scope}]`, err)
  return fallback
}

const buildFileUrl = (filename) => {
  if (!filename) return ''
  if (/^https?:\/\//i.test(filename) || filename.startsWith('/')) return filename
  return `${API_BASE_URL.replace(/\/+$/, '')}/files/${filename}`
}

// ============ Banners ============

export const getActiveBanners = async () => {
  try {
    const data = unwrapData(await apiClient.get('/banners/active'))
    return Array.isArray(data) ? data : []
  } catch (error) {
    return handleError('banners/active', error, [])
  }
}

// ============ Berita ============

export const getPublicBeritaList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/berita/list', { params })
    return normalizeList(unwrapList(response, 'berita'))
  } catch (error) {
    return handleError('berita/list', error, [])
  }
}

export const getPublicBeritaById = async (id) => {
  try {
    const response = await apiClient.get(`/public/berita/${id}`)
    return normalize(unwrapData(response))
  } catch (error) {
    return handleError(`berita/${id}`, error, null)
  }
}

// ============ Desa ============

export const getPublicDesa = async () => {
  try {
    return unwrapData(await apiClient.get('/public/desa'))
  } catch (error) {
    return handleError('desa', error, null)
  }
}

// ============ Fasilitas ============

export const getPublicFasilitasList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/fasilitas/list', { params })
    return normalizeList(unwrapList(response, 'fasilitas'))
  } catch (error) {
    return handleError('fasilitas/list', error, [])
  }
}

// ============ PPID ============

export const getPublicPPIDList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/ppid/list', { params })
    return normalizeList(unwrapList(response, 'ppid'))
  } catch (error) {
    return handleError('ppid/list', error, [])
  }
}

export const getPublicPPIDById = async (id) => {
  try {
    const response = await apiClient.get(`/public/ppid/${id}`)
    return normalize(unwrapData(response))
  } catch (error) {
    return handleError(`ppid/${id}`, error, null)
  }
}

export const createPPIDRequest = async (id, data) => {
  try {
    return unwrapData(await apiClient.post(`/public/ppid/${id}/requests`, data))
  } catch (error) {
    return handleError(`ppid/${id}/requests`, error, null)
  }
}

// ============ Struktur ============

export const getPublicStrukturList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/struktur/list', { params })
    return normalizeList(unwrapList(response, 'struktur'))
  } catch (error) {
    return handleError('struktur/list', error, [])
  }
}

// ============ UMKM ============

export const getPublicUMKMList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/umkm/list', { params })
    return normalizeList(unwrapList(response, 'umkm'))
  } catch (error) {
    return handleError('umkm/list', error, [])
  }
}

// ============ Profile ============

export const getPublicProfileList = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/profile/list', { params })
    return normalizeList(unwrapList(response, 'profile'))
  } catch (error) {
    return handleError('profile/list', error, [])
  }
}

// ============ Infografik (Metabase embeds) ============

export const getPublicInfographicList = async ({ page = 1, limit = 50 } = {}) => {
  try {
    const response = await apiClient.get('/public/infographic/list', {
      params: { page, limit }
    })
    const { items, pagination } = unwrapPaginated(response, 'infographic')
    return { infographics: normalizeList(items), pagination }
  } catch (error) {
    return handleError('infographic/list', error, { infographics: [], pagination: null })
  }
}

export const getPublicInfographicWithToken = async (id) => {
  try {
    const response = await apiClient.get(`/public/infographic/${id}`)
    const data = unwrapData(response)
    if (!data) return null
    return { infographic: normalize(data), token: data.token || null }
  } catch (error) {
    if (error?.status === 429) {
      const err = new Error('rate_limited')
      err.code = 'rate_limited'
      err.status = 429
      throw err
    }
    if (error?.status === 404) {
      const err = new Error('not_found')
      err.code = 'not_found'
      err.status = 404
      throw err
    }
    return handleError(`infographic/${id}`, error, null)
  }
}

export const getInfographicEmbedUrl = (metabaseUrl, infographic, token) => {
  if (!metabaseUrl || !token || !infographic) return null
  const base = metabaseUrl.replace(/\/+$/, '')
  const path = infographic.component_type === 'dashboard'
    ? `/embed/dashboard/${token}`
    : `/embed/question/${token}`
  return `${base}${path}#bordered=false&titled=false`
}

// ============ Galeri ============

export const getPublicGalleryFolders = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/gallery/folders', { params })
    const { items, pagination } = unwrapPaginated(response, 'folders')
    return { folders: items, pagination }
  } catch (error) {
    return handleError('gallery/folders', error, { folders: [], pagination: null })
  }
}

export const getPublicGalleryFolderById = async (id) => {
  try {
    const response = await apiClient.get(`/public/gallery/folders/${id}`)
    const data = unwrapData(response) || {}
    return {
      folder: data.folder || null,
      media: Array.isArray(data.media) ? data.media : [],
      pagination: data.pagination || null
    }
  } catch (error) {
    return handleError(`gallery/folders/${id}`, error, { folder: null, media: [], pagination: null })
  }
}

export const resolveGalleryAssetUrl = (value) => {
  if (!value) return ''
  const str = String(value).trim()
  if (!str) return ''
  if (/^https?:\/\//i.test(str)) return str
  if (str.startsWith('/')) return `${API_BASE_URL.replace(/\/+$/, '')}${str}`
  return str
}

export const resolveMediaThumbnailUrl = (media) => {
  if (!media) return ''
  return resolveGalleryAssetUrl(media.thumbnail_url)
}

export const resolveMediaContentUrl = (media) => {
  if (!media) return ''
  return resolveGalleryAssetUrl(media.content_url)
}

export const resolveFolderCoverUrl = (folder) => {
  if (!folder) return ''
  return resolveGalleryAssetUrl(folder.cover_thumbnail_url)
}

export const fileSrc = (filename) => buildFileUrl(filename)

export const formatMediaDate = (dateString) => {
  if (!dateString) return '-'
  try {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric'
    })
  } catch {
    return '-'
  }
}
