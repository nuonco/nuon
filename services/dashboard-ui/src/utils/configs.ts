export const API_URL =
  process?.env?.NEXT_PUBLIC_API_URL ||
  process?.env?.NUON_API_URL ||
  'https://api.nuon.co'
export const ADMIN_API_URL =
  process?.env?.NEXT_PUBLIC_ADMIN_API_URL || 'http://localhost:8082'
export const POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_POLL_DURATION as unknown as number) || 10000
export const SHORT_POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_SHORT_POLL_DURATION as unknown as number) || 5000
export const LOG_POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_LOG_POLL_DURATION as unknown as number) || 1000
export const GITHUB_APP_NAME = process?.env?.GITHUB_APP_NAME || 'nuon-connect'
export const WORKFLOWS =
  Boolean(process?.env?.NUON_WORKFLOWS === 'true') || false
export const RUNNERS = Boolean(process?.env?.NUON_RUNNERS === 'true') || false
export const USER_REPROVISION =
  Boolean(process?.env?.NUON_INSTALL_REPROVISION === 'true') || false
export const DEPLOY_INTERMEDIATE_DATA =
  Boolean(process?.env?.NUON_DEPLOY_DATA === 'true') || false
export const CANCEL_RUNNER_JOBS =
  Boolean(process?.env?.NUON_CANCEL_JOBS === 'true') || false
export const INSTALL_UPDATE =
  Boolean(process?.env?.NUON_INSTALL_UPDATE === 'true') || false
export const VERSION = process?.env?.VERSION || '0.1.0'
export const ORG_DASHBOARD =
  Boolean(process?.env?.NUON_ORG_DASHBOARD === 'true') || false
export const ORG_RUNNER =
  Boolean(process?.env?.NUON_ORG_RUNNER === 'true') || false
export const ORG_SETTINGS =
  Boolean(process?.env?.NUON_ORG_SETTINGS === 'true') || false
export const ORG_SUPPORT =
  Boolean(process?.env?.NUON_ORG_SUPPORT === 'true') || false
