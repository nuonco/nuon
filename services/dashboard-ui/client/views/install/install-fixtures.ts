import type { TFixture, TFixtureReply } from './install-fixture'
import { VIEW_APP_ID, VIEW_INSTALL_ID, VIEW_ORG_ID } from './install-fixture'

export const APP_CONFIG_ID = 'cfg-1'
const RUNNER_ID = 'runner-1'

const minutesAgo = (minutes: number) =>
  new Date(Date.now() - minutes * 60_000).toISOString()

const reply = (
  body: unknown,
  paginated = false,
  status = 200
): TFixtureReply => ({ body, paginated, status })

export const pendingReply = (): TFixtureReply => ({ body: null, pending: true })

const viewOrg = {
  id: VIEW_ORG_ID,
  name: 'Acme',
  features: {
    'new-app-ia': true,
  },
}

const branchRun = {
  id: 'run-1',
  status: 'succeeded',
  created_at: '2026-09-18T15:04:00Z',
  app_branch: { id: 'br-1', name: 'main' },
  head_sha: 'a1b2c3d4e5f6a7b8',
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6a7b8',
    message: 'Pin the payments chart',
    author_name: 'Ada Lovelace',
    created_at: '2026-09-18T15:02:00Z',
  },
}

const actionConfig = (
  id: string,
  actionId: string,
  triggers: { id: string; type: string; cron_schedule?: string }[],
  stepName: string,
  command: string
) => ({
  id,
  action_workflow_id: actionId,
  app_config_id: APP_CONFIG_ID,
  image: 'ghcr.io/acme/payments-worker:2.4.0',
  role: 'acme-action',
  timeout: 120_000_000_000,
  enable_kube_config: { bool: true, valid: true },
  triggers,
  steps: [{ id: `${id}-step`, idx: 0, name: stepName, command }],
})

const actionRun = (
  id: string,
  status: string,
  description: string,
  trigger: string,
  createdAt: string
) => ({
  id,
  created_at: createdAt,
  execution_time: 90_000_000_000,
  status,
  status_v2: { status, status_human_description: description },
  triggered_by_type: trigger,
  workflow_id: `wf-${id}`,
  created_by: { email: 'ada@example.com' },
})

const actions = [
  {
    id: 'act-1',
    name: 'rotate-keys',
    config: actionConfig(
      'acfg-1',
      'act-1',
      [{ id: 'trg-1', type: 'manual' }],
      'Rotate signing keys',
      'acme keys rotate'
    ),
    run: actionRun('arun-1', 'success', 'Action finished.', 'manual', '2026-10-04T15:00:00Z'),
  },
  {
    id: 'act-2',
    name: 'drain-queue',
    config: actionConfig(
      'acfg-2',
      'act-2',
      [{ id: 'trg-2', type: 'cron', cron_schedule: '0 */6 * * *' }],
      'Drain the payments queue',
      'acme queue drain'
    ),
    run: actionRun('arun-2', 'success', 'Action finished.', 'cron', '2026-10-05T08:00:00Z'),
  },
  {
    id: 'act-3',
    name: 'notify-on-call',
    config: actionConfig(
      'acfg-3',
      'act-3',
      [{ id: 'trg-3', type: 'post-deploy-component' }],
      'Page the on-call rotation',
      'acme page on-call'
    ),
    run: actionRun(
      'arun-3',
      'error',
      'The on-call webhook returned 502.',
      'post-deploy-component',
      '2026-10-05T11:20:00Z'
    ),
  },
  {
    id: 'act-4',
    name: 'backup-ledger',
    config: actionConfig(
      'acfg-4',
      'act-4',
      [
        { id: 'trg-4', type: 'manual' },
        { id: 'trg-5', type: 'cron', cron_schedule: '0 2 * * *' },
      ],
      'Snapshot the ledger',
      'acme ledger snapshot'
    ),
    run: actionRun('arun-4', 'running', 'Snapshot is in progress.', 'manual', '2026-10-05T14:10:00Z'),
  },
  {
    id: 'act-5',
    name: 'collect-metrics',
    config: actionConfig(
      'acfg-5',
      'act-5',
      [{ id: 'trg-6', type: 'cron', cron_schedule: '*/15 * * * *' }],
      'Scrape install metrics',
      'acme metrics scrape'
    ),
    run: undefined,
  },
  {
    id: 'act-6',
    name: 'warm-cache',
    removed: true,
    config: actionConfig(
      'acfg-6',
      'act-6',
      [{ id: 'trg-7', type: 'manual' }],
      'Warm the API cache',
      'acme cache warm'
    ),
    run: actionRun('arun-6', 'success', 'Action finished.', 'manual', '2026-09-28T12:00:00Z'),
  },
]

