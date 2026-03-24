import { api } from '@/lib/api'
import type { TWorkflow, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getOrgWorkflows = ({
  finished,
  limit,
  offset,
  orgId,
  planonly = true,
  type = '',
}: {
  finished?: boolean
  orgId: string
  planonly?: boolean
  type?: string
} & TPaginationParams) =>
  api<TWorkflow[]>({
    path: `workflows${buildQueryParams({ limit, offset, planonly, type, finished })}`,
    orgId,
    paginated: true,
  })
