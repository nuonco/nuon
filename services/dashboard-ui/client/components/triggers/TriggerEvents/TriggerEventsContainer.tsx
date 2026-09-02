import {
  keepPreviousData,
  useInfiniteQuery,
  useQuery,
} from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { getTriggerEventsForTrigger, getTriggerEventTypes } from '@/lib'
import { TriggerEvents } from './TriggerEvents'

export const TriggerEventsContainer = ({
  triggerId,
}: {
  triggerId: string
}) => {
  const { org } = useOrg()
  const [params, setParams] = useSearchParams()
  const filters = {
    q: params.get('q') ?? '',
    eventType: params.get('event_type') ?? '',
    outcome: params.get('outcome') ?? '',
    order: params.get('order') === 'asc' ? 'asc' : 'desc',
  }
  const events = useInfiniteQuery({
    queryKey: [
      'event-trigger-events',
      org?.id,
      triggerId,
      filters.q,
      filters.eventType,
      filters.outcome,
      filters.order,
    ],
    queryFn: ({ pageParam }) =>
      getTriggerEventsForTrigger({
        orgId: org!.id,
        triggerId: triggerId,
        cursor: pageParam,
        query: filters.q.trim(),
        eventType: filters.eventType,
        outcome: filters.outcome,
        order: filters.order as 'asc' | 'desc',
      }),
    initialPageParam: '',
    getNextPageParam: (page) => page?.next_cursor || undefined,
    enabled: !!org?.id && !!triggerId,
    refetchInterval: 5000,
  })
  const facets = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['event-trigger-event-types', org?.id, triggerId],
    queryFn: () =>
      getTriggerEventTypes({ orgId: org!.id, triggerId: triggerId }),
    enabled: !!org?.id && !!triggerId,
  })
  const onFilterChange = (key: string, value: string) => {
    const next = new URLSearchParams(params)
    if (value) next.set(key, value)
    else next.delete(key)
    setParams(next, { replace: true })
  }
  return (
    <TriggerEvents
      data={events.data?.pages.flatMap((page) => page?.items ?? []) ?? []}
      eventTypes={facets.data ?? []}
      filters={filters}
      hasError={!!events.error}
      isLoading={events.isLoading}
      isLoadingMore={events.isFetchingNextPage}
      onFilterChange={onFilterChange}
      onLoadMore={
        events.hasNextPage ? () => void events.fetchNextPage() : undefined
      }
      onRetry={() => void events.refetch()}
      orgId={org?.id ?? ''}
      triggerId={triggerId}
    />
  )
}
