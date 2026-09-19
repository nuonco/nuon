import type {
  TPlaygroundInstall,
  TActivityEvent,
  TBranchTracking,
  TPlaygroundConfiguration,
} from './types'
import type {
  THealthTimelineDay,
  TInstallHealthTimeline,
  TInstallHealthTimelineComponent,
} from '@/types'

const NOW = new Date('2026-09-18T18:00:00Z').getTime()
const h = (hours: number) => new Date(NOW - hours * 3600_000).toISOString()
const DAY_SECONDS = 86_400
const HEALTH_WINDOW_DAYS = 30

const isoDateDaysAgo = (daysAgo: number) =>
  new Date(NOW - daysAgo * DAY_SECONDS * 1000).toISOString().slice(0, 10)

const buildHealthDaily = (
  days: number,
  patches: Record<number, Partial<THealthTimelineDay>> = {}
): THealthTimelineDay[] =>
  Array.from({ length: days }, (_, i) => {
    const daysAgo = days - 1 - i
    const observedSeconds = daysAgo === 0 ? 18 * 3600 : DAY_SECONDS
    return {
      date: isoDateDaysAgo(daysAgo),
      health: 'healthy',
      unhealthy_seconds: 0,
      degraded_seconds: 0,
      unknown_seconds: 0,
      observed_seconds: observedSeconds,
      ...patches[daysAgo],
    }
  })

const healthFromDaily = (
  daily: THealthTimelineDay[],
  currentHealth: string,
  components: TInstallHealthTimelineComponent[]
): TInstallHealthTimeline => {
  const observed_seconds = daily.reduce(
    (sum, day) => sum + day.observed_seconds,
    0
  )
  const bad = daily.reduce(
    (sum, day) => sum + day.unhealthy_seconds + day.degraded_seconds,
    0
  )
  return {
    days: daily.length,
    uptime_percent: observed_seconds
      ? Math.round(((observed_seconds - bad) / observed_seconds) * 10000) / 100
      : 0,
    observed_seconds,
    current_health: currentHealth,
    daily,
    components,
  }
}

const HEALTH_COMPONENTS_CURRENT: TInstallHealthTimelineComponent[] = [
  {
    install_component_id: 'icmp-api',
    component_id: 'cmp-1',
    component_name: 'api',
    current_health: 'healthy',
    uptime_percent: 99.98,
    observed_seconds: 30 * DAY_SECONDS,
  },
  {
    install_component_id: 'icmp-worker',
    component_id: 'cmp-2',
    component_name: 'worker',
    current_health: 'healthy',
    uptime_percent: 99.94,
    observed_seconds: 30 * DAY_SECONDS,
  },
  {
    install_component_id: 'icmp-frontend',
    component_id: 'cmp-3',
    component_name: 'frontend',
    current_health: 'healthy',
    uptime_percent: 100,
    observed_seconds: 30 * DAY_SECONDS,
  },
  {
    install_component_id: 'icmp-cache',
    component_id: 'cmp-4',
    component_name: 'cache',
    current_health: 'not-applicable',
    uptime_percent: 0,
    observed_seconds: 0,
  },
]

const HEALTH_CURRENT = healthFromDaily(
  buildHealthDaily(HEALTH_WINDOW_DAYS, {
    12: {
      health: 'degraded',
      degraded_seconds: 900,
      observed_seconds: DAY_SECONDS,
    },
  }),
  'healthy',
  HEALTH_COMPONENTS_CURRENT
)

const HEALTH_RESOURCE_LAG = healthFromDaily(
  buildHealthDaily(HEALTH_WINDOW_DAYS, {
    0: {
      health: 'degraded',
      degraded_seconds: 2400,
      observed_seconds: 18 * 3600,
    },
    1: {
      health: 'unhealthy',
      unhealthy_seconds: 1800,
      observed_seconds: DAY_SECONDS,
    },
  }),
  'degraded',
  [
    {
      install_component_id: 'icmp-api',
      component_id: 'cmp-1',
      component_name: 'api',
      current_health: 'degraded',
      uptime_percent: 98.4,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-worker',
      component_id: 'cmp-2',
      component_name: 'worker',
      current_health: 'unhealthy',
      uptime_percent: 97.1,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-frontend',
      component_id: 'cmp-3',
      component_name: 'frontend',
      current_health: 'healthy',
      uptime_percent: 100,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-cache',
      component_id: 'cmp-4',
      component_name: 'cache',
      current_health: 'not-applicable',
      uptime_percent: 0,
      observed_seconds: 0,
    },
  ]
)

