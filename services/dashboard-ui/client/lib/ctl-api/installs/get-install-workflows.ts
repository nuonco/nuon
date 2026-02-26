import { api } from '@/lib/api'
import type { TWorkflow, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallWorkflows = ({
  installId,
  limit,
  offset,
  orgId,
  planonly = true,
  type = '',
}: {
  installId: string
  orgId: string
  planonly?: boolean
  type?: string
} & TPaginationParams) =>
  api<TWorkflow[]>({
    path: `installs/${installId}/workflows${buildQueryParams({ limit, offset, planonly, type })}`,
    orgId,
  })
