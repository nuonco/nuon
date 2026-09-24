import { api } from '@/lib/api'
import type { TWorkflow, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallWorkflows = ({
  installId,
  finished,
  limit,
  offset,
  orgId,
  planonly = true,
  type = '',
  status = '',
  search = '',
  created_at_gte,
  created_at_lte,
}: {
  installId: string
  finished?: boolean
  orgId: string
  planonly?: boolean
  type?: string
  status?: string
  search?: string
  created_at_gte?: string
  created_at_lte?: string
} & TPaginationParams) =>
  api<TWorkflow[]>({
    path: `installs/${installId}/workflows${buildQueryParams({
      limit,
      offset,
      planonly,
          type,
      status,
      finished,
      search,
      created_at_gte,
      created_at_lte,
    })}`,
    orgId,
    paginated: true,
  })
