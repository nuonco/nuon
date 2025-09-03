import { queryData } from '@/utils'
import type { IGetWorkspace } from '../shared-interfaces'

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