'use client'

import { createContext, useEffect, useState, useRef, type ReactNode } from 'react'
import { useLogStream } from '@/hooks/use-log-stream'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import type { TOTELLog, TAPIError } from '@/types'

const useUnifiedLogData = ({ 
  initLogs 
}: { 
  initLogs: TOTELLog[] | null 
}) => {
  const { org } = useOrg()
  const { logStream } = useLogStream()
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

  const isStreamOpen = logStream?.open || false
  const params = useQueryParams({ order: isStreamOpen ? 'asc' : 'desc' })

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
          setLogs(prev => {
            const logMap = new Map(prev.map(log => [log.id, log]))
            newLogs.forEach(log => logMap.set(log.id, log))
            return Array.from(logMap.values())
          })
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

  const pollingResults = usePolling<TOTELLog[]>({
    path: `/api/orgs/${org.id}/log-streams/${logStream?.id}/logs`,
    dependencies: [offset],
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
    initData: initLogs,
    pollInterval: 2000,
    shouldPoll: false,
  })

  const staticResults = useQuery<TOTELLog[]>({
    dependencies: [staticTrigger],
    path: `/api/orgs/${org.id}/log-streams/${logStream?.id}/logs${params}`,
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
    initData: initLogs,
    initIsLoading: false,
    enabled: staticEnabled && !isStreamOpen,
  })

  const paginationCheckResults = useQuery<TOTELLog[]>({
    dependencies: [needsPaginationCheck],
    path: `/api/orgs/${org.id}/log-streams/${logStream?.id}/logs${params}`,
    headers: logs.length > 0 ? {
      'X-Nuon-API-Offset': String(new Date(logs[logs.length - 1]?.timestamp).getTime() * 1000000)
    } : {},
    initData: [],
    initIsLoading: false,
    enabled: needsPaginationCheck && !isStreamOpen,
  })

  const finalFetchResults = useQuery<TOTELLog[]>({
    dependencies: [needsFinalFetch],
    path: `/api/orgs/${org.id}/log-streams/${logStream?.id}/logs`,
    headers: logs.length > 0 ? {
      'X-Nuon-API-Offset': String(new Date(logs[logs.length - 1]?.timestamp).getTime() * 1000000)
    } : {},
    initData: [],
    initIsLoading: false,
    enabled: needsFinalFetch && !isStreamOpen,
  })

  useEffect(() => {
    if (!isStreamOpen && staticResults?.data) {
      setLogs((prev) => {
        const logMap = new Map(prev.map((log) => [log.id, log]))
        staticResults.data.forEach((log) => logMap.set(log.id, log))
        return Array.from(logMap.values())
      })

      if (staticResults?.headers) {
        const logOffset = staticResults?.headers?.['x-nuon-api-next']
        setOffset(logOffset)
        setHasMore(!!logOffset)
      }
    }
  }, [staticResults?.data, staticResults?.headers, isStreamOpen])

  useEffect(() => {
    if (!isStreamOpen && finalFetchResults?.data && needsFinalFetch) {
      if (finalFetchResults.data.length > 0) {
        setLogs((prev) => {
          const logMap = new Map(prev.map((log) => [log.id, log]))
          finalFetchResults.data.forEach((log) => logMap.set(log.id, log))
          return Array.from(logMap.values())
        })
      }

      if (finalFetchResults?.headers) {
        const nextOffset = finalFetchResults?.headers?.['x-nuon-api-next']
        const hasMoreLogs = !!nextOffset
        setHasMore(hasMoreLogs)
        if (hasMoreLogs && nextOffset) {
          setOffset(nextOffset)
        }
      } else {
        setHasMore(false)
      }
      setNeedsFinalFetch(false)
    }
  }, [finalFetchResults?.data, finalFetchResults?.headers, needsFinalFetch, isStreamOpen])

  useEffect(() => {
    if (paginationCheckResults?.headers && needsPaginationCheck) {
      const nextOffset = paginationCheckResults?.headers?.['x-nuon-api-next']
      const hasMoreLogs = !!nextOffset && paginationCheckResults.data.length > 0
      setHasMore(hasMoreLogs)
      if (hasMoreLogs && nextOffset) {
        setOffset(nextOffset)
      }
      setNeedsPaginationCheck(false)
    }
  }, [paginationCheckResults?.headers, paginationCheckResults?.data, needsPaginationCheck])

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
    ? connectionState === 'connecting' || connectionState === 'reconnecting'
    : staticResults?.isLoading || false
    
  const currentError = isStreamOpen ? error : staticResults?.error || null

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
  initLogs: TOTELLog[]
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