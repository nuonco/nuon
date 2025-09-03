import { queryData } from '@/utils'
import type { IGetWorkspace } from '../shared-interfaces'

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