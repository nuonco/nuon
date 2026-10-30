import { DateTime } from 'luxon'
import type { TInstallScheduledReconcile } from '@/types'

export const MIN_SCHEDULE_OFFSET_MS = 60 * 1000
export const MAX_SCHEDULE_OFFSET_MS = 90 * 24 * 60 * 60 * 1000

/**
 * Mirrors the ctl-api window check (now+1 minute, now+90 days) so the form
 * can surface the same constraint before submitting.
 */
export const validateScheduledAt = (
  scheduledAt: DateTime,
  now: DateTime = DateTime.now()
): string | undefined => {
  if (!scheduledAt.isValid) return 'Enter a valid date and time.'
  const diffMs = scheduledAt.diff(now).as('milliseconds')
  if (diffMs < MIN_SCHEDULE_OFFSET_MS) {
    return 'Schedule at least 1 minute from now.'
  }
  if (diffMs > MAX_SCHEDULE_OFFSET_MS) {
    return 'Schedule within the next 90 days.'
  }
  return undefined
}

/** `datetime-local` inputs are naive (no timezone); interpret them in the browser's local zone. */
export const parseLocalDateTimeInput = (value: string): DateTime =>
  DateTime.fromISO(value, { zone: 'local' })

export const SCHEDULED_RECONCILE_DISABLED_REASON =
  'This install already has a scheduled reconcile. Cancel it to schedule another.'

export const scheduleDisabledReason = (
  pending: TInstallScheduledReconcile | undefined
): string | undefined =>
  pending ? SCHEDULED_RECONCILE_DISABLED_REASON : undefined

export const SCHEDULED_RECONCILE_STATUS_THEME: Record<
  NonNullable<TInstallScheduledReconcile['status']>,
  'neutral' | 'success' | 'error' | 'info' | 'warn'
> = {
  pending: 'info',
  fired: 'success',
  'no-changes': 'neutral',
  cancelled: 'neutral',
  failed: 'error',
}

export const SCHEDULED_RECONCILE_STATUS_LABEL: Record<
  NonNullable<TInstallScheduledReconcile['status']>,
  string
> = {
  pending: 'Pending',
  fired: 'Fired',
  'no-changes': 'No changes',
  cancelled: 'Cancelled',
  failed: 'Failed',
}
