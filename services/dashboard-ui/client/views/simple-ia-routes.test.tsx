import { expect, test } from 'bun:test'
import { isValidElement } from 'react'
import { matchRoutes } from 'react-router'
import { orgRoutes } from '@/views/org/routes'
import { SimpleIAGate } from '@/views/SimpleIAGate'

const matchedRoutes = (path: string) => matchRoutes(orgRoutes, path) ?? []

const expectSimpleIARoute = (path: string, routePath: string) => {
  const matches = matchedRoutes(path)
  const leaf = matches.at(-1)?.route
  const gate = matches.at(-2)?.route.element

  expect(leaf?.path).toBe(routePath)
  expect(isValidElement(gate) && gate.type === SimpleIAGate).toBe(true)
}

test('registers simple IA setup routes behind the feature gate', () => {
  expectSimpleIARoute('/org-1/apps/setup', ':orgId/apps/setup')
  expectSimpleIARoute('/org-1/installs/setup', ':orgId/installs/setup')
})

test('registers simple IA activity and config routes behind the feature gate', () => {
  expectSimpleIARoute(
    '/org-1/apps/app-1/branches/branch-1/activity',
    'activity'
  )
  expectSimpleIARoute('/org-1/apps/app-1/branches/branch-1/config', 'config')
  expectSimpleIARoute(
    '/org-1/installs/install-1/activity',
    ':orgId/installs/:installId/activity'
  )
})
