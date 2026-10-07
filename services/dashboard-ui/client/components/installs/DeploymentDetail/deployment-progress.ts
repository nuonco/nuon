import type {
  TInstallDeploymentAffectedResources,
  TInstallDeploymentRecord,
  TInstallDeploymentSummary,
  TWorkflow,
  TWorkflowStep,
} from '@/types'
import { getStepKind, isRetryChain } from '@/utils/workflow-utils'

export type TDeploymentOutcome = {
  category: 'Stack' | 'Sandbox' | 'Inputs' | 'Secrets' | 'Components' | 'Images'
  name: string
  status: string
  detail: string
  applied?: boolean
}

export const DEPLOYMENT_TABS = {
  template: { path: '/template-updates', text: 'Template updates' },
  workflow: { path: '/workflow', text: 'Workflow' },
  changes: { path: '/', text: 'Change summary' },
}
export type TDeploymentTab = keyof typeof DEPLOYMENT_TABS

export const deploymentTabOrder = (status?: string): TDeploymentTab[] =>
  status === 'success'
    ? ['changes', 'workflow', 'template']
    : ['template', 'workflow', 'changes']

export const ACTIVE_DEPLOYMENT_STATUSES = new Set([
  '',
  'pending',
  'queued',
  'in-progress',
  'retrying',
  'approved',
  'approval-awaiting',
  'failed-pending-retry',
])

export const isDeploymentRunning = (status: string) =>
  [
    'in-progress',
    'retrying',
    'approved',
    'approval-awaiting',
    'awaiting-approval',
  ].includes(status)

export const recoveredDeploymentResources = (
  outcomes: TDeploymentOutcome[],
  previousOutcomes: TDeploymentOutcome[][]
): TDeploymentOutcome['category'][] => [
  ...new Set(
    outcomes.flatMap((outcome) => {
      if (outcome.status !== 'success') return []
      const previous = previousOutcomes
        .map((resources) =>
          resources.find(
            ({ category, name }) =>
              category === outcome.category && name === outcome.name
          )
        )
        .find((resource) => resource !== undefined)
      return previous?.status === 'error' ? [outcome.category] : []
    })
  ),
]

export type TDeploymentStep = {
  id?: string
  name?: string
  status?: { status?: string }
  execution_type?: string
  retried?: boolean
  group_idx?: number
  group_retry_idx?: number
  idx?: number
  step_target_type?: string
  metadata?: { component_name?: string }
  approval?: { type?: string; response?: unknown }
}

export type TDeploymentEvidence<S extends TDeploymentStep = TDeploymentStep> = {
  finished?: boolean
  plan_only?: boolean
  status?: { status?: string }
  steps?: S[]
}

export const summaryDeploymentEvidence = (
  deployment: TInstallDeploymentSummary
): Required<Pick<TDeploymentEvidence, 'finished' | 'status' | 'steps'>> => ({
  finished: !!deployment.finished,
  status: { status: deployment.status },
  steps: deployment.steps.map((step) => ({
    id: step.id,
    name: step.name,
    status: { status: step.status },
    execution_type: step.execution_type,
    retried: step.retried,
    group_idx: step.group_idx || undefined,
    group_retry_idx: step.group_retry_idx ?? 0,
    idx: step.idx || undefined,
    step_target_type: step.step_target_type,
    metadata: step.component_name
      ? { component_name: step.component_name }
      : undefined,
    approval: step.approval_response_id
      ? { response: { id: step.approval_response_id } }
      : undefined,
  })),
})

const isStackStep = (step: TDeploymentStep) =>
  step.step_target_type === 'install_stack_versions' ||
  /\b(install stack|stack policy)\b/i.test(step.name ?? '')

const isSandboxStep = (step: TDeploymentStep) =>
  step.step_target_type === 'install_sandbox_runs' ||
  /\bsandbox\b/i.test(step.name ?? '')

const isInputsStep = (step: TDeploymentStep) =>
  step.name === 'update install state inputs' ||
  /\((pre|post)-update-inputs\)$/.test(step.name ?? '')

const isSecretsStep = (step: TDeploymentStep) =>
  step.name === 'sync secrets' ||
  /\((pre|post)-secrets-sync\)$/.test(step.name ?? '')

