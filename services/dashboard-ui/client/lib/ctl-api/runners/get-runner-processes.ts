import { api } from '@/lib/api'
import type { TRunnerProcess, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getRunnerProcesses = ({
  runnerId,
  orgId,
  type,
  status,
  limit = 20,
  offset,
}: {
  runnerId: string
  orgId: string
  type?: string
  status?: string
} & TPaginationParams) =>
  api<TRunnerProcess[]>({
    path: `runners/${runnerId}/processes${buildQueryParams({ limit, offset, type, status })}`,
    orgId,
    paginated: true,
  })
