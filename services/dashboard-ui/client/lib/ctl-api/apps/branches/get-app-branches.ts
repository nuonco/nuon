import { api } from '@/lib/api'
import type { TAppBranch, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getAppBranches = ({
  appId,
  orgId,
  limit,
  offset,
  q,
}: {
  appId: string
  orgId: string
  q?: string
} & TPaginationParams) =>
  api<TAppBranch[]>({
    path: `apps/${appId}/branches${buildQueryParams({ limit, offset, q })}`,
    orgId,
    paginated: true,
  })
