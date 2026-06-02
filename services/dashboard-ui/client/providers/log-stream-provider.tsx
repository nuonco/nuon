import { createContext, useEffect, useState, useRef, type ReactNode } from 'react'
import { useOrg } from '@/hooks/use-org'
import { LogsPageSkeleton } from '@/components/log-stream/SSELogs'
import type { TOTELLog, TAPIError } from '@/types'

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

export type LogStreamContextValue = {
  logs: TOTELLog[]
  logStreamId: string
  isLoading: boolean
  isCatchingUp: boolean
  error: TAPIError | null
  connectionState: ConnectionState
}

export const LogStreamContext = createContext<LogStreamContextValue | undefined>(undefined)

export function LogStreamProvider({
  children,
  logStreamId,
}: {
  children: ReactNode
  logStreamId?: string
}) {
  const { org } = useOrg()

  const [logs, setLogs] = useState<TOTELLog[]>([])
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected')
  const [error, setError] = useState<TAPIError | null>(null)
  const [isCatchingUp, setIsCatchingUp] = useState(false)

  const seenIdsRef = useRef(new Set<string>())
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCompleteRef = useRef(false)
  const reconnectAttemptRef = useRef(0)
  const connStateRef = useRef<ConnectionState>('disconnected')

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
    seenIdsRef.current = new Set()
    setLogs([])
    setError(null)
    setIsCatchingUp(false)

    const connectSSE = () => {
      if (eventSourceRef.current) return

      setConnState('connecting')
      setError(null)

      const url = `/api/orgs/${org.id}/log-streams/${logStreamId}/logs/sse`
      const eventSource = new EventSource(url)
      eventSourceRef.current = eventSource

      eventSource.onmessage = (event) => {
        try {
          const newLogs: TOTELLog[] = JSON.parse(event.data)
          const unique = newLogs.filter(log => {
            if (seenIdsRef.current.has(log.id)) return false
            seenIdsRef.current.add(log.id)
            return true
          })
          if (unique.length > 0) {
            setLogs(prev => [...prev, ...unique])
          }
          setConnState('connected')
          reconnectAttemptRef.current = 0
        } catch {
          setError({
            error: 'Failed to parse log data',
            description: 'The log data received from the server could not be parsed as valid JSON',
            user_error: false,
          })
        }
      }

      eventSource.addEventListener('status', (event: MessageEvent) => {
        if (event.data === 'catching-up') {
          setIsCatchingUp(true)
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
            description: errorData.description || 'An error was received from the log streaming server',
            user_error: errorData.user_error || false,
            meta: errorData.meta,
          })
        } catch {
          setError({
            error: 'Server error occurred',
            description: 'Failed to parse error message from the log streaming server',
            user_error: false,
          })
        }
      })

      eventSource.onerror = () => {
        eventSource.close()
        eventSourceRef.current = null

        if (isCompleteRef.current) return

        setConnState('reconnecting')
        const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttemptRef.current), 30000)
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
  }, [logStreamId, org?.id])

  if (!logStreamId) return <LogsPageSkeleton />

  const isLoading =
    (logs.length === 0 && connectionState === 'connecting') ||
    connectionState === 'reconnecting'

  return (
    <LogStreamContext.Provider
      value={{
        logs,
        logStreamId,
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
