import type { TPlaygroundInstall } from './types'

const NOW = new Date('2026-09-18T18:00:00Z').getTime()
const h = (hours: number) => new Date(NOW - hours * 3600_000).toISOString()

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

const APPLIED_BRANCH: TPlaygroundInstall['appliedBranch'] = {
  id: 'branch-main',
  name: 'main',
  sha: 'a1b2c3d4',
  repo: 'acme/platform-configs',
  repoBranch: 'main',
  directory: 'apps/acme',
  commitMessage: 'Add Redis cache component',
  author: 'alice',
  createdAt: h(4),
}

const EXPECTED_BRANCH_MAIN = { ...APPLIED_BRANCH }

const EXPECTED_BRANCH_FEAT: TPlaygroundInstall['expectedBranch'] = {
  id: 'branch-feat',
  name: 'feat/multi-region',
  sha: 'ff001234',
  repo: 'acme/platform-configs',
  repoBranch: 'feat/multi-region',
  directory: 'apps/acme',
  commitMessage: 'Add secondary region support',
  author: 'carol',
  createdAt: h(1),
}

const COMMON_UPDATES: TPlaygroundInstall['updates'] = [
  {
    id: 'upd-5',
    type: 'deploy',
    status: 'active',
    createdAt: h(4),
    title: 'feat: add Redis cache component',
    trigger: {
      source: 'pr',
      prNumber: 218,
      branch: 'feat/cache',
      sha: 'a1b2c3d4',
      author: 'alice',
    },
  },
  {
    id: 'upd-4',
    type: 'branch_update',
    status: 'active',
    createdAt: h(10),
    title: 'Branch updated to main',
    trigger: { source: 'push', branch: 'main', sha: 'e5f6a7b8', author: 'bob' },
  },
  {
    id: 'upd-3',
    type: 'config_update',
    status: 'active',
    createdAt: h(24),
    title: 'fix: bump frontend image tag to 1.14.0',
    trigger: {
      source: 'push',
      branch: 'main',
      sha: 'c9d0e1f2',
      author: 'alice',
    },
  },
  {
    id: 'upd-2',
    type: 'inputs_update',
    status: 'active',
    createdAt: h(48),
    title: 'Inputs updated',
    details: 'license, region',
  },
  {
    id: 'upd-1',
    type: 'stack_update',
    status: 'active',
    createdAt: h(96),
    title: 'Stack version generated',
    trigger: {
      source: 'pr',
      prNumber: 201,
      branch: 'feat/permissions-refactor',
      sha: 'b3c4d5e6',
      author: 'carol',
    },
  },
]

// Scenario 1: Config-file-managed, everything current, no drift
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

  appliedBranch: APPLIED_BRANCH,
  expectedBranch: EXPECTED_BRANCH_MAIN,
  branchIsCurrent: true,

  runnerStatus: 'active',
  sandboxStatus: 'active',
  componentStatus: 'active',

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
  },

  driftedObjects: [],

  updates: COMMON_UPDATES,

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
}

// Scenario 2: Expected branch has moved (new branch target, old applied branch still running)
export const branchMovedFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  id: 'inst-01hzacmeprod',
  name: 'acme-prod',
  updatedAt: h(1),

  appliedBranch: APPLIED_BRANCH,
  expectedBranch: EXPECTED_BRANCH_FEAT,
  branchIsCurrent: false,

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

  updates: [
    {
      id: 'upd-6',
      type: 'branch_update',
      status: 'in-progress',
      createdAt: h(1),
      title: 'Branch target changed to feat/multi-region',
      trigger: {
        source: 'manual',
        branch: 'feat/multi-region',
        sha: 'ff001234',
        author: 'carol',
      },
    },
    ...COMMON_UPDATES,
  ],
}

// Scenario 3: Config is current but resources are lagging (some components/images pending apply)
export const resourceLagFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  id: 'inst-01hzacmeprod',
  name: 'acme-prod',
  updatedAt: h(0.5),

  appliedBranch: APPLIED_BRANCH,
  expectedBranch: EXPECTED_BRANCH_MAIN,
  branchIsCurrent: true,

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

  updates: [
    {
      id: 'upd-6',
      type: 'deploy',
      status: 'in-progress',
      createdAt: h(0.5),
      title: 'fix: patch CVE in base image',
      trigger: {
        source: 'push',
        branch: 'main',
        sha: 'dd112233',
        author: 'alice',
      },
    },
    ...COMMON_UPDATES,
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

// Scenario 4: Infrastructure drift detected while config is current
export const infraDriftFixture: TPlaygroundInstall = {
  ...configCurrentFixture,
  id: 'inst-01hzacmeprod',
  name: 'acme-prod',
  updatedAt: h(2),

  appliedBranch: APPLIED_BRANCH,
  expectedBranch: EXPECTED_BRANCH_MAIN,
  branchIsCurrent: true,

  runnerStatus: 'active',
  sandboxStatus: 'warn',
  componentStatus: 'active',

  // Config is fully applied — lag is zero
  configLag: configCurrentFixture.configLag,

  // Drift: someone changed sandbox Terraform state manually
  driftedObjects: [
    { id: 'dft-1', targetType: 'sandbox', componentName: undefined },
    { id: 'dft-2', targetType: 'install_deploy', componentName: 'cache' },
  ],

  updates: [
    {
      id: 'upd-7',
      type: 'deploy',
      status: 'warn',
      createdAt: h(2),
      title: 'Drift detected — sandbox and cache component',
      details: 'Drift scan found 2 objects with unexpected changes.',
    },
    ...COMMON_UPDATES,
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
