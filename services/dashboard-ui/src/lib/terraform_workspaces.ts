import { queryData } from '@/utils'

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
    path: `runners/terraform-workspace/${workspaceId}/states`,
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
    path: `runners/terraform-workspace/${workspaceId}/states/${stateId}`,
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
    errorMessage: 'Unable to retrieve workspace state.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/states/${stateId}/resources`,
  })
}

export interface ILockWorkspace extends IGetWorkspace {}

export async function lockWorkspace({ workspaceId, orgId }: ILockWorkspace) {
  return queryData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/lock`,
  })
}

export interface IUnlockWorkspace extends IGetWorkspace {}

export async function unlockWorkspace({
  workspaceId,
  orgId,
}: IUnlockWorkspace) {
  return queryData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `runners/terraform-workspace/${workspaceId}/unlock`,
  })
}
