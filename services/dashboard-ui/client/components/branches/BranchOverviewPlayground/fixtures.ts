import type { TRunSource } from '@/components/branches/BranchOverview/run-source'

export type TRolloutState =
  | 'in-progress'
  | 'approval-awaiting'
  | 'success'
  | 'error'
  | 'cancelled'

export type TStageState =
  | 'success'
  | 'in-progress'
  | 'approval-awaiting'
  | 'error'
  | 'pending'
  | 'auto-skipped'

export type TInstallState = 'success' | 'in-progress' | 'error' | 'pending'

export type TStageInstall = {
  id: string
  name: string
  region: string
  state: TInstallState
  durationNs?: number
}

export type TStage = {
  id: string
  kind: 'commit' | 'build' | 'group'
  name: string
  state: TStageState
  detail?: string
  durationNs?: number
  groupId?: string
  installs?: TStageInstall[]
}

export type TChangeOp = 'add' | 'change' | 'remove'

export type TChange = {
  op: TChangeOp
  name: string
  detail: string
}

export type TRollout = {
  id: string
  number: number
  source: TRunSource
  title: string
  sha: string
  author: string
  state: TRolloutState
  startedAt: string
  finishedAt?: string
  durationNs?: number
  stages: TStage[]
  changes: TChange[]
  activity?: string
}

export type TPlanGroup = {
  id: string
  name: string
  selector?: Record<string, string>
  installCount: number
  maxParallel: number
  approval: 'auto' | 'manual'
}

export type TRecentRun = {
  id: string
  number: number
  title: string
  sha: string
  author: string
  createdAt: string
  state: TRolloutState
  outcome: string
}

export type TTemplateEntry = {
  name: string
  detail: string
}

export type TTemplateItem = {
  id: string
  label: string
  description: string
  entries: TTemplateEntry[]
}

export type TBranchOverview = {
  appName: string
  branchName: string
  repo: string
  directory: string
  trigger: string
  configVersion: number
  groups: TPlanGroup[]
  rollout?: TRollout
  recentRuns: TRecentRun[]
  template: TTemplateItem[]
}

const MINUTE = 60_000
const SECOND_NS = 1_000_000_000
const MINUTE_NS = 60 * SECOND_NS

const ago = (minutes: number) =>
  new Date(Date.now() - minutes * MINUTE).toISOString()

const ns = (minutes: number, seconds = 0) =>
  minutes * MINUTE_NS + seconds * SECOND_NS

export const PLAN_GROUPS: TPlanGroup[] = [
  {
    id: 'grp_canary',
    name: 'Canary',
    selector: { env: 'prod', tier: 'canary' },
    installCount: 2,
    maxParallel: 2,
    approval: 'auto',
  },
  {
    id: 'grp_primary',
    name: 'Primary region',
    selector: { env: 'prod', region: '*' },
    installCount: 5,
    maxParallel: 2,
    approval: 'manual',
  },
  {
    id: 'grp_rest',
    name: 'Remaining',
    installCount: 14,
    maxParallel: 4,
    approval: 'auto',
  },
]

const CANARY_INSTALLS: TStageInstall[] = [
  {
    id: 'inst_alpha',
    name: 'alpha',
    region: 'us-east-1',
    state: 'success',
    durationNs: ns(1, 58),
  },
  {
    id: 'inst_bravo',
    name: 'bravo',
    region: 'us-west-2',
    state: 'success',
    durationNs: ns(2, 14),
  },
]

const primaryInstalls = (states: TInstallState[]): TStageInstall[] =>
  [
    {
      id: 'inst_charlie',
      name: 'charlie',
      region: 'us-east-1',
      durationNs: ns(1, 52),
    },
    {
      id: 'inst_echo',
      name: 'echo',
      region: 'us-west-2',
      durationNs: ns(2, 10),
    },
    {
      id: 'inst_golf',
      name: 'golf',
      region: 'eu-west-1',
      durationNs: ns(1, 47),
    },
    {
      id: 'inst_delta',
      name: 'delta',
      region: 'us-east-2',
      durationNs: ns(2, 34),
    },
    {
      id: 'inst_hotel',
      name: 'hotel',
      region: 'ap-south-1',
      durationNs: ns(1, 59),
    },
  ].map((install, idx) => ({
    ...install,
    state: states[idx],
    durationNs: states[idx] === 'success' ? install.durationNs : undefined,
  }))

const REST_NAMES = [
  'india',
  'juliet',
  'kilo',
  'lima',
  'mike',
  'november',
  'oscar',
  'papa',
  'quebec',
  'romeo',
  'sierra',
  'tango',
  'uniform',
  'victor',
]
const REST_REGIONS = [
  'us-east-1',
  'us-west-2',
  'eu-central-1',
  'ap-southeast-2',
]

const restInstalls = (state: TInstallState): TStageInstall[] =>
  REST_NAMES.map((name, idx) => ({
    id: `inst_${name}`,
    name,
    region: REST_REGIONS[idx % REST_REGIONS.length],
    state,
    durationNs: state === 'success' ? ns(1, 30 + idx * 3) : undefined,
  }))

