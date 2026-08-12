import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useRoleTitles } from '@/hooks/use-roles'
import { listServiceAccounts } from '@/lib'
import { ServiceAccountsTable, SERVICE_ACCOUNTS_TABLE_LIMIT } from './ServiceAccountsTable'

export const ServiceAccountsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const roleTitles = useRoleTitles()
  const offset = Number(searchParams.get('offset') ?? 0)
  const includeRunners = searchParams.get('runners') === 'true'

  const { data: result, isLoading } = useQuery({
    queryKey: ['service-accounts', org.id, offset, includeRunners],
    queryFn: () =>
      listServiceAccounts({
        orgId: org.id,
        offset,
        limit: SERVICE_ACCOUNTS_TABLE_LIMIT + 1,
        includeRunners,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const accounts = (result ?? []).slice(0, SERVICE_ACCOUNTS_TABLE_LIMIT)
  const hasNext = (result?.length ?? 0) > SERVICE_ACCOUNTS_TABLE_LIMIT

  return (
    <ServiceAccountsTable
      data={accounts}
      roleTitles={roleTitles}
      isLoading={isLoading}
      pagination={{ hasNext, offset, limit: SERVICE_ACCOUNTS_TABLE_LIMIT }}
    />
  )
}