const HEALTH_BRANCH_MOVED = healthFromDaily(
  buildHealthDaily(HEALTH_WINDOW_DAYS, {
    0: {
      health: 'degraded',
      degraded_seconds: 1200,
      observed_seconds: 18 * 3600,
    },
  }),
  'degraded',
  [
    {
      install_component_id: 'icmp-api',
      component_id: 'cmp-1',
      component_name: 'api',
      current_health: 'degraded',
      uptime_percent: 99.2,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-worker',
      component_id: 'cmp-2',
      component_name: 'worker',
      current_health: 'degraded',
      uptime_percent: 99.1,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-frontend',
      component_id: 'cmp-3',
      component_name: 'frontend',
      current_health: 'healthy',
      uptime_percent: 100,
      observed_seconds: 30 * DAY_SECONDS,
    },
    {
      install_component_id: 'icmp-cache',
      component_id: 'cmp-4',
      component_name: 'cache',
      current_health: 'not-applicable',
      uptime_percent: 0,
      observed_seconds: 0,
    },
  ]
)

// ─── Shared resource fixtures ─────────────────────────────────────────────────

const COMMON_RESOURCES = {
  roles: [
    {
      id: 'role-1',
      name: 'acme-prod-admin',
      type: 'iam_role',
      status: 'active' as const,
    },
    {
      id: 'role-2',
      name: 'acme-prod-readonly',
      type: 'iam_role',
      status: 'active' as const,
    },
    {
      id: 'role-3',
      name: 'acme-prod-s3-access',
      type: 'iam_role',
      status: 'active' as const,
    },
  ],
  components: [
    {
      id: 'cmp-1',
      name: 'api',
      type: 'helm_chart',
      status: 'active' as const,
      deployedAt: h(4),
      sha: 'a1b2c3d4',
    },
    {
      id: 'cmp-2',
      name: 'worker',
      type: 'helm_chart',
      status: 'active' as const,
      deployedAt: h(4),
      sha: 'a1b2c3d4',
    },
    {
      id: 'cmp-3',
      name: 'frontend',
      type: 'helm_chart',
      status: 'active' as const,
      deployedAt: h(6),
      sha: 'e5f6a7b8',
    },
    {
      id: 'cmp-4',
      name: 'cache',
      type: 'terraform_module',
      status: 'active' as const,
      deployedAt: h(24),
      sha: 'c9d0e1f2',
    },
  ],
  images: [
    {
      id: 'img-1',
      repository: 'acme/api',
      tag: '1.14.2',
      status: 'active' as const,
      builtAt: h(3),
      sha: 'sha256:a1b2c3d4',
    },
    {
      id: 'img-2',
      repository: 'acme/worker',
      tag: '1.14.2',
      status: 'active' as const,
      builtAt: h(3),
      sha: 'sha256:a1b2c3d4',
    },
    {
      id: 'img-3',
      repository: 'acme/frontend',
      tag: '1.14.0',
      status: 'active' as const,
      builtAt: h(26),
      sha: 'sha256:e5f6a7b8',
    },
  ],
  actions: [
    {
      id: 'act-1',
      name: 'Scale API pods',
      description: 'Adjust the replica count for the API deployment.',
      lastRunAt: h(12),
      lastRunStatus: 'active' as const,
    },
    {
      id: 'act-2',
      name: 'Clear cache',
      description: 'Flush the Redis cache for this install.',
      lastRunAt: h(36),
      lastRunStatus: 'active' as const,
    },
    {
      id: 'act-3',
      name: 'Rotate secrets',
      description: 'Rotate API keys and update Kubernetes secrets.',
    },
  ],
  runbooks: [
    {
      id: 'rb-1',
      name: 'Incident response',
      description: 'Step-by-step guide for responding to production incidents.',
      stepCount: 8,
      lastRunAt: h(72),
      lastRunStatus: 'active' as const,
    },
    {
      id: 'rb-2',
      name: 'Database migration',
      description: 'Run schema migrations safely with a pre-migration backup.',
      stepCount: 5,
    },
    {
      id: 'rb-3',
      name: 'Failover to secondary region',
      description: 'Switch traffic to the secondary AWS region.',
      stepCount: 12,
    },
  ],
}

// ─── Readme fixture ───────────────────────────────────────────────────────────

