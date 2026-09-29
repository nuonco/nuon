import { api } from '@/lib/api'
import type { TAppBranchConfig } from '@/types'

export const getBranchLatestConfig = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api<TAppBranchConfig>({
    path: `apps/${appId}/branches/${branchId}/latest-config`,
    orgId,
  })
