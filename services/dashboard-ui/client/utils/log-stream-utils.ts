import type { TBadgeTheme } from '@/components/common/Badge'

export function getSeverityBgClasses(severityNumber: number): string {
  if (severityNumber <= 4) return 'bg-primary-400 dark:bg-primary-300'
  if (severityNumber <= 8) return 'bg-cool-grey-600 dark:bg-cool-grey-500'
  if (severityNumber <= 12) return 'bg-blue-600 dark:bg-blue-500'
  if (severityNumber <= 16) return 'bg-orange-600 dark:bg-orange-500'
  if (severityNumber <= 20) return 'bg-red-600 dark:bg-red-500'
  if (severityNumber <= 24) return 'bg-red-700 dark:bg-red-600'

  // Fallback for values > 24
  return 'bg-red-700 dark:bg-red-600'
}

type TBorderPosition = 't' | 'l' | 'b' | 'r' | ''
type TSeverityLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal'

const SEVERITY_BORDERS = {
  // Top borders
  t: {
    trace: '!border-t-primary-400 dark:!border-t-primary-300',
    debug: '!border-t-cool-grey-600 dark:!border-t-cool-grey-500',
    info: '!border-t-blue-600 dark:!border-t-blue-500',
    warn: '!border-t-orange-600 dark:!border-t-orange-500',
    error: '!border-t-red-600 dark:!border-t-red-500',
    fatal: '!border-t-red-700 dark:!border-t-red-600',
  },
  // Left borders
  l: {
    trace: '!border-l-primary-400 dark:!border-l-primary-300',
    debug: '!border-l-cool-grey-600 dark:!border-l-cool-grey-500',
    info: '!border-l-blue-600 dark:!border-l-blue-500',
    warn: '!border-l-orange-600 dark:!border-l-orange-500',
    error: '!border-l-red-600 dark:!border-l-red-500',
    fatal: '!border-l-red-700 dark:!border-l-red-600',
  },
  // Bottom borders
  b: {
    trace: '!border-b-primary-400 dark:!border-b-primary-300',
    debug: '!border-b-cool-grey-600 dark:!border-b-cool-grey-500',
    info: '!border-b-blue-600 dark:!border-b-blue-500',
    warn: '!border-b-orange-600 dark:!border-b-orange-500',
    error: '!border-b-red-600 dark:!border-b-red-500',
    fatal: '!border-b-red-700 dark:!border-b-red-600',
  },
  // Right borders
  r: {
    trace: '!border-r-primary-400 dark:!border-r-primary-300',
    debug: '!border-r-cool-grey-600 dark:!border-r-cool-grey-500',
    info: '!border-r-blue-600 dark:!border-r-blue-500',
    warn: '!border-r-orange-600 dark:!border-r-orange-500',
    error: '!border-r-red-600 dark:!border-r-red-500',
    fatal: '!border-r-red-700 dark:!border-r-red-600',
  },
  // Standard borders (no position)
  '': {
    trace: '!border-primary-400 dark:!border-primary-300',
    debug: '!border-cool-grey-600 dark:!border-cool-grey-500',
    info: '!border-blue-600 dark:!border-blue-500',
    warn: '!border-orange-600 dark:!border-orange-500',
    error: '!border-red-600 dark:!border-red-500',
    fatal: '!border-red-700 dark:!border-red-600',
  },
} as const

export function getSeverityBorderClasses(
  severityNumber: number,
  position: TBorderPosition = ''
): string {
  // Map severity number to severity level name
  let severityLevel: TSeverityLevel
  if (severityNumber <= 4) severityLevel = 'trace'
  else if (severityNumber <= 8) severityLevel = 'debug'
  else if (severityNumber <= 12) severityLevel = 'info'
  else if (severityNumber <= 16) severityLevel = 'warn'
  else if (severityNumber <= 20) severityLevel = 'error'
  else severityLevel = 'fatal'

  return SEVERITY_BORDERS[position][severityLevel]
}

export function getBadgeThemeFromSeverity(severityNumber: number): TBadgeTheme {
  if (severityNumber <= 4) return 'brand'
  if (severityNumber <= 8) return 'neutral'
  if (severityNumber <= 12) return 'info'
  if (severityNumber <= 16) return 'warn'
  if (severityNumber <= 20) return 'error'
  if (severityNumber <= 24) return 'error'

  // Fallback for values > 24
  return 'error'
}
