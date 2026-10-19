import type {
  TInstallAlignmentResource,
  TInstallBranchTracking,
  TInstallOverview,
  TInstallOverviewCommit,
} from '@/types'

const hoursAgo = (hours: number) =>
  new Date(Date.now() - hours * 3_600_000).toISOString()

const messages = [
  'Raise the checkout worker retry budget',
  'Pin the ledger module to 2.4.0',
  'Add a webhook signing secret input',
  'Move scheduler jobs to the shared queue',
  'Enable request tracing for the API',
  'Bump the API chart to 1.4.0',
]

export const alignmentCommit = (
  index: number,
  overrides: Partial<TInstallOverviewCommit> = {}
): TInstallOverviewCommit => ({
  sha: `${(index + 10).toString(16).padStart(2, '0')}c4e9a1b7d3f05e2a8c6b1d4f7a9e3c5b2d8f1a`,
  run_id: `abr-run-${index}`,
  workflow_id: `wf-branch-run-${index}`,
  branch_id: 'br-1',
  app_config_id: `cfg-${index}`,
  message: messages[index % messages.length],
  author: 'Example Developer',
  created_at: hoursAgo(index * 6 + 2),
  run_status: 'success',
  ...overrides,
})

export const selectedCommit = alignmentCommit(6)

export const alignmentPlan = (resourceId: string, planOnly = false) => ({
  workflow_id: `wf-plan-${resourceId}`,
  step_id: `step-plan-${resourceId}`,
  approval_id: `apr-${resourceId}`,
  owner_id: `dep-${resourceId}`,
  owner_type: 'install_deploys',
  type: 'terraform_plan' as const,
  plan_only: planOnly,
  created_at: hoursAgo(5),
  changes_create: 1,
  changes_update: 1,
  changes_delete: 0,
  changes_replace: 0,
  changes_noop: 0,
})

export const alignmentResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
): TInstallAlignmentResource => ({
  id: 'ic-api',
  kind: 'component',
  name: 'api',
  component_id: 'cmp-api',
  component_type: 'helm_chart',
  status: 'active',
  category: 'in-sync',
  ...overrides,
})

export const unappliedResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  alignmentResource({
    category: 'changes-not-applied',
    reason: 'This component hasn’t been deployed yet.',
    status: 'error',
    last_attempt: {
      id: `dep-${overrides.id ?? 'ic-api'}`,
      status: 'error',
      created_at: hoursAgo(5),
      workflow_id: `wf-attempt-${overrides.id ?? 'ic-api'}`,
      branch_run: selectedCommit,
    },
    newer_branch_runs: 3,
    plan: alignmentPlan(overrides.id ?? 'ic-api'),
    explanation: {
      summary: 'The latest deployment failed.',
      status: 'error',
      workflow_id: `wf-attempt-${overrides.id ?? 'ic-api'}`,
    },
    ...overrides,
  })

export const openDeploymentResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  unappliedResource({
    status: 'in-progress',
    open_deployment: {
      id: `wf-open-${overrides.id ?? 'ic-api'}`,
      status: 'failed-pending-retry',
      created_at: hoursAgo(1),
      type: 'manual_deploy',
    },
    explanation: {
      summary: 'The deployment failed and is waiting to be retried.',
      status: 'failed-pending-retry',
      workflow_id: `wf-open-${overrides.id ?? 'ic-api'}`,
    },
    ...overrides,
  })

export const inputUpdateResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  unappliedResource({
    status: 'active',
    reason: 'Saved input changes haven’t taken effect.',
    expected_build_id: 'bld-selected',
    applied_build_id: 'bld-selected',
    expected_config_id: 'cfg-selected',
    applied_config_id: 'cfg-selected',
    last_attempt: {
      id: `wf-input-${overrides.id ?? 'ic-api'}`,
      workflow_id: `wf-input-${overrides.id ?? 'ic-api'}`,
      type: 'input_update',
      status: 'cancelled',
      created_at: hoursAgo(2),
    },
    explanation: {
      summary:
        'Applying the input changes was cancelled. The saved values haven’t taken effect.',
      status: 'cancelled',
      workflow_id: `wf-input-${overrides.id ?? 'ic-api'}`,
    },
    newer_branch_runs: null,
    plan: undefined,
    ...overrides,
  })

export const stateDriftResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  alignmentResource({
    category: 'state-drift-detected',
    reason: 'Live resources differ from what was last applied.',
    last_attempt: {
      id: `dep-${overrides.id ?? 'ic-api'}`,
      status: 'success',
      created_at: hoursAgo(30),
      workflow_id: `wf-attempt-${overrides.id ?? 'ic-api'}`,
      branch_run: selectedCommit,
    },
    newer_branch_runs: 0,
    plan: alignmentPlan(overrides.id ?? 'ic-api', true),
    ...overrides,
  })

const names = [
  'api',
  'worker',
  'ledger',
  'webhooks',
  'scheduler',
  'migrate',
  'search',
  'billing',
  'reports',
  'gateway',
  'notifier',
  'exporter',
  'indexer',
  'mailer',
  'cache',
  'processor',
  'frontend',
]

const types = ['helm_chart', 'terraform_module', 'kubernetes_manifest']

type TAttemptType = NonNullable<
  TInstallAlignmentResource['last_attempt']
>['type']

const openWorkflow = (resourceId: string, status: string) => ({
  id: `wf-open-${resourceId}`,
  status,
  created_at: hoursAgo(1),
  type: 'manual_deploy' as TAttemptType,
})

const explained = (summary: string, workflowId: string, status: string) => ({
  summary,
  status,
  workflow_id: workflowId,
  step_id: `step-${workflowId}`,
})

