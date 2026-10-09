import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useServiceAccountFilters } from '@/hooks/use-service-account-filters'
import { listServiceAccounts } from '@/lib'
import { ServiceAccountFiltersContainer } from '@/components/service-accounts/ServiceAccountFilters'
import type { TAPIError } from '@/types'
import {
  ServiceAccountsTable,
  SERVICE_ACCOUNTS_TABLE_LIMIT,
} from './ServiceAccountsTable'

export const ServiceAccountsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()
  const orgId = org?.id
  const { management, purpose, q, offset } = useServiceAccountFilters()

  const {
    data: result,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['service-accounts', orgId, { management, purpose, q, offset }],
    queryFn: () =>
      listServiceAccounts({
        orgId: orgId!,
        offset,
        limit: SERVICE_ACCOUNTS_TABLE_LIMIT + 1,
        management,
        purpose,
        q,
      }),
    enabled: !!orgId,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const accounts = (result ?? []).slice(0, SERVICE_ACCOUNTS_TABLE_LIMIT)
  const hasNext = (result?.length ?? 0) > SERVICE_ACCOUNTS_TABLE_LIMIT

  return (
    <ServiceAccountsTable
      data={accounts}
      orgId={orgId ?? ''}
      isLoading={isLoading}
      error={
        error
          ? (error as TAPIError).description || 'Refresh the page to try again.'
          : undefined
      }
      management={management}
      hasActiveFilters={!!purpose || !!q}
      filterActions={<ServiceAccountFiltersContainer />}
      pagination={{ hasNext, offset, limit: SERVICE_ACCOUNTS_TABLE_LIMIT }}
    />
  )
}
