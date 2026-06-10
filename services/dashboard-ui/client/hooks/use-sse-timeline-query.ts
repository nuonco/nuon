import { useMemo } from 'react'
import { useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { createSSEQueryListener, type TSSEListenerMap } from '@/lib/sse-listeners'

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
}: IUseSSETimelineQuery<TData>) {
  const queryClient = useQueryClient()

  const listeners = useMemo(
    () => ({
      [eventName]: createSSEQueryListener(queryClient, queryKey, { transform }),
      ...extraListeners,
    }),
    [queryClient, eventName, transform, extraListeners, ...queryKey]
  )

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn,
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled,
  })

  return { data, isLoading, error, sseConnected }
}
