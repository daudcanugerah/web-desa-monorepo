const IMAGE_MIME_TYPES = ['image/jpeg', 'image/png', 'image/webp']
const DOCUMENT_MIME_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
]

export function isValidImageType(file) {
  return IMAGE_MIME_TYPES.includes(file.type)
}

export function isValidDocumentType(file) {
  return DOCUMENT_MIME_TYPES.includes(file.type)
}

export function isWithinSize(file, maxMB) {
  return file.size <= maxMB * 1024 * 1024
}

export function formatFileSize(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