const COMMON_README = `# acme-prod

This install runs Acme in the customer's AWS account \`111122223333\` in
\`us-east-1\`. It is managed from \`installs/acme-prod.toml\` on the \`main\`
app branch.

## Access

- Dashboard: [https://acme-prod.example.com](https://acme-prod.example.com)
- API: [https://api.acme-prod.example.com](https://api.acme-prod.example.com)

Log in with your Acme SSO account. If SSO is unavailable, use the break-glass
role \`acme-prod-admin\` and note the reason in the incident channel.

## Components

| Component | Purpose |
| --- | --- |
| \`api\` | Public REST API and webhook receiver |
| \`worker\` | Async jobs, retries, and scheduled syncs |
| \`frontend\` | Customer-facing web app |
| \`cache\` | Redis used by the API for sessions and rate limits |

## Operations

1. Check the 30-day health bar above before making changes.
2. Use the **Scale API pods** action to change replica counts.
3. Follow the **Incident response** runbook for production incidents.

> Updates are applied automatically when the tracked app branch changes. If a
> component is lagging, open the Activity tab and inspect the latest deploy.
`

// ─── Configuration fixtures ───────────────────────────────────────────────────

const CONFIG_FILE_CONTENTS = `[install]
name = "acme-prod"
app = "acme-byoc"

[install.inputs]
region = "us-east-1"
cluster_size = "large"
enable_backups = "true"
retention_days = "30"

[install.aws]
iam_role_arn = "arn:aws:iam::111122223333:role/nuon-acme-prod"
`

const COMMON_CONFIGURATION: TPlaygroundConfiguration = {
  inputs: [
    {
      name: 'region',
      displayName: 'AWS region',
      value: 'us-east-1',
      group: 'Cloud',
    },
    {
      name: 'iam_role_arn',
      displayName: 'IAM role ARN',
      value: 'arn:aws:iam::111122223333:role/nuon-acme-prod',
      group: 'Cloud',
    },
    {
      name: 'cluster_size',
      displayName: 'Cluster size',
      value: 'large',
      group: 'Platform',
    },
    {
      name: 'enable_backups',
      displayName: 'Enable backups',
      value: 'true',
      group: 'Platform',
    },
    {
      name: 'retention_days',
      displayName: 'Backup retention (days)',
      value: '30',
      group: 'Platform',
    },
    {
      name: 'api_token',
      displayName: 'API token',
      value: '••••••••',
      group: 'Secrets',
      isRedacted: true,
    },
  ],
  configFile: {
    path: 'installs/acme-prod.toml',
    repo: 'acme/platform-configs',
    gitBranch: 'main',
    version: 'v14',
    syncedAt: h(4),
    contents: CONFIG_FILE_CONTENTS,
  },
  overrides: [
    {
      id: 'ovr-1',
      componentName: 'api',
      inputName: 'replica_count',
      value: '6',
      updatedAt: h(30),
    },
    {
      id: 'ovr-2',
      componentName: 'cache',
      inputName: 'instance_type',
      value: 'cache.r6g.large',
      updatedAt: h(96),
    },
  ],
}

// ─── Branch tracking fixtures ─────────────────────────────────────────────────

const BRANCH_TRACKING_CURRENT: TBranchTracking = {
  targetBranch: 'main',
  branchId: 'br-acme-main',
  repo: 'acme/platform-configs',
  gitBranch: 'main',
  directory: 'apps/acme',
  expectedCommit: {
    sha: 'a1b2c3d4',
    message: 'feat: add Redis cache component',
    author: 'alice',
    createdAt: h(4),
    runStatus: 'active',
  },
  appliedCommit: {
    sha: 'a1b2c3d4',
    message: 'feat: add Redis cache component',
    author: 'alice',
    createdAt: h(4),
    runStatus: 'active',
  },
  status: 'current',
}

const BRANCH_TRACKING_MOVED: TBranchTracking = {
  targetBranch: 'feat/multi-region',
  branchId: 'br-acme-feat-multi-region',
  repo: 'acme/platform-configs',
  gitBranch: 'feat/multi-region',
  directory: 'apps/acme',
  expectedCommit: {
    sha: 'ff001234',
    message: 'feat: add secondary region support',
    author: 'carol',
    createdAt: h(1),
    runStatus: 'in-progress',
  },
  appliedCommit: {
    sha: 'a1b2c3d4',
    message: 'feat: add Redis cache component',
    author: 'alice',
    createdAt: h(4),
    runStatus: 'active',
  },
  status: 'updating',
}

