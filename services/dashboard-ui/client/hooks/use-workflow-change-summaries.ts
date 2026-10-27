import { useMemo } from 'react'
import { emptyCounts } from '@/components/workflows/WorkflowChangesSummary/change-summary-utils'
import { inputChangeSummaries } from '@/components/workflows/WorkflowInputChanges'
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

// Sandbox provisioning always plans through terraform (see
// services/ctl-api/internal/app/installs/worker/sandbox/execute_plan.go),
// so these can be labeled correctly before the approval exists.
const SANDBOX_PLAN_STEP_NAMES = new Set([
  'provision sandbox plan',
  'deprovision sandbox plan',
  'reprovision sandbox plan',
])

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

// execution_type 'approval' is assigned at step-generation time, before the
// workflow runs (see getSignalStepMetadata in
// services/ctl-api/internal/app/installs/workflows/v2/shared.go), so it's a
// reliable signal for "this step will plan" even if the workflow never reaches it.
export const toPendingChangeSummary = (
  step: TWorkflowStep
): TStepChangeSummary | null => {
  if (step?.execution_type !== 'approval' || step?.approval || step?.retried) {
    return null
  }

  const planType: TStepChangePlanType = SANDBOX_PLAN_STEP_NAMES.has(
    step.name ?? ''
  )
    ? 'terraform_plan'
    : 'pending'

  return {
    stepId: step.id!,
    stepName: step.name ?? '',
    componentName: step.metadata?.component_name,
    planType,
    status: 'not-generated',
    counts: emptyCounts(),
    countsState: 'unsupported',
    hasDetail: false,
  }
}

export function useWorkflowChangeSummaries(workflow?: TWorkflow) {
  return useMemo(() => {
    const steps = workflow?.steps ?? []
    const planSummaries = steps
      .map((step) => toChangeSummary(step) ?? toPendingChangeSummary(step))
      .filter((summary): summary is TStepChangeSummary => summary !== null)
    return [...planSummaries, ...inputChangeSummaries(workflow)]
  }, [workflow])
}
