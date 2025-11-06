import { api } from '@/lib/api'
import type { TRunnerHealthCheck, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getRunnerRecentHealthChecks = ({
  runnerId,
  limit,
  offset,
  orgId,
  window,
}: { runnerId: string; orgId: string; window?: string } & TPaginationParams) =>
  api<TRunnerHealthCheck[]>({
    path: `runners/${runnerId}/recent-health-checks${buildQueryParams({ limit, offset, window })}`,
    orgId,
  })
