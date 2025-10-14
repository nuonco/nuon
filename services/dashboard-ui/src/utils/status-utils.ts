import type { TIconVariant } from '@/components/common/Icon'

type TStatusTheme = 'success' | 'warn' | 'neutral' | 'error' | 'info' | 'brand'

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
  unknown: 'error',
  unhealthy: 'error',
  'not connected': 'error',
  'not-connected': 'error',

  // Warn
  'approval-denied': 'warn',
  'approval-awaiting': 'warn',
  cancelled: 'warn',
  outdated: 'warn',
  warn: 'warn',
  drifted: 'warn',

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

  // Neutral
  noop: 'neutral',
  inactive: 'neutral',
  pending: 'neutral',
  offline: 'neutral',
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
  active: 'CheckCircle',
  ok: 'CheckCircle',
  finished: 'CheckCircle',
  healthy: 'CheckCircle',
  connected: 'CheckCircle',
  approved: 'CheckCircle',
  success: 'CheckCircle',

  // Error
  failed: 'XCircle',
  error: 'XCircle',
  bad: 'XCircle',
  'access-error': 'XCircle',
  access_error: 'XCircle',
  'timed-out': 'XCircle',
  unknown: 'XCircle',
  unhealthy: 'XCircle',
  'not connected': 'XCircle',
  'not-connected': 'XCircle',

  // Warn
  'approval-denied': 'Warning',
  'approval-awaiting': 'Warning',
  cancelled: 'Warning',
  outdated: 'Warning',
  warn: 'Warning',

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
  noop: 'ClockCountdown',
  inactive: 'ClockCountdown',
  pending: 'ClockCountdown',
  offline: 'ClockCountdown',
  'Not deployed': 'ClockCountdown',
  'No build': 'ClockCountdown',
  deprovisioned: 'ClockCountdown',

  // skipped
  'user-skipped': 'MinusCircle',
  retried: 'Repeat',

  // Brand
  special: 'Prohibit',
  'not-attempted': 'Prohibit',
  discarded: 'Prohibit',

  // skeleton
  skeleton: 'none' as TIconVariant,
}

export function getStatusTheme(status: string): TStatusTheme {
  return STATUS_THEME_MAP[status] ?? 'neutral'
}

export function getStatusIconVariant(status: string): TIconVariant {
  return STATUS_ICON_MAP[status] ?? 'ClockCountdown'
}
