const RAW_HOST = import.meta.env.VITE_API_BASE_URL

if (!RAW_HOST) {
  console.error(
    '[imageUrl] VITE_API_BASE_URL is not defined. Set it in .env (e.g. VITE_API_BASE_URL=http://localhost:8080). Image URLs will be relative until fixed.'
  )
}

const HOST = (RAW_HOST || '').replace(/\/+$/, '')
export const API_V1_PREFIX = '/api/v1'

export function getImageUrl(imageNameOrUrl) {
  if (!imageNameOrUrl) return null

  if (/^https?:\/\//i.test(imageNameOrUrl)) {
    return imageNameOrUrl
  }

  let path = imageNameOrUrl

  if (path.startsWith('/uploads/')) {
    path = path.slice('/uploads/'.length)
  } else if (path.startsWith('uploads/')) {
    path = path.slice('uploads/'.length)
  } else if (path.startsWith('/files/')) {
    path = path.slice('/files/'.length)
  } else if (path.startsWith('files/')) {
    path = path.slice('files/'.length)
  } else if (path.startsWith('/')) {
    return `${HOST}${path}`
  }

  return `${HOST}${API_V1_PREFIX}/files/${path.replace(/^\/+/, '')}`
}

export const API_BASE_URL = `${HOST}${API_V1_PREFIX}`

// Resolve any media URL emitted by the API to an absolute URL on the API
// origin. Signed media paths are relative (/api/v1/media/{id}/...?jwt=...)
// and must never hit the SPA origin — the Vite dev server would answer with
// index.html instead of the binary.
export function resolveMediaUrl(value) {
  if (!value) return ''
  const str = String(value).trim()
  if (!str) return ''
  if (/^(https?:|blob:|data:)/i.test(str)) return str
  if (str.startsWith('/')) return `${HOST}${str}`
  return getImageUrl(str)
}