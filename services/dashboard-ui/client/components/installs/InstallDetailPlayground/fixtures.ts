import type {
  TPlaygroundInstall,
  TActivityEvent,
  TBranchTracking,
  TDeploymentRecord,
  TPlaygroundConfiguration,
} from './types'
import type {
  THealthTimelineDay,
  TInstallComponentHealthTimeline,
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

const componentHealth = ({
  currentHealth,
  uptimePercent,
  patches = {},
  transitions = [],
}: {
  currentHealth: string
  uptimePercent: number
  patches?: Record<number, Partial<THealthTimelineDay>>
  transitions?: TInstallComponentHealthTimeline['transitions']
}): TInstallComponentHealthTimeline => {
  const hasSignal = currentHealth !== 'not-applicable'
  const daily = hasSignal ? buildHealthDaily(HEALTH_WINDOW_DAYS, patches) : []

  return {
    days: HEALTH_WINDOW_DAYS,
    uptime_percent: hasSignal ? uptimePercent : 0,
    observed_seconds: hasSignal ? HEALTH_WINDOW_DAYS * DAY_SECONDS : 0,
    current_health: currentHealth,
    daily,
    transitions,
  }
}

const COMPONENT_HEALTH_CURRENT = {
  api: componentHealth({
    currentHealth: 'healthy',
    uptimePercent: 99.98,
    transitions: [
      {
        from_health: 'degraded',
        to_health: 'healthy',
        message: 'API readiness checks recovered',
        correlated_deploy_id: 'deploy-api-14',
        observed_at: h(4),
      },
    ],
  }),
  worker: componentHealth({
    currentHealth: 'healthy',
    uptimePercent: 99.94,
    transitions: [
      {
        from_health: 'unknown',
        to_health: 'healthy',
        message: 'Worker health signal restored',
        observed_at: h(18),
      },
    ],
  }),
  frontend: componentHealth({
    currentHealth: 'healthy',
    uptimePercent: 100,
  }),
  cache: componentHealth({
    currentHealth: 'not-applicable',
    uptimePercent: 0,
  }),
}

const COMPONENT_HEALTH_BRANCH_MOVED = {
  api: componentHealth({
    currentHealth: 'degraded',
    uptimePercent: 99.2,
    patches: {
      0: {
        health: 'degraded',
        degraded_seconds: 1200,
        observed_seconds: 18 * 3600,
      },
    },
    transitions: [
      {
        from_health: 'healthy',
        to_health: 'degraded',
        message: 'Readiness checks degraded while the branch update applies',
        correlated_deploy_id: 'deploy-api-15',
        observed_at: h(1),
      },
    ],
  }),
  worker: componentHealth({
    currentHealth: 'degraded',
    uptimePercent: 99.1,
    patches: {
      0: {
        health: 'degraded',
        degraded_seconds: 1500,
        observed_seconds: 18 * 3600,
      },
    },
    transitions: [
      {
        from_health: 'healthy',
        to_health: 'degraded',
        message: 'Worker queue latency exceeded the threshold',
        observed_at: h(1),
      },
    ],
  }),
  frontend: COMPONENT_HEALTH_CURRENT.frontend,
  cache: COMPONENT_HEALTH_CURRENT.cache,
}

const COMPONENT_HEALTH_RESOURCE_LAG = {
  api: componentHealth({
    currentHealth: 'degraded',
    uptimePercent: 98.4,
    patches: {
      0: {
        health: 'degraded',
        degraded_seconds: 2400,
        observed_seconds: 18 * 3600,
      },
    },
    transitions: [
      {
        from_health: 'healthy',
        to_health: 'degraded',
        message: 'New API pods are not ready',
        correlated_deploy_id: 'deploy-api-16',
        observed_at: h(0.5),
      },
    ],
  }),
  worker: componentHealth({
    currentHealth: 'unhealthy',
    uptimePercent: 97.1,
    patches: {
      0: {
        health: 'unhealthy',
        unhealthy_seconds: 3600,
        observed_seconds: 18 * 3600,
      },
      1: {
        health: 'unhealthy',
        unhealthy_seconds: 1800,
        observed_seconds: DAY_SECONDS,
      },
    },
    transitions: [
      {
        from_health: 'degraded',
        to_health: 'unhealthy',
        message: 'Worker pods are failing readiness checks',
        correlated_deploy_id: 'deploy-worker-16',
        observed_at: h(0.5),
      },
    ],
  }),
  frontend: COMPONENT_HEALTH_CURRENT.frontend,
  cache: COMPONENT_HEALTH_CURRENT.cache,
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
      health: COMPONENT_HEALTH_CURRENT.api,
    },
    {
      id: 'cmp-2',
      name: 'worker',
      type: 'helm_chart',
      status: 'active' as const,
      deployedAt: h(4),
      sha: 'a1b2c3d4',
      health: COMPONENT_HEALTH_CURRENT.worker,
    },
    {
      id: 'cmp-3',
      name: 'frontend',
      type: 'helm_chart',
      status: 'active' as const,
      deployedAt: h(6),
      sha: 'e5f6a7b8',
      health: COMPONENT_HEALTH_CURRENT.frontend,
    },
    {
      id: 'cmp-4',
      name: 'cache',
      type: 'terraform_module',
      status: 'active' as const,
      deployedAt: h(24),
      sha: 'c9d0e1f2',
      health: COMPONENT_HEALTH_CURRENT.cache,
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

const COMMON_POLICIES = [
  {
    id: 'policy-1',
    name: 'Images must use immutable tags',
    componentName: 'api',
    status: 'active' as const,
    evaluatedAt: h(3),
  },
  {
    id: 'policy-2',
    name: 'Production workloads require limits',
    componentName: 'worker',
    status: 'active' as const,
    evaluatedAt: h(3),
  },
  {
    id: 'policy-3',
    name: 'Public services require TLS',
    componentName: 'frontend',
    status: 'warn' as const,
    evaluatedAt: h(3),
  },
]

const COMMON_RUNNER = {
  id: 'runner-01hzacmeprod',
  version: 'v0.22.4',
  status: 'active' as const,
  processes: [
    {
      id: 'proc-1',
      name: 'runner-primary',
      status: 'active' as const,
      startedAt: h(26),
    },
    {
      id: 'proc-2',
      name: 'runner-secondary',
      status: 'active' as const,
      startedAt: h(18),
    },
  ],
  recentJobs: [
    {
      id: 'job-1',
      name: 'Deploy cache',
      status: 'active' as const,
      createdAt: h(4),
    },
    {
      id: 'job-2',
      name: 'Apply sandbox',
      status: 'active' as const,
      createdAt: h(10),
    },
    {
      id: 'job-3',
      name: 'Scan infrastructure drift',
      status: 'active' as const,
      createdAt: h(24),
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
  inputVersions: [
    {
      id: 'inputs-v14',
      version: 'v14',
      title: 'Increase backup retention',
      createdAt: h(4),
      actor: 'alice',
      source: 'Install config',
      changes: [
        {
          path: 'Platform.cluster_size',
          operation: 'change',
          previousValue: 'medium',
          nextValue: 'large',
        },
        {
          path: 'Platform.retention_days',
          operation: 'change',
          previousValue: '14',
          nextValue: '30',
        },
      ],
    },
    {
      id: 'inputs-v13',
      version: 'v13',
      title: 'Rotate API credentials',
      createdAt: h(72),
      actor: 'bob',
      source: 'Dashboard',
      changes: [
        {
          path: 'Secrets.api_token',
          operation: 'change',
          isRedacted: true,
        },
      ],
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
  configFileVersions: [
    {
      id: 'config-v14',
      version: 'v14',
      title: 'Increase production capacity',
      createdAt: h(4),
      actor: 'alice',
      source: 'a1b2c3d4',
      changes: [
        {
          path: 'install.inputs.cluster_size',
          operation: 'change',
          previousValue: 'medium',
          nextValue: 'large',
        },
        {
          path: 'install.inputs.retention_days',
          operation: 'change',
          previousValue: '14',
          nextValue: '30',
        },
        {
          path: 'install.aws.iam_role_arn',
          operation: 'add',
          nextValue: 'arn:aws:iam::111122223333:role/nuon-acme-prod',
        },
      ],
      fileDiff: `-[install.inputs]
-cluster_size = "medium"
-retention_days = "14"
+[install.inputs]
+cluster_size = "large"
+retention_days = "30"
+
+[install.aws]
+iam_role_arn = "arn:aws:iam::111122223333:role/nuon-acme-prod"`,
    },
    {
      id: 'config-v13',
      version: 'v13',
      title: 'Enable production backups',
      createdAt: h(96),
      actor: 'carol',
      source: 'b3c4d5e6',
      changes: [
        {
          path: 'install.inputs.enable_backups',
          operation: 'add',
          nextValue: 'true',
        },
      ],
      fileDiff: `+[install.inputs]
+enable_backups = "true"
+retention_days = "14"`,
    },
  ],
  appBranchVersions: [
    {
      id: 'branch-run-v14',
      version: 'a1b2c3d4',
      title: 'Add Redis cache component',
      createdAt: h(4),
      actor: 'alice',
      source: 'main',
      changes: [
        {
          path: 'components.cache',
          operation: 'add',
          nextValue: 'terraform_module',
        },
        {
          path: 'inputs.cluster_size.default',
          operation: 'change',
          previousValue: 'medium',
          nextValue: 'large',
        },
      ],
    },
    {
      id: 'branch-run-v13',
      version: 'e5f6a7b8',
      title: 'Update frontend image',
      createdAt: h(48),
      actor: 'bob',
      source: 'main',
      changes: [
        {
          path: 'components.frontend.image.tag',
          operation: 'change',
          previousValue: '1.13.8',
          nextValue: '1.14.0',
        },
      ],
    },
  ],
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

const BRANCH_MOVED_CONFIGURATION: TPlaygroundConfiguration = {
  ...COMMON_CONFIGURATION,
  inputs: COMMON_CONFIGURATION.inputs.map((input) =>
    input.name === 'region' ? { ...input, value: 'us-east-1,us-west-2' } : input
  ),
  configFile: COMMON_CONFIGURATION.configFile && {
    ...COMMON_CONFIGURATION.configFile,
    gitBranch: 'feat/multi-region',
    version: 'v15',
    syncedAt: h(1),
    contents: CONFIG_FILE_CONTENTS.replace(
      'region = "us-east-1"',
      'region = "us-east-1,us-west-2"'
    ),
  },
  inputVersions: [
    {
      id: 'inputs-v15',
      version: 'v15',
      title: 'Add secondary region',
      createdAt: h(1),
      actor: 'carol',
      source: 'Install config',
      changes: [
        {
          path: 'Cloud.region',
          operation: 'change',
          previousValue: 'us-east-1',
          nextValue: 'us-east-1,us-west-2',
        },
      ],
    },
    ...COMMON_CONFIGURATION.inputVersions,
  ],
  configFileVersions: [
    {
      id: 'config-v15',
      version: 'v15',
      title: 'Add secondary region',
      createdAt: h(1),
      actor: 'carol',
      source: 'ff001234',
      changes: [
        {
          path: 'install.inputs.region',
          operation: 'change',
          previousValue: 'us-east-1',
          nextValue: 'us-east-1,us-west-2',
        },
      ],
      fileDiff: `-region = "us-east-1"
+region = "us-east-1,us-west-2"`,
    },
    ...COMMON_CONFIGURATION.configFileVersions,
  ],
  appBranchVersions: [
    {
      id: 'branch-run-v15',
      version: 'ff001234',
      title: 'Add secondary region support',
      createdAt: h(1),
      actor: 'carol',
      source: 'feat/multi-region',
      changes: [
        {
          path: 'install_inputs.region',
          operation: 'change',
          previousValue: 'us-east-1',
          nextValue: 'us-east-1,us-west-2',
        },
        {
          path: 'components.worker.env.AWS_REGION',
          operation: 'add',
          nextValue: 'us-west-2',
        },
      ],
    },
    ...COMMON_CONFIGURATION.appBranchVersions,
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

// ─── Deployment change prototypes ─────────────────────────────────────────────

const COMMON_DEPLOYMENTS: TDeploymentRecord[] = [
  {
    id: 'dep-install-config-8',
    type: 'install_config_update',
    status: 'active',
    createdAt: h(2),
    title: 'Install config updated',
    summary: 'Updated production inputs and triggered an install update.',
    workflow: {
      id: 'wf-install-update-8',
      name: 'Install update',
      type: 'install_update',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-8',
      sha: 'a1b2c3d4',
    },
    affectedResources: {
      components: ['api', 'worker'],
      images: [],
    },
    changeGroups: [
      {
        id: 'install-config-inputs',
        scope: 'install_config',
        label: 'Install config',
        summary: 'Two production inputs changed.',
        changes: [
          {
            path: 'inputs.replicas',
            operation: 'change',
            previousValue: '3',
            nextValue: '5',
          },
          {
            path: 'inputs.log_level',
            operation: 'change',
            previousValue: 'info',
            nextValue: 'warn',
          },
        ],
        fileDiff:
          '- replicas = 3\n+ replicas = 5\n- log_level = "info"\n+ log_level = "warn"',
        diffLanguage: 'toml',
      },
      {
        id: 'install-config-workflow',
        scope: 'workflow',
        label: 'Triggered workflow',
        summary: 'The install update workflow redeployed api and worker.',
        changes: [
          {
            path: 'steps.deploy_components.targets',
            operation: 'change',
            previousValue: 'api',
            nextValue: 'api, worker',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-stack-7',
    type: 'stack_update',
    status: 'active',
    createdAt: h(8),
    title: 'Stack changed',
    summary: 'Added a cache dependency and regenerated the stack workflow.',
    workflow: {
      id: 'wf-stack-update-7',
      name: 'Stack update',
      type: 'stack_update',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-7',
      sha: 'c9d0e1f2',
    },
    affectedResources: {
      stack: true,
      components: ['cache'],
      images: [],
    },
    changeGroups: [
      {
        id: 'stack-plan',
        scope: 'stack',
        label: 'Stack plan',
        summary: 'The shared Redis security group was added.',
        changes: [
          {
            path: 'resources.aws_security_group.redis',
            operation: 'add',
            nextValue: 'planned',
          },
          {
            path: 'outputs.redis_security_group_id',
            operation: 'add',
            nextValue: 'known after apply',
          },
        ],
      },
      {
        id: 'stack-workflow',
        scope: 'workflow',
        label: 'Workflow definition',
        summary: 'A cache deploy step was inserted after the stack apply.',
        changes: [
          {
            path: 'steps.deploy_cache',
            operation: 'add',
            nextValue: 'after: apply_stack',
          },
        ],
        fileDiff:
          '+ - id: deploy_cache\n+   after: apply_stack\n+   component: cache',
        diffLanguage: 'yaml',
      },
    ],
  },
  {
    id: 'dep-image-6',
    type: 'image_update',
    status: 'active',
    createdAt: h(18),
    title: 'API image updated',
    summary: 'Updated the API image to the latest patch release.',
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-6',
      sha: 'dd112233',
    },
    componentName: 'api',
    image: {
      repository: 'acme/api',
      previousTag: '1.14.2',
      nextTag: '1.14.3',
    },
    affectedResources: {
      components: ['api'],
      images: ['acme/api'],
    },
    changeGroups: [
      {
        id: 'api-image',
        scope: 'image',
        label: 'acme/api',
        resourceName: 'acme/api',
        summary: 'Image tag changed from 1.14.2 to 1.14.3.',
        changes: [
          {
            path: 'image.tag',
            operation: 'change',
            previousValue: '1.14.2',
            nextValue: '1.14.3',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-component-5',
    type: 'component_deploy',
    status: 'active',
    createdAt: h(30),
    title: 'Frontend deployed',
    summary: 'Deployed the latest frontend build.',
    workflow: {
      id: 'wf-component-deploy-5',
      name: 'Component deploy',
      type: 'component_deploy',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-5',
      sha: 'e5f6a7b8',
    },
    componentName: 'frontend',
    affectedResources: {
      components: ['frontend'],
      images: [],
    },
    changeGroups: [
      {
        id: 'frontend-deploy',
        scope: 'component',
        label: 'frontend',
        resourceName: 'frontend',
        summary: 'The frontend build and Helm values changed.',
        changes: [
          {
            path: 'build.sha',
            operation: 'change',
            previousValue: '8b7c6d5e',
            nextValue: 'e5f6a7b8',
          },
          {
            path: 'helm.values.featureFlags.checkout',
            operation: 'change',
            previousValue: 'false',
            nextValue: 'true',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-app-branch-4',
    type: 'app_branch_update',
    status: 'active',
    createdAt: h(48),
    title: 'App branch updated',
    summary: 'Moved the install from release/1.13 to main.',
    workflow: {
      id: 'wf-app-branch-update-4',
      name: 'App branch update',
      type: 'app_branch_update',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-4',
      sha: 'a1b2c3d4',
    },
    affectedResources: {
      stack: true,
      sandbox: true,
      components: ['api', 'worker', 'frontend'],
      images: [],
    },
    changeGroups: [
      {
        id: 'app-branch-target',
        scope: 'app_branch',
        label: 'App branch',
        summary: 'The tracked app branch and commit changed.',
        changes: [
          {
            path: 'app_branch.name',
            operation: 'change',
            previousValue: 'release/1.13',
            nextValue: 'main',
          },
          {
            path: 'app_branch.commit',
            operation: 'change',
            previousValue: '77aa8899',
            nextValue: 'a1b2c3d4',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-sandbox-3',
    type: 'sandbox_reprovision',
    status: 'active',
    createdAt: h(96),
    title: 'Sandbox reprovisioned',
    summary: 'Recreated the sandbox with the current network configuration.',
    workflow: {
      id: 'wf-sandbox-reprovision-3',
      name: 'Sandbox reprovision',
      type: 'sandbox_reprovision',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-3',
      sha: 'b3c4d5e6',
    },
    affectedResources: {
      sandbox: true,
      components: [],
      images: [],
    },
    changeGroups: [
      {
        id: 'sandbox-plan',
        scope: 'sandbox',
        label: 'Sandbox',
        summary:
          'Replaced the runner node group and updated its instance type.',
        changes: [
          {
            path: 'runner_node_group.instance_type',
            operation: 'change',
            previousValue: 'm6i.large',
            nextValue: 'm6i.xlarge',
          },
          {
            path: 'runner_node_group',
            operation: 'change',
            previousValue: 'existing',
            nextValue: 'replaced',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-reprovision-2',
    type: 'reprovision',
    status: 'success',
    createdAt: h(240),
    title: 'Install reprovisioned',
    summary: 'Rebuilt the stack, sandbox, and all install components.',
    workflow: {
      id: 'wf-reprovision-2',
      name: 'Install reprovision',
      type: 'reprovision',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-2',
      sha: '90ab12cd',
    },
    affectedResources: {
      stack: true,
      sandbox: true,
      components: ['api', 'worker', 'frontend'],
      images: [],
    },
    changeGroups: [
      {
        id: 'reprovision-stack',
        scope: 'stack',
        label: 'Stack',
        summary: 'Re-applied shared IAM and networking resources.',
        changes: [
          {
            path: 'stack.version',
            operation: 'change',
            previousValue: 'stkv-7b2e',
            nextValue: 'stkv-8a3f',
          },
        ],
      },
      {
        id: 'reprovision-sandbox',
        scope: 'sandbox',
        label: 'Sandbox',
        summary: 'Recreated the sandbox workspace.',
        changes: [
          {
            path: 'sandbox.generation',
            operation: 'change',
            previousValue: '18',
            nextValue: '19',
          },
        ],
      },
      {
        id: 'reprovision-components',
        scope: 'component',
        label: 'Components',
        summary: 'Redeployed api, worker, and frontend in dependency order.',
        changes: [
          {
            path: 'components.api',
            operation: 'change',
            previousValue: 'generation 41',
            nextValue: 'generation 42',
          },
          {
            path: 'components.worker',
            operation: 'change',
            previousValue: 'generation 27',
            nextValue: 'generation 28',
          },
          {
            path: 'components.frontend',
            operation: 'change',
            previousValue: 'generation 16',
            nextValue: 'generation 17',
          },
        ],
      },
    ],
  },
  {
    id: 'dep-provision-1',
    type: 'provision',
    status: 'success',
    createdAt: h(700),
    title: 'Install provisioned',
    summary: 'Created the stack, sandbox, and initial install components.',
    workflow: {
      id: 'wf-provision-1',
      name: 'Install provision',
      type: 'provision',
    },
    appBranch: {
      id: 'br-acme-main',
      name: 'main',
      runId: 'abr-1',
      sha: '1234abcd',
    },
    affectedResources: {
      stack: true,
      sandbox: true,
      components: ['api', 'worker', 'frontend'],
      images: [],
    },
    changeGroups: [
      {
        id: 'provision-stack',
        scope: 'stack',
        label: 'Stack',
        summary: 'Created shared IAM and network resources.',
        changes: [
          {
            path: 'stack',
            operation: 'add',
            nextValue: 'stkv-1a2b',
          },
        ],
      },
      {
        id: 'provision-sandbox',
        scope: 'sandbox',
        label: 'Sandbox',
        summary: 'Created the initial sandbox and runner.',
        changes: [
          {
            path: 'sandbox',
            operation: 'add',
            nextValue: 'sbx-01hzacmeprod',
          },
        ],
      },
      {
        id: 'provision-components',
        scope: 'component',
        label: 'Components',
        summary: 'Deployed api, worker, and frontend.',
        changes: [
          {
            path: 'components.api',
            operation: 'add',
            nextValue: 'deployed',
          },
          {
            path: 'components.worker',
            operation: 'add',
            nextValue: 'deployed',
          },
          {
            path: 'components.frontend',
            operation: 'add',
            nextValue: 'deployed',
          },
        ],
      },
    ],
  },
]

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
  orgName: 'acme',
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
  deployments: COMMON_DEPLOYMENTS,

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
    policies: COMMON_POLICIES,
    runner: COMMON_RUNNER,
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

  configuration: BRANCH_MOVED_CONFIGURATION,

  health: HEALTH_BRANCH_MOVED,
  resources: {
    ...configCurrentFixture.resources,
    components: configCurrentFixture.resources.components.map((component) => ({
      ...component,
      health:
        COMPONENT_HEALTH_BRANCH_MOVED[
          component.name as keyof typeof COMPONENT_HEALTH_BRANCH_MOVED
        ],
    })),
  },

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
        health: COMPONENT_HEALTH_RESOURCE_LAG.api,
      },
      {
        id: 'cmp-2',
        name: 'worker',
        type: 'helm_chart',
        status: 'pending',
        deployedAt: h(0.5),
        sha: 'dd112233',
        health: COMPONENT_HEALTH_RESOURCE_LAG.worker,
      },
      {
        id: 'cmp-3',
        name: 'frontend',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(6),
        sha: 'e5f6a7b8',
        health: COMPONENT_HEALTH_RESOURCE_LAG.frontend,
      },
      {
        id: 'cmp-4',
        name: 'cache',
        type: 'terraform_module',
        status: 'active',
        deployedAt: h(24),
        sha: 'c9d0e1f2',
        health: COMPONENT_HEALTH_RESOURCE_LAG.cache,
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
        health: COMPONENT_HEALTH_CURRENT.api,
      },
      {
        id: 'cmp-2',
        name: 'worker',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(4),
        sha: 'a1b2c3d4',
        health: COMPONENT_HEALTH_CURRENT.worker,
      },
      {
        id: 'cmp-3',
        name: 'frontend',
        type: 'helm_chart',
        status: 'active',
        deployedAt: h(6),
        sha: 'e5f6a7b8',
        health: COMPONENT_HEALTH_CURRENT.frontend,
      },
      {
        id: 'cmp-4',
        name: 'cache',
        type: 'terraform_module',
        status: 'warn',
        deployedAt: h(4),
        sha: 'c9d0e1f2',
        health: COMPONENT_HEALTH_CURRENT.cache,
      },
    ],
  },
}