const runbookRun = (
  id: string,
  status: string,
  description: string,
  createdAt: string,
  updatedAt: string
) => ({
  id,
  created_at: createdAt,
  updated_at: updatedAt,
  status,
  created_by: { email: 'ada@example.com' },
  install_workflow_id: `wf-${id}`,
  install_workflow: {
    id: `wf-${id}`,
    status: { status, status_human_description: description },
  },
})

const runbook = ({
  id,
  name,
  description,
  readme,
  steps,
  run,
}: {
  id: string
  name: string
  description: string
  readme?: string
  steps: string[]
  run?: ReturnType<typeof runbookRun>
}) => ({
  id: `irb-${id}`,
  runbook_id: id,
  runs: run ? [run] : [],
  runbook: {
    id,
    name,
    description,
    configs: [
      {
        id: `rbc-${id}`,
        runbook_id: id,
        app_config_id: APP_CONFIG_ID,
        readme,
        inputs:
          id === 'rb-1'
            ? [
                {
                  id: 'in-1',
                  name: 'region',
                  display_name: 'Region',
                  required: true,
                  idx: 0,
                },
              ]
            : [],
        steps: steps.map((stepName, index) => ({
          id: `${id}-stp-${index}`,
          idx: index,
          name: stepName,
          type: 'action',
          action_workflow_id: 'act-1',
        })),
      },
    ],
  },
})

const runbooks = [
  runbook({
    id: 'rb-1',
    name: 'failover-database',
    description: 'Fails over the primary database to the replica.',
    readme: 'Fail over the primary database to the replica.',
    steps: [
      'Promote the replica',
      'Repoint the API',
      'Verify write traffic',
      'Page the database owner',
    ],
    run: runbookRun(
      'rrun-1',
      'success',
      'Runbook finished.',
      '2026-10-03T15:00:00Z',
      '2026-10-03T15:08:00Z'
    ),
  }),
  runbook({
    id: 'rb-2',
    name: 'drain-nodes',
    description: 'Moves pods off a node before it is replaced.',
    readme: 'Drain one node at a time and wait for pods to reschedule.',
    steps: ['Cordon the node', 'Evict the pods', 'Confirm the node is empty'],
    run: runbookRun(
      'rrun-2',
      'error',
      'A pod would not evict.',
      '2026-10-05T09:12:00Z',
      '2026-10-05T09:18:00Z'
    ),
  }),
  runbook({
    id: 'rb-3',
    name: 'scale-workers',
    description: 'Scales the payments workers for a traffic spike.',
    readme: 'Scale the worker deployment, then wait for the new pods to become ready.',
    steps: ['Scale the worker', 'Wait for ready pods'],
  }),
  runbook({
    id: 'rb-4',
    name: 'rotate-credentials',
    description: 'Rotates database and cache credentials together.',
    steps: [
      'Issue new credentials',
      'Update the worker',
      'Update the API',
      'Revoke the old credentials',
      'Confirm connections',
    ],
    run: runbookRun(
      'rrun-4',
      'success',
      'Runbook finished.',
      '2026-10-02T18:00:00Z',
      '2026-10-02T18:11:00Z'
    ),
  }),
  runbook({
    id: 'rb-5',
    name: 'restore-snapshot',
    description: 'Restores the ledger from the latest snapshot.',
    readme: 'Restore into a new volume, then cut the API over once checks pass.',
    steps: [
      'Select the snapshot',
      'Restore the volume',
      'Run consistency checks',
      'Cut over the API',
      'Drop the old volume',
      'Notify finance',
    ],
    run: runbookRun(
      'rrun-5',
      'running',
      'Restore is in progress.',
      '2026-10-05T13:40:00Z',
      '2026-10-05T13:40:00Z'
    ),
  }),
]

const appConfig = {
  id: APP_CONFIG_ID,
  action_workflow_configs: actions
    .filter((action) => !action.removed)
    .map((action) => action.config),
  component_config_connections: [],
}

const activity = {
  activity: [
    {
      id: 'evt-1',
      type: 'action_run',
      status: 'finished',
      created_at: '2026-10-05T14:10:00Z',
      title: 'backup-ledger',
      summary: 'Started a ledger snapshot.',
      action: { run_id: 'arun-4', name: 'backup-ledger', trigger_type: 'manual' },
    },
    {
      id: 'evt-2',
      type: 'action_run',
      status: 'error',
      created_at: '2026-10-05T11:20:00Z',
      title: 'notify-on-call',
      summary: 'The on-call webhook returned 502.',
      action: {
        run_id: 'arun-3',
        name: 'notify-on-call',
        trigger_type: 'post-deploy-component',
      },
    },
    {
      id: 'evt-3',
      type: 'runbook_run',
      status: 'error',
      created_at: '2026-10-05T09:12:00Z',
      title: 'drain-nodes',
      summary: 'A pod would not evict.',
      runbook: { run_id: 'rrun-2', runbook_id: 'rb-2', name: 'drain-nodes' },
    },
    {
      id: 'evt-4',
      type: 'action_run',
      status: 'finished',
      created_at: '2026-10-04T15:00:00Z',
      title: 'rotate-keys',
      summary: 'Rotated API keys for the payments service.',
      action: { run_id: 'arun-1', name: 'rotate-keys', trigger_type: 'manual' },
    },
  ],
  page: 0,
  offset: 0,
  limit: 20,
  has_more: false,
}

