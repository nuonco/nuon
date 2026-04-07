import { api } from '@/lib/api'
import type { TRunnerHealthCheck, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getRunnerRecentHealthChecks = ({
  runnerId,
  limit,
  offset,
  orgId,
  processId,
  window,
}: { runnerId: string; orgId: string; processId?: string; window?: string } & TPaginationParams) =>
  api<TRunnerHealthCheck[]>({
    path: `runners/${runnerId}/recent-health-checks${buildQueryParams({ limit, offset, process_id: processId, window })}`,
    orgId,
  })
