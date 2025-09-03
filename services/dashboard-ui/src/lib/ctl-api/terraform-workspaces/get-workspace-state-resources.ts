import { queryData } from '@/utils'
import type { IGetWorkspace } from '../shared-interfaces'

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