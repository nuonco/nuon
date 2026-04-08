import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getOrgAccounts } from '@/lib'
import { TeamTable, TEAM_TABLE_LIMIT } from './TeamTable'

export const TeamTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['org-accounts', org.id, offset],
    queryFn: () => getOrgAccounts({ orgId: org.id, offset, limit: TEAM_TABLE_LIMIT + 1 }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const members = (result ?? []).slice(0, TEAM_TABLE_LIMIT)
  const hasNext = (result?.length ?? 0) > TEAM_TABLE_LIMIT

  return (
    <TeamTable
      data={members}
      isLoading={isLoading}
      pagination={{ hasNext, offset, limit: TEAM_TABLE_LIMIT }}
    />
  )
}
