export type TGroupInstallStatus =
  | 'success'
  | 'failed'
  | 'in-progress'
  | 'awaiting-approval'
  | 'pending'

export interface TCommitDetails {
  message: string
  author: string
  sha: string
  createdAt: string
  repo: string
  baseBranch: string
  pullRequest: {
    number: number
    title: string
  }
}

export interface TConfigChange {
  id: string
  kind: string
  name: string
  field?: string
  before: string
  after: string
}

export interface TGroupInstall {
  id: string
  name: string
  status: TGroupInstallStatus
  error?: string
}

export interface TDeploymentGroup {
  id: string
  name: string
  label: string
  version?: string
  attempts?: number
  startedAt?: string
  plannedCount: number
  changes: TConfigChange[]
  installs: TGroupInstall[]
}

export interface TDeploymentOverview {
  commit: TCommitDetails
  changes: TConfigChange[]
  groups: TDeploymentGroup[]
}

const imageChange: TConfigChange = {
  id: 'chg-image',
  kind: 'Image',
  name: 'runner',
  before: '1.4.1',
  after: '1.4.2',
}

const replicasChange: TConfigChange = {
  id: 'chg-replicas',
  kind: 'Component',
  name: 'api',
  field: 'replicas',
  before: '2',
  after: '3',
}

const logLevelChange: TConfigChange = {
  id: 'chg-log-level',
  kind: 'Input',
  name: 'log_level',
  before: 'info',
  after: 'debug',
}

const commit: TCommitDetails = {
  message: 'Release runner 1.4.2',
  author: 'Avery Chen',
  sha: 'a4c91e2b7d0f4c8a91e2b7d0f4c8a91e2b7d0f4c',
  createdAt: '2026-10-15T09:30:00.000Z',
  repo: 'acme/platform',
  baseBranch: 'main',
  pullRequest: {
    number: 128,
    title: 'Release runner 1.4.2',
  },
}

const use1 = (
  suffix: string,
  status: TGroupInstallStatus,
  error?: string
): TGroupInstall => ({
  id: `inst-use1-${suffix}`,
  name: `acme-use1-${suffix}`,
  status,
  error,
})

export const rollingOutOverview: TDeploymentOverview = {
  commit,
  changes: [imageChange, replicasChange, logLevelChange],
  groups: [
    {
      id: 'd0',
      name: 'd0',
      label: 'Canary',
      version: '1.4.2',
      attempts: 1,
      plannedCount: 1,
      changes: [imageChange],
      installs: [{ id: 'inst-canary', name: 'acme-canary', status: 'success' }],
    },
    {
      id: 'd1',
      name: 'd1',
      label: 'US East',
      version: '1.4.2',
      attempts: 3,
      startedAt: '2026-10-15T13:40:00.000Z',
      plannedCount: 9,
      changes: [imageChange, replicasChange, logLevelChange],
      installs: [
        use1('a', 'success'),
        use1('b', 'success'),
        use1('c', 'success'),
        use1('d', 'success'),
        use1('e', 'failed', 'Helm upgrade timed out waiting for api'),
        use1('f', 'failed', 'Image pull failed for runner:1.4.2'),
        use1('g', 'in-progress'),
        use1('h', 'awaiting-approval'),
        use1('i', 'awaiting-approval'),
      ],
    },
    {
      id: 'd2',
      name: 'd2',
      label: 'EU Central',
      version: '1.4.1',
      attempts: 2,
      plannedCount: 12,
      changes: [],
      installs: [],
    },
    {
      id: 'd3',
      name: 'd3',
      label: 'Holdout',
      plannedCount: 10,
      changes: [],
      installs: [],
    },
  ],
}

const regionLabels = [
  'Canary',
  'US East',
  'US West',
  'EU Central',
  'EU West',
  'AP Northeast',
  'AP Southeast',
  'CA Central',
  'SA East',
  'AF South',
  'ME Central',
  'IL Central',
  'AU East',
  'JP East',
  'KR Central',
  'Holdout',
]

const completedGroup = (index: number, label: string): TDeploymentGroup => {
  const name = `d${index}`
  return {
    id: name,
    name,
    label,
    version: '1.4.2',
    attempts: 1,
    plannedCount: 2,
    changes: [imageChange],
    installs: [
      { id: `${name}-a`, name: `acme-${name}-a`, status: 'success' },
      { id: `${name}-b`, name: `acme-${name}-b`, status: 'success' },
    ],
  }
}

const waitingGroup = (index: number, label: string): TDeploymentGroup => {
  const name = `d${index}`
  const attempted = index % 2 === 0
  return {
    id: name,
    name,
    label,
    version: attempted ? '1.4.1' : undefined,
    attempts: attempted ? 2 : undefined,
    plannedCount: attempted ? 6 : 4,
    changes: [],
    installs: [],
  }
}

const activeLongGroup = (index: number, label: string): TDeploymentGroup => {
  const name = `d${index}`
  const install = (
    suffix: string,
    status: TGroupInstallStatus,
    error?: string
  ): TGroupInstall => ({
    id: `${name}-${suffix}`,
    name: `acme-${name}-${suffix}`,
    status,
    error,
  })
  return {
    id: name,
    name,
    label,
    version: '1.4.2',
    attempts: 3,
    startedAt: '2026-10-15T13:40:00.000Z',
    plannedCount: 9,
    changes: [imageChange, replicasChange, logLevelChange],
    installs: [
      install('a', 'success'),
      install('b', 'success'),
      install('c', 'success'),
      install('d', 'success'),
      install('e', 'failed', 'Helm upgrade timed out waiting for api'),
      install('f', 'failed', 'Image pull failed for runner:1.4.2'),
      install('g', 'in-progress'),
      install('h', 'awaiting-approval'),
      install('i', 'awaiting-approval'),
    ],
  }
}

export const longPlanOverview: TDeploymentOverview = {
  commit,
  changes: [imageChange, replicasChange, logLevelChange],
  groups: regionLabels.map((label, index) => {
    if (index < 8) return completedGroup(index, label)
    if (index === 8) return activeLongGroup(index, label)
    return waitingGroup(index, label)
  }),
}

export const awaitingApprovalOverview: TDeploymentOverview = {
  commit,
  changes: [imageChange, replicasChange],
  groups: [
    {
      id: 'd0',
      name: 'd0',
      label: 'Canary',
      version: '1.4.2',
      attempts: 1,
      plannedCount: 1,
      changes: [imageChange],
      installs: [{ id: 'inst-canary', name: 'acme-canary', status: 'success' }],
    },
    {
      id: 'd1',
      name: 'd1',
      label: 'US East',
      version: '1.4.2',
      attempts: 1,
      plannedCount: 3,
      changes: [imageChange, replicasChange],
      installs: [
        use1('a', 'awaiting-approval'),
        use1('b', 'awaiting-approval'),
        use1('c', 'awaiting-approval'),
      ],
    },
    {
      id: 'd2',
      name: 'd2',
      label: 'EU Central',
      plannedCount: 4,
      changes: [],
      installs: [],
    },
  ],
}
