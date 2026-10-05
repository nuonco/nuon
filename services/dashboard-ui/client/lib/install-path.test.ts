import { expect, test } from 'bun:test'
import {
  installHref,
  installPathnameSuffix,
  isNewInstallIAEnabled,
  withAppInstallBreadcrumbs,
} from './install-path'

test('install href stays on the org when the new install IA is off', () => {
  expect(
    installHref({
      orgId: 'org-1',
      appId: 'app-1',
      installId: 'inst-1',
      suffix: '/components/comp-1',
    })
  ).toBe('/org-1/installs/inst-1/components/comp-1')
})

test('install href nests under the app when the new install IA is on', () => {
  expect(
    installHref({
      orgId: 'org-1',
      appId: 'app-1',
      installId: 'inst-1',
      nested: true,
      suffix: '/health',
    })
  ).toBe('/org-1/apps/app-1/installs/inst-1/health')
})

test('install href keeps the org path when the app id is missing', () => {
  expect(
    installHref({
      orgId: 'org-1',
      installId: 'inst-1',
      nested: true,
    })
  ).toBe('/org-1/installs/inst-1')
})

test('workflow detail nests under deployments when the new install IA is on', () => {
  expect(
    installHref({
      orgId: 'org-1',
      appId: 'app-1',
      installId: 'inst-1',
      nested: true,
      suffix: '/history/wf-1',
    })
  ).toBe('/org-1/apps/app-1/installs/inst-1/deployments/wf-1')
  expect(
    installHref({
      orgId: 'org-1',
      appId: 'app-1',
      installId: 'inst-1',
      nested: true,
      suffix: '/workflows/wf-1?panel=step-1',
    })
  ).toBe('/org-1/apps/app-1/installs/inst-1/deployments/wf-1?panel=step-1')
  expect(
    installHref({
      orgId: 'org-1',
      appId: 'app-1',
      installId: 'inst-1',
      nested: true,
      suffix: '/workflows?type=drift_run',
    })
  ).toBe('/org-1/apps/app-1/installs/inst-1/workflows?type=drift_run')
  expect(
    installHref({
      orgId: 'org-1',
      installId: 'inst-1',
      suffix: '/history/wf-1',
    })
  ).toBe('/org-1/installs/inst-1/history/wf-1')
})

test('install pathname suffix keeps the page after the install id', () => {
  expect(
    installPathnameSuffix(
      '/org-1/installs/inst-1/components/comp-1/deploys/dep-1',
      'inst-1'
    )
  ).toBe('/components/comp-1/deploys/dep-1')
  expect(
    installPathnameSuffix(
      '/org-1/apps/app-1/installs/inst-1/resources/sandbox',
      'inst-1'
    )
  ).toBe('/resources/sandbox')
})

test('new install IA follows new-app-ia', () => {
  expect(isNewInstallIAEnabled({ 'new-app-ia': true })).toBe(true)
  expect(isNewInstallIAEnabled({ 'new-app-ia': false })).toBe(false)
})

test('breadcrumbs gain the app when the new install IA is on', () => {
  expect(
    withAppInstallBreadcrumbs(
      [
        { path: '/org-1', text: 'Acme' },
        { path: '/org-1/installs', text: 'Installs' },
        { path: '/org-1/installs/inst-1', text: 'prod' },
        { path: '/org-1/installs/inst-1/health', text: 'Health' },
      ],
      {
        nested: true,
        orgId: 'org-1',
        appId: 'app-1',
        appName: 'Payments',
        installId: 'inst-1',
      }
    )
  ).toEqual([
    { path: '/org-1', text: 'Acme' },
    { path: '/org-1/apps', text: 'Apps' },
    { path: '/org-1/apps/app-1', text: 'Payments' },
    { path: '/org-1/apps/app-1/installs/inst-1', text: 'prod' },
    { path: '/org-1/apps/app-1/installs/inst-1/health', text: 'Health' },
  ])
})

test('breadcrumbs stay put when the new install IA is off', () => {
  const crumbs = [
    { path: '/org-1', text: 'Acme' },
    { path: '/org-1/installs', text: 'Installs' },
    { path: '/org-1/installs/inst-1', text: 'prod' },
  ]
  expect(
    withAppInstallBreadcrumbs(crumbs, {
      nested: false,
      orgId: 'org-1',
      appId: 'app-1',
      appName: 'Payments',
      installId: 'inst-1',
    })
  ).toBe(crumbs)
})
