import { mutateData } from '@/utils'
import type { IGetWorkspace } from '../shared-interfaces'

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