const commitStage = (sha: string): TStage => ({
  id: 'stage_commit',
  kind: 'commit',
  name: 'Commit',
  state: 'success',
  detail: sha.slice(0, 7),
  durationNs: ns(0, 4),
})

const buildStage = (state: TStageState = 'success'): TStage => ({
  id: 'stage_build',
  kind: 'build',
  name: 'Build',
  state,
  detail: '3 components',
  durationNs: state === 'success' ? ns(2, 31) : undefined,
})

const groupStage = (
  group: TPlanGroup,
  state: TStageState,
  installs: TStageInstall[],
  durationNs?: number
): TStage => ({
  id: `stage_${group.id}`,
  kind: 'group',
  name: group.name,
  state,
  groupId: group.id,
  installs,
  durationNs,
})

const CACHE_CHANGES: TChange[] = [
  { op: 'add', name: 'cache', detail: 'Terraform module' },
  { op: 'change', name: 'api', detail: 'Image tag 1.4.2' },
  { op: 'change', name: 'worker', detail: 'Helm values' },
]

const CREDENTIAL_CHANGES: TChange[] = [
  { op: 'change', name: 'database', detail: 'Rotated credentials' },
  { op: 'change', name: 'worker', detail: 'Secret ref' },
]

const [CANARY, PRIMARY, REST] = PLAN_GROUPS

const RECENT_RUNS: TRecentRun[] = [
  {
    id: 'run_183',
    number: 183,
    title: 'Bump api to 1.4.2',
    sha: '9f8e7d6c5b4a',
    author: 'ci@example.com',
    createdAt: ago(180),
    state: 'success',
    outcome: '21 installs updated',
  },
  {
    id: 'run_182',
    number: 182,
    title: 'Rotate db credentials',
    sha: '4c5d6e7f8a9b',
    author: 'sam@example.com',
    createdAt: ago(60 * 26),
    state: 'error',
    outcome: 'Failed in Primary region',
  },
  {
    id: 'run_181',
    number: 181,
    title: 'Add worker autoscaling',
    sha: '1a2b3c4d5e6f',
    author: 'jane@example.com',
    createdAt: ago(60 * 50),
    state: 'success',
    outcome: '21 installs updated',
  },
]

const TEMPLATE: TTemplateItem[] = [
  {
    id: 'components',
    label: 'Components',
    description: 'What gets built and deployed to every install.',
    entries: [
      { name: 'api', detail: 'Helm chart' },
      { name: 'worker', detail: 'Helm chart' },
      { name: 'cache', detail: 'Terraform module' },
      { name: 'database', detail: 'Terraform module' },
      { name: 'ingress', detail: 'Kubernetes manifest' },
      { name: 'api-image', detail: 'Docker build' },
    ],
  },
  {
    id: 'inputs',
    label: 'Inputs',
    description: 'Values each install provides at setup.',
    entries: [
      { name: 'domain', detail: 'Required' },
      { name: 'instance_type', detail: 'Default m6i.large' },
      { name: 'replica_count', detail: 'Default 3' },
      { name: 'enable_cache', detail: 'Default true' },
    ],
  },
  {
    id: 'actions',
    label: 'Actions',
    description: 'Scripts that run on a trigger or on demand.',
    entries: [
      { name: 'migrate-db', detail: 'Before deploy' },
      { name: 'healthcheck', detail: 'Every 5 minutes' },
      { name: 'rotate-keys', detail: 'Manual' },
      { name: 'warm-cache', detail: 'After deploy' },
    ],
  },
  {
    id: 'runbooks',
    label: 'Runbooks',
    description: 'Multi-step operations for day-2 tasks.',
    entries: [
      { name: 'restore-backup', detail: '4 steps' },
      { name: 'scale-out', detail: '2 steps' },
    ],
  },
  {
    id: 'sandbox',
    label: 'Sandboxes',
    description: 'The base infrastructure each install runs in.',
    entries: [{ name: 'aws-eks', detail: 'nuonco/aws-eks-sandbox v0.9.2' }],
  },
  {
    id: 'policies',
    label: 'Policies',
    description: 'Checks a plan must pass before it applies.',
    entries: [
      { name: 'no-public-buckets', detail: 'Terraform' },
      { name: 'require-resource-limits', detail: 'Kubernetes' },
      { name: 'allowed-regions', detail: 'Terraform' },
    ],
  },
  {
    id: 'roles',
    label: 'Roles',
    description: 'Cloud permissions the runner assumes.',
    entries: [
      { name: 'provision', detail: 'Setup and teardown' },
      { name: 'maintenance', detail: 'Deploys and actions' },
    ],
  },
  {
    id: 'labels',
    label: 'Labels',
    description: 'Keys used to group installs.',
    entries: [
      { name: 'env', detail: 'prod, staging' },
      { name: 'tier', detail: 'canary, standard' },
      { name: 'region', detail: '6 values' },
      { name: 'customer_size', detail: 'small, large' },
      { name: 'team', detail: '3 values' },
    ],
  },
  {
    id: 'readme',
    label: 'README',
    description: 'Notes shown to operators on every install.',
    entries: [{ name: 'README.md', detail: '48 lines' }],
  },
]

