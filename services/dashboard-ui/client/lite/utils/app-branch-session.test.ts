import { afterEach, describe, expect, test } from 'bun:test'
import {
  LAST_APP_BRANCH_STORAGE_KEY,
  getLastAppBranch,
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
