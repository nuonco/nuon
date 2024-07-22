// general utils
export const API_URL =
  (process?.env?.NEXT_PUBLIC_API_URL || process?.env?.NUON_API_URL) || 'https://api.nuon.co'
export const POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_POLL_DURATION as unknown as number) || 45000
export const SHORT_POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_SHORT_POLL_DURATION as unknown as number) || 15000
export const GITHUB_APP_NAME =
  process?.env?.NEXT_PUBLIC_GITHUB_APP_NAME || 'nuon-connect'

export const sentanceCase = (s = '') => s.charAt(0).toUpperCase() + s.slice(1)
export const titleCase = (s = '') =>
  s.replace(/^_*(.)|_+(.)/g, (s, c, d) => (c ? c.toUpperCase() : ' ' + d))

export function getFlagEmoji(countryCode = 'us') {
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

export * from './install-regions'
export * from './get-fetch-opts'
export * from './datadog-logs'
export * from './datadog-rum'
export * from './posthog-analytics'
