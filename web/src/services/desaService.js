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

// Fetches every page of a paginated public list endpoint, capped by `pageSize`
// (server MaxLimit is 100) and `maxPages` as a safety valve.
const fetchAllPages = async (endpoint, key, params = {}, { pageSize = 100, maxPages = 50 } = {}) => {
  const first = await apiClient.get(endpoint, { params: { ...params, page: 1, limit: pageSize } })
  const firstPage = unwrapPaginated(first, key)
  let items = firstPage.items
  const pagination = firstPage.pagination
  const totalPages = Number(pagination?.total_pages) || 1
  const pagesToFetch = Math.min(totalPages, maxPages)
  for (let page = 2; page <= pagesToFetch; page++) {
    const res = await apiClient.get(endpoint, { params: { ...params, page, limit: pageSize } })
    items = items.concat(unwrapPaginated(res, key).items)
  }
  return { items, pagination }
}

export const mediaUrl = (item, variant = 'single') => {
  if (!item) return ''
  // `profile_media` is the struktur-specific single-media field.
  const m = item.media || item.profile_media
  if (variant === 'array') {
    if (Array.isArray(m) && m.length > 0) return resolveGalleryAssetUrl(m[0]?.url || m[0]?.thumbnail_url || '')
    if (Array.isArray(item.images) && item.images.length > 0) return resolveGalleryAssetUrl(item.images[0])
    return ''
  }
  if (m && !Array.isArray(m)) return resolveGalleryAssetUrl(m.url || m.thumbnail_url || '')
  return resolveGalleryAssetUrl(item.image_url || item.profile_image_url || '')
}

export const thumbnailUrl = (item) => {
  if (!item) return ''
  const t = item.thumbnail
  // Signed media URLs are relative (/api/v1/media/...?jwt=...); resolve
  // them against the API base so they don't hit the SPA origin.
  if (t && typeof t === 'object') return resolveGalleryAssetUrl(t.url || t.thumbnail_url || '')
  return resolveGalleryAssetUrl(item.thumbnail_url || '')
}

export const documentUrl = (item) => {
  if (!item) return ''
  const d = item.document
  if (d && typeof d === 'object') return resolveGalleryAssetUrl(d.url || '')
  return resolveGalleryAssetUrl(item.document_url || '')
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
  // `/files/{name}` is mounted at the API root, not under `/api/v1`.
  const origin = (() => {
    try {
      return new URL(API_BASE_URL).origin
    } catch {
      return API_BASE_URL.replace(/\/api\/v\d+\/?$/, '').replace(/\/+$/, '')
    }
  })()
  return `${origin}/files/${filename}`
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

// Resolves the UUID of a berita category by name (case-insensitive), so the
// Home pengumuman ticker and other views can filter by a known category
// without hardcoding its id.
export const getPublicBeritaCategoryIdByName = async (name) => {
  if (!name) return null
  try {
    const response = await apiClient.get('/public/berita/categories', { params: { limit: 100 } })
    const data = unwrapData(response)
    const list = Array.isArray(data?.categories) ? data.categories : unwrapList(response, 'categories')
    const target = String(name).trim().toLowerCase()
    const found = (list || []).find(c => String(c.name || '').trim().toLowerCase() === target)
    return found?.id || null
  } catch (error) {
    return handleError('berita/categories', error, null)
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

// Fetches ALL public fasilitas (walks pagination) so the map is not capped at
// the server default page size. Returns { items, pagination }.
export const getPublicFasilitasAll = async (params = {}) => {
  try {
    const { items, pagination } = await fetchAllPages('/public/fasilitas/list', 'fasilitas', params)
    return { items: normalizeList(items), pagination }
  } catch (error) {
    return handleError('fasilitas/list/all', error, { items: [], pagination: null })
  }
}

// Public fasilitas categories. Returns [{ id, name, usage_count }].
export const getPublicFasilitasCategories = async () => {
  try {
    const response = await apiClient.get('/public/fasilitas/categories', { params: { limit: 100 } })
    const data = unwrapData(response)
    const list = Array.isArray(data?.categories) ? data.categories : unwrapList(response, 'categories')
    return Array.isArray(list) ? list : []
  } catch (error) {
    return handleError('fasilitas/categories', error, [])
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

// Fetches ALL struktur entries (officials) by walking pagination, so the
// Perangkat list is not capped at the server default page size.
export const getPublicStrukturAll = async (params = {}) => {
  try {
    const { items } = await fetchAllPages('/public/struktur/list', 'struktur', params)
    return normalizeList(items)
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

// Fetches every profile section by walking pagination, so the public Profil
// nav is not capped at the server's default/max page size.
export const getPublicProfileAll = async (params = {}) => {
  try {
    const { items } = await fetchAllPages('/public/profile/list', 'profile', params)
    return normalizeList(items)
  } catch (error) {
    return handleError('profile/list', error, [])
  }
}

// Public profile section categories. Returns [{ id, name, usage_count }].
export const getPublicProfileCategories = async (params = {}) => {
  try {
    const response = await apiClient.get('/public/profile/categories', { params: { limit: 100, ...params } })
    const data = unwrapData(response)
    const list = Array.isArray(data?.categories) ? data.categories : unwrapList(response, 'categories')
    return Array.isArray(list) ? list : []
  } catch (error) {
    return handleError('profile/categories', error, [])
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
  if (str.startsWith('/api/')) {
    // Path already carries the /api/v1 prefix (signed media URLs) —
    // resolve against the API origin to avoid double-prefixing.
    try {
      return `${new URL(API_BASE_URL).origin}${str}`
    } catch {
      return `${API_BASE_URL.replace(/\/+$/, '')}${str}`
    }
  }
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
