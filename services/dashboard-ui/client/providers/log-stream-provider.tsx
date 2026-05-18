import { createContext, useEffect, useState, useRef, type ReactNode } from 'react'
import { useOrg } from '@/hooks/use-org'
import { LogsPageSkeleton } from '@/components/log-stream/SSELogs'
import type { TOTELLog, TAPIError } from '@/types'

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

export type LogStreamContextValue = {
  logs: TOTELLog[]
  logStreamId: string
  isLoading: boolean
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
  const [reconnectAttempt, setReconnectAttempt] = useState(0)

  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCatchingUpRef = useRef(false)
  const catchUpBufferRef = useRef<TOTELLog[]>([])
  const isCompleteRef = useRef(false)

  const connectSSE = () => {
    if (!logStreamId || !org?.id || eventSourceRef.current) return

    setConnectionState('connecting')
    setError(null)

    const url = `/api/orgs/${org.id}/log-streams/${logStreamId}/logs/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      try {
        const newLogs: TOTELLog[] = JSON.parse(event.data)
        if (newLogs.length > 0) {
          if (isCatchingUpRef.current) {
            catchUpBufferRef.current.push(...newLogs)
          } else {
            setLogs(prev => {
              const logMap = new Map(prev.map(log => [log.id, log]))
              newLogs.forEach(log => logMap.set(log.id, log))
              return Array.from(logMap.values())
            })
          }
        }
        setConnectionState('connected')
        setReconnectAttempt(0)
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
        isCatchingUpRef.current = true
        setIsCatchingUp(true)
        catchUpBufferRef.current = []
      } else if (event.data === 'live') {
        const buffered = catchUpBufferRef.current
        catchUpBufferRef.current = []
        isCatchingUpRef.current = false
        setIsCatchingUp(false)
        if (buffered.length > 0) {
          setLogs(prev => {
            const logMap = new Map(prev.map(log => [log.id, log]))
            buffered.forEach(log => logMap.set(log.id, log))
            return Array.from(logMap.values())
          })
        }
      } else if (event.data === 'complete') {
        isCompleteRef.current = true
        eventSource.close()
        eventSourceRef.current = null
        setConnectionState('disconnected')
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

      setConnectionState('reconnecting')
      const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttempt), 30000)
      setReconnectAttempt(prev => prev + 1)

      reconnectTimeoutRef.current = setTimeout(() => {
        connectSSE()
      }, backoffDelay)
    }

    eventSource.onopen = () => {
      setConnectionState('connected')
      setError(null)
      setReconnectAttempt(0)
    }
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
    setConnectionState('disconnected')
  }

  useEffect(() => {
    isCompleteRef.current = false
    if (logStreamId && org?.id) {
      connectSSE()
    }
    return () => {
      disconnect()
    }
  }, [logStreamId, org?.id])

  if (!logStreamId) return <LogsPageSkeleton />

  const isLoading = isCatchingUp || connectionState === 'connecting' || connectionState === 'reconnecting'

  return (
    <LogStreamContext.Provider
      value={{
        logs,
        logStreamId,
        isLoading,
        error,
        connectionState,
      }}
    >
      {children}
    </LogStreamContext.Provider>
  )
}
