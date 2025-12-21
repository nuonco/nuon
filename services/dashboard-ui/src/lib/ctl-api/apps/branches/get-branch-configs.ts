import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

export const getBranchConfigs = ({
  appId,
  branchId,
  orgId,
  limit,
  offset,
}: {
  appId: string
  branchId: string
  orgId: string
  limit?: number
  offset?: number
}) =>
  api<TAppBranchConfig[]>({
    path: `apps/${appId}/branches/${branchId}/configs`,
    orgId,
  })
