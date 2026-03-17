import { api } from '@/lib/api'
import type { TAppBranch } from '@/types'

export type TUpdateBranchRequest = {
  name?: string
}

export const updateBranch = ({
  appId,
  branchId,
  orgId,
  request,
}: {
  appId: string
  branchId: string
  orgId: string
  request: TUpdateBranchRequest
}) =>
  api<TAppBranch>({
    path: `apps/${appId}/branches/${branchId}`,
    orgId,
    method: 'PATCH',
    body: request,
  })
