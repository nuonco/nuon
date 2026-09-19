export type TBranchRef = {
  id: string
  name: string
  sha: string
  repo?: string
  repoBranch?: string
  directory?: string
  commitMessage?: string
  author?: string
  createdAt?: string
}

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

export type TUpdateTriggerSource = 'push' | 'pr' | 'manual'

export type TUpdateTrigger = {
  source: TUpdateTriggerSource
  prNumber?: number
  author?: string
  branch?: string
  sha?: string
}

export type TUpdateEventType =
  | 'deploy'
  | 'branch_update'
  | 'config_update'
  | 'inputs_update'
  | 'stack_update'

export type TUpdateEvent = {
  id: string
  type: TUpdateEventType
  status: string
  createdAt: string
  title: string
  details?: string
  trigger?: TUpdateTrigger
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
}

export type TPlaygroundInstall = {
  id: string
  name: string
  orgId: string
  appId: string
  appName: string
  labels: Record<string, string>
  isManagedByConfig: boolean
  configFilePath?: string
  createdAt: string
  updatedAt: string

  // Branch state — applied is what's running, expected is what config targets
  appliedBranch?: TBranchRef
  expectedBranch?: TBranchRef
  branchIsCurrent: boolean

  // Aggregate health statuses
  runnerStatus: TResourceStatus
  sandboxStatus: TResourceStatus
  componentStatus: TResourceStatus

  // Config lag: version drift between applied and expected for each resource class
  configLag: TConfigLag

  // Infrastructure drift: terraform/cloud state diverged from expected
  driftedObjects: TDriftedObject[]

  // Updates rail data
  updates: TUpdateEvent[]

  resources: TPlaygroundResources
  operations: TPlaygroundOperations
}
