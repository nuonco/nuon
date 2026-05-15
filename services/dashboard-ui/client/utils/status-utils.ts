import type { TIconVariant } from '@/components/common/Icon'

export type TStatusTheme =
  | 'success'
  | 'warn'
  | 'neutral'
  | 'error'
  | 'info'
  | 'brand'

const STATUS_THEME_MAP: Record<string, TStatusTheme> = {
  // Success
  active: 'success',
  ok: 'success',
  finished: 'success',
  healthy: 'success',
  connected: 'success',
  approved: 'success',
  success: 'success',

  // Error
  failed: 'error',
  error: 'error',
  bad: 'error',
  'access-error': 'error',
  access_error: 'error',
  'timed-out': 'error',
  unhealthy: 'error',
  'not connected': 'error',
  'not-connected': 'error',
  suspended: 'error',
  policy_failed: 'error',
  'failed-pending-retry': 'error',

  // Warn
  'approval-denied': 'warn',
  'approval-awaiting': 'warn',
  cancelled: 'warn',
  outdated: 'warn',
  warn: 'warn',
  drifted: 'warn',
  expired: 'warn',
  'pending-shutdown': 'warn',
  'shutting-down': 'warn',
  offline: 'warn',

  // Info
  executing: 'info',
  waiting: 'info',
  started: 'info',
  'in-progress': 'info',
  building: 'info',
  queued: 'info',
  planning: 'info',
  provisioning: 'info',
  syncing: 'info',
  deploying: 'info',
  available: 'info',
  'pending-approval': 'info',
  info: 'info',
  retried: 'info',
  applying: 'info',
  'awaiting-user-run': 'info',

  // Neutral
  noop: 'neutral',
  'shut-down': 'neutral',
  unknown: 'neutral',
  inactive: 'neutral',
  pending: 'neutral',
  'Not deployed': 'neutral',
  'No build': 'neutral',
  'not-attempted': 'neutral',
  deprovisioned: 'neutral',
  skeleton: 'neutral',

  // Brand
  special: 'brand',
  brand: 'brand',
}

const STATUS_ICON_MAP: Record<string, TIconVariant> = {
  // Success
  active: 'CheckCircleIcon',
  ok: 'CheckCircleIcon',
  finished: 'CheckCircleIcon',
  healthy: 'CheckCircleIcon',
  connected: 'CheckCircleIcon',
  approved: 'CheckCircleIcon',
  success: 'CheckCircleIcon',

  // Error
  failed: 'XCircleIcon',
  error: 'XCircleIcon',
  bad: 'XCircleIcon',
  'access-error': 'XCircleIcon',
  access_error: 'XCircleIcon',
  'timed-out': 'XCircleIcon',
  unknown: 'XCircleIcon',
  unhealthy: 'XCircleIcon',
  'not connected': 'XCircleIcon',
  'not-connected': 'XCircleIcon',
  policy_failed: 'XCircleIcon',
  'failed-pending-retry': 'XCircleIcon',

  // Warn
  'approval-denied': 'WarningIcon',
  'approval-awaiting': 'WarningIcon',
  cancelled: 'WarningIcon',
  outdated: 'WarningIcon',
  warn: 'WarningIcon',

  // Info
  executing: 'Loading',
  waiting: 'Loading',
  started: 'Loading',
  'in-progress': 'Loading',
  building: 'Loading',
  queued: 'Loading',
  planning: 'Loading',
  provisioning: 'Loading',
  syncing: 'Loading',
  deploying: 'Loading',
  available: 'Loading',
  'pending-approval': 'Loading',
  info: 'Loading',

  // Neutral
  noop: 'ClockCountdownIcon',
  inactive: 'ClockCountdownIcon',
  pending: 'ClockCountdownIcon',
  offline: 'ClockCountdownIcon',
  'Not deployed': 'ClockCountdownIcon',
  'No build': 'ClockCountdownIcon',
  deprovisioned: 'ClockCountdownIcon',

  // skipped
  'auto-skipped': 'MinusCircleIcon',
  'user-skipped': 'MinusCircleIcon',
  retried: 'RepeatIcon',

  // Brand
  special: 'ProhibitIcon',
  'not-attempted': 'ProhibitIcon',
  discarded: 'ProhibitIcon',

  // skeleton
  skeleton: 'none' as TIconVariant,
}

export function getStatusTheme(status: string): TStatusTheme {
  return STATUS_THEME_MAP[status] ?? 'neutral'
}

export function getStatusIconVariant(status: string): TIconVariant {
  return STATUS_ICON_MAP[status] ?? 'ClockCountdownIcon'
}

const THEME_PRIORITY: TStatusTheme[] = [
  'error',
  'warn',
  'info',
  'neutral',
  'success',
]

export function getWorstStatusTheme(
  statuses: (string | undefined)[]
): { theme: TStatusTheme; worstStatus: string } {
  const defined = statuses.filter((s): s is string => s !== undefined)
  if (defined.length === 0) return { theme: 'neutral', worstStatus: 'unknown' }

  let worstIdx = THEME_PRIORITY.length
  let worstStatus = defined[0]

  for (const status of defined) {
    const theme = getStatusTheme(status)
    const idx = THEME_PRIORITY.indexOf(theme)
    if (idx < worstIdx) {
      worstIdx = idx
      worstStatus = status
    }
  }

  return {
    theme: THEME_PRIORITY[worstIdx] ?? 'neutral',
    worstStatus,
  }
}
