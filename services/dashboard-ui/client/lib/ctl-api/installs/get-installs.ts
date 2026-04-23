import { api } from '@/lib/api'
import type { TInstall, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstalls = ({
  limit,
  offset,
  orgId,
  q,
  labels,
}: { orgId: string; q?: string; labels?: string } & TPaginationParams) =>
  api<TInstall[]>({
    path: `installs${buildQueryParams({ limit, offset, q, labels })}`,
    orgId,
    paginated: true,
  })
