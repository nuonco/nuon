import { api } from '@/lib/api'
import type { TAppConfig } from '@/types'

export const getBranchConfigs = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api<TAppConfig[]>({
    path: `apps/${appId}/branches/${branchId}/configs`,
    orgId,
  })
