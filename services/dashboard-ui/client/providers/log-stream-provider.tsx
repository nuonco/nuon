import {
  createContext,
  useCallback,
  useEffect,
  useState,
  useRef,
  type ReactNode,
} from 'react'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { LogsPageSkeleton } from '@/components/log-stream/SSELogs'
import { getLogStreamLogs } from '@/lib'
import type { TOTELLog, TAPIError } from '@/types'

type ConnectionState =
  | 'disconnected'
  | 'connecting'
  | 'connected'
  | 'reconnecting'

export type LogStreamContextValue = {
  logs: TOTELLog[]
  logStreamId: string
  runnerJobId?: string
  isLoading: boolean
  isCatchingUp: boolean
  error: TAPIError | null
  connectionState: ConnectionState
}

export const LogStreamContext = createContext<
  LogStreamContextValue | undefined
>(undefined)

export function LogStreamProvider({
  children,
  logStreamId,
  runnerJobId,
  renderWhilePending = false,
}: {
  children: ReactNode
  logStreamId?: string
  runnerJobId?: string
  renderWhilePending?: boolean
}) {
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const isNewestFirst = searchParams.get('sort') !== 'asc'

  const [logs, setLogs] = useState<TOTELLog[]>([])
  const [connectionState, setConnectionState] =
    useState<ConnectionState>('disconnected')
  const [error, setError] = useState<TAPIError | null>(null)
  const [isCatchingUp, setIsCatchingUp] = useState(false)

  const seenIdsRef = useRef(new Set<string>())
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCompleteRef = useRef(false)
  const reconnectAttemptRef = useRef(0)
  const connStateRef = useRef<ConnectionState>('disconnected')
  const seedRequestedRef = useRef(false)
  const isNewestFirstRef = useRef(isNewestFirst)
  const logStreamIdRef = useRef(logStreamId)

  useEffect(() => {
    isNewestFirstRef.current = isNewestFirst
  }, [isNewestFirst])

  const appendLogs = useCallback((incoming: TOTELLog[]) => {
    const unique = incoming.filter((log) => {
      if (seenIdsRef.current.has(log.id)) return false
      seenIdsRef.current.add(log.id)
      return true
    })
    if (unique.length > 0) {
      setLogs((prev) => [...prev, ...unique])
    }
  }, [])

  // The stream itself must stay ASC: the tail cursor only moves forward, so an
  // `order=desc` stream would page backwards into history and never surface new
  // lines on a running job. Instead fetch the newest page once, so newest-first
  // has a correct top immediately and the ASC stream backfills below it.
  const seedNewestPage = useCallback(
    (streamId: string, orgId: string, jobId?: string) => {
      if (seedRequestedRef.current || !isNewestFirstRef.current) return
      seedRequestedRef.current = true

      getLogStreamLogs({ logStreamId: streamId, orgId, order: 'desc' })
        .then((newest) => {
          if (streamId !== logStreamIdRef.current) return
          if (!Array.isArray(newest)) return
          appendLogs(
            jobId
              ? newest.filter((log) => log?.runner_job_id === jobId)
              : newest
          )
        })
        .catch(() => {})
    },
    [appendLogs]
  )

  const setConnState = (state: ConnectionState) => {
    if (connStateRef.current === state) return
    connStateRef.current = state
    setConnectionState(state)
  }

  const disconnect = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    setConnState('disconnected')
  }

  useEffect(() => {
    if (!logStreamId || !org?.id) return

    isCompleteRef.current = false
    reconnectAttemptRef.current = 0
    seedRequestedRef.current = false
    logStreamIdRef.current = logStreamId
    seenIdsRef.current = new Set()
    setLogs([])
    setError(null)
    setIsCatchingUp(false)

    const connectSSE = () => {
      if (eventSourceRef.current) return

      setConnState('connecting')
      setError(null)

      const params = new URLSearchParams()
      if (runnerJobId) params.set('runner_job_id', runnerJobId)
      const query = params.toString()
      const url = `/api/orgs/${org.id}/log-streams/${logStreamId}/logs/sse${query ? `?${query}` : ''}`
      const eventSource = new EventSource(url)
      eventSourceRef.current = eventSource

      eventSource.onmessage = (event) => {
        try {
          const newLogs: TOTELLog[] = JSON.parse(event.data)
          appendLogs(newLogs)
          setConnState('connected')
          reconnectAttemptRef.current = 0
        } catch {
          setError({
            error: 'Failed to parse log data',
            description:
              'The log data received from the server could not be parsed as valid JSON',
            user_error: false,
          })
        }
      }

      eventSource.addEventListener('status', (event: MessageEvent) => {
        if (event.data === 'catching-up') {
          setIsCatchingUp(true)
          seedNewestPage(logStreamId, org.id, runnerJobId)
        } else if (event.data === 'live') {
          setIsCatchingUp(false)
        } else if (event.data === 'complete') {
          isCompleteRef.current = true
          setIsCatchingUp(false)
          eventSource.close()
          eventSourceRef.current = null
          setConnState('disconnected')
        }
      })

      eventSource.addEventListener('error', (event: MessageEvent) => {
        try {
          const errorData = JSON.parse(event.data)
          setError({
            error: errorData.error || 'Server error occurred',
            description:
              errorData.description ||
              'An error was received from the log streaming server',
            user_error: errorData.user_error || false,
            meta: errorData.meta,
          })
        } catch {
          setError({
            error: 'Server error occurred',
            description:
              'Failed to parse error message from the log streaming server',
            user_error: false,
          })
        }
      })

      eventSource.onerror = () => {
        eventSource.close()
        eventSourceRef.current = null

        if (isCompleteRef.current) return

        setConnState('reconnecting')
        const backoffDelay = Math.min(
          1000 * Math.pow(2, reconnectAttemptRef.current),
          30000
        )
        reconnectAttemptRef.current += 1

        reconnectTimeoutRef.current = setTimeout(() => {
          connectSSE()
        }, backoffDelay)
      }

      eventSource.onopen = () => {
        setConnState('connected')
        setError(null)
        reconnectAttemptRef.current = 0
      }
    }

    connectSSE()
    return () => {
      disconnect()
    }
  }, [logStreamId, org?.id, runnerJobId])

  // Switching to newest-first while a long catch-up is still draining needs the
  // same seed; the `catching-up` event has already come and gone by then.
  useEffect(() => {
    if (!logStreamId || !org?.id || !isNewestFirst || !isCatchingUp) return
    seedNewestPage(logStreamId, org.id, runnerJobId)
  }, [
    logStreamId,
    org?.id,
    runnerJobId,
    isNewestFirst,
    isCatchingUp,
    seedNewestPage,
  ])

  if (!logStreamId) {
    if (!renderWhilePending) return <LogsPageSkeleton />
    return (
      <LogStreamContext.Provider
        value={{
          logs: [],
          logStreamId: '',
          runnerJobId,
          isLoading: false,
          isCatchingUp: false,
          error: null,
          connectionState: 'disconnected',
        }}
      >
        {children}
      </LogStreamContext.Provider>
    )
  }

  const isLoading =
    (logs.length === 0 && connectionState === 'connecting') ||
    connectionState === 'reconnecting'

  return (
    <LogStreamContext.Provider
      value={{
        logs,
        logStreamId,
        runnerJobId,
        isLoading,
        isCatchingUp,
        error,
        connectionState,
      }}
    >
      {children}
    </LogStreamContext.Provider>
  )
}
