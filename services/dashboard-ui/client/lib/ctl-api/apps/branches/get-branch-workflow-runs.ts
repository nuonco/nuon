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
  preview,
  q,
  status,
  type,
  created_at_gte,
  created_at_lte,
}: {
  appId: string
  branchId: string
  orgId: string
  limit?: number
  offset?: number
  planonly?: boolean
  preview?: boolean
  q?: string
  status?: string
  type?: string
  created_at_gte?: string
  created_at_lte?: string
}) =>
  api<TInstallWorkflow[]>({
    path: `apps/${appId}/branches/${branchId}/runs${buildQueryParams({
      limit,
      offset,
      planonly,
      preview,
      q,
      status,
      type,
      created_at_gte,
      created_at_lte,
    })}`,
    orgId,
    paginated: true,
  })
