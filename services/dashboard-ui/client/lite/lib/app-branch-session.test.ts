import { afterEach, describe, expect, test } from 'bun:test'
import {
  LAST_APP_BRANCH_STORAGE_KEY,
  appSetupHref,
  getLastAppBranch,
  resolveAppHref,
  setLastAppBranch,
} from './app-branch-session'

afterEach(() => {
  localStorage.clear()
})

describe('last app branch session', () => {
  test('stores the last viewed branch per org and app', () => {
    setLastAppBranch('org_a', 'app_payments', 'br_main')
    setLastAppBranch('org_a', 'app_billing', 'br_next')
    setLastAppBranch('org_b', 'app_payments', 'br_other')

    expect(getLastAppBranch('org_a', 'app_payments')).toBe('br_main')
    expect(getLastAppBranch('org_a', 'app_billing')).toBe('br_next')
    expect(getLastAppBranch('org_b', 'app_payments')).toBe('br_other')
    expect(getLastAppBranch('org_a', 'app_missing')).toBeUndefined()
  })

  test('ignores unreadable storage', () => {
    localStorage.setItem(LAST_APP_BRANCH_STORAGE_KEY, '{')

    expect(getLastAppBranch('org_a', 'app_payments')).toBeUndefined()
  })
})

describe('resolveAppHref', () => {
  test('points a new app at the setup path', () => {
    expect(appSetupHref('org_a')).toBe('/org_a/apps/setup')
  })

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
