export const GITHUB_APP_NAME =
  typeof window !== 'undefined' && window?.['GITHUB_APP_NAME']
    ? window?.['GITHUB_APP_NAME']
    : process?.env?.GITHUB_APP_NAME || 'nuon-connect'
