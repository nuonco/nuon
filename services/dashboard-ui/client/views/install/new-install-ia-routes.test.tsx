import { expect, test } from 'bun:test'
import { matchRoutes, type RouteObject } from 'react-router'
import { orgRoutes } from '@/views/org/routes'
import { NEW_INSTALL_NAV_LINKS } from './InstallLayout'
import {
  NEW_INSTALL_CONFIGURATION_TABS,
  NEW_INSTALL_OPERATIONS_TABS,
  NEW_INSTALL_RESOURCES_TABS,
} from './NewInstallSectionLayout'
import { installRoutes } from './routes'

const paths = (routes: RouteObject[]): string[] =>
  routes.flatMap((route) => [
    ...(route.path ? [route.path] : []),
    ...paths(route.children ?? []),
  ])

const tabLabels = (tabs: { text: string }[]) => tabs.map((tab) => tab.text)

test('registers the new install IA navigation and routes', () => {
  expect(
    NEW_INSTALL_NAV_LINKS.map((link) => ('text' in link ? link.text : null))
  ).toEqual([
    'Overview',
    'Deployments',
    'Resources',
    'Health',
    'Operations',
    'Configuration',
  ])

  expect(tabLabels(NEW_INSTALL_RESOURCES_TABS)).toEqual([
    'Stack',
    'Sandbox',
    'Components',
    'Images',
  ])
  expect(tabLabels(NEW_INSTALL_OPERATIONS_TABS)).toEqual([
    'Activity',
    'Actions',
    'Runbooks',
    'Policies',
    'Runner',
  ])
  expect(tabLabels(NEW_INSTALL_CONFIGURATION_TABS)).toEqual([
    'App branch',
    'Inputs',
    'Config file',
    'Overrides',
    'State',
  ])

  expect(paths(installRoutes)).toEqual(
    expect.arrayContaining([
      ':orgId/installs/:installId',
      ':orgId/apps/:appId/installs/:installId',
      'resources',
      'sandbox',
      'components',
      'images',
      'state',
      'deployments',
      'health',
      'operations',
      'actions',
      'runbooks',
      'policies',
      'runner',
      'configuration',
      'inputs',
      'config-file',
      'overrides',
    ])
  )
})

const matchedPaths = (pathname: string) =>
  matchRoutes(orgRoutes, pathname)?.map((match) => match.route.path) ?? []

test('install detail matches beside the app layout', () => {
  const detail = matchedPaths('/org-1/apps/app-1/installs/inst-1')
  expect(detail).toContain(':orgId/apps/:appId/installs/:installId')
  expect(detail).not.toContain(':orgId/apps/:appId/installs')

  const list = matchedPaths('/org-1/apps/app-1/installs')
  expect(list).toContain(':orgId/apps/:appId/installs')
  expect(list).not.toContain(':orgId/apps/:appId/installs/:installId')

  const legacy = matchedPaths('/org-1/installs/inst-1/components/comp-1')
  expect(legacy).toContain(':orgId/installs/:installId')
  expect(legacy).toContain('components/:componentId')

  const deployment = matchedPaths(
    '/org-1/apps/app-1/installs/inst-1/deployments/wf-1'
  )
  expect(deployment).toContain('deployments/:workflowId')
  expect(deployment).not.toContain('deployments')

  const history = matchedPaths('/org-1/installs/inst-1/history/wf-1')
  expect(history).toContain('history/:workflowId')
})
