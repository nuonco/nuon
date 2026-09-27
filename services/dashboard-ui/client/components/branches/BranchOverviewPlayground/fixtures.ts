import type { TRunSource } from '@/components/branches/BranchOverview/run-source'
import type { TComponentType } from '@/types'

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

export type TComponentFixture = {
  id: string
  type: TComponentType
  version: number
  buildTimeout: string
  deployTimeout: string
  config: { label: string; value: string }[]
  dependents?: string[]
  source: {
    message: string
    author: string
    sha: string
    createdAt: string
  }
}

export type TTemplateDetail = {
  summary: string
  fields: { label: string; value: string }[]
  dependencies?: string[]
  steps?: { name: string; detail: string }[]
  builds?: { sha: string; status: string; when: string }[]
  component?: TComponentFixture
}

export type TTemplateEntry = {
  name: string
  cells: string[]
  detail?: TTemplateDetail
}

export type TTemplateItem = {
  id: string
  label: string
  description: string
  columns: string[]
  entries: TTemplateEntry[]
}

export type TBranchOverview = {
  appName: string
  branchName: string
  repo: string
  gitBranch: string
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

const COMPONENT_TYPE_LABEL: Partial<Record<TComponentType, string>> = {
  helm_chart: 'Helm chart',
  terraform_module: 'Terraform module',
  kubernetes_manifest: 'Kubernetes manifest',
  docker_build: 'Docker build',
}

const componentEntry = ({
  name,
  id,
  type,
  status,
  summary,
  config,
  dependencies,
  dependents,
  minutesAgo,
}: {
  name: string
  id: string
  type: TComponentType
  status: 'success' | 'in-progress' | 'error'
  summary: string
  config: { label: string; value: string }[]
  dependencies?: string[]
  dependents?: string[]
  minutesAgo: number
}): TTemplateEntry => ({
  name,
  cells: [
    COMPONENT_TYPE_LABEL[type] ?? type,
    status === 'in-progress' ? 'Building' : 'Succeeded',
  ],
  detail: {
    summary,
    fields: [],
    dependencies,
    builds: [
      {
        sha: 'a1b2c3d4',
        status,
        when: `${minutesAgo} minutes ago`,
      },
    ],
    component: {
      id,
      type,
      version: 14,
      buildTimeout: '1h',
      deployTimeout: '1h',
      config,
      dependents,
      source: {
        message: 'Bump api image to 1.4.2',
        author: 'jane@example.com',
        sha: 'a1b2c3d4e5f6',
        createdAt: ago(minutesAgo),
      },
    },
  },
})

const TEMPLATE: TTemplateItem[] = [
  {
    id: 'components',
    label: 'Components',
    description: 'The components defined by this branch\u2019s configuration.',
    columns: ['Component', 'Type', 'Latest build'],
    entries: [
      componentEntry({
        name: 'api',
        id: 'cmpxk2a9apiv8d3h1q7rn0wz',
        type: 'helm_chart',
        status: 'success',
        summary: 'Helm chart deployed to every install.',
        config: [
          { label: 'Chart name', value: 'api' },
          { label: 'Namespace', value: 'acme' },
          { label: 'Storage driver', value: 'secret' },
        ],
        dependencies: ['database', 'api-image'],
        dependents: ['ingress'],
        minutesAgo: 12,
      }),
      componentEntry({
        name: 'worker',
        id: 'cmpw4rk8e2rq9s1t6m3bz5yu',
        type: 'helm_chart',
        status: 'success',
        summary: 'Background worker chart.',
        config: [
          { label: 'Chart name', value: 'worker' },
          { label: 'Namespace', value: 'acme' },
          { label: 'Storage driver', value: 'secret' },
        ],
        dependencies: ['database'],
        minutesAgo: 12,
      }),
      componentEntry({
        name: 'cache',
        id: 'cmpc4ch3m0d7l2k9p5x8vq1n',
        type: 'terraform_module',
        status: 'in-progress',
        summary: 'Terraform module for the shared cache.',
        config: [{ label: 'Terraform version', value: '1.9.5' }],
        minutesAgo: 4,
      }),
      componentEntry({
        name: 'database',
        id: 'cmpd8b4s3e6m1q0r7t2y9w5k',
        type: 'terraform_module',
        status: 'success',
        summary: 'Terraform module for the primary database.',
        config: [{ label: 'Terraform version', value: '1.9.5' }],
        dependents: ['api', 'worker'],
        minutesAgo: 180,
      }),
      componentEntry({
        name: 'ingress',
        id: 'cmpi9n6g2r5e8s1s4x7z0c3v',
        type: 'kubernetes_manifest',
        status: 'success',
        summary: 'Ingress manifest for the public API.',
        config: [{ label: 'Namespace', value: 'acme' }],
        dependencies: ['api'],
        minutesAgo: 12,
      }),
      componentEntry({
        name: 'api-image',
        id: 'cmpa1m5g8e3i6m9a2g4e7b0d',
        type: 'docker_build',
        status: 'success',
        summary: 'Container image for the API.',
        config: [
          { label: 'Dockerfile name', value: 'Dockerfile' },
          { label: 'Target', value: 'release' },
        ],
        dependents: ['api'],
        minutesAgo: 14,
      }),
    ],
  },
  {
    id: 'inputs',
    label: 'Inputs',
    description: 'Values each install provides at setup.',
    columns: ['Input', 'Required', 'Default'],
    entries: [
      { name: 'domain', cells: ['Yes', 'None'] },
      { name: 'instance_type', cells: ['No', 'm6i.large'] },
      { name: 'replica_count', cells: ['No', '3'] },
      { name: 'enable_cache', cells: ['No', 'true'] },
    ],
  },
  {
    id: 'actions',
    label: 'Actions',
    description: 'Scripts that run on a trigger or on demand.',
    columns: ['Action', 'Triggers', 'Steps'],
    entries: [
      {
        name: 'migrate-db',
        cells: ['Before deploy', '2'],
        detail: {
          summary: 'Runs before each deploy.',
          fields: [{ label: 'Trigger', value: 'Before deploy' }],
          steps: [
            { name: 'snapshot', detail: 'Take a database snapshot' },
            { name: 'migrate', detail: 'Apply pending migrations' },
          ],
        },
      },
      {
        name: 'healthcheck',
        cells: ['Every 5 minutes', '1'],
        detail: {
          summary: 'Checks that the API is serving traffic.',
          fields: [{ label: 'Trigger', value: 'Every 5 minutes' }],
          steps: [{ name: 'probe', detail: 'GET /healthz' }],
        },
      },
      {
        name: 'rotate-keys',
        cells: ['Manual', '3'],
        detail: {
          summary: 'Rotates signing keys on demand.',
          fields: [{ label: 'Trigger', value: 'Manual' }],
          steps: [
            { name: 'generate', detail: 'Create a new key' },
            { name: 'publish', detail: 'Publish the public key' },
            { name: 'retire', detail: 'Retire the previous key' },
          ],
        },
      },
      {
        name: 'warm-cache',
        cells: ['After deploy', '1'],
        detail: {
          summary: 'Warms the cache after a deploy.',
          fields: [{ label: 'Trigger', value: 'After deploy' }],
          steps: [{ name: 'warm', detail: 'Request the top routes' }],
        },
      },
    ],
  },
  {
    id: 'runbooks',
    label: 'Runbooks',
    description: 'Multi-step operations for day-2 tasks.',
    columns: ['Runbook', 'Steps'],
    entries: [
      {
        name: 'restore-backup',
        cells: ['4'],
        detail: {
          summary: 'Restores a database backup.',
          fields: [{ label: 'Steps', value: '4' }],
          steps: [
            { name: 'select', detail: 'Choose a snapshot' },
            { name: 'stop', detail: 'Stop writers' },
            { name: 'restore', detail: 'Restore the snapshot' },
            { name: 'verify', detail: 'Check row counts' },
          ],
        },
      },
      {
        name: 'scale-out',
        cells: ['2'],
        detail: {
          summary: 'Adds capacity to the worker group.',
          fields: [{ label: 'Steps', value: '2' }],
          steps: [
            { name: 'resize', detail: 'Raise the replica count' },
            { name: 'wait', detail: 'Wait until the new pods are ready' },
          ],
        },
      },
    ],
  },
  {
    id: 'sandbox',
    label: 'Sandboxes',
    description: 'The base infrastructure each install runs in.',
    columns: ['Sandbox', 'Version'],
    entries: [{ name: 'aws-eks', cells: ['nuonco/aws-eks-sandbox v0.9.2'] }],
  },
  {
    id: 'policies',
    label: 'Policies',
    description: 'Checks a plan must pass before it applies.',
    columns: ['Policy', 'Applies to'],
    entries: [
      {
        name: 'no-public-buckets',
        cells: ['Terraform'],
        detail: {
          summary: 'Denies plans that make a storage bucket public.',
          fields: [{ label: 'Applies to', value: 'Terraform' }],
        },
      },
      {
        name: 'require-resource-limits',
        cells: ['Kubernetes'],
        detail: {
          summary: 'Requires CPU and memory limits on every container.',
          fields: [{ label: 'Applies to', value: 'Kubernetes' }],
        },
      },
      {
        name: 'allowed-regions',
        cells: ['Terraform'],
        detail: {
          summary: 'Allows resources only in the configured regions.',
          fields: [{ label: 'Applies to', value: 'Terraform' }],
        },
      },
    ],
  },
  {
    id: 'roles',
    label: 'Roles',
    description: 'Cloud permissions the runner assumes.',
    columns: ['Role', 'Used for'],
    entries: [
      { name: 'provision', cells: ['Setup and teardown'] },
      { name: 'maintenance', cells: ['Deploys and actions'] },
    ],
  },
  {
    id: 'labels',
    label: 'Labels',
    description: 'Keys used to group installs.',
    columns: ['Label', 'Values'],
    entries: [
      { name: 'env', cells: ['prod, staging'] },
      { name: 'tier', cells: ['canary, standard'] },
      { name: 'region', cells: ['6 values'] },
      { name: 'customer_size', cells: ['small, large'] },
      { name: 'team', cells: ['3 values'] },
    ],
  },
  {
    id: 'readme',
    label: 'README',
    description: 'Notes shown to operators on every install.',
    columns: ['File', 'Length'],
    entries: [{ name: 'README.md', cells: ['48 lines'] }],
  },
]

const BASE: Omit<TBranchOverview, 'rollout' | 'recentRuns'> = {
  appName: 'acme-platform',
  branchName: 'main',
  repo: 'acme/platform',
  gitBranch: 'main',
  directory: '/deploy',
  trigger: 'Every push',
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
    source: { kind: 'manual' },
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
