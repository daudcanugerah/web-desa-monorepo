const KEYS = {
  accessToken: 'access_token',
  refreshToken: 'refresh_token',
  expiresAt: 'expires_at',
}

export function getAccessToken() {
  return localStorage.getItem(KEYS.accessToken)
}

export function getRefreshToken() {
  return localStorage.getItem(KEYS.refreshToken)
}

export function setTokens({ access_token, refresh_token, expires_at }) {
  if (access_token) localStorage.setItem(KEYS.accessToken, access_token)
  if (refresh_token) localStorage.setItem(KEYS.refreshToken, refresh_token)
  if (expires_at) localStorage.setItem(KEYS.expiresAt, expires_at)
}

export function clearTokens() {
  Object.values(KEYS).forEach((k) => localStorage.removeItem(k))
}
