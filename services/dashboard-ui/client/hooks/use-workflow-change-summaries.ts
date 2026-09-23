import { useMemo } from 'react'
import type {
  TStepChangePlanType,
  TStepChangeState,
  TStepChangeStatus,
  TStepChangeSummary,
  TWorkflow,
  TWorkflowStep,
} from '@/types'

const SUMMARY_PLAN_TYPES: TStepChangePlanType[] = [
  'terraform_plan',
  'pulumi_plan',
  'helm_approval',
  'kubernetes_manifest_approval',
  'app_branch_plan',
  'install_creation',
]

const STEP_STATUS_MAP: Record<string, TStepChangeStatus> = {
  'approval-awaiting': 'pending-approval',
  approved: 'approved',
  'approval-denied': 'denied',
  success: 'applied',
  'user-skipped': 'applied',
  'auto-skipped': 'applied',
  discarded: 'applied',
  error: 'error',
  cancelled: 'error',
  'failed-pending-retry': 'error',
}

const toChangeStatus = (status?: string): TStepChangeStatus =>
  STEP_STATUS_MAP[status ?? ''] ?? 'generating'

const toCountsState = (state?: string): TStepChangeState =>
  state === 'ok' || state === 'unsupported' ? state : 'unknown'

const isSummaryPlanType = (type?: string): type is TStepChangePlanType =>
  SUMMARY_PLAN_TYPES.includes(type as TStepChangePlanType)

export const toChangeSummary = (
  step: TWorkflowStep
): TStepChangeSummary | null => {
  const approval = step?.approval
  if (!approval || !isSummaryPlanType(approval.type)) return null

  return {
    stepId: step.id!,
    stepName: step.name ?? '',
    componentName: step.metadata?.component_name,
    planType: approval.type,
    status: toChangeStatus(step.status?.status),
    counts: {
      create: approval.changes_create ?? 0,
      update: approval.changes_update ?? 0,
      delete: approval.changes_delete ?? 0,
      replace: approval.changes_replace ?? 0,
      noop: approval.changes_noop ?? 0,
    },
    countsState: toCountsState(approval.changes_state),
    hasDetail: approval.type !== 'install_creation',
  }
}

export function useWorkflowChangeSummaries(workflow?: TWorkflow) {
  return useMemo(() => {
    const steps = workflow?.steps ?? []
    return steps
      .map(toChangeSummary)
      .filter((summary): summary is TStepChangeSummary => summary !== null)
  }, [workflow?.steps])
}
