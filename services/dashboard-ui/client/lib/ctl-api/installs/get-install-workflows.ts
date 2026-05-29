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
  search = '',
}: {
  installId: string
  finished?: boolean
  orgId: string
  planonly?: boolean
  type?: string
  search?: string
} & TPaginationParams) =>
  api<TWorkflow[]>({
    path: `installs/${installId}/workflows${buildQueryParams({ limit, offset, planonly, type, finished, search })}`,
    orgId,
    paginated: true
  })
