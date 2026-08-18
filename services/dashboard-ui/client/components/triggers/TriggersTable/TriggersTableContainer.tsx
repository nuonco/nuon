import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { getTriggers } from '@/lib'
import { filterTriggers } from '@/utils/trigger-utils'
import {
  AUTH_TYPE_PARAM,
  ENVELOPE_PARAM,
  TriggerFiltersContainer,
  SOURCE_PARAM,
} from '../TriggerFilters/TriggerFiltersContainer'
import { TriggersTable } from './TriggersTable'

export const TriggersTableContainer = () => {
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const { data, error, isLoading, refetch } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['triggers', org?.id],
    queryFn: () => getTriggers({ orgId: org!.id }),
    enabled: !!org?.id,
  })
  const filtered = filterTriggers(data ?? [], {
    source: searchParams.get(SOURCE_PARAM) || undefined,
    authType: searchParams.get(AUTH_TYPE_PARAM) || undefined,
    envelope: searchParams.get(ENVELOPE_PARAM) || undefined,
  })
  return (
    <TriggersTable
      data={filtered}
      error={!!error}
      filterActions={<TriggerFiltersContainer />}
      isLoading={isLoading}
      onRetry={() => void refetch()}
      orgId={org?.id ?? ''}
    />
  )
}