const policyReport = {
  id: 'rpt-1',
  app_id: VIEW_APP_ID,
  component_id: 'cmp-api',
  component_name: 'api',
  owner_type: 'install_deploys',
  evaluated_at: '2026-10-04T16:00:00Z',
  deny_count: 1,
  status: { status: 'error', status_human_description: 'Policy denied.' },
  policies: [
    { policy_id: 'pol-1', policy_name: 'no-privileged', status: 'deny', deny_count: 1 },
  ],
  violations: [
    {
      policy_id: 'pol-1',
      severity: 'deny',
      message: 'Container is running as privileged',
    },
  ],
}

const healthchecks = (failingFrom?: number) =>
  Array.from({ length: 30 }, (_, index) => ({
    id: `hc-${index}`,
    status_code: failingFrom !== undefined && index >= failingFrom ? 900 : 0,
    minute_bucket: minutesAgo(30 - index),
  }))

const processes = (offline: boolean) => [
  {
    id: 'rpr-install',
    runner_id: RUNNER_ID,
    type: 'install',
    composite_status: { status: offline ? 'offline' : 'active' },
    labels: [],
    started_at: minutesAgo(6 * 60),
    version: 'v1.2.3',
    warnings: offline
      ? ['Runner is offline and will be marked inactive in 5 minutes']
      : [],
  },
  {
    id: 'rpr-mng',
    runner_id: RUNNER_ID,
    type: 'mng',
    composite_status: { status: 'active' },
    labels: [],
    started_at: minutesAgo(6 * 60),
    version: 'v1.2.3',
    warnings: [],
  },
]

const installStatus = {
  deployments: { status: 'success', status_human_description: 'Up to date' },
  resources: { status: 'active', status_human_description: 'Healthy' },
  health_checks: { status: 'active', status_human_description: 'Passing' },
}

export const viewInstall = (overrides: Record<string, unknown> = {}) => ({
  id: VIEW_INSTALL_ID,
  name: 'payments',
  org_id: VIEW_ORG_ID,
  app_id: VIEW_APP_ID,
  app: { id: VIEW_APP_ID, name: 'Payments' },
  app_config_id: APP_CONFIG_ID,
  app_branch: { id: 'br-1', name: 'main' },
  app_branch_id: 'br-1',
  runner_id: RUNNER_ID,
  runner_status: 'active',
  sandbox_status: 'active',
  composite_component_status: 'active',
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-10-05T12:00:00Z',
  ...overrides,
})

const chrome = (
  url: URL,
  install: Record<string, unknown>
): TFixtureReply | undefined => {
  const path = url.pathname
  if (path.endsWith('/orgs/current')) return reply(viewOrg)
  if (path.endsWith('/version')) {
    return reply({ api: { version: '0.0.0', git_ref: 'dev' }, ui: { version: '0.0.0' } })
  }
  if (path === `/v1/installs/${VIEW_INSTALL_ID}`) return reply(install)
  if (path === `/v1/installs/${VIEW_INSTALL_ID}/status`) return reply(installStatus)
  if (path === `/v1/apps/${VIEW_APP_ID}`) {
    return reply({ id: VIEW_APP_ID, name: 'Payments' })
  }
  if (path === `/v1/apps/${VIEW_APP_ID}/labels`) {
    return reply({ labels: [], label_colors: {}, default_colors: [] })
  }
  if (path.endsWith('/app-config-versions')) {
    return reply([
      {
        id: 'ver-1',
        new_app_config_id: APP_CONFIG_ID,
        created_at: branchRun.created_at,
        app_branch_run: branchRun,
      },
    ])
  }
  if (path.includes('/updates')) {
    return reply({
      updates: [],
      current_app_branch_run: install.app_branch ? branchRun : undefined,
      page: 0,
      limit: 1,
      has_more: false,
    })
  }
  if (path === '/v1/workflows/pending-approvals') return reply([])
  if (path === '/v1/workflows') return reply([], true)
  if (path === '/v1/orgs') return reply([viewOrg], true)
  return undefined
}

