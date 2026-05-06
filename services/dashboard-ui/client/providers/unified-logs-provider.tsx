import { createContext, useEffect, useState, useRef, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useLogServerFilters } from '@/hooks/use-log-server-filters'
import { useLogStream } from '@/hooks/use-log-stream'
import { useOrg } from '@/hooks/use-org'
import { getLogStreamLogsWithMeta } from '@/lib'
import type { TOTELLog, TAPIError } from '@/types'

const useUnifiedLogData = ({
  initLogs,
}: {
  initLogs?: TOTELLog[]
}) => {
  const { org } = useOrg()
  const { logStream } = useLogStream()
  const serverFilters = useLogServerFilters()
  const [logs, setLogs] = useState<TOTELLog[]>(initLogs || [])
  const [offset, setOffset] = useState<string>()
  const [hasMore, setHasMore] = useState(true)
  const [staticEnabled, setStaticEnabled] = useState(false)
  const [staticTrigger, setStaticTrigger] = useState(0)
  const [needsPaginationCheck, setNeedsPaginationCheck] = useState(false)
  const [needsFinalFetch, setNeedsFinalFetch] = useState(false)

  const [connectionState, setConnectionState] = useState<'disconnected' | 'connecting' | 'connected' | 'reconnecting'>('disconnected')
  const [error, setError] = useState<TAPIError | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [isCatchingUp, setIsCatchingUp] = useState(false)
  const isCatchingUpRef = useRef(false)
  const catchUpBufferRef = useRef<TOTELLog[]>([])

  const isStreamOpen = logStream?.open || false

  const connectSSE = () => {
    if (!logStream?.id || eventSourceRef.current) return

    setConnectionState('connecting')
    setError(null)

    const url = `/api/orgs/${org.id}/log-streams/${logStream.id}/logs/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      try {
        const newLogs: TOTELLog[] = JSON.parse(event.data)
        if (newLogs.length > 0) {
          if (catchUpBufferRef.current !== null && isCatchingUpRef.current) {
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
      } catch (err) {
        setError({
          error: 'Failed to parse log data',
          description: 'The log data received from the server could not be parsed as valid JSON',
          user_error: false
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
      }
    })

    eventSource.addEventListener('error', (event: MessageEvent) => {
      try {
        const errorData = JSON.parse(event.data)
        setError({
          error: errorData.error || 'Server error occurred',
          description: errorData.description || 'An error was received from the log streaming server',
          user_error: errorData.user_error || false,
          meta: errorData.meta
        })
      } catch (parseErr) {
        setError({
          error: 'Server error occurred',
          description: 'Failed to parse error message from the log streaming server',
          user_error: false
        })
      }
    })

    eventSource.onerror = () => {
      eventSource.close()
      eventSourceRef.current = null

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

  const loadMore = () => {
    if (!isStreamOpen) {
      if (!staticEnabled) {
        setStaticEnabled(true)
      }
      setStaticTrigger(prev => prev + 1)
    }
  }

  const { data: staticResult, isLoading: staticIsLoading } = useQuery({
    queryKey: ['log-stream-logs-static', logStream?.id, org.id, staticTrigger, serverFilters],
    queryFn: () => getLogStreamLogsWithMeta({ logStreamId: logStream!.id, orgId: org.id, offset, order: 'desc', filters: serverFilters }),
    enabled: staticEnabled && !isStreamOpen && !!logStream?.id,
  })

  const { data: paginationCheckResult } = useQuery({
    queryKey: ['log-stream-logs-pagination-check', logStream?.id, org.id, needsPaginationCheck, serverFilters],
    queryFn: () => getLogStreamLogsWithMeta({
      logStreamId: logStream!.id,
      orgId: org.id,
      offset: logs.length > 0 ? String(new Date(logs[logs.length - 1]?.timestamp).getTime() * 1000000) : undefined,
      order: 'desc',
      filters: serverFilters,
    }),
    enabled: needsPaginationCheck && !isStreamOpen && !!logStream?.id,
  })

  const { data: finalFetchResult } = useQuery({
    queryKey: ['log-stream-logs-final', logStream?.id, org.id, needsFinalFetch, serverFilters],
    queryFn: () => getLogStreamLogsWithMeta({
      logStreamId: logStream!.id,
      orgId: org.id,
      offset: logs.length > 0 ? String(new Date(logs[logs.length - 1]?.timestamp).getTime() * 1000000) : undefined,
      filters: serverFilters,
    }),
    enabled: needsFinalFetch && !isStreamOpen && !!logStream?.id,
  })

  useEffect(() => {
    if (!isStreamOpen && staticResult?.data) {
      setLogs((prev) => {
        const logMap = new Map(prev.map((log) => [log.id, log]))
        staticResult.data.forEach((log) => logMap.set(log.id, log))
        return Array.from(logMap.values())
      })
      setOffset(staticResult.nextOffset ?? undefined)
      setHasMore(!!staticResult.nextOffset)
    }
  }, [staticResult, isStreamOpen])

  useEffect(() => {
    if (!isStreamOpen && finalFetchResult?.data && needsFinalFetch) {
      if (finalFetchResult.data.length > 0) {
        setLogs((prev) => {
          const logMap = new Map(prev.map((log) => [log.id, log]))
          finalFetchResult.data.forEach((log) => logMap.set(log.id, log))
          return Array.from(logMap.values())
        })
      }
      const hasMoreLogs = !!finalFetchResult.nextOffset
      setHasMore(hasMoreLogs)
      if (finalFetchResult.nextOffset) {
        setOffset(finalFetchResult.nextOffset)
      }
      setNeedsFinalFetch(false)
    }
  }, [finalFetchResult, needsFinalFetch, isStreamOpen])

  useEffect(() => {
    if (paginationCheckResult && needsPaginationCheck) {
      const hasMoreLogs = !!paginationCheckResult.nextOffset && paginationCheckResult.data.length > 0
      setHasMore(hasMoreLogs)
      if (hasMoreLogs && paginationCheckResult.nextOffset) {
        setOffset(paginationCheckResult.nextOffset)
      }
      setNeedsPaginationCheck(false)
    }
  }, [paginationCheckResult, needsPaginationCheck])

  const hasConnectedSSE = useRef(false)
  const prevIsStreamOpen = useRef(isStreamOpen)

  useEffect(() => {
    if (isStreamOpen) {
      if (!hasConnectedSSE.current) {
        connectSSE()
        hasConnectedSSE.current = true
      }
      setError(null)
    } else {
      if (hasConnectedSSE.current && prevIsStreamOpen.current && !staticEnabled && logs.length > 0) {
        setNeedsPaginationCheck(true)
      }

      if (!hasConnectedSSE.current && !staticEnabled) {
        setStaticEnabled(true)
        setStaticTrigger(1)
      }
    }

    prevIsStreamOpen.current = isStreamOpen
  }, [logStream?.id, isStreamOpen, org.id, staticEnabled])

  useEffect(() => {
    return () => {
      disconnect()
      hasConnectedSSE.current = false
    }
  }, [logStream?.id, org.id])

  useEffect(() => {
    if (isStreamOpen) {
      setHasMore(false)
    }
  }, [isStreamOpen])

  const isLoading = isStreamOpen
    ? isCatchingUp || connectionState === 'connecting' || connectionState === 'reconnecting'
    : staticIsLoading || false

  const currentError = isStreamOpen ? error : null

  return {
    logs,
    isLoading,
    error: currentError,
    connectionState,
    loadMore,
    hasMore: isStreamOpen ? false : hasMore,
    isStreamOpen,
  }
}

type UnifiedLogsContextValue = {
  logs: TOTELLog[]
  isLoading: boolean
  error: TAPIError | null
  connectionState: 'disconnected' | 'connecting' | 'connected' | 'reconnecting'
  loadMore: () => void
  hasMore: boolean
  isStreamOpen: boolean
}

export const UnifiedLogsContext = createContext<UnifiedLogsContextValue | undefined>(undefined)

export function UnifiedLogsProvider({
  children,
  initLogs,
}: {
  children: ReactNode
  initLogs?: TOTELLog[]
}) {
  const logData = useUnifiedLogData({ initLogs })

  return (
    <UnifiedLogsContext.Provider value={logData}>
      {children}
    </UnifiedLogsContext.Provider>
  )
}

export const useUnifiedLogs = () => {
  const context = UnifiedLogsContext
  if (context === undefined) {
    throw new Error('useUnifiedLogs must be used within a UnifiedLogsProvider')
  }
  return context
}
