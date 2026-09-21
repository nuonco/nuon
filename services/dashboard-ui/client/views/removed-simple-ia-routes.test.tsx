import { expect, test } from 'bun:test'
import type { RouteObject } from 'react-router'
import { orgRoutes } from '@/views/org/routes'

const removedPaths = [
  ':orgId/apps/setup',
  ':orgId/installs/setup',
  ':orgId/apps/:appId/branches/:branchId/activity',
  ':orgId/apps/:appId/branches/:branchId/config',
  ':orgId/installs/:installId/activity',
]

const collectPaths = (routes: RouteObject[], parentPath = ''): string[] =>
  routes.flatMap((route) => {
    const path = [parentPath, route.path].filter(Boolean).join('/')
    return [path, ...collectPaths(route.children ?? [], path)]
  })

test('does not register simple IA routes', () => {
  const paths = collectPaths(orgRoutes)
  removedPaths.forEach((path) => expect(paths).not.toContain(path))
})
