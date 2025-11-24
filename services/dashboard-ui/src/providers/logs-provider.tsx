'use client'

import { createContext, useEffect, useState, type ReactNode } from 'react'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useLogFilters, type TLogFiltersProps } from '@/hooks/use-log-filters'
import { useLogStream } from '@/hooks/use-log-stream'
import { useLogs } from '@/hooks/use-logs'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TOTELLog, TAPIError } from '@/types'

const useLoadLogs = ({
  initLogs,
  initOffset,
}: {
  initLogs: TOTELLog[] | null
  initOffset?: string
}) => {
  const { org } = useOrg()
  const { logStream } = useLogStream()
  const shouldPoll = logStream?.open || false
  const params = useQueryParams({ order: shouldPoll ? 'asc' : 'desc' })
  const [offset, setOffset] = useState<string>(initOffset)
  const [staticEnabled, setStaticEnabled] = useState(false)
  const [staticTrigger, setStaticTrigger] = useState(0)

  const loadNextLogs = () => {
    if (!staticEnabled) {
      setStaticEnabled(true)
    }
    setStaticTrigger(prev => prev + 1)
  }

  const pollingResults = usePolling<TOTELLog[]>({
    path: `/api/orgs/${org.id}/log-streams/${logStream.id}/logs`,
    dependencies: [offset],
    headers: offset
      ? {
          'X-Nuon-API-Offset': offset,
        }
      : {},
    initData: initLogs,
    pollInterval: 2000,
    shouldPoll,
  })

  const staticResults = useQuery<TOTELLog[]>({
    dependencies: [staticTrigger],
    path: `/api/orgs/${org.id}/log-streams/${logStream.id}/logs${params}`,
    headers: offset
      ? {
          'X-Nuon-API-Offset': offset,
        }
      : {},
    initData: initLogs,
    initIsLoading: false,
    enabled: staticEnabled,
  })

  const results = shouldPoll ? pollingResults : staticResults

  const [logs, setLogs] = useState<TOTELLog[]>(results?.data)

  useEffect(() => {
    setLogs((prev) => {
      const logMap = new Map(prev.map((log) => [log.id, log]))
      results?.data.forEach((log) => logMap.set(log.id, log))
      return Array.from(logMap.values())
    })

    if (results?.headers) {
      const logOffset = results?.headers?.['x-nuon-api-next']
      setOffset(logOffset)
    }
  }, [results?.data, results?.headers])

  return {
    logs,
    isLoading: results?.isLoading,
    error: results?.error,
    loadNextLogs,
    offset,
  }
}

type LogsContextValue = {
  activeLog?: TOTELLog
  error: TAPIError
  filters: Omit<TLogFiltersProps, 'filteredLogs'>
  handleActiveLog: (id: string) => void
  isLoading: boolean
  logs: TOTELLog[] | null
  loadNextLogs: () => void
  offset?: string
}

export const LogsContext = createContext<LogsContextValue | undefined>(
  undefined
)

export function LogsProvider({
  children,
  initLogs,
  initOffset,
}: {
  children: ReactNode
  initLogs: TOTELLog[]
  initOffset?: string
} & Omit<IPollingProps, 'shouldPoll'>) {
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const { logs, isLoading, error, loadNextLogs, offset } = useLoadLogs({
    initLogs,
    initOffset,
  })
  const { filteredLogs, ...filters } = useLogFilters(logs)

  function handleActiveLog(id?: string) {
    setActiveLog(id ? filteredLogs.find((l) => l.id === id) : undefined)
  }

  return (
    <LogsContext.Provider
      value={{
        activeLog,
        filters,
        error,
        handleActiveLog,
        isLoading,
        logs: filteredLogs,
        loadNextLogs,
        offset,
      }}
    >
      <LogViewer>{children}</LogViewer>
    </LogsContext.Provider>
  )
}

const LogViewer = ({ children }) => {
  const { addPanel, removePanel } = useSurfaces()
  const { activeLog, handleActiveLog, logs } = useLogs()

  useArrowKeys({
    onDownArrow() {
      if (activeLog) {
        removePanel(activeLog?.id)
        const activeLogIndex = logs.findIndex((l) => l.id === activeLog.id)
        const nextLogIndex = activeLogIndex + 1
        const nextLog = logs?.at(
          nextLogIndex === logs?.length ? 0 : nextLogIndex
        )
        setTimeout(() => {
          handleActiveLog(nextLog?.id)
        }, 160)
      }
    },
    onUpArrow() {
      if (activeLog) {
        removePanel(activeLog?.id)
        const activeLogIndex = logs.findIndex((l) => l.id === activeLog.id)
        const prevLog = logs?.at(activeLogIndex - 1)
        setTimeout(() => {
          handleActiveLog(prevLog?.id)
        }, 160)
      }
    },
  })

  useEffect(() => {
    if (activeLog) {
      addPanel(
        <LogPanel
          log={activeLog}
          onClose={() => {
            handleActiveLog(undefined)
          }}
        />,
        activeLog.id,
        activeLog.id
      )
    }
  }, [activeLog])

  return <>{children}</>
}
