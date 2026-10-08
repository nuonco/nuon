import type { TFixture, TFixtureReply } from '@/views/install/install-fixture'
import { VIEW_APP_ID, VIEW_ORG_ID } from '@/views/install/install-fixture'
import { pendingReply, viewInstall, withChrome } from '@/views/install/install-fixtures'

const ok = (body: unknown, paginated = false): TFixtureReply => ({
  body,
  paginated,
})

export const appPath = (suffix = '') => `/${VIEW_ORG_ID}/apps/${VIEW_APP_ID}${suffix}`

const appOrg = {
  id: VIEW_ORG_ID,
  name: 'Acme',
  features: {
    'new-app-ia': true,
  },
}

const app = {
  id: VIEW_APP_ID,
  name: 'Payments',
  org_id: VIEW_ORG_ID,
  runner_config: { type: 'cloudformation' },
}

const commit = (
  sha: string,
  message: string,
  createdAt: string
) => ({
  sha,
  message,
  author_name: 'Ada Lovelace',
  created_at: createdAt,
})

const branchRun = (
  id: string,
  branchId: string,
  branchName: string,
  sha: string,
  message: string,
  createdAt: string,
  workflowId: string
) => ({
  id,
  status: 'succeeded',
  created_at: createdAt,
  head_sha: sha,
  workflow_id: workflowId,
  app_branch: { id: branchId, name: branchName },
  app_config_id: 'bcfg-1',
  vcs_connection_commit: commit(sha, message, createdAt),
  metadata: { trigger: 'push' },
})

const config = (
  id: string,
  branchName: string,
  repoBranch: string,
  groups: { id: string; name: string; order: number; install_ids?: string[] }[]
) => ({
  id,
  config_number: 4,
  version: 4,
  connected_github_vcs_config: {
    repo: 'acme/payments',
    branch: repoBranch,
  },
  run_config: { mode: 'all' },
  install_groups: groups,
  component_ids: ['cmp-api', 'cmp-worker', 'cmp-ledger', 'cmp-webhooks'],
  action_ids: ['act-1', 'act-2', 'act-3'],
  runbook_ids: ['rb-1', 'rb-2'],
  app_branch: { name: branchName },
})

const mainConfig = config('bcfg-1', 'main', 'main', [
  { id: 'grp-staging', name: 'staging', order: 0, install_ids: ['inst-staging'] },
  {
    id: 'grp-customers',
    name: 'customers',
    order: 1,
    install_ids: ['inst-acme', 'inst-globex'],
  },
])

const mainLatest = branchRun(
  'brun-1',
  'br-1',
  'main',
  'a1b2c3d4e5f6a7b8',
  'Pin the payments chart',
  '2026-10-04T15:04:00Z',
  'wf-run-1'
)

const branchRecord = (
  id: string,
  name: string,
  branchConfig: ReturnType<typeof config>,
  latest: ReturnType<typeof branchRun>
) => ({
  id,
  name,
  app_id: VIEW_APP_ID,
  configs: [branchConfig],
  latest_run: latest,
})

const mainBranch = branchRecord('br-1', 'main', mainConfig, mainLatest)

const releaseBranch = branchRecord(
  'br-2',
  'release',
  config('bcfg-2', 'release', 'release', [
    { id: 'grp-release', name: 'customers', order: 0 },
  ]),
  branchRun(
    'brun-2',
    'br-2',
    'release',
    'b2c3d4e5f6a7b8c9',
    'Cut the 2.4 release',
    '2026-10-02T11:00:00Z',
    'wf-run-2'
  )
)

const previewBranch = branchRecord(
  'br-3',
  'preview',
  config('bcfg-3', 'preview', 'preview', [
    { id: 'grp-preview', name: 'previews', order: 0 },
  ]),
  branchRun(
    'brun-3',
    'br-3',
    'preview',
    'c3d4e5f6a7b8c9d0',
    'Try the ledger snapshot',
    '2026-10-05T09:12:00Z',
    'wf-run-3'
  )
)

