export interface IBranchFixture {
  id: string
  label: string
}

export interface IAppFixture {
  id: string
  nameWidth: number
  platformWidth: number
  branches: IBranchFixture[]
}

const mainBranch: IBranchFixture[] = [{ id: 'br-main', label: 'main' }]

const manyBranches: IBranchFixture[] = [
  { id: 'br-release', label: 'release' },
  { id: 'br-main', label: 'main' },
  { id: 'br-staging', label: 'staging' },
]

export const apps: IAppFixture[] = [
  { id: 'app-01', nameWidth: 140, platformWidth: 48, branches: manyBranches },
  { id: 'app-02', nameWidth: 96, platformWidth: 56, branches: mainBranch },
  { id: 'app-03', nameWidth: 172, platformWidth: 48, branches: manyBranches },
  { id: 'app-04', nameWidth: 118, platformWidth: 64, branches: mainBranch },
  { id: 'app-05', nameWidth: 156, platformWidth: 48, branches: mainBranch },
  { id: 'app-06', nameWidth: 88, platformWidth: 56, branches: manyBranches },
  { id: 'app-07', nameWidth: 132, platformWidth: 64, branches: [] },
]

export const appById = (appId: string) =>
  apps.find((app) => app.id === appId) ?? apps[0]

export const defaultBranch = (appId: string) => appById(appId).branches.at(0)

export interface IInstallFixture {
  id: string
  nameWidth: number
  appWidth: number
  platformWidth: number
  regionWidth: number
  labels: number[]
  branchWidth: number
}

export const installs: IInstallFixture[] = [
  {
    id: 'install-01',
    nameWidth: 150,
    appWidth: 96,
    platformWidth: 48,
    regionWidth: 72,
    labels: [44, 60],
    branchWidth: 64,
  },
  {
    id: 'install-02',
    nameWidth: 118,
    appWidth: 72,
    platformWidth: 56,
    regionWidth: 64,
    labels: [52],
    branchWidth: 88,
  },
  {
    id: 'install-03',
    nameWidth: 176,
    appWidth: 110,
    platformWidth: 48,
    regionWidth: 80,
    labels: [40, 56, 48],
    branchWidth: 56,
  },
  {
    id: 'install-04',
    nameWidth: 132,
    appWidth: 88,
    platformWidth: 64,
    regionWidth: 68,
    labels: [64],
    branchWidth: 72,
  },
  {
    id: 'install-05',
    nameWidth: 96,
    appWidth: 104,
    platformWidth: 48,
    regionWidth: 76,
    labels: [48, 44],
    branchWidth: 60,
  },
  {
    id: 'install-06',
    nameWidth: 160,
    appWidth: 80,
    platformWidth: 56,
    regionWidth: 64,
    labels: [],
    branchWidth: 84,
  },
  {
    id: 'install-07',
    nameWidth: 124,
    appWidth: 96,
    platformWidth: 48,
    regionWidth: 72,
    labels: [56, 40],
    branchWidth: 68,
  },
  {
    id: 'install-08',
    nameWidth: 142,
    appWidth: 68,
    platformWidth: 64,
    regionWidth: 80,
    labels: [44],
    branchWidth: 76,
  },
]

export interface IRunnable {
  label: string
  kind: 'action' | 'runbook'
  pinned?: boolean
}

export const runnables: IRunnable[] = [
  { label: 'Deploy', kind: 'runbook', pinned: true },
  { label: 'Rotate credentials', kind: 'action', pinned: true },
  { label: 'Restore backup', kind: 'runbook', pinned: true },
  { label: 'Health check', kind: 'action', pinned: true },
  { label: 'Sync secrets', kind: 'action', pinned: true },
  { label: 'Teardown', kind: 'runbook' },
  { label: 'Drain traffic', kind: 'action' },
  { label: 'Resume traffic', kind: 'action' },
  { label: 'Reindex search', kind: 'action' },
  { label: 'Flush cache', kind: 'action' },
  { label: 'Rotate certificates', kind: 'action' },
  { label: 'Scale up', kind: 'action' },
  { label: 'Scale down', kind: 'action' },
  { label: 'Snapshot database', kind: 'runbook' },
  { label: 'Migrate database', kind: 'runbook' },
  { label: 'Rollback release', kind: 'runbook' },
  { label: 'Collect diagnostics', kind: 'action' },
  { label: 'Verify connectivity', kind: 'action' },
  { label: 'Refresh inputs', kind: 'action' },
  { label: 'Recycle runner', kind: 'runbook' },
  { label: 'Reconcile drift', kind: 'runbook' },
  { label: 'Export audit log', kind: 'action' },
]

export interface IActivityGroup {
  label: string
  rows: { title: number }[]
}

export const activityGroups: IActivityGroup[] = [
  {
    label: 'Today',
    rows: [{ title: 180 }, { title: 148 }, { title: 210 }],
  },
  {
    label: 'Yesterday',
    rows: [{ title: 164 }, { title: 196 }, { title: 132 }, { title: 176 }],
  },
  {
    label: 'Earlier',
    rows: [{ title: 152 }, { title: 188 }],
  },
]

