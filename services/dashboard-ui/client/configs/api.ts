function getApiUrl(): string {
  if (typeof window !== 'undefined' && window.__NUON_CONFIG__ !== undefined) {
    return window.__NUON_CONFIG__.apiUrl ?? ''
  }
  return ''
}

export const API_URL = getApiUrl()
export const POLLING_TIMEOUT = 10000
export const POLLING_TIMEOUT_SHORT = 5000
export const POLLING_TIMEOUT_LOGS = 2000
