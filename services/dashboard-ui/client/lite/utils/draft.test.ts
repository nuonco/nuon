import { describe, expect, test, afterEach } from 'bun:test'
import {
  DRAFT_STORAGE_PREFIX,
  clearDraft,
  loadDraft,
  migrateDraft,
  saveDraft,
} from './draft'

afterEach(() => {
  localStorage.clear()
})

describe('draft storage', () => {
  test('round-trips field values without a step index', () => {
    saveDraft('org_a', 'app', { name: 'Payments' })

    expect(loadDraft('org_a', 'app')).toEqual({
      values: { name: 'Payments' },
      updatedAt: expect.any(String),
    })
    expect(localStorage.getItem(`${DRAFT_STORAGE_PREFIX}:org_a:app:new`)).toBeTruthy()
  })

  test('keeps two incomplete resources on separate keys', () => {
    saveDraft('org_a', 'app', { name: 'Payments' }, 'app_payments')
    saveDraft('org_a', 'app', { name: 'Billing' }, 'app_billing')

    expect(loadDraft('org_a', 'app', 'app_payments')?.values).toEqual({
      name: 'Payments',
    })
    expect(loadDraft('org_a', 'app', 'app_billing')?.values).toEqual({
      name: 'Billing',
    })
  })

  test('migrates the new key onto a resource id without overwriting', () => {
    saveDraft('org_a', 'app', { name: 'Payments' })
    migrateDraft('org_a', 'app', 'app_payments')

    expect(loadDraft('org_a', 'app')).toBeUndefined()
    expect(loadDraft('org_a', 'app', 'app_payments')?.values).toEqual({
      name: 'Payments',
    })

    saveDraft('org_a', 'app', { name: 'Fresh' })
    saveDraft('org_a', 'app', { name: 'Existing' }, 'app_payments')
    migrateDraft('org_a', 'app', 'app_payments')

    expect(loadDraft('org_a', 'app')).toBeUndefined()
    expect(loadDraft('org_a', 'app', 'app_payments')?.values).toEqual({
      name: 'Existing',
    })
  })

  test('clears on completion', () => {
    saveDraft('org_a', 'app', { name: 'Payments' }, 'app_payments')
    clearDraft('org_a', 'app', 'app_payments')
    expect(loadDraft('org_a', 'app', 'app_payments')).toBeUndefined()
  })
})
