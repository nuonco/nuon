import { getStatusTheme } from '@/utils/status-utils'

export type TStepStatusCategory =
  | 'success'
  | 'error'
  | 'active'
  | 'awaiting'
  | 'pending'

const SUCCESS_STATUSES = new Set(['success', 'succeeded'])
const AWAITING_STATUSES = new Set(['approval-awaiting', 'pending-approval'])

export function stepStatusCategory(status?: string): TStepStatusCategory {
  if (!status) return 'pending'
  if (SUCCESS_STATUSES.has(status)) return 'success'
  if (AWAITING_STATUSES.has(status)) return 'awaiting'

  const theme = getStatusTheme(status)
  if (theme === 'success') return 'success'
  if (theme === 'error') return 'error'
  if (theme === 'info') return 'active'
  return 'pending'
}

export function isActiveStepStatus(status?: string): boolean {
  const category = stepStatusCategory(status)
  return category === 'active' || category === 'awaiting'
}
