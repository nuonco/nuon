import { api } from '@/lib/api'
import type { TAppConfig } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getBranchConfigs = ({
  appId,
  branchId,
  orgId,
  limit,
}: {
  appId: string
  branchId: string
  orgId: string
  limit?: number
}) =>
  api<TAppConfig[]>({
    path: `apps/${appId}/branches/${branchId}/configs${buildQueryParams({
      limit,
    })}`,
    orgId,
  })
