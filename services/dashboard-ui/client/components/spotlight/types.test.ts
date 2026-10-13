import { expect, test } from 'bun:test'
import {
  APP_SUB_PAGES,
  NEW_INSTALL_IA_SUB_PAGES,
  spotlightAppNavigation,
  spotlightCommands,
  spotlightInstallActionPath,
  spotlightInstallComponentPath,
  spotlightInstallPath,
  spotlightOrgPages,
} from './types'

const labels = (newAppIA: boolean) =>
  spotlightOrgPages(newAppIA).map((page) => page.label)

test('org spotlight pages hide installs and the build runner under new-app-ia', () => {
  expect(labels(false)).toEqual([
    'Dashboard',
    'Apps',
    'Installs',
    'Team',
    'Build runner',
    'Settings',
  ])
  expect(labels(true)).toEqual([
    'Dashboard',
    'Apps',
    'Team',
    'Settings',
  ])
})

test('app spotlight navigation lists branches under new-app-ia', () => {
  const legacy = spotlightAppNavigation({
    appId: 'app-1',
    appName: 'Payments',
    newAppIA: false,
    query: '',
    subPages: APP_SUB_PAGES,
  }).map((result) => result.path)

  expect(legacy).toEqual([
    '/apps/app-1',
    '/apps/app-1/components',
    '/apps/app-1/actions',
    '/apps/app-1/roles',
    '/apps/app-1/policies',
    '/apps/app-1/installs',
  ])

  const nested = spotlightAppNavigation({
    appId: 'app-1',
    appName: 'Payments',
    branches: [
      { id: 'br-main', name: 'main' },
      { id: 'br-dev', name: 'dev' },
    ],
    newAppIA: true,
    query: 'pay',
    subPages: APP_SUB_PAGES,
  }).map((result) => [result.label, result.path])

  expect(nested).toEqual([
    ['Payments', '/apps/app-1'],
    ['Payments › main', '/apps/app-1/branches/br-main'],
    ['Payments › dev', '/apps/app-1/branches/br-dev'],
  ])
})

test('app spotlight commands use trigger run under new-app-ia', () => {
  expect(spotlightCommands(false).app).toEqual(['build all components'])
  expect(spotlightCommands(false).component).toEqual([
    'build',
    'deploy',
    'drift scan',
    'teardown',
  ])
  expect(spotlightCommands(true).app).toEqual(['trigger run'])
  expect(spotlightCommands(true).component).toEqual([
    'deploy',
    'drift scan',
    'teardown',
  ])
  expect(spotlightCommands(true).install).toEqual(spotlightCommands(false).install)
  expect(spotlightCommands(true).action).toEqual(spotlightCommands(false).action)
})

test('install spotlight opens the nested app route under new-app-ia', () => {
  expect(
    spotlightInstallPath({ appId: 'app-1', installId: 'inst-1', newAppIA: false })
  ).toBe('/installs/inst-1')
  expect(
    spotlightInstallPath({ appId: 'app-1', installId: 'inst-1', newAppIA: true })
  ).toBe('/apps/app-1/installs/inst-1')
  expect(
    spotlightInstallPath({ installId: 'inst-1', newAppIA: true })
  ).toBe('/installs/inst-1')
})

test('new install pages and resource searches use the nested routes', () => {
  expect(NEW_INSTALL_IA_SUB_PAGES).toEqual([
    { label: 'Components', suffix: '/resources/components' },
    { label: 'Images', suffix: '/resources/images' },
    { label: 'Actions', suffix: '/operations/actions' },
    { label: 'Runner', suffix: '/operations/runner' },
    { label: 'Deployments', suffix: '/deployments' },
    { label: 'Stack', suffix: '/resources' },
    { label: 'Inputs', suffix: '/configuration/inputs' },
    { label: 'State', suffix: '/configuration/state' },
  ])

  const nested = { appId: 'app-1', installId: 'inst-1', newAppIA: true }
  expect(
    spotlightInstallComponentPath({
      ...nested,
      componentId: 'cmp-1',
      name: 'api server',
      type: 'helm_chart',
    })
  ).toBe('/apps/app-1/installs/inst-1/resources/components?q=api%20server')
  expect(
    spotlightInstallComponentPath({
      ...nested,
      componentId: 'cmp-2',
      name: 'worker',
      type: 'docker_build',
    })
  ).toBe('/apps/app-1/installs/inst-1/resources/images?q=worker')
  expect(
    spotlightInstallComponentPath({
      ...nested,
      componentId: 'cmp-3',
      name: 'proxy',
      type: 'external_image',
    })
  ).toBe('/apps/app-1/installs/inst-1/resources/images?q=proxy')
  expect(
    spotlightInstallActionPath({
      ...nested,
      actionId: 'act-1',
      name: 'restart api',
    })
  ).toBe('/apps/app-1/installs/inst-1/operations/actions?q=restart%20api')
})