export const stepAffectedResources = (
  steps: TDeploymentStep[]
): TInstallDeploymentAffectedResources & {
  inputs: boolean
  secrets: boolean
} => ({
  stack: steps.some(isStackStep),
  sandbox: steps.some(isSandboxStep),
  inputs: steps.some(isInputsStep),
  secrets: steps.some(isSecretsStep),
  components: [
    ...new Set(
      steps.flatMap((step) =>
        step.metadata?.component_name ? [step.metadata.component_name] : []
      )
    ),
  ],
  images: [],
})

export function deploymentSteps(workflow?: TWorkflow): TWorkflowStep[]
export function deploymentSteps<S extends TDeploymentStep>(
  run?: TDeploymentEvidence<S>
): S[]
export function deploymentSteps(run?: TDeploymentEvidence): TDeploymentStep[] {
  const steps = (run?.steps ?? []).filter(
    (step) => step.execution_type !== 'hidden' && !step.retried
  )
  const rounds = new Map<number, number>()
  for (const step of steps) {
    if (step.group_idx != null)
      rounds.set(
        step.group_idx,
        Math.max(rounds.get(step.group_idx) ?? 0, step.group_retry_idx ?? 0)
      )
  }
  const kinds = new Map<string, TDeploymentStep[]>()
  for (const step of steps) {
    if (
      step.group_idx != null &&
      (step.group_retry_idx ?? 0) < rounds.get(step.group_idx)!
    )
      continue
    const key = getStepKind(step)
    kinds.set(key, [...(kinds.get(key) ?? []), step])
  }
  return [...kinds.values()]
    .flatMap((attempts) =>
      isRetryChain(attempts) ? [attempts[attempts.length - 1]] : attempts
    )
    .sort(
      (a, b) =>
        (a.group_idx ?? 0) - (b.group_idx ?? 0) || (a.idx ?? 0) - (b.idx ?? 0)
    )
}

export const isPendingDeploymentStep = (step: TDeploymentStep) =>
  ['pending', 'not-attempted'].includes(step.status?.status ?? '')

export const isCompletedDeploymentStep = (step: TDeploymentStep) =>
  ['success', 'auto-skipped', 'user-skipped'].includes(
    step.status?.status ?? ''
  ) ||
  (step.execution_type === 'approval' && step.status?.status === 'approved')

export const isAwaitingDeploymentRetry = (status?: string) =>
  status === 'failed-pending-retry'

export const isQueuedDeployment = (status: string) =>
  ['pending', 'queued'].includes(status)

export const isAwaitingDeploymentApproval = (step: TDeploymentStep) =>
  step.status?.status === 'approval-awaiting' && !step.approval?.response

const ACTIVE_STEP_DETAILS: Record<string, string> = {
  queued: 'queued',
  'checking-plan': 'checking plan',
  'approval-retry': 're-planning',
}

export const activeStepDetail = (step?: TDeploymentStep) =>
  step?.status?.status && Object.hasOwn(ACTIVE_STEP_DETAILS, step.status.status)
    ? ACTIVE_STEP_DETAILS[step.status.status]
    : undefined

const isActiveDeploymentStep = (step: TDeploymentStep) =>
  step.status?.status === 'in-progress' ||
  !!activeStepDetail(step) ||
  (step.status?.status === 'approval-awaiting' && !!step.approval?.response)

const isOperatorSkippedStep = (step: TDeploymentStep) =>
  step.status?.status === 'user-skipped' &&
  ['system', 'user'].includes(step.execution_type ?? '')

const FAILED_STATUSES = new Set([
  'error',
  'failed-pending-retry',
  'approval-denied',
  'approval-expired',
])

export const deploymentStepContext = <S extends TDeploymentStep>(
  status: string,
  steps: S[]
) => {
  const awaiting = steps.some(isAwaitingDeploymentApproval)
  const current = isDeploymentRunning(status)
    ? steps.find((step) =>
        awaiting
          ? isAwaitingDeploymentApproval(step)
          : isActiveDeploymentStep(step)
      )
    : status === 'success'
      ? steps.findLast(isCompletedDeploymentStep)
      : status === 'cancelled'
        ? (steps.find((step) => step.status?.status === 'cancelled') ??
          steps.findLast((step) => !isPendingDeploymentStep(step)))
        : steps.find((step) => FAILED_STATUSES.has(step.status?.status))
  const currentIndex = current ? steps.indexOf(current) : -1
  const next = steps.find(
    (step, index) => isPendingDeploymentStep(step) && index > currentIndex
  )
  return { current, next }
}

