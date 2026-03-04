import { api } from '@/lib/api'
import type { TOrg, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getOrgs = ({
  limit,
  offset,
  q,
}: { q?: string } & TPaginationParams = {}) =>
  api<TOrg[]>({
    path: `orgs${buildQueryParams({ limit, offset, q })}`,
  })
