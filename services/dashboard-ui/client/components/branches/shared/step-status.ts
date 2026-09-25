import { getStatusTheme } from '@/utils/status-utils'
import type { TInstallWorkflowStep } from '@/types'

export type TStepStatusCategory =
  | 'success'
  | 'error'
  | 'active'
  | 'awaiting'
  | 'pending'

const SUCCESS_STATUSES = new Set(['success', 'succeeded'])
const AWAITING_STATUSES = new Set(['approval-awaiting', 'pending-approval'])

const RESPONSE_STATUS: Record<string, string> = {
  approve: 'approved',
  'auto-approve': 'approved',
  deny: 'approval-denied',
  'deny-skip-current': 'user-skipped',
  'deny-skip-current-and-dependents': 'user-skipped',
  skip: 'user-skipped',
  'auto-skipped': 'auto-skipped',
  retry: 'retried',
}

/** Prefer approval response over a stale approval-awaiting step status. */
export function getStepDisplayStatus(step: TInstallWorkflowStep): string {
  const responseType = step.approval?.response?.type
  if (responseType && RESPONSE_STATUS[responseType]) {
    return RESPONSE_STATUS[responseType]
  }
  if (step.approval?.response && step.status?.status === 'approval-awaiting') {
    return 'approved'
  }
  return step.status?.status || 'pending'
}

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
