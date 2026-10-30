import { describe, expect, test } from 'bun:test'
import { DateTime } from 'luxon'
import {
  scheduleDisabledReason,
  SCHEDULED_RECONCILE_DISABLED_REASON,
  validateScheduledAt,
} from './schedule-utils'
import type { TInstallScheduledReconcile } from '@/types'

const now = DateTime.fromISO('2026-01-01T00:00:00Z')

describe('validateScheduledAt', () => {
  test('rejects times less than 1 minute from now', () => {
    expect(validateScheduledAt(now.plus({ seconds: 30 }), now)).toBe(
      'Schedule at least 1 minute from now.'
    )
  })

  test('rejects times in the past', () => {
    expect(validateScheduledAt(now.minus({ minutes: 5 }), now)).toBe(
      'Schedule at least 1 minute from now.'
    )
  })

  test('accepts exactly 1 minute from now', () => {
    expect(validateScheduledAt(now.plus({ minutes: 1 }), now)).toBeUndefined()
  })

  test('accepts times well within the window', () => {
    expect(validateScheduledAt(now.plus({ days: 5 }), now)).toBeUndefined()
  })

  test('accepts exactly 90 days from now', () => {
    expect(validateScheduledAt(now.plus({ days: 90 }), now)).toBeUndefined()
  })

  test('rejects times beyond 90 days from now', () => {
    expect(validateScheduledAt(now.plus({ days: 91 }), now)).toBe(
      'Schedule within the next 90 days.'
    )
  })

  test('rejects an invalid date', () => {
    expect(validateScheduledAt(DateTime.invalid('bad'), now)).toBe(
      'Enter a valid date and time.'
    )
  })
})

describe('scheduleDisabledReason', () => {
  test('returns undefined when there is no pending schedule', () => {
    expect(scheduleDisabledReason(undefined)).toBeUndefined()
  })

  test('returns the disabled reason when a schedule is pending', () => {
    const pending = { status: 'pending' } as TInstallScheduledReconcile
    expect(scheduleDisabledReason(pending)).toBe(
      SCHEDULED_RECONCILE_DISABLED_REASON
    )
  })
})
