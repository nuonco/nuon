import { useEffect, useMemo, useRef } from 'react'
import { useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'
import { useResourceSSE } from '@/lib/sse/use-resource-sse'
import { createSSEQueryListener, type TSSEListenerMap } from '@/lib/sse-listeners'
import type { TAPIError } from '@/types'

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
  onError?: (message: string) => void
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
  onError,
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

  const { connected: sseConnected, suspended: sseSuspended, disconnect } = useResourceSSE({
    url: sseUrl,
    enabled: sseEnabled ?? shouldPoll,
    listeners,
    onError,
  })

  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn,
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      if (isFinished?.(query.state.data)) return finishedPollMs
      return sseSuspended ? finishedPollMs : fallbackPollMs
    },
    enabled,
  })

  const onErrorRef = useRef(onError)
  onErrorRef.current = onError

  useEffect(() => {
    if (error && data) {
      onErrorRef.current?.((error as TAPIError)?.error ?? 'Connection issue')
    }
  }, [error])

  return { data, isLoading, error, refetch, sseConnected, disconnect }
}