const install = (
  id: string,
  name: string,
  group: string,
  region: string,
  health: 'active' | 'degraded',
  createdAt: string
) => ({
  id,
  name,
  org_id: VIEW_ORG_ID,
  app_id: VIEW_APP_ID,
  app_branch_id: 'br-1',
  app_branch_group: group,
  app_branch: { id: 'br-1', name: 'main' },
  cloud_platform: 'aws',
  aws_account: { region },
  labels: { env: group === 'staging' ? 'staging' : 'prod', app: 'payments' },
  runner_status: 'active',
  sandbox_status: health,
  sandbox_health_status: health === 'active' ? 'healthy' : 'degraded',
  composite_component_status: health,
  composite_health_status: health === 'active' ? 'healthy' : 'degraded',
  created_at: createdAt,
  updated_at: '2026-10-04T16:00:00Z',
  created_by: { email: 'ada@example.com', account_type: 'user' },
})

const installs = [
  install('inst-staging', 'staging', 'staging', 'us-west-2', 'active', '2026-08-01T00:00:00Z'),
  install('inst-acme', 'acme-prod', 'customers', 'us-east-1', 'active', '2026-06-12T00:00:00Z'),
  install('inst-globex', 'globex-prod', 'customers', 'eu-west-1', 'degraded', '2026-07-03T00:00:00Z'),
]

type TPickerGroup = { name: string; statuses: Record<string, number> }

const pickerRollouts: Record<string, TPickerGroup[]> = {
  'br-1': [
    { name: 'staging', statuses: { success: 3 } },
    { name: 'canary', statuses: { success: 5, error: 1 } },
    {
      name: 'customers',
      statuses: { success: 7, 'in-progress': 3, queued: 8 },
    },
  ],
  'br-2': [
    { name: 'customers-us', statuses: { success: 8 } },
    {
      name: 'customers-eu',
      statuses: { success: 4, 'approval-awaiting': 6 },
    },
    { name: 'enterprise', statuses: { pending: 5 } },
  ],
  'br-3': [
    { name: 'previews', statuses: { success: 2, error: 1, cancelled: 1 } },
    { name: 'qa', statuses: { success: 3, error: 2, pending: 1 } },
  ],
}

const pickerGroupInstalls = (branchId: string, group: TPickerGroup) =>
  Object.entries(group.statuses).flatMap(([status, count]) =>
    Array.from({ length: count }, (_, index) => ({
      id: `inst-${branchId}-${group.name}-${status}-${index + 1}`,
      name: `${group.name}-${status}-${index + 1}`,
      status,
    }))
  )

const pickerInstalls = Object.entries(pickerRollouts).flatMap(
  ([branchId, groups]) =>
    groups.flatMap((group) =>
      pickerGroupInstalls(branchId, group).map((item) => ({
        ...install(
          item.id,
          item.name,
          group.name,
          'us-east-1',
          'active',
          '2026-08-01T00:00:00Z'
        ),
        app_branch_id: branchId,
        app_branch: { id: branchId },
      }))
    )
)

const pickerGroupRuns = (branchId: string) =>
  (pickerRollouts[branchId] ?? []).map((group) => {
    const groupInstalls = pickerGroupInstalls(branchId, group)
    return {
      id: `igr-${branchId}-${group.name}`,
      install_group_id: `grp-${branchId}-${group.name}`,
      install_group_name: group.name,
      status: { status: 'in-progress' },
      completed_installs: groupInstalls.filter(
        (item) => item.status === 'success'
      ).length,
      total_installs: groupInstalls.length,
      installs: groupInstalls.map((item) => ({
        install_id: item.id,
        workflow_id: `wf-${item.id}`,
        status: item.status,
      })),
    }
  })

