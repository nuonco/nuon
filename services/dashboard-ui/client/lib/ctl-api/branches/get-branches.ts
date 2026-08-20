import { api } from '@/lib/api'
import type { TAppBranch, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getBranches = ({
  orgId,
  limit,
  offset,
}: {
  orgId: string
} & TPaginationParams) =>
  api<TAppBranch[]>({
    path: `branches${buildQueryParams({ limit, offset })}`,
    orgId,
    paginated: true,
  })
