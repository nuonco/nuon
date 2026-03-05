/**
 * Dynamically discovers routes from the app directory filesystem
 * This is for getting a route name for datadog metrics.
 */

import { readdirSync, statSync } from 'fs'
import { join } from 'path'

interface RoutePattern {
  pattern: string
  regex: RegExp
  paramNames: string[]
}

// Cache of compiled route patterns
let routePatterns: RoutePattern[] | null = null

/**
 * Recursively scan the app directory to discover all routes
 */
function discoverRoutes(baseDir: string, currentPath = ''): string[] {
  const routes: string[] = []

  try {
    const entries = readdirSync(baseDir)

    for (const entry of entries) {
      const fullPath = join(baseDir, entry)
      const stat = statSync(fullPath)

      if (stat.isDirectory()) {
        // Skip special Next.js directories that aren't routes
        if (entry.startsWith('_') || entry === 'node_modules') {
          continue
        }

        // Process directory as a route segment
        const routeSegment = currentPath + '/' + entry

        // Recursively scan subdirectories
        const childRoutes = discoverRoutes(fullPath, routeSegment)
        routes.push(...childRoutes)
      } else if (stat.isFile()) {
        // Check if this is a route-defining file (page.tsx, route.ts, layout.tsx)
        if (
          entry === 'page.tsx' ||
          entry === 'page.ts' ||
          entry === 'page.jsx' ||
          entry === 'page.js' ||
          entry === 'route.ts' ||
          entry === 'route.tsx' ||
          entry === 'route.js' ||
          entry === 'route.jsx'
        ) {
          // This directory defines a route
          routes.push(currentPath || '/')
        }
      }
    }
  } catch (error) {
    console.error(`Error scanning directory ${baseDir}:`, error)
  }

  return routes
}

/**
 * Build route patterns from Next.js App Router file structure
 * This runs once at startup to create matchers for all routes
 */
function buildRoutePatterns(): RoutePattern[] {
  // Discover routes from the app directory
  const appDir = join(process.cwd(), 'src', 'app')
  let routes: string[] = []

  try {
    routes = discoverRoutes(appDir)
  } catch (error) {
    console.error('Error discovering routes:', error)
    // Fallback to empty array if discovery fails
    routes = []
  }

  // Sort routes by specificity (more specific routes first)
  const sortedRoutes = routes.sort((a, b) => {
    const aSegments = a.split('/').filter((s) => s)
    const bSegments = b.split('/').filter((s) => s)

    // More segments = more specific
    if (aSegments.length !== bSegments.length) {
      return bSegments.length - aSegments.length
    }

    // Fewer dynamic segments = more specific
    const aDynamic = (a.match(/\[/g) || []).length
    const bDynamic = (b.match(/\[/g) || []).length
    if (aDynamic !== bDynamic) {
      return aDynamic - bDynamic
    }

    // Catch-all routes ([...slug]) are less specific
    const aCatchAll = a.includes('[...')
    const bCatchAll = b.includes('[...')
    if (aCatchAll !== bCatchAll) {
      return aCatchAll ? 1 : -1
    }

    return 0
  })

  // Convert route patterns to regex matchers
  return sortedRoutes.map((pattern) => {
    const paramNames: string[] = []

    // Convert Next.js pattern to regex
    // [param] -> capture group
    // [...slug] -> capture rest
    const regexPattern = pattern
      .split('/')
      .map((segment) => {
        if (segment.startsWith('[...') && segment.endsWith(']')) {
          // Catch-all route
          const paramName = segment.slice(4, -1)
          paramNames.push(paramName)
          return '.*'
        } else if (segment.startsWith('[') && segment.endsWith(']')) {
          // Dynamic segment
          const paramName = segment.slice(1, -1)
          paramNames.push(paramName)
          return '[^/]+'
        } else {
          // Static segment - escape special regex characters
          return segment.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
        }
      })
      .join('/')

    return {
      pattern,
      regex: new RegExp(`^${regexPattern}$`),
      paramNames,
    }
  })
}

/**
 * Match a URL path to its Next.js route pattern
 * Returns the pattern (e.g., "/api/[org-id]/apps/[app-id]") or the path itself if no match
 */
export function matchRoutePattern(url: string): string {
  // Lazy initialize patterns
  if (!routePatterns) {
    routePatterns = buildRoutePatterns()
  }

  // Remove query params and hash
  const path = url.split('?')[0].split('#')[0]

  // Special cases for common prefixes
  if (path.startsWith('/_next/')) {
    return '/_next'
  }

  // Try to match against known patterns
  for (const { pattern, regex } of routePatterns) {
    if (regex.test(path)) {
      return pattern
    }
  }

  // Fallback: return the path as-is if no pattern matches
  return 'not-found'
}