export interface IResourceFixture {
  id: string
  nameWidth: number
  typeWidth: number
}

export const resources: IResourceFixture[] = [
  { id: 'res-01', nameWidth: 190, typeWidth: 110 },
  { id: 'res-02', nameWidth: 140, typeWidth: 86 },
  { id: 'res-03', nameWidth: 212, typeWidth: 124 },
  { id: 'res-04', nameWidth: 168, typeWidth: 96 },
  { id: 'res-05', nameWidth: 124, typeWidth: 138 },
  { id: 'res-06', nameWidth: 202, typeWidth: 104 },
  { id: 'res-07', nameWidth: 156, typeWidth: 92 },
]

export interface IComponentFixture {
  id: string
  nameWidth: number
  typeWidth: number
}

export const installComponents: IComponentFixture[] = [
  { id: 'cmp-01', nameWidth: 168, typeWidth: 96 },
  { id: 'cmp-02', nameWidth: 132, typeWidth: 118 },
  { id: 'cmp-03', nameWidth: 196, typeWidth: 88 },
  { id: 'cmp-04', nameWidth: 148, typeWidth: 104 },
  { id: 'cmp-05', nameWidth: 178, typeWidth: 96 },
  { id: 'cmp-06', nameWidth: 120, typeWidth: 126 },
  { id: 'cmp-07', nameWidth: 162, typeWidth: 92 },
  { id: 'cmp-08', nameWidth: 140, typeWidth: 110 },
]

export interface IInputGroup {
  title: string
  rows: { key: number; value: number }[]
}

export const inputGroups: IInputGroup[] = [
  {
    title: 'Install inputs',
    rows: [
      { key: 120, value: 200 },
      { key: 96, value: 164 },
      { key: 140, value: 232 },
      { key: 108, value: 188 },
    ],
  },
  {
    title: 'Secrets',
    rows: [
      { key: 132, value: 120 },
      { key: 104, value: 120 },
    ],
  },
  {
    title: 'Derived values',
    rows: [
      { key: 116, value: 244 },
      { key: 152, value: 176 },
      { key: 128, value: 210 },
    ],
  },
]

export interface IInfrastructureStage {
  id: string
  label: string
  slug: string
  rows: number
}

export const infrastructureStages: IInfrastructureStage[] = [
  { id: 'stack', label: 'Stack', slug: 'stack', rows: 4 },
  { id: 'access', label: 'Roles & policies', slug: 'access', rows: 5 },
  { id: 'sandbox', label: 'Sandbox', slug: 'sandbox', rows: 3 },
  { id: 'runner', label: 'Runner', slug: 'runner', rows: 3 },
]

export const runRows = [180, 148, 212, 164, 196, 132, 176, 152]

export const settingsActions = [
  'Sync secrets',
  'Rotate credentials',
  'Reprovision',
  'Pause deploys',
  'Delete install',
]

export interface IActionFixture {
  id: string
  nameWidth: number
  typeWidth: number
}

export const actions: IActionFixture[] = [
  { id: 'act-01', nameWidth: 156, typeWidth: 88 },
  { id: 'act-02', nameWidth: 124, typeWidth: 104 },
  { id: 'act-03', nameWidth: 188, typeWidth: 92 },
  { id: 'act-04', nameWidth: 142, typeWidth: 110 },
  { id: 'act-05', nameWidth: 170, typeWidth: 86 },
  { id: 'act-06', nameWidth: 116, typeWidth: 120 },
]

export const configLines = [
  { key: 124, value: 210 },
  { key: 96, value: 178 },
  { key: 148, value: 240 },
  { key: 108, value: 164 },
  { key: 132, value: 196 },
]

export const outputRows = [
  { key: 118, value: 224 },
  { key: 142, value: 186 },
  { key: 104, value: 248 },
]

export const traceSpans = [
  { indent: 0, width: '92%' },
  { indent: 1, width: '74%' },
  { indent: 2, width: '52%' },
  { indent: 2, width: '46%' },
  { indent: 1, width: '68%' },
  { indent: 2, width: '38%' },
  { indent: 1, width: '58%' },
]

export interface IAccessFixture {
  id: string
  nameWidth: number
  metaWidth: number
}

export const roles: IAccessFixture[] = [
  { id: 'role-01', nameWidth: 176, metaWidth: 104 },
  { id: 'role-02', nameWidth: 142, metaWidth: 88 },
  { id: 'role-03', nameWidth: 198, metaWidth: 120 },
  { id: 'role-04', nameWidth: 130, metaWidth: 96 },
]

export const policies: IAccessFixture[] = [
  { id: 'pol-01', nameWidth: 164, metaWidth: 112 },
  { id: 'pol-02', nameWidth: 190, metaWidth: 92 },
  { id: 'pol-03', nameWidth: 126, metaWidth: 128 },
]

export interface IConfigRun {
  id: string
  labelWidth: number
}

