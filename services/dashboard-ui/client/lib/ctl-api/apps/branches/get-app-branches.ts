import { api } from '@/lib/api'
import type { TAppBranch, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getAppBranches = ({
  appId,
  orgId,
  limit,
  offset,
}: {
  appId: string
  orgId: string
} & TPaginationParams) =>
  api<TAppBranch[]>({
    path: `apps/${appId}/branches${buildQueryParams({ limit, offset })}`,
    orgId,
    paginated: true,
  })
