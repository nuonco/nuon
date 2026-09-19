import type {
  TInstallComponentHealthTimeline,
  TInstallHealthTimeline,
} from '@/types'

export type TResourceStatus =
  | 'active'
  | 'pending'
  | 'in-progress'
  | 'error'
  | 'warn'
  | 'deprovisioned'
  | 'unknown'

export type TLagItem = {
  name: string
  appliedVersion: string
  expectedVersion: string
  isCurrent: boolean
}

export type TConfigLag = {
  stack?: TLagItem
  sandbox?: TLagItem
  components: TLagItem[]
  images: TLagItem[]
}

export type TDriftedObject = {
  id: string
  targetType: 'install_deploy' | 'sandbox'
  componentName?: string
}

export type TStackVersion = {
  id: string
  version: string
  status: TResourceStatus
  planType: string
  createdAt: string
}

export type TRoleEntry = {
  id: string
  name: string
  type: string
  status: TResourceStatus
}

export type TSandboxInfo = {
  id: string
  status: TResourceStatus
  runType: string
  lastRunAt: string
  workspaceUrl?: string
}

export type TComponentEntry = {
  id: string
  name: string
  type: string
  status: TResourceStatus
  deployedAt: string
  sha?: string
  health: TInstallComponentHealthTimeline
}

export type TImageEntry = {
  id: string
  repository: string
  tag: string
  status: TResourceStatus
  builtAt: string
  sha?: string
}

export type TActionEntry = {
  id: string
  name: string
  description: string
  lastRunAt?: string
  lastRunStatus?: TResourceStatus
}

export type TRunbookEntry = {
  id: string
  name: string
  description: string
  stepCount: number
  lastRunAt?: string
  lastRunStatus?: TResourceStatus
}

export type TPolicyReportEntry = {
  id: string
  name: string
  componentName: string
  status: TResourceStatus
  evaluatedAt: string
}

export type TRunnerProcessEntry = {
  id: string
  name: string
  status: TResourceStatus
  startedAt: string
}

export type TRunnerJobEntry = {
  id: string
  name: string
  status: TResourceStatus
  createdAt: string
}

export type TRunnerInfo = {
  id: string
  version: string
  status: TResourceStatus
  processes: TRunnerProcessEntry[]
  recentJobs: TRunnerJobEntry[]
}

export type TPlaygroundResources = {
  stackVersions: TStackVersion[]
  roles: TRoleEntry[]
  sandbox?: TSandboxInfo
  components: TComponentEntry[]
  images: TImageEntry[]
}

export type TPlaygroundOperations = {
  actions: TActionEntry[]
  runbooks: TRunbookEntry[]
  policies: TPolicyReportEntry[]
  runner: TRunnerInfo
}

// ─── Configuration ────────────────────────────────────────────────────────────

export type TInputEntry = {
  name: string
  displayName: string
  value: string
  group: string
  isRedacted?: boolean
}

export type TConfigFileInfo = {
  path: string
  repo?: string
  gitBranch?: string
  version: string
  syncedAt: string
  contents: string
}

export type TOverrideEntry = {
  id: string
  componentName: string
  inputName: string
  value: string
  updatedAt: string
}

export type TConfigurationChange = {
  path: string
  operation: 'add' | 'remove' | 'change'
  previousValue?: string
  nextValue?: string
  isRedacted?: boolean
}

export type TConfigurationVersion = {
  id: string
  version: string
  title: string
  createdAt: string
  actor?: string
  source?: string
  changes: TConfigurationChange[]
  fileDiff?: string
}

export type TPlaygroundConfiguration = {
  inputs: TInputEntry[]
  inputVersions: TConfigurationVersion[]
  configFile?: TConfigFileInfo
  configFileVersions: TConfigurationVersion[]
  appBranchVersions: TConfigurationVersion[]
  overrides: TOverrideEntry[]
}

// ─── Activity event types ─────────────────────────────────────────────────────

export type TAppBranchSourceType = 'push' | 'pr' | 'tag' | 'commit' | 'manual'

export type TAppBranchSource = {
  type: TAppBranchSourceType
  branch?: string
  sha?: string
  author?: string
  prNumber?: number
  tag?: string
}

export type TActivityEventType =
  | 'app_branch_run'
  | 'deploy'
  | 'config_update'
  | 'inputs_update'
  | 'stack_update'
  | 'drift_scan'

export type TActivityEvent = {
  id: string
  type: TActivityEventType
  status: string
  createdAt: string
  title: string
  details?: string
  /** Present on app_branch_run and deploy events */
  source?: TAppBranchSource
  /** Name of the affected component, for deploy and drift_scan events */
  componentName?: string
}

// ─── Branch tracking ──────────────────────────────────────────────────────────

export type TBranchCommitRef = {
  sha: string
  message?: string
  author?: string
  createdAt?: string
  runStatus?: string
}

export type TBranchTrackingStatus = 'current' | 'pending' | 'updating'

export type TBranchTracking = {
  targetBranch: string
  branchId: string
  repo?: string
  gitBranch?: string
  directory?: string
  expectedCommit?: TBranchCommitRef
  appliedCommit?: TBranchCommitRef
  status: TBranchTrackingStatus
}

// ─── Install ──────────────────────────────────────────────────────────────────

export type TPlaygroundInstall = {
  id: string
  name: string
  orgId: string
  orgName: string
  appId: string
  appName: string
  labels: Record<string, string>
  isManagedByConfig: boolean
  configFilePath?: string
  createdAt: string
  updatedAt: string

  branchTracking: TBranchTracking

  runnerStatus: TResourceStatus
  sandboxStatus: TResourceStatus
  componentStatus: TResourceStatus

  configLag: TConfigLag
  driftedObjects: TDriftedObject[]
  activity: TActivityEvent[]

  resources: TPlaygroundResources
  operations: TPlaygroundOperations
  configuration: TPlaygroundConfiguration
  health: TInstallHealthTimeline
  readme?: string
}
