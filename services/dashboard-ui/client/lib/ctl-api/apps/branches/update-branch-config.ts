import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

export type TUpdateBranchConfigRequest = {
  disable_branch_triggers?: boolean
}

export const updateBranchConfig = ({
  appId,
  branchId,
  configId,
  orgId,
  request,
}: {
  appId: string
  branchId: string
  configId: string
  orgId: string
  request: TUpdateBranchConfigRequest
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/configs/${configId}`,
    orgId,
    method: 'PATCH',
    body: request,
  })
