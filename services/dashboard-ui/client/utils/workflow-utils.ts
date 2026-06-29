import type { TBadgeTheme } from '@/components/common/Badge'
import type { TBannerTheme } from '@/components/common/Banner'
import type { TWorkflow, TWorkflowStep, TWorkflowStepApproval } from '@/types'

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
  'auto-approved': { children: 'Auto approved', theme: 'neutral' },
  approved: { children: 'Plan approved', theme: 'success' },
  'approval-awaiting': { children: 'Awaiting approval', theme: 'warn' },
  'approval-denied': { children: 'Plan denied', theme: 'warn' },
  'approval-retry': { children: 'Plan retried', theme: 'info' },
  error: { children: 'Failed', theme: 'error' },
  'failed-pending-retry': { children: 'Failed — awaiting retry', theme: 'error' },
  'not-attempted': { children: 'Not attempted' },
  noop: { children: 'NOOP' },
  cancelled: { children: 'Cancelled', theme: 'warn' },
  'stale-plan': { children: 'Plan stale', theme: 'warn' },
  superseded: { children: 'Plan superseded', theme: 'warn' },
}

export function getWorkflowBadge(workflow: TWorkflow): TBadgeCfg {
  const status = workflow?.status?.status
  // fallback to empty object if status not found
  return status && WORKFLOW_BADGE_MAP[status] ? WORKFLOW_BADGE_MAP[status] : {}
}

// The retry "lineage" badge — how this attempt relates to its retries. Distinct
// from the outcome badge (below) so a retry row can show both.
function getStepLineageBadge(step: TWorkflowStep): TBadgeCfg | undefined {
  const metadata = step?.status?.metadata

  if (metadata?.auto_retried) {
    return { children: 'Auto-retried', theme: 'info' }
  }

  // An approved plan that was later retried (e.g. its apply failed and the
  // whole group re-ran) is no longer valid — surface that the approval was
  // superseded rather than a generic "Retried".
  const wasApprovedPlan =
    step?.execution_type === 'approval' &&
    (metadata?.status === 'approved' ||
      step?.approval?.response?.type === 'approve')
  if (step?.retried && wasApprovedPlan) {
    return { children: 'Plan superseded', theme: 'warn' }
  }

  if (step?.retried) {
    return { children: 'Retried', theme: 'info' }
  }

  if (metadata?.is_retry) {
    const retryIdx = Number(metadata.retry_idx ?? metadata.group_retry_idx ?? 0)
    const retryLabel = metadata.retry_type === 'manual' ? 'Manual retry' : metadata.retry_type === 'auto' ? 'Auto retry' : 'Retry'
    return { children: `${retryLabel} #${retryIdx}`, theme: 'info' }
  }

  return undefined
}

// The step's outcome badge — its terminal result (Completed / Failed / Plan
// approved / etc.), independent of any retry lineage.
function getStepOutcomeBadge(
  step: TWorkflowStep,
  isApprovalPrompt?: boolean,
  planOnly?: boolean
): TBadgeCfg {
  if (step?.execution_type === 'skipped') {
    return { children: 'Skipped' }
  }

  if (step?.status?.status === 'auto-skipped') {
    return { children: 'Auto skipped' }
  }

  if (step?.execution_type === 'approval' && !isApprovalPrompt) {
    const metadata = step?.status?.metadata as Record<string, unknown> | undefined
    if (metadata?.auto_approved && metadata?.check === 'policy-auto-approve') {
      return { children: 'Auto-approved (policies)', theme: 'success' as TBadgeTheme }
    }
    if (step?.status?.status === 'approved') {
      if (planOnly) {
        return { children: 'Plan created', theme: 'success' }
      }
      return WORKFLOW_BADGE_MAP['approved']
    } else if (step?.status?.status === 'approval-awaiting') {
      return WORKFLOW_BADGE_MAP['approval-awaiting']
    } else if (step?.status?.status === 'pending') {
      return WORKFLOW_BADGE_MAP['auto-approved']
    } else if (
      step?.approval?.type === 'approve-all' ||
      step?.approval?.response?.type === 'auto-approve'
    ) {
      return WORKFLOW_BADGE_MAP['auto-approved']
    }
  }
  const status = step?.status?.status
  const checkType = (step?.status?.metadata as Record<string, unknown>)?.check

  if (status === 'error' && (checkType === 'stale-plan' || checkType === 'superseded')) {
    return WORKFLOW_BADGE_MAP[checkType === 'stale-plan' ? 'stale-plan' : 'superseded']
  }

  // Show retryable/skippable hints for failed steps awaiting user action.
  if (status === 'failed-pending-retry' || (status === 'error' && step?.retryable && !step?.retried && !step?.status?.metadata?.retries_exhausted)) {
    const hints: string[] = []
    if (step?.retryable) hints.push('retryable')
    if (step?.skippable) hints.push('skippable')
    const suffix = hints.length > 0 ? ` (${hints.join(' / ')})` : ''
    return { children: `Failed${suffix}`, theme: 'error' as TBadgeTheme }
  }

  return status && WORKFLOW_BADGE_MAP[status] ? WORKFLOW_BADGE_MAP[status] : {}
}

