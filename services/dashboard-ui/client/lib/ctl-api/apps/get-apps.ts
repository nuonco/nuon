import { api } from '@/lib/api'
import type { TApp, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getApps = ({
  limit,
  offset,
  orgId,
  q,
}: { orgId: string; q?: string } & TPaginationParams) =>
  api<TApp[]>({
    path: `apps${buildQueryParams({ limit, offset, q })}`,
    orgId,
  })