export const withChrome = (
  install: Record<string, unknown>,
  page: TFixture
): TFixture => (url, init) => page(url, init) ?? chrome(url, install)

type TOperationsState = 'active' | 'empty' | 'offline'

const installFor = (state: TOperationsState) =>
  viewInstall({
    runner_id: state === 'empty' ? undefined : RUNNER_ID,
    runner_status:
      state === 'offline' ? 'offline' : state === 'empty' ? undefined : 'active',
  })

const segmentAfter = (path: string, marker: string) => {
  const start = path.indexOf(marker)
  if (start === -1) return undefined
  const segment = path.slice(start + marker.length).split('/')[0]
  return segment || undefined
}

const includes = (value: string | undefined, query: string) =>
  !query || (value ?? '').toLowerCase().includes(query.toLowerCase())

export const operationsFixture = (state: TOperationsState): TFixture => {
  const populated = state !== 'empty'
  const offline = state === 'offline'

  return (url) => {
    const path = url.pathname
    const params = url.searchParams

    if (path.includes('/configs/') && path.includes('/apps/')) return reply(appConfig)
    if (path.includes('/branches/') && path.includes('/runs')) return reply([], true)
    if (path.endsWith('/activity')) {
      if (!populated) return reply({ ...activity, activity: [] })
      const search = params.get('search') ?? ''
      const status = params.get('status')
      const type = params.get('type')
      return reply({
        ...activity,
        activity: activity.activity.filter((item) => {
          if (search && !includes(item.title, search) && !includes(item.summary, search)) {
            return false
          }
          if (status && !status.split(',').includes(item.status)) return false
          if (type && !type.split(',').includes(item.type)) return false
          return true
        }),
      })
    }
    if (path.includes('/action-workflows/latest-runs')) {
      const query = params.get('q') ?? ''
      const trigger = params.get('trigger_types') ?? ''
      const rows = (populated ? actions : []).filter((action) => {
        if (!includes(action.name, query)) return false
        if (trigger && !action.config.triggers.some((item) => item.type === trigger)) {
          return false
        }
        return true
      })
      return reply(
        rows.map((action) => ({
          id: `ia-${action.id}`,
          action_workflow_id: action.id,
          action_workflow: {
            id: action.id,
            name: action.name,
            configs: [action.config],
          },
          runs: action.run ? [action.run] : [],
        })),
        true
      )
    }
    if (path.includes('/recent-runs')) {
      const id = segmentAfter(path, '/action-workflows/')
      const action = actions.find((item) => item.id === id)
      return reply({ runs: action?.run ? [action.run] : [] })
    }
    if (path.includes('/action-workflows/configs/')) {
      const id = segmentAfter(path, '/configs/')
      const action = actions.find((item) => item.config.id === id)
      return reply(action?.config ?? {})
    }
    if (path.includes('/runbooks/') && !path.endsWith('/runbooks')) {
      const id = segmentAfter(path, '/runbooks/')
      return reply(runbooks.find((item) => item.runbook_id === id) ?? {})
    }
    if (path.endsWith('/runbooks')) {
      const query = params.get('q') ?? ''
      return reply(
        (populated ? runbooks : []).filter((item) => includes(item.runbook.name, query)),
        true
      )
    }
    if (path.includes('/policy-reports')) return reply(populated ? [policyReport] : [])
    if (path.endsWith('/policies-configs')) {
      return reply([{ id: 'pc-1', policies: [{ id: 'pol-1', name: 'no-privileged' }] }])
    }
    if (path === `/v1/runners/${RUNNER_ID}`) {
      return reply({
        id: RUNNER_ID,
        name: 'acme-runner',
        status: offline ? 'offline' : 'active',
        warnings: offline ? ['The runner has not sent a heartbeat in 5 minutes.'] : [],
      })
    }
    if (path.endsWith('/settings') && path.includes('/runners/')) {
      return reply({
        id: 'rs-1',
        container_image_tag: 'v1.2.3',
        binary_version: 'v1.2.3',
      })
    }
    if (path.includes('/processes') && path.includes('/heart-beats/latest')) {
      const id = segmentAfter(path, '/processes/')
      return reply({
        created_at: minutesAgo(id === 'rpr-install' && offline ? 5 : 0),
        version: 'v1.2.3',
      })
    }
    if (path.includes('/recent-health-checks')) {
      const processId = params.get('process_id')
      return reply(
        healthchecks(offline && processId === 'rpr-install' ? 25 : undefined)
      )
    }
    if (path.includes('/processes') && path.includes('/runners/')) {
      return reply(populated ? processes(offline) : [], true)
    }
    if (path.includes('/heart-beats/latest')) {
      return reply({ created_at: minutesAgo(offline ? 5 : 0), version: 'v1.2.3' })
    }
    return chrome(url, installFor(state))
  }
}