// Badges for a step, in render order: the retry lineage badge (if any) followed
// by the outcome badge, so a retry attempt shows both "Retry #2" and its result.
export function getStepBadges(
  step: TWorkflowStep,
  isApprovalPrompt?: boolean,
  planOnly?: boolean
): TBadgeCfg[] {
  const lineage = getStepLineageBadge(step)

  // "Plan superseded" already conveys the outcome — don't double up.
  if (lineage?.children === 'Plan superseded') {
    return [lineage]
  }

  const outcome = getStepOutcomeBadge(step, isApprovalPrompt, planOnly)
  const badges = [lineage, outcome].filter(
    (badge): badge is TBadgeCfg => !!badge?.children
  )

  return badges.filter(
    (badge, idx) => badges.findIndex((b) => b.children === badge.children) === idx
  )
}

export function getStepBadge(
  step: TWorkflowStep,
  isApprovalPrompt?: boolean,
  planOnly?: boolean
): TBadgeCfg {
  return getStepBadges(step, isApprovalPrompt, planOnly)[0] ?? {}
}
/**
 * A step's logical "kind" — stable across retry attempts. Retrying a step
 * creates a new step row that shares the same group and name as the prior
 * attempt(s), so callers can group attempts together (e.g. to only show retry
 * controls on the latest attempt of a kind).
 */
export function getStepKind(step: TWorkflowStep): string {
  return `${step?.group_idx ?? ''}:${step?.step_target_type ?? ''}:${step?.name ?? step?.id ?? ''}`
}

export type TStepButtonsCfg = {
  cancel: boolean
  approval: boolean
  retry: boolean
}

