import type { InfiniteData } from '@tanstack/react-query'
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useParams } from 'react-router'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import type { TTriggerEvent } from '@/types'
import {
  getTriggerEvent,
  getTriggerEventDispatches,
  getTriggerEventRaw,
  replayTriggerEvent,
  retryTriggerEventDispatch,
} from '@/lib'
import { EventDetails } from './EventDetails'
import {
  isTriggerEventTerminal,
  markTriggerEventDispatchPending,
  shouldPollTriggerEventDispatches,
} from '../events'
import type { TTriggerEventDispatchPage } from '@/types'

export const EventDetailsContainer = ({
  expectedTriggerId,
}: {
  expectedTriggerId?: string
}) => {
  const { eventId } = useParams()
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const {
    data: event,
    isFetching,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['trigger-event', org?.id, eventId],
    queryFn: () => getTriggerEvent({ eventId: eventId!, orgId: org!.id }),
    enabled: !!org?.id && !!eventId,
    refetchInterval: (query) =>
      isTriggerEventTerminal({
        ...query.state.data,
        dispatches: [],
        dispatches_truncated: false,
      })
        ? false
        : 5000,
  })
  const triggerMatches =
    !expectedTriggerId || event?.trigger_id === expectedTriggerId
  const dispatchesQuery = useInfiniteQuery({
    queryKey: ['trigger-event-dispatches', org?.id, eventId],
    queryFn: ({ pageParam }) =>
      getTriggerEventDispatches({
        cursor: pageParam,
        eventId: eventId!,
        orgId: org!.id,
      }),
    initialPageParam: '',
    getNextPageParam: (lastPage) => lastPage?.next_cursor || undefined,
    enabled:
      !!org?.id &&
      !!eventId &&
      (!expectedTriggerId || (!!event && triggerMatches)),
    refetchInterval: (query) =>
      shouldPollTriggerEventDispatches({
        event,
        pages: query.state.data?.pages,
      })
        ? 5000
        : false,
  })
  const dispatches = dispatchesQuery.data?.pages.flatMap(
    (page) => page?.items ?? []
  )
  const eventWithDispatches =
    event && triggerMatches
      ? {
          ...event,
          dispatches: dispatches ?? [],
          dispatches_truncated: !!dispatchesQuery.hasNextPage,
        }
      : undefined
  const {
    data: rawRequest,
    error: rawError,
    isFetching: isRawLoading,
    refetch: revealRaw,
  } = useQuery({
    queryKey: ['trigger-event-raw', org?.id, eventId],
    queryFn: () => getTriggerEventRaw({ eventId: eventId!, orgId: org!.id }),
    enabled: false,
  })
  const { mutate: replay, isPending: isReplaying } = useMutation({
    mutationFn: () =>
      replayTriggerEvent({ eventId: eventId!, orgId: org!.id }),
    onSuccess: () => {
      queryClient.setQueryData<TTriggerEvent | undefined>(
        ['trigger-event', org?.id, eventId],
        (current) =>
          current ? { ...current, routing_status: 'routing' } : current
      )
      queryClient.invalidateQueries({
        queryKey: ['trigger-event', org?.id, eventId],
      })
      queryClient.resetQueries({
        queryKey: ['trigger-event-dispatches', org?.id, eventId],
        exact: true,
      })
      queryClient.invalidateQueries({
        queryKey: ['trigger-events', org?.id],
      })
      addToast(
        <Toast heading="Event replay queued" theme="success">
          <Text>
            The event will be evaluated against the current trigger rules.
          </Text>
        </Toast>
      )
    },
    onError: (err) =>
      addToast(
        <Toast heading="Event replay failed" theme="error">
          <Text>{err?.error || 'Unable to replay the event.'}</Text>
        </Toast>
      ),
  })
  const {
    mutate: retryDispatch,
    variables: retryDispatchId,
    isPending: isRetryingDispatch,
  } = useMutation({
    mutationFn: (dispatchId: string) =>
      retryTriggerEventDispatch({ dispatchId, orgId: org!.id }),
    onMutate: (dispatchId) => {
      const dispatchesQueryKey = ['trigger-event-dispatches', org?.id, eventId]
      queryClient.setQueryData<InfiniteData<TTriggerEventDispatchPage>>(
        dispatchesQueryKey,
        (data) =>
          data
            ? {
                ...data,
                pages: markTriggerEventDispatchPending(data.pages, dispatchId),
              }
            : data
      )
    },
    onSuccess: () => {
      const dispatchesQueryKey = ['trigger-event-dispatches', org?.id, eventId]
      queryClient.invalidateQueries({
        queryKey: dispatchesQueryKey,
        exact: true,
      })
      addToast(
        <Toast heading="Dispatch retry queued" theme="success">
          <Text>The dispatch will run again.</Text>
        </Toast>
      )
    },
    onError: (err) => {
      queryClient.invalidateQueries({
        queryKey: ['trigger-event-dispatches', org?.id, eventId],
        exact: true,
      })
      addToast(
        <Toast heading="Dispatch retry failed" theme="error">
          <Text>{err?.error || 'Unable to retry the dispatch.'}</Text>
        </Toast>
      )
    },
  })

  if (isLoading)
    return (
      <div className="flex justify-center py-12">
        <Loading variant="large" />
      </div>
    )
  return (
    <EventDetails
      event={eventWithDispatches}
      orgId={org?.id ?? ''}
      hasDispatchError={!!dispatchesQuery.error}
      hasError={!!error}
      hasRawError={!!rawError}
      rawRequest={rawRequest}
      isRawLoading={isRawLoading}
      isRetrying={isFetching && !!error}
      isReplaying={isReplaying}
      onRevealRaw={() => void revealRaw()}
      onRetry={() => void refetch()}
      onReplay={() => replay()}
      onLoadMoreDispatches={
        dispatchesQuery.hasNextPage
          ? () => void dispatchesQuery.fetchNextPage()
          : undefined
      }
      onRetryDispatch={(dispatchId) => retryDispatch(dispatchId)}
      retryingDispatchId={isRetryingDispatch ? retryDispatchId : undefined}
      isLoadingMoreDispatches={dispatchesQuery.isFetchingNextPage}
    />
  )
}