const pickerBranch = (
  base: ReturnType<typeof branchRecord>,
  status: string,
  awaitingApproval = false
) => ({
  ...base,
  configs: [
    {
      ...base.configs[0],
      install_groups: (pickerRollouts[base.id] ?? []).map((group, index) => ({
        id: `grp-${base.id}-${group.name}`,
        name: group.name,
        order: index,
      })),
    },
  ],
  latest_run: {
    ...base.latest_run,
    status,
    awaiting_approval: awaitingApproval,
  },
})

const workflowRun = {
  id: 'wf-run-1',
  name: 'Pin the payments chart',
  type: 'app_branch_config_update',
  created_at: mainLatest.created_at,
  finished: true,
  status: {
    status: 'success',
    status_human_description: 'Rollout finished.',
  },
  app_branch_runs: [mainLatest],
  steps: [
    { id: 'step-fetch', name: 'Fetch commit', group_idx: 0, status: { status: 'success' }, finished: true },
    { id: 'step-config', name: 'Sync app config', group_idx: 1, status: { status: 'success' }, finished: true },
    { id: 'step-build', name: 'Build components', group_idx: 2, status: { status: 'success' }, finished: true },
    { id: 'step-staging', name: 'Deploy install group: staging', group_idx: 3, status: { status: 'success' }, finished: true },
    { id: 'step-customers', name: 'Deploy install group: customers', group_idx: 4, status: { status: 'success' }, finished: true },
  ],
}

const groupRuns = [
  {
    id: 'igr-staging',
    install_group_id: 'grp-staging',
    install_group_name: 'staging',
    status: { status: 'success' },
    completed_installs: 1,
    total_installs: 1,
    installs: [
      { install_id: 'inst-staging', workflow_id: 'wf-inst-staging', status: 'success' },
    ],
  },
  {
    id: 'igr-customers',
    install_group_id: 'grp-customers',
    install_group_name: 'customers',
    status: { status: 'success' },
    completed_installs: 2,
    total_installs: 2,
    installs: [
      { install_id: 'inst-acme', workflow_id: 'wf-inst-acme', status: 'success' },
      { install_id: 'inst-globex', workflow_id: 'wf-inst-globex', status: 'success' },
    ],
  },
]

const olderRun = {
  ...workflowRun,
  id: 'wf-run-0',
  name: 'Raise the checkout worker retry budget',
  created_at: '2026-09-18T15:04:00Z',
  status: {
    status: 'success',
    status_human_description: 'Rollout finished.',
  },
  app_branch_runs: [
    branchRun(
      'brun-0',
      'br-1',
      'main',
      'd4e5f6a7b8c9d0e1',
      'Raise the checkout worker retry budget',
      '2026-09-18T15:04:00Z',
      'wf-run-0'
    ),
  ],
}

const built = (status: string, description: string) => ({
  id: `bld-${status}`,
  status_v2: {
    status,
    status_human_description: description,
    created_at_ts: 1759590000,
  },
})

const components = [
  {
    id: 'cmp-api',
    name: 'api',
    type: 'helm_chart',
    app_id: VIEW_APP_ID,
    labels: { tier: 'edge' },
    dependencies: [{ id: 'cmp-ledger', name: 'ledger' }],
    latest_build: built('success', 'Image published.'),
  },
  {
    id: 'cmp-worker',
    name: 'worker',
    type: 'helm_chart',
    app_id: VIEW_APP_ID,
    labels: { tier: 'jobs' },
    dependencies: [{ id: 'cmp-api', name: 'api' }],
    latest_build: built('success', 'Image published.'),
  },
  {
    id: 'cmp-ledger',
    name: 'ledger',
    type: 'terraform_module',
    app_id: VIEW_APP_ID,
    latest_build: built('success', 'Module published.'),
  },
  {
    id: 'cmp-webhooks',
    name: 'webhooks',
    type: 'kubernetes_manifest',
    app_id: VIEW_APP_ID,
    latest_build: built('error', 'Manifest validation failed.'),
  },
  {
    id: 'cmp-scheduler',
    name: 'scheduler',
    type: 'pulumi',
    app_id: VIEW_APP_ID,
    latest_build: built('success', 'Program published.'),
  },
]