const causeBase = (index: number) => {
  const name = names[index % names.length]
  return {
    id: `ic-${name}`,
    name,
    component_id: `cmp-${name}`,
    component_type: types[index % types.length],
  }
}

const pastAttempt = (resourceId: string, status: string) => ({
  id: `dep-${resourceId}`,
  status,
  created_at: hoursAgo(5),
  workflow_id: `wf-attempt-${resourceId}`,
  branch_run: selectedCommit,
})

const openCause = (index: number, status: string, summary: string) => {
  const base = causeBase(index)
  return openDeploymentResource({
    ...base,
    last_attempt: undefined,
    open_deployment: openWorkflow(base.id, status),
    explanation: explained(summary, `wf-open-${base.id}`, status),
  })
}

const attemptCause = (index: number, status: string, summary: string) => {
  const base = causeBase(index)
  return unappliedResource({
    ...base,
    last_attempt: pastAttempt(base.id, status),
    explanation: explained(summary, `wf-attempt-${base.id}`, status),
  })
}

const cancelledCause = (index: number, type: TAttemptType, summary: string) => {
  const base = causeBase(index)
  const workflowId = `wf-${type}-${base.id}`
  return inputUpdateResource({
    ...base,
    reason:
      type === 'component_enabled'
        ? 'This component is enabled, but hasn’t been deployed.'
        : type === 'component_disabled'
          ? 'This component is disabled, but is still deployed.'
          : 'Saved input changes haven’t taken effect.',
    last_attempt: {
      id: workflowId,
      workflow_id: workflowId,
      type,
      status: 'cancelled',
      created_at: hoursAgo(2),
    },
    explanation: explained(summary, workflowId, 'cancelled'),
  })
}

export const causeResources = (): TInstallAlignmentResource[] => [
  unappliedResource({
    ...causeBase(0),
    last_attempt: undefined,
    plan: undefined,
    newer_branch_runs: null,
    explanation: { summary: 'No deployment has started for these changes.' },
  }),
  openCause(1, 'queued', 'The deployment is queued and hasn’t started yet.'),
  {
    ...openCause(
      2,
      'in-progress',
      'The deployment is in progress. These changes haven’t taken effect yet.'
    ),
    last_attempt: pastAttempt(causeBase(2).id, 'error'),
  },
  openCause(3, 'approval-awaiting', 'These changes are waiting for approval.'),
  attemptCause(
    4,
    'approval-denied',
    'These changes weren’t applied because approval was declined.'
  ),
  attemptCause(5, 'error', 'The latest deployment failed.'),
  openCause(
    6,
    'failed-pending-retry',
    'The deployment failed and is waiting to be retried.'
  ),
  cancelledCause(
    7,
    'input_update',
    'Applying the input changes was cancelled. The saved values haven’t taken effect.'
  ),
  cancelledCause(
    8,
    'component_enabled',
    'Enabling this component was cancelled. It hasn’t been deployed.'
  ),
  cancelledCause(
    9,
    'component_disabled',
    'Disabling this component was cancelled. It is still deployed.'
  ),
  attemptCause(10, 'user-skipped', 'Deployment of this resource was skipped.'),
  {
    ...attemptCause(
      11,
      'not-attempted',
      'The deployment stopped before it reached this resource.'
    ),
    plan: undefined,
  },
  stateDriftResource({
    ...causeBase(12),
    explanation: {
      summary:
        'A drift check found differences between live resources and what was last applied.',
      workflow_id: `wf-plan-${causeBase(12).id}`,
      step_id: `step-plan-${causeBase(12).id}`,
    },
  }),
  stateDriftResource(causeBase(13)),
  unappliedResource({
    ...causeBase(14),
    last_attempt: undefined,
    plan: undefined,
    newer_branch_runs: null,
    explanation: undefined,
  }),
  attemptCause(
    15,
    'error',
    'An earlier operation failed, so deployment of this resource couldn’t start.'
  ),
  attemptCause(
    16,
    'cancelled',
    'The deployment was cancelled before these changes took effect.'
  ),
]

export const mixedAlignmentResources = (
  count = 12
): TInstallAlignmentResource[] => {
  const causes = causeResources()
  return Array.from({ length: count }, (_, index) => {
    const resource = causes[index % causes.length]
    if (index < causes.length) return resource
    const name = `${resource.name}-${index}`
    return { ...resource, id: `ic-${name}`, name, component_id: `cmp-${name}` }
  })
}

export const inSyncResources = (count = 3): TInstallAlignmentResource[] =>
  Array.from({ length: count }, (_, index) =>
    alignmentResource({
      id: `ic-synced-${index}`,
      name: `synced-${index}`,
      component_id: `cmp-synced-${index}`,
    })
  )

export const branchTracking = (
  overrides: Partial<TInstallBranchTracking> = {}
): TInstallBranchTracking => ({
  branch_id: 'br-1',
  target_branch: 'main',
  repo: 'acme/payments',
  status: 'behind',
  selected_commit: selectedCommit,
  commits_behind: 4,
  pending_commits: [0, 1, 2, 3].map((index) =>
    alignmentCommit(index, {
      run_status: index === 0 ? 'in-progress' : 'success',
      awaiting_approval: index === 1,
    })
  ),
  ...overrides,
})

export const installOverview = (
  resources: TInstallAlignmentResource[],
  tracking: Partial<TInstallBranchTracking> = {}
): TInstallOverview => ({
  branch_tracking: branchTracking(tracking),
  config_drift: { components: [] },
  alignment: { resources },
})
