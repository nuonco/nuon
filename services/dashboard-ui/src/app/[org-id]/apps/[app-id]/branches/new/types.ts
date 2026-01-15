export interface IInstallGroup {
  id: string
  name: string
  installIds: string[]
  order: number
  requiresApproval: boolean
  rollbackOnFailure: boolean
  maxParallel: number
}

export interface IAppBranchConfig {
  name: string
  description?: string
  vcsEnabled: boolean
  vcsConnectionId?: string
  repository?: string
  gitBranch?: string
  directory?: string
  pathFilter?: string
  installGroups: IInstallGroup[]
}

export interface IFormData {
  branchName: string
  description?: string
  isManualOnly: boolean
  vcsConnection: string
  repo: string
  gitBranch: string
  directory: string
  pathFilter: string
  installGroups: IInstallGroup[]
  ungroupedInstalls: string[]
}

export interface IGroupTemplate {
  id: string
  name: string
  description: string
  groups: Array<{
    name: string
    requiresApproval: boolean
    rollbackOnFailure: boolean
    maxParallel: number
  }>
}