const isAppliedDeploymentStep = (step: TDeploymentStep) =>
  ['system', 'user'].includes(step.execution_type ?? '') &&
  step.status?.status === 'success' &&
  (/^(apply |teardown apply plan |await install stack$|(?:re)?provision sandbox apply( plan)?$|sync secrets$|update install state inputs$)/i.test(
    step.name ?? ''
  ) ||
    (!!step.metadata?.component_name &&
      step.name === `teardown ${step.metadata.component_name}`))

const isImageSyncStep = (step: TDeploymentStep) =>
  ['system', 'user'].includes(step.execution_type ?? '') &&
  step.status?.status === 'success' &&
  !!step.metadata?.component_name &&
  step.name === `sync ${step.metadata.component_name}`

const isHelmRecoveryStep = (step: TDeploymentStep) =>
  ['system', 'user'].includes(step.execution_type ?? '') &&
  step.status?.status === 'success' &&
  !!step.metadata?.component_name &&
  step.name === `recover helm release ${step.metadata.component_name}`

const outcomeForSteps = (
  category: TDeploymentOutcome['category'],
  name: string,
  steps: TDeploymentStep[],
  workflow: TDeploymentEvidence,
  history: TDeploymentStep[]
): TDeploymentOutcome => {
  const failed = steps.find((step) => FAILED_STATUSES.has(step.status?.status))
  const current = steps.find(isActiveDeploymentStep)
  const awaiting = steps.find(isAwaitingDeploymentApproval)
  const cancelled = steps.find((step) => step.status?.status === 'cancelled')
  const skipped = steps.find(isOperatorSkippedStep)
  const applied = !workflow.plan_only && history.some(isAppliedDeploymentStep)
  const currentApplied =
    !workflow.plan_only && steps.some(isAppliedDeploymentStep)
  const completed = steps.length > 0 && steps.every(isCompletedDeploymentStep)
  const stopped =
    workflow.finished || isAwaitingDeploymentRetry(workflow.status?.status)
  if (failed)
    return {
      category,
      name,
      status: 'error',
      detail: `${applied ? 'Applied; ' : ''}${failed.name ?? 'Resource step'} ${
        failed.status?.status === 'approval-denied'
          ? 'denied'
          : failed.status?.status === 'approval-expired'
            ? 'approval expired'
            : 'failed'
      }`,
      applied,
    }
  if (current)
    return {
      category,
      name,
      status: 'in-progress',
      detail:
        current.status?.status === 'approval-awaiting'
          ? 'Plan approved; continuing'
          : activeStepDetail(current)
            ? `${current.name ?? 'Step'} ${activeStepDetail(current)}`
            : (current.name ?? 'In progress'),
      applied,
    }
  if (awaiting)
    return {
      category,
      name,
      status: applied ? 'in-progress' : 'not-started',
      detail: applied
        ? 'Previously applied; awaiting plan approval'
        : 'Not started — awaiting plan approval',
      applied,
    }
  if (cancelled)
    return applied
      ? {
          category,
          name,
          status: 'warn',
          detail: 'Applied; cancelled before finishing',
          applied,
        }
      : { category, name, status: 'cancelled', detail: 'Cancelled' }
  if (skipped)
    return currentApplied
      ? {
          category,
          name,
          status: 'warn',
          detail: `Applied; ${skipped.name ?? 'a step'} skipped`,
          applied,
        }
      : {
          category,
          name,
          status: 'user-skipped',
          detail: 'Skipped by operator',
        }
  if (completed && currentApplied)
    return { category, name, status: 'success', detail: 'Completed' }
  if (completed && !workflow.plan_only && steps.some(isImageSyncStep))
    return { category, name, status: 'success', detail: 'Image synced' }
  if (completed && !workflow.plan_only && steps.some(isHelmRecoveryStep))
    return {
      category,
      name,
      status: 'success',
      detail: 'Helm release recovered',
    }
  if (completed && steps.some((step) => step.status?.status === 'auto-skipped'))
    return { category, name, status: 'success', detail: 'No changes' }
  if (!applied && steps.length && steps.every(isPendingDeploymentStep))
    return { category, name, status: 'not-started', detail: 'Not started' }
  if (applied)
    return {
      category,
      name,
      status: stopped ? 'warn' : 'in-progress',
      detail: 'Applied; remaining steps not completed',
      applied: true,
    }
  if (
    steps.some(
      (step) =>
        step.execution_type === 'approval' &&
        ['success', 'approved'].includes(step.status?.status ?? '')
    )
  )
    return {
      category,
      name,
      status: 'not-started',
      detail: 'Planned; not applied',
    }
  return { category, name, status: 'unknown', detail: 'Outcome unknown' }
}