const action = (
  id: string,
  name: string,
  steps: string[],
  trigger: { type: string; cron_schedule?: string; component?: { id: string; name: string } }
) => ({
  id,
  name,
  app_id: VIEW_APP_ID,
  labels: { owner: 'payments' },
  configs: [
    {
      id: `${id}-cfg`,
      steps: steps.map((step, index) => ({ id: `${id}-s${index}`, idx: index, name: step })),
      triggers: [{ id: `${id}-t`, ...trigger }],
    },
  ],
})

const actions = [
  action('act-1', 'rotate-keys', ['Issue a new key', 'Restart the api'], { type: 'manual' }),
  action('act-2', 'drain-queue', ['Stop consumers', 'Wait for the queue to empty'], {
    type: 'cron',
    cron_schedule: '0 4 * * *',
  }),
  action('act-3', 'backup-ledger', ['Snapshot the volume', 'Copy the snapshot'], {
    type: 'post-deploy-component',
    component: { id: 'cmp-ledger', name: 'ledger' },
  }),
]

const runbook = (id: string, name: string, description: string, steps: string[]) => ({
  id,
  name,
  description,
  app_id: VIEW_APP_ID,
  updated_at: '2026-09-20T12:00:00Z',
  labels: { owner: 'payments' },
  configs: [
    {
      id: `${id}-cfg`,
      steps: steps.map((step, index) => ({ id: `${id}-s${index}`, idx: index, name: step })),
    },
  ],
})

const runbooks = [
  runbook('rb-1', 'failover-database', 'Fails over the primary database to the replica.', [
    'Promote the replica',
    'Point the api at the new primary',
  ]),
  runbook('rb-2', 'drain-nodes', 'Moves pods off a node before it is replaced.', [
    'Cordon the node',
    'Evict the pods',
  ]),
  runbook('rb-3', 'restore-snapshot', 'Restores the ledger from the latest snapshot.', [
    'Select the snapshot',
    'Restore the volume',
  ]),
]

const sandboxBuilds = [
  {
    id: 'sbld-1',
    app_id: VIEW_APP_ID,
    app_branch_id: 'br-1',
    created_at: '2026-10-04T14:00:00Z',
    updated_at: '2026-10-04T14:12:00Z',
    status: 'active',
    status_description: 'Sandbox is current.',
    status_v2: { status: 'active', status_human_description: 'Sandbox is current.', metadata: {} },
    created_by: { email: 'ada@example.com' },
    vcs_connection_commit: commit('a1b2c3d4e5f6a7b8', 'Pin the payments chart', '2026-10-04T15:04:00Z'),
  },
  {
    id: 'sbld-0',
    app_id: VIEW_APP_ID,
    app_branch_id: 'br-1',
    created_at: '2026-09-18T14:00:00Z',
    updated_at: '2026-09-18T14:20:00Z',
    status: 'succeeded',
    status_v2: { status: 'succeeded', metadata: {} },
    created_by: { email: 'ada@example.com' },
  },
]

