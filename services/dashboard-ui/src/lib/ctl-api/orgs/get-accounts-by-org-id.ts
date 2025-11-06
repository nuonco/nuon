import { api } from '@/lib/api'
import type { TAccount, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getAccountsByOrgId = ({
  orgId,
  limit,
  offset,
}: { orgId: string } & TPaginationParams) =>
  api<TAccount[]>({
    path: `orgs/current/accounts${buildQueryParams({ limit, offset })}`,
    orgId,
  })
