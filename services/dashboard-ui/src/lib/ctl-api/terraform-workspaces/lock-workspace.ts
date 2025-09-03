import { mutateData } from '@/utils'
import type { IGetWorkspace } from '../shared-interfaces'

export interface ILockWorkspace extends IGetWorkspace {}

export async function lockWorkspace({ workspaceId, orgId }: ILockWorkspace) {
  return mutateData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `terraform-workspaces/${workspaceId}/lock`,
  })
}