// ─── Shared activity events ───────────────────────────────────────────────────

// Newest first. One record per workflow — no duplicates.
const COMMON_ACTIVITY: TActivityEvent[] = [
  {
    id: 'ev-deploy-5',
    type: 'deploy',
    status: 'active',
    createdAt: h(4),
    title: 'feat: add Redis cache component',
    source: {
      type: 'pr',
      prNumber: 218,
      branch: 'feat/cache',
      sha: 'a1b2c3d4',
      author: 'alice',
    },
    componentName: 'cache',
  },
  {
    id: 'ev-branch-4',
    type: 'app_branch_run',
    status: 'active',
    createdAt: h(10),
    title: 'App branch run: main',
    source: { type: 'push', branch: 'main', sha: 'e5f6a7b8', author: 'bob' },
  },
  {
    id: 'ev-config-3',
    type: 'config_update',
    status: 'active',
    createdAt: h(24),
    title: 'fix: bump frontend image tag to 1.14.0',
    source: { type: 'push', branch: 'main', sha: 'c9d0e1f2', author: 'alice' },
  },
  {
    id: 'ev-inputs-2',
    type: 'inputs_update',
    status: 'active',
    createdAt: h(48),
    title: 'Inputs updated',
    details: 'Fields changed: license, region',
  },
  {
    id: 'ev-stack-1',
    type: 'stack_update',
    status: 'active',
    createdAt: h(96),
    title: 'Stack version generated',
    source: {
      type: 'pr',
      prNumber: 201,
      branch: 'feat/permissions-refactor',
      sha: 'b3c4d5e6',
      author: 'carol',
    },
  },
  {
    id: 'ev-drift-scan-0',
    type: 'drift_scan',
    status: 'active',
    createdAt: h(120),
    title: 'Drift scan: no drift detected',
  },
  {
    id: 'ev-branch-tag',
    type: 'app_branch_run',
    status: 'active',
    createdAt: h(144),
    title: 'App branch run: v1.14.0',
    source: { type: 'tag', tag: 'v1.14.0', author: 'carol' },
  },
]

// ─── Config lag — current ─────────────────────────────────────────────────────

const CONFIG_LAG_CURRENT = {
  stack: {
    name: 'stack',
    appliedVersion: 'stkv-8a3f',
    expectedVersion: 'stkv-8a3f',
    isCurrent: true,
  },
  sandbox: {
    name: 'sandbox',
    appliedVersion: 'sbxv-3c19',
    expectedVersion: 'sbxv-3c19',
    isCurrent: true,
  },
  components: [
    {
      name: 'api',
      appliedVersion: 'a1b2c3d4',
      expectedVersion: 'a1b2c3d4',
      isCurrent: true,
    },
    {
      name: 'worker',
      appliedVersion: 'a1b2c3d4',
      expectedVersion: 'a1b2c3d4',
      isCurrent: true,
    },
    {
      name: 'frontend',
      appliedVersion: 'e5f6a7b8',
      expectedVersion: 'e5f6a7b8',
      isCurrent: true,
    },
  ],
  images: [
    {
      name: 'acme/api',
      appliedVersion: '1.14.2',
      expectedVersion: '1.14.2',
      isCurrent: true,
    },
    {
      name: 'acme/worker',
      appliedVersion: '1.14.2',
      expectedVersion: '1.14.2',
      isCurrent: true,
    },
    {
      name: 'acme/frontend',
      appliedVersion: '1.14.0',
      expectedVersion: '1.14.0',
      isCurrent: true,
    },
  ],
}

// ─── Scenario 1: Everything in sync ──────────────────────────────────────────

