import type { TStatusTheme } from '@/utils/status-utils'
import type { TIconVariant } from '../components/atoms/Icon'

export type TSurfaceTheme = TStatusTheme | 'default'

export const STATUS_THEME_COLOR: Record<TSurfaceTheme, string> = {
  default: 'var(--text-tertiary)',
  success: 'var(--status-success)',
  error: 'var(--status-error)',
  warn: 'var(--status-warn)',
  info: 'var(--status-info)',
  brand: 'var(--status-brand)',
  neutral: 'var(--status-neutral)',
}

export const STATUS_THEME_ICON: Record<TSurfaceTheme, TIconVariant> = {
  default: 'InfoIcon',
  success: 'CheckCircleIcon',
  error: 'XCircleIcon',
  warn: 'WarningIcon',
  info: 'ClockCountdownIcon',
  brand: 'SparkleIcon',
  neutral: 'MinusCircleIcon',
}
