import { api } from '@/lib/api'
import type { TInstallWorkflow } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getBranchWorkflowRuns = ({
  appId,
  branchId,
  orgId,
  limit,
  offset,
  planonly,
}: {
  appId: string
  branchId: string
  orgId: string
  limit?: number
  offset?: number
  planonly?: boolean
}) =>
  api<TInstallWorkflow[]>({
    path: `apps/${appId}/branches/${branchId}/runs${buildQueryParams({ limit, offset, planonly })}`,
    orgId,
    paginated: true,
  })