export const configRuns: IConfigRun[] = Array.from({ length: 8 }, (_, i) => ({
  id: `run-${String(8 - i).padStart(2, '0')}`,
  labelWidth: 150 + ((i * 31) % 70),
}))

export type TDiffKind = 'added' | 'removed' | 'changed'

export interface IDiffEntry {
  id: string
  label: string
  section: string
  kind: TDiffKind
}

export const diffEntries: IDiffEntry[] = [
  { id: 'cmp-01', label: 'api', section: 'Components', kind: 'changed' },
  { id: 'cmp-04', label: 'scheduler', section: 'Components', kind: 'added' },
  { id: 'cmp-03', label: 'web', section: 'Components', kind: 'removed' },
  { id: 'sandbox', label: 'Sandbox', section: 'Sandbox', kind: 'changed' },
  { id: 'act-02', label: 'Health check', section: 'Actions', kind: 'changed' },
  { id: 'rbk-01', label: 'Restore', section: 'Runbooks', kind: 'added' },
]

export type TConfigKind =
  | 'stack'
  | 'role'
  | 'sandbox'
  | 'component'
  | 'action'
  | 'runbook'
  | 'policy'

export interface IBranchConfigItem {
  id: string
  label: string
  kind: TConfigKind
}

const componentNames = [
  'api',
  'worker',
  'web',
  'scheduler',
  'gateway',
  'ingest',
  'billing',
  'search',
  'notifier',
  'reporting',
  'auth',
  'cache',
  'queue',
  'migrator',
  'proxy',
  'exporter',
  'collector',
  'indexer',
  'streamer',
  'archiver',
  'audit',
  'metrics',
  'tracing',
  'docs',
]

const actionNames = [
  'Rotate credentials',
  'Health check',
  'Drain traffic',
  'Resume traffic',
  'Flush cache',
  'Reindex search',
  'Scale up',
  'Scale down',
  'Collect diagnostics',
  'Verify connectivity',
  'Refresh inputs',
  'Export audit log',
]

const runbookNames = [
  'Restore backup',
  'Snapshot database',
  'Migrate database',
  'Rollback release',
  'Recycle runner',
  'Reconcile drift',
  'Teardown',
  'Failover region',
]

const roleNames = [
  'deploy-role',
  'runner-role',
  'readonly-role',
  'bootstrap-role',
]

const policyNames = [
  'require-approval',
  'block-public-buckets',
  'enforce-tagging',
  'no-privileged-containers',
  'image-signature',
  'cost-ceiling',
]

export const branchConfigItems: IBranchConfigItem[] = [
  { id: 'stack', label: 'Stack', kind: 'stack' as const },
  ...componentNames.map((label, i) => ({
    id: `cmp-${String(i + 1).padStart(2, '0')}`,
    label,
    kind: 'component' as const,
  })),
  { id: 'sandbox', label: 'Sandbox', kind: 'sandbox' as const },
  ...actionNames.map((label, i) => ({
    id: `act-${String(i + 1).padStart(2, '0')}`,
    label,
    kind: 'action' as const,
  })),
  ...runbookNames.map((label, i) => ({
    id: `rbk-${String(i + 1).padStart(2, '0')}`,
    label,
    kind: 'runbook' as const,
  })),
  ...roleNames.map((label, i) => ({
    id: `rol-${String(i + 1).padStart(2, '0')}`,
    label,
    kind: 'role' as const,
  })),
  ...policyNames.map((label, i) => ({
    id: `pol-${String(i + 1).padStart(2, '0')}`,
    label,
    kind: 'policy' as const,
  })),
]

export interface IPolicyCheck {
  id: string
  label: string
  severity: string
  passed: boolean
}

export const policyChecks: IPolicyCheck[] = [
  {
    id: 'pol-01',
    label: 'require-approval',
    severity: 'blocking',
    passed: true,
  },
  {
    id: 'pol-02',
    label: 'block-public-buckets',
    severity: 'blocking',
    passed: true,
  },
  { id: 'pol-03', label: 'enforce-tagging', severity: 'warn', passed: false },
  {
    id: 'pol-04',
    label: 'no-privileged-containers',
    severity: 'blocking',
    passed: true,
  },
  { id: 'pol-06', label: 'cost-ceiling', severity: 'warn', passed: true },
]

export interface IApprovalFixture {
  id: string
  label: string
  installId: string
  waitingWidth: number
}

export const approvals: IApprovalFixture[] = [
  {
    id: 'inw-01',
    label: 'Deploy api',
    installId: 'install-02',
    waitingWidth: 96,
  },
  {
    id: 'inw-02',
    label: 'Branch config updated',
    installId: 'install-04',
    waitingWidth: 110,
  },
  {
    id: 'inw-03',
    label: 'Reconcile drift',
    installId: 'install-01',
    waitingWidth: 88,
  },
  {
    id: 'inw-04',
    label: 'Provision install',
    installId: 'install-07',
    waitingWidth: 120,
  },
]
