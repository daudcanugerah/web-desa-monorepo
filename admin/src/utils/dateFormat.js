const LOCALE = 'id-ID'

function toDate(value) {
  if (!value) return null
  if (value instanceof Date) return value
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? null : d
}

export function formatDate(value) {
  const d = toDate(value)
  if (!d) return '—'
  return d.toLocaleDateString(LOCALE, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

export function formatDateTime(value) {
  const d = toDate(value)
  if (!d) return '—'
  const dateStr = d.toLocaleDateString(LOCALE, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
  const timeStr = d.toLocaleTimeString(LOCALE, {
    hour: '2-digit',
    minute: '2-digit',
  })
  return `${dateStr} ${timeStr}`
}

export function formatLongDate(value) {
  const d = toDate(value)
  if (!d) return '—'
  return d.toLocaleDateString(LOCALE, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export function formatRelative(value) {
  const d = toDate(value)
  if (!d) return '—'
  const now = Date.now()
  const diff = now - d.getTime()
  const sec = Math.round(diff / 1000)
  const min = Math.round(sec / 60)
  const hr = Math.round(min / 60)
  const day = Math.round(hr / 24)
  if (sec < 60) return 'baru saja'
  if (min < 60) return `${min} menit lalu`
  if (hr < 24) return `${hr} jam lalu`
  if (day < 7) return `${day} hari lalu`
  return formatDate(value)
}