export const deploymentOutcomes = (
  deployment?: Pick<
    TInstallDeploymentRecord,
    'affected_resources' | 'image' | 'component_name'
  >,
  workflow?: TDeploymentEvidence
): TDeploymentOutcome[] => {
  if (!workflow) return []
  const steps = deploymentSteps(workflow)
  const stepResources = stepAffectedResources(workflow.steps ?? [])
  const recordResources = deployment?.affected_resources
  const hasSteps = steps.length > 0
  const resources = {
    stack: stepResources.stack || (!hasSteps && !!recordResources?.stack),
    sandbox: stepResources.sandbox || (!hasSteps && !!recordResources?.sandbox),
    components: recordResources?.components ?? [],
    images: recordResources?.images ?? [],
  }
  const outcomes: TDeploymentOutcome[] = []
  const resourceOutcome = (
    category: TDeploymentOutcome['category'],
    name: string,
    belongsToResource: (step: TDeploymentStep) => boolean
  ) =>
    outcomeForSteps(
      category,
      name,
      steps.filter(belongsToResource),
      workflow,
      (workflow.steps ?? []).filter(belongsToResource)
    )
  if (resources.stack)
    outcomes.push(resourceOutcome('Stack', 'Stack', isStackStep))
  if (resources.sandbox)
    outcomes.push(resourceOutcome('Sandbox', 'Sandbox', isSandboxStep))
  if (stepResources.inputs)
    outcomes.push(resourceOutcome('Inputs', 'Inputs', isInputsStep))
  if (stepResources.secrets)
    outcomes.push(resourceOutcome('Secrets', 'Secrets', isSecretsStep))
  const componentNames = new Set([
    ...(resources.components ?? []),
    ...steps.flatMap((step) =>
      step.metadata?.component_name ? [step.metadata.component_name] : []
    ),
  ])
  for (const name of componentNames) {
    const componentSteps = steps.filter(
      (step) => step.metadata?.component_name === name
    )
    const groups = new Set(
      componentSteps.map((step) => step.group_idx).filter((idx) => idx != null)
    )
    outcomes.push(
      resourceOutcome(
        'Components',
        name,
        (step) =>
          step.metadata?.component_name === name ||
          (!step.metadata?.component_name &&
            step.group_idx != null &&
            groups.has(step.group_idx))
      )
    )
  }
  for (const name of resources.images ?? []) {
    outcomes.push(
      resourceOutcome(
        'Images',
        name,
        (step) =>
          name === deployment?.image?.repository &&
          !!deployment.component_name &&
          step.metadata?.component_name === deployment.component_name
      )
    )
  }
  return outcomes
}

export const deploymentChangeDescription = (
  workflow: TWorkflow,
  outcomes: TDeploymentOutcome[]
) => {
  if (!workflow.finished && !isAwaitingDeploymentRetry(workflow.status?.status))
    return 'Planned changes. This deployment is still running; these are not final results.'
  if (workflow.plan_only)
    return 'No changes applied. This was a plan-only deployment.'
  if (workflow.status?.status !== 'success' && outcomes.length) {
    if (
      outcomes.some(
        (outcome) => outcome.applied || outcome.status === 'success'
      )
    )
      return 'Some changes were applied. Review resource outcomes for what completed and what failed.'
    if (
      outcomes.every((outcome) =>
        ['error', 'not-started'].includes(outcome.status)
      )
    )
      return 'No completed resource changes were confirmed. A failed apply may have made partial changes.'
  }
  return 'Resource plans from this deployment. Check resource outcomes for what was applied.'
}