export const configCurrentFixture: TPlaygroundInstall = {
  id: 'inst-01hzacmeprod',
  name: 'acme-prod',
  orgId: 'org-acme',
  appId: 'app-acme-byoc',
  appName: 'acme-byoc',
  labels: { env: 'prod', region: 'us-east-1', tier: 'enterprise' },
  isManagedByConfig: true,
  configFilePath: 'installs/acme-prod.toml',
  createdAt: h(720),
  updatedAt: h(4),

  branchTracking: BRANCH_TRACKING_CURRENT,

  runnerStatus: 'active',
  sandboxStatus: 'active',
  componentStatus: 'active',

  configLag: CONFIG_LAG_CURRENT,
  driftedObjects: [],

  activity: COMMON_ACTIVITY,

  resources: {
    stackVersions: [
      {
        id: 'stkv-8a3f',
        version: 'stkv-8a3f',
        status: 'active',
        planType: 'apply',
        createdAt: h(4),
      },
      {
        id: 'stkv-7b2e',
        version: 'stkv-7b2e',
        status: 'active',
        planType: 'apply',
        createdAt: h(48),
      },
    ],
    sandbox: {
      id: 'sbx-01hzacmeprod',
      status: 'active',
      runType: 'apply',
      lastRunAt: h(4),
      workspaceUrl:
        'https://app.terraform.io/acme/workspaces/acme-prod-sandbox',
    },
    ...COMMON_RESOURCES,
  },
  operations: {
    actions: COMMON_RESOURCES.actions,
    runbooks: COMMON_RESOURCES.runbooks,
  },
  configuration: COMMON_CONFIGURATION,
  health: HEALTH_CURRENT,
  readme: COMMON_README,
}

// ─── Scenario 2: Branch target moved to feat/multi-region ────────────────────

export const branchMovedFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  updatedAt: h(1),

  branchTracking: BRANCH_TRACKING_MOVED,

  configuration: {
    ...COMMON_CONFIGURATION,
    configFile: COMMON_CONFIGURATION.configFile && {
      ...COMMON_CONFIGURATION.configFile,
      gitBranch: 'feat/multi-region',
      version: 'v15',
      syncedAt: h(1),
    },
  },

  health: HEALTH_BRANCH_MOVED,

  runnerStatus: 'active',
  sandboxStatus: 'active',
  componentStatus: 'pending',

  configLag: {
    stack: {
      name: 'stack',
      appliedVersion: 'stkv-8a3f',
      expectedVersion: 'stkv-9b1c',
      isCurrent: false,
    },
    sandbox: {
      name: 'sandbox',
      appliedVersion: 'sbxv-3c19',
      expectedVersion: 'sbxv-4d20',
      isCurrent: false,
    },
    components: [
      {
        name: 'api',
        appliedVersion: 'a1b2c3d4',
        expectedVersion: 'ff001234',
        isCurrent: false,
      },
      {
        name: 'worker',
        appliedVersion: 'a1b2c3d4',
        expectedVersion: 'ff001234',
        isCurrent: false,
      },
      {
        name: 'frontend',
        appliedVersion: 'e5f6a7b8',
        expectedVersion: 'e5f6a7b8',
        isCurrent: true,
      },
    ],
    images: [
      {
        name: 'acme/api',
        appliedVersion: '1.14.2',
        expectedVersion: '1.15.0-beta.1',
        isCurrent: false,
      },
      {
        name: 'acme/worker',
        appliedVersion: '1.14.2',
        expectedVersion: '1.15.0-beta.1',
        isCurrent: false,
      },
      {
        name: 'acme/frontend',
        appliedVersion: '1.14.0',
        expectedVersion: '1.14.0',
        isCurrent: true,
      },
    ],
  },

  driftedObjects: [],

  activity: [
    {
      id: 'ev-branch-6',
      type: 'app_branch_run',
      status: 'in-progress',
      createdAt: h(1),
      title: 'App branch run: feat/multi-region',
      source: {
        type: 'manual',
        branch: 'feat/multi-region',
        sha: 'ff001234',
        author: 'carol',
      },
    },
    ...COMMON_ACTIVITY,
  ],
}

// ─── Scenario 3: Branch current, api/worker components mid-deploy ─────────────

