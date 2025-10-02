import type { TBadgeTheme } from '@/components/common/Badge'
import type { TBannerTheme } from '@/components/common/Banner'
import type { TWorkflow, TWorkflowStep } from '@/types'

export type TBadgeCfg = {
  children?: string
  theme?: TBadgeTheme
}

const WORKFLOW_BADGE_MAP: Record<
  string,
  { children: string; theme?: TBadgeTheme }
> = {
  'user-skipped': { children: 'Skipped' },
  discarded: { children: 'Discarded' },
  success: { children: 'Completed', theme: 'success' },
  approved: { children: 'Plan approved', theme: 'success' },
  'approval-awaiting': { children: 'Awaiting approval', theme: 'warn' },
  'approval-denied': { children: 'Plan denied', theme: 'warn' },
  'approval-retry': { children: 'Plan retried', theme: 'info' },
  error: { children: 'Failed', theme: 'error' },
  'not-attempted': { children: 'Not attempted' },
  cancelled: { children: 'Cancelled', theme: 'warn' },
}

export function getWorkflowBadge(workflow: TWorkflow): TBadgeCfg {
  const status = workflow?.status?.status
  // fallback to empty object if status not found
  return status && WORKFLOW_BADGE_MAP[status] ? WORKFLOW_BADGE_MAP[status] : {}
}

export function getStepBadge(step: TWorkflowStep): TBadgeCfg {
  if (step?.retried) {
    return { children: 'Retried', theme: 'info' }
  }
  if (step?.execution_type === 'skipped') {
    return { children: 'Skipped' }
  }
  const status = step?.status?.status
  return status && WORKFLOW_BADGE_MAP[status] ? WORKFLOW_BADGE_MAP[status] : {}
}
export type TStepButtonsCfg = {
  cancel: boolean
  approval: boolean
  retry: boolean
}

export function getStepButtons(step: TWorkflowStep): TStepButtonsCfg {
  const status = step?.status?.status
  return {
    retry: status === 'error' && !!step?.retryable && !step?.retried,
    cancel: status === 'in-progress' || status === 'approval-awaiting',
    approval: status === 'approval-awaiting',
  }
}

export type TStepBannerCfg = {
  copy: string
  theme: TBannerTheme
  title: string
}

export function getStepBanner(step: TWorkflowStep): TStepBannerCfg | undefined {
  if (!step?.status?.status) return undefined

  const { status, status_human_description } = step.status
  const email = step?.created_by?.email

  if (status === 'error') {
    return {
      copy: `Step encountered an error: ${status_human_description}`,
      theme: 'error',
      title: 'Step failed',
    }
  }

  if (status === 'cancelled') {
    return {
      copy: `Step was cancelled: ${status_human_description}`,
      theme: 'warn',
      title: 'Step cancelled',
    }
  }

  if (status === 'discarded') {
    return {
      copy: `Step was discarded: ${status_human_description}`,
      theme: 'default',
      title: 'Step discarded',
    }
  }

  if (status === 'user-skipped') {
    return {
      copy: `Step was skipped by ${email}: ${status_human_description}`,
      theme: 'default',
      title: 'Step skipped',
    }
  }

  if (step.execution_type === 'skipped') {
    return {
      copy: `Step was skipped due to being a plan only workflow`,
      theme: 'default',
      title: 'Step skipped',
    }
  }

  if (step?.retryable && step?.retried) {
    return {
      copy: `Step was retried by ${email}: ${status_human_description}`,
      theme: 'info',
      title: 'Step retried',
    }
  }

  return undefined
}
