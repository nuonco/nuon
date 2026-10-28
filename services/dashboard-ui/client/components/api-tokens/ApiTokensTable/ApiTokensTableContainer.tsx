import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { listTokens } from '@/lib'
import { ApiTokensTable, API_TOKENS_TABLE_LIMIT } from './ApiTokensTable'

export const ApiTokensTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') ?? ''

  const { data: result, isLoading } = useQuery({
    queryKey: ['tokens', org.id, { q, offset }],
    queryFn: () =>
      listTokens({
        orgId: org.id,
        q,
        offset,
        limit: API_TOKENS_TABLE_LIMIT + 1,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const tokens = (result ?? []).slice(0, API_TOKENS_TABLE_LIMIT)
  const hasNext = (result?.length ?? 0) > API_TOKENS_TABLE_LIMIT

  return (
    <ApiTokensTable
      data={tokens}
      isLoading={isLoading}
      pagination={{ hasNext, offset, limit: API_TOKENS_TABLE_LIMIT }}
    />
  )
}