const appConfig = {
  id: 'bcfg-1',
  version: 4,
  input: {
    input_groups: [
      {
        id: 'group-cloud',
        name: 'cloud',
        display_name: 'Cloud',
        description: 'Where installs of this branch run.',
      },
      {
        id: 'group-platform',
        name: 'platform',
        display_name: 'Platform',
      },
    ],
    inputs: [
      {
        id: 'in-region',
        group_id: 'group-cloud',
        name: 'region',
        display_name: 'Region',
        default: 'us-west-2',
      },
      {
        id: 'in-account',
        group_id: 'group-cloud',
        name: 'account_id',
        display_name: 'AWS account',
        default: '000000000000',
      },
      {
        id: 'in-domain',
        group_id: 'group-platform',
        name: 'domain',
        display_name: 'Domain',
        default: 'payments.example.com',
      },
    ],
  },
  permissions: {
    aws_iam_roles: [
      {
        id: 'role-1',
        name: 'acme-provision',
        display_name: 'Provision',
        description: 'Creates the stack resources for an install.',
        type: 'provision',
        created_at: '2026-04-02T00:00:00Z',
        named_policy_names: ['acme-provision'],
        policies: [{ id: 'p-1', name: 'acme-provision' }],
      },
      {
        id: 'role-2',
        name: 'acme-deprovision',
        display_name: 'Deprovision',
        description: 'Removes the stack resources for an install.',
        type: 'deprovision',
        created_at: '2026-04-02T00:00:00Z',
        named_policy_names: ['acme-deprovision'],
        policies: [{ id: 'p-2', name: 'acme-deprovision' }],
      },
    ],
  },
  policies: {
    policies: [
      { id: 'pol-1', name: 'no-privileged', type: 'admission_control', engine: 'kyverno' },
      { id: 'pol-2', name: 'read-only-root', type: 'admission_control', engine: 'opa' },
    ],
  },
  sandbox: {
    id: 'sbxcfg-1',
    terraform_version: '1.9.0',
    aws_region_type: 'us-west-2',
    connected_github_vcs_config: {
      repo: 'acme/payments',
      branch: 'main',
      directory: 'sandbox',
    },
  },
  readme: [
    'Payments',
    '',
    'Deploys the payments API, workers, and ledger into a customer account.',
    '',
    'The public hostname is payments.example.com.',
  ].join('\n'),
}

const matches = (value: string | undefined, query: string) =>
  !query || (value ?? '').toLowerCase().includes(query.toLowerCase())

const longCommitMessage = [
  'Pin the payments chart to 2.4.1 and widen the checkout worker retry budget',
  '',
  'This also bumps the ledger module so staging can take the new snapshot path,',
  'and documents the payments.example.com hostname cutover for customer installs.',
  '',
  'Follow-up is still needed for the webhooks manifest validation failure on globex-prod.',
].join('\n')

const withCommitMessage = <T extends { name?: string; app_branch_runs: ReturnType<typeof branchRun>[] }>(
  run: T,
  message: string
): T => ({
  ...run,
  name: message.split('\n')[0],
  app_branch_runs: run.app_branch_runs.map((branchRun) => ({
    ...branchRun,
    vcs_connection_commit: {
      ...branchRun.vcs_connection_commit,
      message,
    },
  })),
})

const runComparison = {
  id: 'cmp-1',
  head_run_id: 'brun-1',
  base_run_id: 'brun-0',
  base_sha: 'd4e5f6a7b8c9d0e1',
  head_sha: 'a1b2c3d4e5f6a7b8',
  config_diff_content: {
    additions: 1,
    removals: 0,
    changed: 3,
    sections: [
      {
        name: 'Components',
        additions: 0,
        removals: 0,
        changed: 2,
        entries: [
          {
            op: 'change',
            name: 'api',
            source_changed: true,
            description: 'chart version 2.3.0 → 2.4.1',
          },
          {
            op: 'change',
            name: 'worker',
            source_changed: true,
            description: 'retry budget 3 → 8',
          },
        ],
      },
      {
        name: 'Install inputs',
        additions: 0,
        removals: 0,
        changed: 1,
        entries: [
          {
            op: 'change',
            name: 'domain',
            description: 'payments.example.com',
          },
        ],
      },
      {
        name: 'Actions',
        additions: 1,
        removals: 0,
        changed: 0,
        entries: [{ op: 'add', name: 'backup-ledger' }],
      },
    ],
  },
}

const bareConfig = { ...mainConfig, install_groups: [] }
const bareBranch = { ...mainBranch, configs: [bareConfig] }