const BASE: Omit<TBranchOverview, 'rollout' | 'recentRuns'> = {
  appName: 'acme-platform',
  branchName: 'main',
  repo: 'acme/platform',
  directory: '/deploy',
  trigger: 'Every push to main',
  configVersion: 14,
  groups: PLAN_GROUPS,
  template: TEMPLATE,
}

const LIVE_SHA = 'a1b2c3d4e5f6'

export const rollingOutFixture: TBranchOverview = {
  ...BASE,
  recentRuns: RECENT_RUNS,
  rollout: {
    id: 'run_184',
    number: 184,
    source: {
      kind: 'pull-request',
      number: 482,
      url: 'https://github.com/acme/platform/pull/482',
      label: 'deploy',
      baseBranch: 'main',
    },
    title: 'Add cache component',
    sha: LIVE_SHA,
    author: 'jane@example.com',
    state: 'in-progress',
    startedAt: ago(12),
    activity: 'delta started deploying',
    changes: CACHE_CHANGES,
    stages: [
      commitStage(LIVE_SHA),
      buildStage(),
      groupStage(CANARY, 'success', CANARY_INSTALLS, ns(4, 12)),
      groupStage(
        PRIMARY,
        'in-progress',
        primaryInstalls([
          'success',
          'success',
          'success',
          'in-progress',
          'pending',
        ])
      ),
      groupStage(REST, 'pending', restInstalls('pending')),
    ],
  },
}

export const awaitingApprovalFixture: TBranchOverview = {
  ...BASE,
  recentRuns: RECENT_RUNS,
  rollout: {
    id: 'run_184',
    number: 184,
    source: {
      kind: 'pull-request',
      number: 482,
      url: 'https://github.com/acme/platform/pull/482',
      label: 'deploy',
      baseBranch: 'main',
    },
    title: 'Add cache component',
    sha: LIVE_SHA,
    author: 'jane@example.com',
    state: 'approval-awaiting',
    startedAt: ago(9),
    activity: 'Plan ready for 5 installs in Primary region',
    changes: CACHE_CHANGES,
    stages: [
      commitStage(LIVE_SHA),
      buildStage(),
      groupStage(CANARY, 'success', CANARY_INSTALLS, ns(4, 12)),
      groupStage(
        PRIMARY,
        'approval-awaiting',
        primaryInstalls(['pending', 'pending', 'pending', 'pending', 'pending'])
      ),
      groupStage(REST, 'pending', restInstalls('pending')),
    ],
  },
}

export const succeededFixture: TBranchOverview = {
  ...BASE,
  recentRuns: RECENT_RUNS.slice(1),
  rollout: {
    id: 'run_183',
    number: 183,
    source: {
      kind: 'tag',
      tag: 'v1.4.2',
      url: 'https://github.com/acme/platform/releases/tag/v1.4.2',
    },
    title: 'Bump api to 1.4.2',
    sha: '9f8e7d6c5b4a',
    author: 'ci@example.com',
    state: 'success',
    startedAt: ago(199),
    finishedAt: ago(180),
    durationNs: ns(18, 42),
    activity: '21 installs updated, 0 failed',
    changes: CACHE_CHANGES,
    stages: [
      commitStage('9f8e7d6c5b4a'),
      buildStage(),
      groupStage(CANARY, 'success', CANARY_INSTALLS, ns(4, 12)),
      groupStage(
        PRIMARY,
        'success',
        primaryInstalls([
          'success',
          'success',
          'success',
          'success',
          'success',
        ]),
        ns(6, 20)
      ),
      groupStage(REST, 'success', restInstalls('success'), ns(5, 37)),
    ],
  },
}

export const failedFixture: TBranchOverview = {
  ...BASE,
  recentRuns: [RECENT_RUNS[0], ...RECENT_RUNS.slice(2)],
  rollout: {
    id: 'run_182',
    number: 182,
    source: { kind: 'commit' },
    title: 'Rotate db credentials',
    sha: '4c5d6e7f8a9b',
    author: 'sam@example.com',
    state: 'error',
    startedAt: ago(60 * 26 + 9),
    finishedAt: ago(60 * 26),
    durationNs: ns(9, 14),
    activity: 'delta failed during deploy',
    changes: CREDENTIAL_CHANGES,
    stages: [
      commitStage('4c5d6e7f8a9b'),
      buildStage(),
      groupStage(CANARY, 'success', CANARY_INSTALLS, ns(3, 48)),
      groupStage(
        PRIMARY,
        'error',
        primaryInstalls(['success', 'success', 'pending', 'error', 'pending']),
        ns(3, 2)
      ),
      groupStage(REST, 'auto-skipped', restInstalls('pending')),
    ],
  },
}

export const idleFixture: TBranchOverview = {
  ...BASE,
  recentRuns: [],
}

export const noPlanFixture: TBranchOverview = {
  ...BASE,
  groups: [],
  recentRuns: [],
}