export const resourceLagFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  updatedAt: h(0.5),

  branchTracking: BRANCH_TRACKING_CURRENT,

  health: HEALTH_RESOURCE_LAG,

  runnerStatus: 'active',
  sandboxStatus: 'active',
  componentStatus: 'pending',

  configLag: {
    stack: {
      name: 'stack',
      appliedVersion: 'stkv-8a3f',
      expectedVersion: 'stkv-8a3f',
      isCurrent: true,
    },
    sandbox: {
      name: 'sandbox',
      appliedVersion: 'sbxv-3c19',
      expectedVersion: 'sbxv-3c19',
      isCurrent: true,
    },
    components: [
      {
        name: 'api',
        appliedVersion: 'a1b2c3d4',
        expectedVersion: 'dd112233',
        isCurrent: false,
      },
      {
        name: 'worker',
        appliedVersion: 'a1b2c3d4',
        expectedVersion: 'dd112233',
        isCurrent: false,
      },
      {
        name: 'frontend',
        appliedVersion: 'e5f6a7b8',
        expectedVersion: 'e5f6a7b8',
        isCurrent: true,
      },
    ],
    images: [
      {
        name: 'acme/api',
        appliedVersion: '1.14.2',
        expectedVersion: '1.14.3',
        isCurrent: false,
      },
      {
        name: 'acme/worker',
        appliedVersion: '1.14.2',
        expectedVersion: '1.14.3',
        isCurrent: false,
      },
      {
        name: 'acme/frontend',
        appliedVersion: '1.14.0',
        expectedVersion: '1.14.0',
        isCurrent: true,
      },
    ],
  },

  driftedObjects: [],

  activity: [
    {
      id: 'ev-deploy-6',
      type: 'deploy',
      status: 'in-progress',
      createdAt: h(0.5),
      title: 'fix: patch CVE in base image',
      source: {
        type: 'push',
        branch: 'main',
        sha: 'dd112233',
        author: 'alice',
      },
      componentName: 'api',
    },
    ...COMMON_ACTIVITY,
  ],

  resources: {
    ...configCurrentFixture.resources,
    components: [
      {
        id: 'cmp-1',
        name: 'api',
        type: 'helm_chart',
        status: 'pending',
        deployedAt: h(0.5),
        sha: 'dd112233',
      },
      {
        id: 'cmp-2',
        name: 'worker',
        type: 'helm_chart',
        status: 'pending',
        deployedAt: h(0.5),
        sha: 'dd112233',
      },
      {
        id: 'cmp-3',
        name: 'frontend',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(6),
        sha: 'e5f6a7b8',
      },
      {
        id: 'cmp-4',
        name: 'cache',
        type: 'terraform_module',
        status: 'active',
        deployedAt: h(24),
        sha: 'c9d0e1f2',
      },
    ],
    images: [
      {
        id: 'img-1',
        repository: 'acme/api',
        tag: '1.14.3',
        status: 'in-progress',
        builtAt: h(0.25),
        sha: 'sha256:dd112233',
      },
      {
        id: 'img-2',
        repository: 'acme/worker',
        tag: '1.14.3',
        status: 'in-progress',
        builtAt: h(0.25),
        sha: 'sha256:dd112233',
      },
      {
        id: 'img-3',
        repository: 'acme/frontend',
        tag: '1.14.0',
        status: 'active',
        builtAt: h(26),
        sha: 'sha256:e5f6a7b8',
      },
    ],
  },
}

// ─── Scenario 4: Config current, infra drift on sandbox + cache ───────────────

export const infraDriftFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  updatedAt: h(2),

  branchTracking: BRANCH_TRACKING_CURRENT,

  runnerStatus: 'active',
  sandboxStatus: 'warn',
  componentStatus: 'active',

  configLag: CONFIG_LAG_CURRENT,

  driftedObjects: [
    { id: 'dft-1', targetType: 'sandbox', componentName: undefined },
    { id: 'dft-2', targetType: 'install_deploy', componentName: 'cache' },
  ],

  activity: [
    {
      id: 'ev-drift-7',
      type: 'drift_scan',
      status: 'warn',
      createdAt: h(2),
      title: 'Drift detected: sandbox, cache',
      details: 'Drift scan found 2 objects with unexpected state.',
    },
    ...COMMON_ACTIVITY,
  ],

  resources: {
    ...configCurrentFixture.resources,
    sandbox: {
      id: 'sbx-01hzacmeprod',
      status: 'warn',
      runType: 'apply',
      lastRunAt: h(4),
      workspaceUrl:
        'https://app.terraform.io/acme/workspaces/acme-prod-sandbox',
    },
    components: [
      {
        id: 'cmp-1',
        name: 'api',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(4),
        sha: 'a1b2c3d4',
      },
      {
        id: 'cmp-2',
        name: 'worker',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(4),
        sha: 'a1b2c3d4',
      },
      {
        id: 'cmp-3',
        name: 'frontend',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(6),
        sha: 'e5f6a7b8',
      },
      {
        id: 'cmp-4',
        name: 'cache',
        type: 'terraform_module',
        status: 'warn',
        deployedAt: h(4),
        sha: 'c9d0e1f2',
      },
    ],
  },
}
