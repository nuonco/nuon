import { api } from '@/lib/api'
import type { TPaginationParams, TToken } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const listTokens = ({
  orgId,
  accountId,
  q,
  limit,
  offset,
}: {
  orgId: string
  accountId?: string
  q?: string
} & TPaginationParams) =>
  api<TToken[]>({
    path: `tokens${buildQueryParams({
      account_id: accountId || undefined,
      q: q || undefined,
      limit,
      offset,
    })}`,
    orgId,
  })