export function getStepButtons(step: TWorkflowStep): TStepButtonsCfg {
  const status = step?.status?.status
  // A step that has already been superseded by a newer retry attempt
  // (`retried === true`) should never offer retry/skip controls.
  const isFailedAwaitingRetry = status === 'failed-pending-retry' && !step?.retried
  const isRetryableError = status === 'error' && !!step?.retryable && !step?.retried && !step?.status?.metadata?.retries_exhausted && step?.status?.metadata?.retry_type !== 'auto'
  return {
    retry: isFailedAwaitingRetry || isRetryableError,
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

  if (status === 'failed-pending-retry') {
    const hints: string[] = []
    if (step?.retryable) hints.push('retry')
    if (step?.skippable) hints.push('skip')
    const actions = hints.length > 0 ? ` You can ${hints.join(' or ')} this step.` : ''
    return {
      copy: `Step failed and is awaiting action.${actions}`,
      theme: 'error',
      title: `Step ${step?.name} failed — awaiting retry`,
    }
  }

  if (status === 'error') {
    const metadata = step?.status?.metadata
    if (metadata?.check === 'stale-plan') {
      return {
        copy: `${metadata?.detail ?? 'The plan was approved after it became stale.'}  The step will be automatically retried with a fresh plan.`,
        theme: 'warn',
        title: `Step ${step?.name} — plan stale`,
      }
    }
    if (metadata?.check === 'superseded') {
      return {
        copy: `${metadata?.detail ?? 'A newer workflow was approved, making this plan outdated.'} The step will be automatically retried with a fresh plan.`,
        theme: 'warn',
        title: `Step ${step?.name} — plan superseded`,
      }
    }
    if (metadata?.retries_exhausted) {
      return {
        copy: `This step has used all ${metadata.max_retries ?? ''} retry attempts. No further retries are possible. Rerun the workflow to start fresh.`,
        theme: 'error',
        title: `Step ${step?.name} failed — retries exhausted (${metadata.retry_index ?? '?'}/${metadata.max_retries ?? '?'})`,
      }
    }
    if (metadata?.auto_retried) {
      const attempt = typeof metadata.retry_idx === 'number' ? metadata.retry_idx : '?'
      return {
        copy: `Step encountered an error and was automatically retried (attempt ${attempt} of ${metadata.max_retries ?? '?'}).`,
        theme: 'warn',
        title: `Step ${step?.name} — auto-retried`,
      }
    }
    if (metadata?.auto_retries_exhausted) {
      const maxAuto = Number(metadata.max_auto_retries ?? 0)
      const autoRetryCopy = maxAuto > 0
        ? `All ${maxAuto} automatic retries have been used.`
        : 'No automatic retries are configured for this step.'
      const manualRetryCopy = metadata.max_retries
        ? ` You can still manually retry this step (${metadata.retry_index ?? 0} of ${metadata.max_retries} total retries used).`
        : ''
      return {
        copy: `${autoRetryCopy}${manualRetryCopy}`,
        theme: 'warn',
        title: `Step ${step?.name} — auto-retries exhausted`,
      }
    }
    const retryInfo =
      metadata?.retry_type
        ? ` (${metadata.retry_type} retry ${metadata.retry_idx ?? ''}/${metadata.max_retries ?? ''})`
        : ''
    return {
      copy: `Step encountered an error: ${status_human_description}${retryInfo}`,
      theme: 'error',
      title: `Step ${step?.name} failed`,
    }
  }

  if (status === 'cancelled') {
    return {
      copy: `Step was cancelled: ${status_human_description}`,
      theme: 'warn',
      title: `Step ${step?.name} cancelled`,
    }
  }

  if (status === 'discarded') {
    return {
      copy: `Step was discarded: ${status_human_description}`,
      theme: 'default',
      title: `Step ${step?.name} discarded`,
    }
  }

  if (status === 'user-skipped') {
    return {
      copy: `Step was skipped by ${email}: ${status_human_description}`,
      theme: 'default',
      title: `Step ${step?.name} skipped`,
    }
  }

  if (status === 'not-attempted') {
    return {
      copy: `Step was not attempted ${status_human_description ? `: ${status_human_description}` : ''}`,
      theme: 'default',
      title: `Step ${step?.name} not attempted`,
    }
  }

  if (step.execution_type === 'skipped') {
    return {
      copy: `Step was skipped due to being a plan only workflow`,
      theme: 'default',
      title: `Step ${step?.name} skipped`,
    }
  }

  const wasApprovedPlan =
    step?.execution_type === 'approval' &&
    (step?.status?.metadata?.status === 'approved' ||
      step?.approval?.response?.type === 'approve')
  if (step?.retried && wasApprovedPlan) {
    return {
      copy: 'This plan was approved, but a later step in the group failed and the group was retried with a fresh plan. This approval no longer applies.',
      theme: 'warn',
      title: `Step ${step?.name} — plan superseded`,
    }
  }

  if (step?.retryable && step?.retried) {
    const metadata = step?.status?.metadata
    const retryType = metadata?.retry_type ? `${metadata.retry_type} ` : ''
    return {
      copy: `Step was ${retryType}retried by ${email}: ${status_human_description}`,
      theme: 'info',
      title: `Step ${step?.name} retried`,
    }
  }

  return undefined
}

export function getWorkflowStep({
  workflow,
  stepTargetId,
}: {
  workflow: TWorkflow
  stepTargetId: string
}): TWorkflowStep | null {
  return workflow
    ? workflow?.steps?.filter((s) => s?.step_target_id === stepTargetId)?.at(-1)
    : null
}

export interface PolicyViolation {
  policy_id: string
  message: string
  severity: 'deny' | 'warn'
}

export interface PolicyViolationCounts {
  denyViolations: PolicyViolation[]
  warnViolations: PolicyViolation[]
  passedPolicyIds: string[]
  denyCount: number
  warnCount: number
  passedCount: number
  hasPolicyData: boolean
  hasViolations: boolean
}

export function getPolicyViolationCounts(
  step: TWorkflowStep
): PolicyViolationCounts {
  const metadata = step?.status?.metadata
  const denyViolations = (metadata?.deny_violations as PolicyViolation[]) || []
  const warnViolations = (metadata?.warn_violations as PolicyViolation[]) || []
  const passedPolicyIds = (metadata?.passed_policy_ids as string[]) || []
  const denyCount = denyViolations.length
  const warnCount = warnViolations.length
  const passedCount = passedPolicyIds.length
  const hasPolicyData =
    metadata?.deny_violations !== undefined ||
    metadata?.warn_violations !== undefined ||
    metadata?.passed_policy_ids !== undefined
  const hasViolations = denyCount > 0 || warnCount > 0

  return {
    denyViolations,
    warnViolations,
    passedPolicyIds,
    denyCount,
    warnCount,
    passedCount,
    hasPolicyData,
    hasViolations,
  }
}

export function getWorkflowPendingApprovals(
  approvals: TWorkflowStepApproval[] | undefined,
  workflowId: string | undefined
): TWorkflowStepApproval[] {
  if (!workflowId) return []
  return (
    approvals?.filter(
      (a) => a?.workflow_step?.install_workflow_id === workflowId
    ) ?? []
  )
}
