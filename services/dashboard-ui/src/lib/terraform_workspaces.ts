import { queryData, mutateData } from '@/utils'

export interface IGetWorkspace {
  orgId: string
  workspaceId: string
}

export interface IGetWorkspaceStates extends IGetWorkspace {}

export async function getWorkspaceStates({
  workspaceId,
  orgId,
}: IGetWorkspaceStates) {
  return queryData<any>({
    errorMessage: 'Unable to retrieve workspace states.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/state-json`,
  })
}

export interface IGetWorkspaceState extends IGetWorkspace {
  stateId: string
}

export async function getWorkspaceState({
  workspaceId,
  orgId,
  stateId,
}: IGetWorkspaceState) {
  return queryData<any>({
    errorMessage: 'Unable to retrieve workspace state.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/state-json/${stateId}`,
    abortTimeout: 15000,
  })
}

export interface IGetWorkspaceStateResources extends IGetWorkspace {
  stateId: string
}

export async function getWorkspaceStateResources({
  workspaceId,
  orgId,
  stateId,
}: IGetWorkspaceStateResources) {
  return queryData<any>({
    errorMessage: 'Unable to retrieve workspace resouces.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/state-json/${stateId}/resources`,
  })
}

export interface ILockWorkspace extends IGetWorkspace {}

export async function lockWorkspace({ workspaceId, orgId }: ILockWorkspace) {
  return mutateData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `terraform-workspaces/${workspaceId}/lock`,
  })
}

export interface IUnlockWorkspace extends IGetWorkspace {}

export async function unlockWorkspace({
  workspaceId,
  orgId,
}: IUnlockWorkspace) {
  return mutateData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `terraform-workspaces/${workspaceId}/unlock`,
  })
}
