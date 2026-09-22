import { expect, test } from 'bun:test'
import type { RouteObject } from 'react-router'
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

test('registers the new install IA navigation and placeholder routes', () => {
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
    'State',
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
  ])

  expect(paths(installRoutes)).toEqual(
    expect.arrayContaining([
      ':orgId/installs/:installId',
      ':orgId/installs/:installId/resources',
      'sandbox',
      'components',
      'images',
      'state',
      ':orgId/installs/:installId/deployments',
      ':orgId/installs/:installId/health',
      ':orgId/installs/:installId/operations',
      'actions',
      'runbooks',
      'policies',
      'runner',
      ':orgId/installs/:installId/configuration',
      'inputs',
      'config-file',
      'overrides',
    ])
  )
})
