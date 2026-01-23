export const USE_AUTH_SERVICE = !!process.env.NUON_AUTH_SERVICE_URL
export const AUTH_SERVICE_URL = 
  process.env.NUON_AUTH_SERVICE_URL || 'http://localhost:8084'
export const APP_URL = process.env.NUON_APP_URL || 'http://localhost:4000'