import { useMemo } from 'react'
import { useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useRefreshErrorToast } from '@/hooks/use-refresh-error-toast'
import { createSSEQueryListener, type TSSEListenerMap } from '@/lib/sse-listeners'

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export const isTerminalStatusV2 = (
  data: { status_v2?: { status?: string } } | undefined
): boolean =>
  ['success', 'error', 'cancelled', 'not-attempted'].includes(
    data?.status_v2?.status ?? ''
  )

interface IUseSSEResourceQuery<TData> {
  sseUrl: string | undefined
  queryKey: QueryKey
  queryFn: () => Promise<TData>
  enabled: boolean
  shouldPoll: boolean
  sseEnabled?: boolean
  eventName: string
  onPrimaryEvent?: (data: TData) => void
  extraListeners?: TSSEListenerMap
  isFinished?: (data: TData | undefined) => boolean
  fallbackPollMs?: number
  finishedPollMs?: number
}

export function useSSEResourceQuery<TData>({
  sseUrl,
  queryKey,
  queryFn,
  enabled,
  shouldPoll,
  sseEnabled,
  eventName,
  onPrimaryEvent,
  extraListeners,
  isFinished,
  fallbackPollMs = FALLBACK_POLL_MS,
  finishedPollMs = FINISHED_POLL_MS,
}: IUseSSEResourceQuery<TData>) {
  const queryClient = useQueryClient()

  const listeners = useMemo(
    () => ({
      [eventName]: createSSEQueryListener<TData>(queryClient, queryKey, {
        onData: onPrimaryEvent,
      }),
      ...extraListeners,
    }),
    [queryClient, eventName, onPrimaryEvent, extraListeners, ...queryKey]
  )

  const { connected: sseConnected, disconnect } = useResourceSSE({
    url: sseUrl,
    enabled: sseEnabled ?? shouldPoll,
    listeners,
  })

  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn,
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      return isFinished?.(query.state.data) ? finishedPollMs : fallbackPollMs
    },
    enabled,
  })

  useRefreshErrorToast(error, data)

  return { data, isLoading, error, refetch, sseConnected, disconnect }
}
