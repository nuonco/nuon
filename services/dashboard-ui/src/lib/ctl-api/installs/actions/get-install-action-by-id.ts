import { api } from '@/lib/api'
import type { TInstallAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallActionById = ({
  installId,
  actionId,
  orgId,
  limit,
  offset,
}: {
  installId: string
  actionId: string
  orgId: string
} & TPaginationParams) =>
  api<TInstallAction>({
    path: `installs/${installId}/action-workflows/${actionId}/recent-runs${buildQueryParams({ limit, offset })}`,
    orgId,
  })