const step = (
  id: string,
  name: string,
  status: string,
  extra: Record<string, unknown> = {}
) => ({
  id,
  name,
  status: { status, ...extra },
  finished: status === 'success',
})

const doneSteps = [
  step('step-fetch', 'Fetch commit', 'success'),
  step('step-config', 'Sync app config', 'success'),
  step('step-build', 'Build components', 'success'),
  step('step-plan-staging', 'Plan install group: staging', 'success'),
  step('step-staging', 'Deploy install group: staging', 'success'),
]

const stagingGroup = {
  id: 'igr-staging',
  install_group_id: 'grp-staging',
  install_group_name: 'staging',
  status: { status: 'success' },
  completed_installs: 1,
  total_installs: 1,
  installs: [
    { install_id: 'inst-staging', workflow_id: 'wf-inst-staging', status: 'success' },
  ],
}

const customersGroup = (
  status: string,
  installs: { install_id: string; workflow_id: string; status: string }[]
) => ({
  id: 'igr-customers',
  install_group_id: 'grp-customers',
  install_group_name: 'customers',
  status: { status },
  completed_installs: installs.filter((item) => item.status === 'success').length,
  total_installs: installs.length,
  installs,
})

const runWith = (
  status: string,
  description: string,
  steps: ReturnType<typeof step>[],
  groups: unknown[]
) => ({
  run: {
    ...workflowRun,
    finished: status === 'success',
    status: { status, status_human_description: description },
    steps,
  },
  groups,
})

const rolloutOf = (state: string) => {
  if (state === 'overview-long-commit') return 'long-commit'
  if (state.startsWith('overview-') || state.startsWith('rollout-')) {
    return state.slice(state.indexOf('-') + 1)
  }
  if (state === 'settings-no-plan') return 'no-plan'
  if (state === 'runs-failed') return 'failed'
  return 'succeeded'
}

const catalogOf = (state: string, page: string) => {
  if (state === `${page}-empty` || state === `${page}-unconfigured`) return 'empty'
  if (state === `${page}-loading`) return 'loading'
  if (state === `${page}-missing`) return 'missing'
  if (state === `${page}-failed`) return 'failed'
  return 'results'
}

const rolloutData = (mode: string) => {
  if (mode === 'rolling-out') {
    return runWith(
      'in-progress',
      'Rolling out customers.',
      [...doneSteps, step('step-customers', 'Deploy install group: customers', 'in-progress')],
      [
        stagingGroup,
        customersGroup('in-progress', [
          { install_id: 'inst-acme', workflow_id: 'wf-inst-acme', status: 'success' },
          { install_id: 'inst-globex', workflow_id: 'wf-inst-globex', status: 'in-progress' },
        ]),
      ]
    )
  }
  if (mode === 'awaiting') {
    return runWith(
      'in-progress',
      'Waiting for approval to roll out customers.',
      [
        ...doneSteps,
        step('step-plan-customers', 'Plan install group: customers', 'approval-awaiting'),
      ],
      [
        stagingGroup,
        customersGroup('approval-awaiting', [
          { install_id: 'inst-acme', workflow_id: 'wf-inst-acme', status: 'pending' },
          { install_id: 'inst-globex', workflow_id: 'wf-inst-globex', status: 'pending' },
        ]),
      ]
    )
  }
  if (mode === 'failed') {
    return runWith(
      'error',
      'The globex-prod deploy failed.',
      [
        ...doneSteps,
        step('step-customers', 'Deploy install group: customers', 'error', {
          composite_error: {
            message: 'The globex-prod deploy failed.',
            type: 'install_group.install_update_failed',
            data: { install_id: 'inst-globex', workflow_id: 'wf-inst-globex' },
          },
        }),
      ],
      [
        stagingGroup,
        customersGroup('error', [
          { install_id: 'inst-acme', workflow_id: 'wf-inst-acme', status: 'success' },
          { install_id: 'inst-globex', workflow_id: 'wf-inst-globex', status: 'error' },
        ]),
      ]
    )
  }
  const succeeded = runWith(
    'success',
    'Rollout finished.',
    [
      ...doneSteps,
      step('step-customers', 'Deploy install group: customers', 'success'),
    ],
    groupRuns
  )
  if (mode === 'long-commit') {
    return {
      run: withCommitMessage(succeeded.run, longCommitMessage),
      groups: succeeded.groups,
    }
  }
  return succeeded
}

