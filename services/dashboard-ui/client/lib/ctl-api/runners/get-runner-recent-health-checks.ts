import { api } from '@/lib/api'
import type { TRunnerHealthCheck, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getRunnerRecentHealthChecks = ({
  runnerId,
  limit,
  offset,
  orgId,
  window,
  process,
}: { runnerId: string; orgId: string; window?: string; process?: string } & TPaginationParams) =>
  api<TRunnerHealthCheck[]>({
    path: `runners/${runnerId}/recent-health-checks${buildQueryParams({ limit, offset, window, process })}`,
    orgId,
  })
