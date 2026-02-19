export const REFRESH_PAGE_INTERVAL =
  process.env.REFRESH_PAGE_INTERVAL || 10 * 60 * 1000 // 10 minutes
export const REFRESH_PAGE_WARNING = process.env.REFRESH_PAGE_WARNING || false
export const SF_TRIAL_ACCESS_ENDPOINT = process?.env?.SF_TRIAL_ACCESS_ENDPOINT
export const VERSION = process?.env?.VERSION || '0.1.0'
export const IS_BYOC: boolean =
  process?.env?.NUON_BYOC === 'true' || process?.env?.NUON_BYOC === '1' || false

export const CUSTOMER_PORTAL_URL = (() => {
  const env = process?.env?.NEXT_PUBLIC_DATADOG_ENV
  if (env === 'stage') return 'https://customers.stage.nuon.co'
  if (env === 'local') return 'http://localhost:8080'
  return 'https://customers.nuon.co'
})()

export const IS_CUSTOMER_PORTAL_ENABLED =
  process?.env?.ENABLE_CUSTOMER_PORTAL === 'true'
