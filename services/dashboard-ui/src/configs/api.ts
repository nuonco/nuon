export const API_URL =
  process?.env?.API_URL ||
  process?.env?.NUON_API_URL ||
  process?.env?.NEXT_PUBLIC_API_URL ||
  'https://api.nuon.co'
export const ADMIN_API_URL =
  process?.env?.ADMIN_API_URL ||
  process?.env?.NUON_CTL_API_ADMIN_URL ||
  'http://ctl.nuon.us-west-2.prod.internal.nuon.co'
export const ADMIN_TEMPORAL_UI_URL =
  process.env.NUON_TEMPORAL_UI_URL ||
  'http://temporal-web.temporal.svc.cluster.local:8080'

export const POLLING_TIMEOUT = 10000
export const POLLING_TIMEOUT_SHORT = 5000
export const POLLING_TIMEOUT_LOGS = 2000