const listReply = (state: string, page: string, rows: unknown[]): TFixtureReply | undefined => {
  const catalog = catalogOf(state, page)
  if (catalog === 'loading') return pendingReply()
  if (catalog === 'empty') return ok([], true)
  if (catalog === 'results' && state !== page && !state.startsWith(`${page}-`)) return undefined
  return ok(rows, true)
}

export const appFixture = (state: string): TFixture =>
  withChrome(viewInstall(), (url) => {
    const path = url.pathname
    const params = url.searchParams
    const rollout = rolloutOf(state)
    const planned = rollout === 'no-plan'
    const current = rolloutData(rollout)
    const branch = planned ? bareBranch : mainBranch
    const configs = planned ? [bareConfig] : [mainConfig]
    if (path.endsWith('/orgs/current')) return ok(appOrg)
    if (path === `/v1/apps/${VIEW_APP_ID}`) return ok(app)
    if (path === `/v1/apps/${VIEW_APP_ID}/labels`) {
      const labels = catalogOf(state, 'labels')
      if (labels === 'loading') return pendingReply()
      if (labels === 'empty') {
        return ok({ labels: [], label_colors: {}, default_colors: [] })
      }
      return ok({
        labels: [
          {
            key: 'env',
            color: '#3b82f6',
            default_color: '#3b82f6',
            is_override: false,
            values: ['staging', 'prod'],
            entity_types: ['install'],
            usage_count: 3,
          },
          {
            key: 'app',
            color: '#10b981',
            default_color: '#10b981',
            is_override: false,
            values: ['payments'],
            entity_types: ['install', 'component'],
            usage_count: 4,
          },
        ],
        label_colors: { env: '#3b82f6', app: '#10b981' },
        default_colors: [],
      })
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/branches`) {
      if (state === 'picker-loading') return pendingReply()
      const rows =
        state === 'picker-empty'
          ? []
          : state === 'picker'
            ? [
                pickerBranch(mainBranch, 'in-progress'),
                pickerBranch(releaseBranch, 'in-progress', true),
                pickerBranch(previewBranch, 'error'),
              ]
            : [branch]
      const query = params.get('q') ?? ''
      return ok(rows.filter((item) => matches(item.name, query)), true)
    }
    if (path.startsWith(`/v1/apps/${VIEW_APP_ID}/branches/br-1`) && !path.includes('/runs') && !path.includes('/configs')) {
      return ok(branch)
    }
    if (path.includes('/branches/br-1/configs')) {
      const inputs = catalogOf(state, 'inputs')
      if (inputs === 'loading') return pendingReply()
      if (inputs === 'missing') return ok([])
      return ok(configs)
    }
    if (path.endsWith('/branches/br-1/runs')) {
      if (rollout === 'loading' || catalogOf(state, 'runs') === 'loading') return pendingReply()
      if (rollout === 'no-runs' || rollout === 'no-plan' || catalogOf(state, 'runs') === 'empty') return ok([], true)
      const latest = rollout === 'succeeded' ? workflowRun : current.run
      return ok([latest, olderRun], true)
    }
    if (path.includes('/comparison')) return ok(runComparison)
    if (path.endsWith('/runs/wf-run-1') || path.endsWith('/runs/wf-run-0')) {
      if (path.endsWith('wf-run-0')) return ok(olderRun)
      return ok(rollout === 'succeeded' ? workflowRun : current.run)
    }
    if (state === 'picker' && path.includes('/install-group-runs')) {
      const branchId = path.match(/\/branches\/([^/]+)\//)?.[1] ?? ''
      return ok(pickerGroupRuns(branchId))
    }
    if (path.includes('/install-group-runs')) {
      return ok(rollout === 'no-runs' || rollout === 'no-plan' || rollout === 'loading' ? [] : current.groups)
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/sandbox/builds`) {
      if (catalogOf(state, 'sandbox') === 'loading') return pendingReply()
      return ok(sandboxBuilds, true)
    }
    if (path.startsWith(`/v1/apps/${VIEW_APP_ID}/sandbox/builds/`)) return ok(sandboxBuilds[0])
    if (path.includes('/builds') && path.includes('/runs/')) return ok([])
    if (/^\/v1\/installs\/[^/]+\/status$/.test(path)) {
      const degraded = path.includes('inst-globex')
      return ok({
        deployments: {
          status: degraded ? 'error' : 'success',
          status_human_description: degraded ? 'One component is behind.' : 'Up to date',
        },
        resources: {
          status: degraded ? 'error' : 'success',
          status_human_description: degraded ? 'A node is not ready.' : 'Healthy',
        },
        health_checks: {
          status: degraded ? 'error' : 'success',
          status_human_description: degraded ? 'A check is failing.' : 'Passing',
        },
      })
    }
    if (state === 'picker' && path === `/v1/apps/${VIEW_APP_ID}/installs`) {
      return ok(pickerInstalls, true)
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/installs` || path === '/v1/installs') {
      return listReply(state, 'installs', installs) ?? ok(installs, true)
    }
    if (path.includes('/configs/') && path.includes('/diff')) {
      return ok({
        config_id: 'bcfg-1',
        old_config_id: 'bcfg-0',
        summary: { has_changed: true, added: 1, removed: 0, changed: 3, unchanged: 8 },
        changed: 'components,inputs,actions',
        diff: { key: 'root', children: [] },
      })
    }
    if (path.includes('/configs/') && path.includes('/apps/')) {
      if (catalogOf(state, 'inputs') === 'failed') {
        return { body: { error: 'The app config could not be read.' }, status: 500 }
      }
      if (
        ['inputs', 'policies', 'roles', 'readme', 'sandbox'].some(
          (page) => catalogOf(state, page) === 'loading'
        )
      ) {
        return pendingReply()
      }
      if (catalogOf(state, 'inputs') === 'empty') {
        return ok({ ...appConfig, input: { input_groups: [], inputs: [] } })
      }
      if (catalogOf(state, 'policies') === 'empty') {
        return ok({ ...appConfig, policies: { policies: [] } })
      }
      if (catalogOf(state, 'roles') === 'empty') {
        return ok({ ...appConfig, permissions: { aws_iam_roles: [] } })
      }
      if (catalogOf(state, 'readme') === 'empty') return ok({ ...appConfig, readme: '' })
      if (catalogOf(state, 'sandbox') === 'empty') {
        return ok({ ...appConfig, sandbox: undefined })
      }
      return ok(appConfig)
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/components`) {
      const listed = listReply(state, 'components', components)
      if (listed && catalogOf(state, 'components') !== 'results') return listed
      const query = params.get('q') ?? ''
      const types = params.get('types') ?? ''
      return ok(
        components.filter((item) => {
          if (!matches(item.name, query)) return false
          if (types && !types.split(',').includes(item.type)) return false
          return true
        }),
        true
      )
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/action-workflows`) {
      return listReply(state, 'actions', actions) ?? ok(actions, true)
    }
    if (path === `/v1/apps/${VIEW_APP_ID}/runbooks`) {
      return listReply(state, 'runbooks', runbooks) ?? ok(runbooks, true)
    }
    return undefined
  })
