export const API_URL =
  process?.env?.NEXT_PUBLIC_API_URL ||
  process?.env?.NUON_API_URL ||
  'https://api.nuon.co'
export const POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_POLL_DURATION as unknown as number) || 45000
export const SHORT_POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_SHORT_POLL_DURATION as unknown as number) || 15000
export const LOG_POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_LOG_POLL_DURATION as unknown as number) || 1000
export const GITHUB_APP_NAME =
  process?.env?.NEXT_PUBLIC_GITHUB_APP_NAME || 'nuon-connect'
