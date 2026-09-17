import { describe, expect, test } from 'bun:test'
import {
  appSetupHref,
  installHref,
  installSetupHref,
  resolveAppHref,
} from './hrefs'

describe('setup hrefs', () => {
  test('starts a new resource on the bare setup path', () => {
    expect(appSetupHref('org_a')).toBe('/org_a/apps/setup')
    expect(installSetupHref('org_a')).toBe('/org_a/installs/setup')
  })

  test('carries an existing resource in the query', () => {
    expect(appSetupHref('org_a', 'app_payments')).toBe(
      '/org_a/apps/setup?appId=app_payments'
    )
    expect(installSetupHref('org_a', 'inst_prod')).toBe(
      '/org_a/installs/setup?installId=inst_prod'
    )
  })

  test('links to an existing install', () => {
    expect(installHref('org_a', 'inst_prod')).toBe(
      '/org_a/installs/inst_prod'
    )
  })
})

describe('resolveAppHref', () => {
  test('sends apps without branches to setup', () => {
    expect(
      resolveAppHref({
        orgId: 'org_a',
        appId: 'app_payments',
        branchIds: [],
      })
    ).toBe('/org_a/apps/setup?appId=app_payments')
  })

  test('sends a single-branch app to that branch', () => {
    expect(
      resolveAppHref({
        orgId: 'org_a',
        appId: 'app_payments',
        branchIds: ['br_main'],
        lastBranchId: 'br_stale',
      })
    ).toBe('/org_a/apps/app_payments/branches/br_main')
  })

  test('prefers the last viewed branch when it still exists', () => {
    expect(
      resolveAppHref({
        orgId: 'org_a',
        appId: 'app_payments',
        branchIds: ['br_main', 'br_next'],
        lastBranchId: 'br_next',
      })
    ).toBe('/org_a/apps/app_payments/branches/br_next')
  })

  test('falls back to the first branch when the last viewed one is gone', () => {
    expect(
      resolveAppHref({
        orgId: 'org_a',
        appId: 'app_payments',
        branchIds: ['br_main', 'br_next'],
        lastBranchId: 'br_deleted',
      })
    ).toBe('/org_a/apps/app_payments/branches/br_main')
  })
})
