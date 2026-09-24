import { useMemo } from 'react'
import { useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'
import { useResourceSSE } from '@/lib/sse/use-resource-sse'
import { createSSEQueryListener, type TSSEListenerMap } from '@/lib/sse-listeners'

const SUSPENDED_POLL_MS = 30_000

interface IUseSSETimelineQuery<TData> {
  sseUrl: string | undefined
  queryKey: QueryKey
  queryFn: () => Promise<TData>
  enabled: boolean
  shouldPoll: boolean
  pollInterval: number
  eventName: string
  transform?: (eventData: any) => unknown
  extraListeners?: TSSEListenerMap
  onError?: (message: string) => void
}

export function useSSETimelineQuery<TData>({
  sseUrl,
  queryKey,
  queryFn,
  enabled,
  shouldPoll,
  pollInterval,
  eventName,
  transform,
  extraListeners,
  onError,
}: IUseSSETimelineQuery<TData>) {
  const queryClient = useQueryClient()

  const listeners = useMemo(
    () => ({
      [eventName]: createSSEQueryListener(queryClient, queryKey, { transform }),
      ...extraListeners,
    }),
    [queryClient, eventName, transform, extraListeners, ...queryKey]
  )

  const { connected: sseConnected, suspended: sseSuspended } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
    onError,
  })

  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn,
    refetchOnMount: 'always',
    refetchInterval:
      !shouldPoll || sseConnected
        ? false
        : sseSuspended
          ? SUSPENDED_POLL_MS
          : pollInterval,
    enabled,
  })

  return { data, isLoading, error, sseConnected }
}
