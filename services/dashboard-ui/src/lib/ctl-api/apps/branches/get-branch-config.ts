import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

export const getBranchConfig = ({
  appId,
  branchId,
  configId,
  orgId,
}: {
  appId: string
  branchId: string
  configId: string
  orgId: string
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/configs/${configId}`,
    orgId,
  })
