import type {
  TInstallDeploymentRecord,
  TWorkflow,
  TWorkflowStep,
} from '@/types'
import { getStepKind, isRetryChain } from '@/utils/workflow-utils'

export type TDeploymentOutcome = {
  category: 'Stack' | 'Sandbox' | 'Components' | 'Images'
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

export const isDeploymentRunning = (status: string) =>
  ['in-progress', 'approval-awaiting', 'awaiting-approval'].includes(status)

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

export const deploymentSteps = (workflow?: TWorkflow): TWorkflowStep[] => {
  const steps = (workflow?.steps ?? []).filter(
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
  const kinds = new Map<string, TWorkflowStep[]>()
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
    .sort((a, b) => (a.idx ?? a.group_idx ?? 0) - (b.idx ?? b.group_idx ?? 0))
}

export const isPendingDeploymentStep = (step: TWorkflowStep) =>
  ['pending', 'not-attempted'].includes(step.status?.status ?? '')

export const isAwaitingDeploymentApproval = (step: TWorkflowStep) =>
  step.status?.status === 'approval-awaiting' && !step.approval?.response

const FAILED_STATUSES = new Set([
  'error',
  'failed-pending-retry',
  'approval-denied',
  'approval-expired',
])

export const deploymentStepContext = (
  status: string,
  steps: TWorkflowStep[]
) => {
  const awaiting = steps.some(isAwaitingDeploymentApproval)
  const current = isDeploymentRunning(status)
    ? steps.find((step) =>
        awaiting
          ? isAwaitingDeploymentApproval(step)
          : step.status?.status === 'in-progress'
      )
    : status === 'success'
      ? steps.findLast((step) => step.status?.status === 'success')
      : steps.find((step) => FAILED_STATUSES.has(step.status?.status))
  const currentIndex = current ? steps.indexOf(current) : -1
  const next = steps.find(
    (step, index) => isPendingDeploymentStep(step) && index > currentIndex
  )
  return { current, next }
}

const isAppliedDeploymentStep = (step: TWorkflowStep) =>
  ['system', 'user'].includes(step.execution_type ?? '') &&
  step.status?.status === 'success' &&
  (/^(apply |await install stack$|(?:re)?provision sandbox apply plan$)/i.test(
    step.name ?? ''
  ) ||
    (!!step.metadata?.component_name &&
      step.name === `sync ${step.metadata.component_name}`))

const outcomeForSteps = (
  category: TDeploymentOutcome['category'],
  name: string,
  steps: TWorkflowStep[],
  workflow: TWorkflow,
  history: TWorkflowStep[]
): TDeploymentOutcome => {
  const failed = steps.find((step) => FAILED_STATUSES.has(step.status?.status))
  const current = steps.find((step) => step.status?.status === 'in-progress')
  const awaiting = steps.find(isAwaitingDeploymentApproval)
  const applied = !workflow.plan_only && history.some(isAppliedDeploymentStep)
  const currentApplied =
    !workflow.plan_only && steps.some(isAppliedDeploymentStep)
  const completed =
    steps.length > 0 && steps.every((step) => step.status?.status === 'success')
  if (failed)
    return {
      category,
      name,
      status: 'error',
      detail: `${applied ? 'Applied; ' : ''}${failed.name ?? 'Resource step'} failed`,
      applied,
    }
  if (current)
    return {
      category,
      name,
      status: 'in-progress',
      detail: current.name ?? 'In progress',
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
  if (completed && currentApplied)
    return { category, name, status: 'success', detail: 'Completed' }
  if (!applied && steps.length && steps.every(isPendingDeploymentStep))
    return { category, name, status: 'not-started', detail: 'Not started' }
  if (applied)
    return {
      category,
      name,
      status: workflow.finished ? 'warn' : 'in-progress',
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
  deployment?: TInstallDeploymentRecord,
  workflow?: TWorkflow
): TDeploymentOutcome[] => {
  if (!deployment || !workflow) return []
  const steps = deploymentSteps(workflow)
  const resources = deployment.affected_resources
  const outcomes: TDeploymentOutcome[] = []
  const resourceOutcome = (
    category: TDeploymentOutcome['category'],
    name: string,
    belongsToResource: (step: TWorkflowStep) => boolean
  ) =>
    outcomeForSteps(
      category,
      name,
      steps.filter(belongsToResource),
      workflow,
      (workflow.steps ?? []).filter(belongsToResource)
    )
  if (resources.stack)
    outcomes.push(
      resourceOutcome(
        'Stack',
        'Stack',
        (step) =>
          step.step_target_type === 'install_stack_versions' ||
          /\b(install stack|stack policy)\b/i.test(step.name ?? '')
      )
    )
  if (resources.sandbox)
    outcomes.push(
      resourceOutcome(
        'Sandbox',
        'Sandbox',
        (step) =>
          step.step_target_type === 'install_sandbox_runs' ||
          /\bsandbox\b/i.test(step.name ?? '')
      )
    )
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
          name === deployment.image?.repository &&
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
  if (!workflow.finished)
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
