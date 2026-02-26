import { api } from '@/lib/api'
import type { TInstallAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallActionsLatestRuns = ({
  installId,
  limit,
  offset,
  orgId,
  q,
  trigger_types,
}: {
  installId: string
  orgId: string
  q?: string
  trigger_types?: string
} & TPaginationParams) =>
  api<TInstallAction[]>({
    path: `installs/${installId}/action-workflows/latest-runs${buildQueryParams({ limit, offset, q, trigger_types })}`,
    orgId,
  })
