import type {
  TInstallAlignmentOrigin,
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

export const selectedCommit = alignmentCommit(6, {
  metadata: { tag: 'v1.4.2' },
})

export const liveCommit = alignmentCommit(1, {
  sha: '4f1d9c2a7be03e5d8a6c1b4f7a9e3c5b2d8f1a0e',
  metadata: { pr_number: 412 },
})

export const runOrigin = (
  commit: TInstallOverviewCommit
): TInstallAlignmentOrigin => ({
  workflow_id: `wf-install-${commit.run_id}`,
  workflow_type: 'app_branch_config_update',
  run: commit,
})

export const manualOrigin: TInstallAlignmentOrigin = {
  workflow_id: 'wf-manual-deploy',
  workflow_type: 'manual_deploy',
}

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

export const liveFields = (
  componentId: string
): Partial<TInstallAlignmentResource> => ({
  expected_build_id: `bld-${componentId}`,
  applied_build_id: `bld-${componentId}`,
  applied_deploy_id: `dep-live-${componentId}`,
  applied_deploy_at: hoursAgo(72),
  applied_by: runOrigin(liveCommit),
})

export const unappliedResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  alignmentResource({
    category: 'changes-not-applied',
    reason: 'The deployed version doesn’t match this install’s configuration.',
    status: 'error',
    expected_build_id: 'bld-selected',
    applied_build_id: 'bld-applied',
    applied_deploy_at: hoursAgo(72),
    applied_by: runOrigin(liveCommit),
    plan: alignmentPlan(overrides.id ?? 'ic-api'),
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
    plan: undefined,
    ...overrides,
  })

export const stateDriftResource = (
  overrides: Partial<TInstallAlignmentResource> = {}
) =>
  alignmentResource({
    ...liveFields(overrides.component_id ?? 'cmp-api'),
    category: 'state-drift-detected',
    reason: 'Live resources differ from what was last applied.',
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
]

const types = ['helm_chart', 'terraform_module', 'kubernetes_manifest']

const causeBase = (index: number) => {
  const name = names[index % names.length]
  return {
    id: `ic-${name}`,
    name,
    component_id: `cmp-${name}`,
    component_type: types[index % types.length],
  }
}

const openCause = (index: number, status: string) => {
  const base = causeBase(index)
  return openDeploymentResource({
    ...base,
    open_deployment: {
      id: `wf-open-${base.id}`,
      status,
      created_at: hoursAgo(1),
      type: 'manual_deploy',
    },
  })
}

export const causeResources = (): TInstallAlignmentResource[] => [
  unappliedResource({ ...causeBase(0), plan: undefined }),
  openCause(1, 'queued'),
  openCause(2, 'in-progress'),
  openCause(3, 'approval-awaiting'),
  openCause(4, 'failed-pending-retry'),
  inputUpdateResource(causeBase(5)),
  inputUpdateResource({
    ...causeBase(6),
    reason: 'Component dependency changes haven’t been applied.',
  }),
  inputUpdateResource({
    ...causeBase(7),
    reason: 'This component is disabled, but is still deployed.',
  }),
  unappliedResource({
    ...causeBase(8),
    reason: 'This component hasn’t been deployed yet.',
    applied_build_id: undefined,
    applied_deploy_at: undefined,
    applied_by: undefined,
    plan: undefined,
  }),
  stateDriftResource(causeBase(9)),
  stateDriftResource({ ...causeBase(10), plan: undefined }),
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
      ...liveFields(`cmp-synced-${index}`),
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
