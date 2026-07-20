import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { listStaticTokens } from '@/lib'
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
  const query = (searchParams.get('q') ?? '').toLowerCase()

  const { data: result, isLoading } = useQuery({
    queryKey: ['static-tokens', org.id],
    queryFn: () => listStaticTokens({ orgId: org.id }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const all = (result ?? []).filter((token) =>
    query ? (token.name ?? '').toLowerCase().includes(query) : true
  )
  const tokens = all.slice(offset, offset + API_TOKENS_TABLE_LIMIT)
  const hasNext = all.length > offset + API_TOKENS_TABLE_LIMIT

  return (
    <ApiTokensTable
      data={tokens}
      isLoading={isLoading}
      pagination={{ hasNext, offset, limit: API_TOKENS_TABLE_LIMIT }}
    />
  )
}
