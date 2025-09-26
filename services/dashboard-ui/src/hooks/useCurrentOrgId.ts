'use client'

import { usePathname } from 'next/navigation'

/**
 * Hook to extract the current organization ID from the route
 * Returns empty string for root routes, orgId for org-scoped routes
 */
export const useCurrentOrgId = (): string => {
  const pathname = usePathname()

  // Root routes (/, /some-path) don't have orgId
  if (pathname === '/' || !pathname.includes('/')) {
    return ''
  }

  // Extract orgId from routes like /{orgId}/apps, /{orgId}/installs/{installId}, etc.
  const segments = pathname.split('/').filter(Boolean)

  // First segment should be the orgId for org-scoped routes
  // Skip if it's a known non-org route (api, auth, etc.)
  const firstSegment = segments[0]

  // Known non-org routes that might appear at root level
  const nonOrgRoutes = ['api', 'auth', 'login', 'logout', 'callback']

  if (nonOrgRoutes.includes(firstSegment)) {
    return ''
  }

  // Return the first segment as orgId, or empty string if not found
  return firstSegment || ''
}