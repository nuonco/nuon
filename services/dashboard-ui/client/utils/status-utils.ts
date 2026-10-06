export type TStatusTheme =
  | 'success'
  | 'warn'
  | 'neutral'
  | 'error'
  | 'info'
  | 'brand'

export type TStatusIconVariant =
  | 'CheckCircleIcon'
  | 'XCircleIcon'
  | 'WarningIcon'
  | 'ClockCountdownIcon'
  | 'MinusCircleIcon'
  | 'ProhibitIcon'
  | 'RepeatIcon'
  | 'Loading'
  | 'none'

const STATUS_THEME_MAP: Record<string, TStatusTheme> = {
  active: 'success',
  ok: 'success',
  finished: 'success',
  healthy: 'success',
  connected: 'success',
  verified: 'success',
  approved: 'success',
  success: 'success',

  failed: 'error',
  error: 'error',
  bad: 'error',
  'access-error': 'error',
  'timed-out': 'error',
  unhealthy: 'error',
  'not-connected': 'error',
  suspended: 'error',
  'policy-failed': 'error',
  'failed-pending-retry': 'error',
  'health-failed': 'error',

  'approval-denied': 'neutral',
  'approval-awaiting': 'warn',
  'awaiting-approval': 'warn',
  'pending-customer': 'warn',
  cancelled: 'neutral',
  rejected: 'neutral',
  outdated: 'warn',
  warn: 'warn',
  warning: 'warn',
  drifted: 'warn',
  expired: 'warn',
  'pending-shutdown': 'warn',
  'shutting-down': 'warn',
  offline: 'warn',
  degraded: 'warn',

  executing: 'info',
  running: 'info',
  waiting: 'info',
  started: 'info',
  'in-progress': 'info',
  building: 'info',
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
  'awaiting-install-stack-run': 'info',
  'awaiting-heartbeat': 'info',
  deprovisioning: 'info',
  reprovisioning: 'info',
  progressing: 'info',

  noop: 'neutral',
  'shut-down': 'neutral',
  unknown: 'neutral',
  inactive: 'neutral',
  disabled: 'neutral',
  pending: 'brand',
  queued: 'brand',
  invited: 'neutral',
  'not-deployed': 'neutral',
  'no-build': 'neutral',
  'not-attempted': 'neutral',
  deprovisioned: 'warn',
  skeleton: 'neutral',

  special: 'brand',
  brand: 'brand',
}

const STATUS_ICON_MAP: Record<string, TStatusIconVariant> = {
  active: 'CheckCircleIcon',
  ok: 'CheckCircleIcon',
  finished: 'CheckCircleIcon',
  healthy: 'CheckCircleIcon',
  connected: 'CheckCircleIcon',
  verified: 'CheckCircleIcon',
  approved: 'CheckCircleIcon',
  success: 'CheckCircleIcon',

  failed: 'XCircleIcon',
  error: 'XCircleIcon',
  bad: 'XCircleIcon',
  'access-error': 'XCircleIcon',
  'timed-out': 'XCircleIcon',
  unknown: 'XCircleIcon',
  unhealthy: 'XCircleIcon',
  'not-connected': 'XCircleIcon',
  'policy-failed': 'XCircleIcon',
  'failed-pending-retry': 'XCircleIcon',
  'health-failed': 'XCircleIcon',

  'approval-denied': 'ProhibitIcon',
  'approval-awaiting': 'WarningIcon',
  'awaiting-approval': 'WarningIcon',
  cancelled: 'ProhibitIcon',
  rejected: 'ProhibitIcon',
  outdated: 'WarningIcon',
  warn: 'WarningIcon',
  warning: 'WarningIcon',
  degraded: 'WarningIcon',

  executing: 'Loading',
  running: 'Loading',
  waiting: 'Loading',
  started: 'Loading',
  'in-progress': 'Loading',
  building: 'Loading',
  planning: 'Loading',
  provisioning: 'Loading',
  syncing: 'Loading',
  deploying: 'Loading',
  available: 'Loading',
  'pending-approval': 'Loading',
  'awaiting-install-stack-run': 'Loading',
  'awaiting-heartbeat': 'Loading',
  info: 'Loading',
  deprovisioning: 'Loading',
  reprovisioning: 'Loading',
  progressing: 'Loading',

  noop: 'ClockCountdownIcon',
  inactive: 'WarningIcon',
  disabled: 'ProhibitIcon',
  pending: 'ClockCountdownIcon',
  queued: 'ClockCountdownIcon',
  offline: 'ClockCountdownIcon',
  'not-deployed': 'ClockCountdownIcon',
  'no-build': 'ClockCountdownIcon',
  deprovisioned: 'WarningIcon',

  'not-applicable': 'MinusCircleIcon',

  'auto-skipped': 'MinusCircleIcon',
  'user-skipped': 'MinusCircleIcon',
  retried: 'RepeatIcon',

  special: 'ProhibitIcon',
  'not-attempted': 'ProhibitIcon',
  discarded: 'ProhibitIcon',

  skeleton: 'none',
}

const normalizeStatusKey = (status: string): string =>
  status.toLowerCase().replace(/[\s_]+/g, '-')

export function getStatusTheme(status: string): TStatusTheme {
  return STATUS_THEME_MAP[normalizeStatusKey(status)] ?? 'neutral'
}

export function getStatusIconVariant(status: string): TStatusIconVariant {
  return STATUS_ICON_MAP[normalizeStatusKey(status)] ?? 'ClockCountdownIcon'
}

const THEME_PRIORITY: TStatusTheme[] = [
  'error',
  'warn',
  'info',
  'brand',
  'neutral',
  'success',
]

export function getWorstStatusTheme(statuses: (string | undefined)[]): {
  theme: TStatusTheme
  worstStatus: string
} {
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
