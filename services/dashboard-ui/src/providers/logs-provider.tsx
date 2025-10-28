'use client'

import { createContext, useEffect, useState, type ReactNode } from 'react'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useLogFilters, type TLogFiltersProps } from '@/hooks/use-log-filters'
import { useLogStream } from '@/hooks/use-log-stream'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TOTELLog, TAPIError } from '@/types'

type LogsContextValue = {
  activeLog?: TOTELLog
  error: TAPIError
  filters: Omit<TLogFiltersProps, 'filteredLogs'>
  handleActiveLog: (id: string) => void
  isLoading: boolean
  logs: TOTELLog[] | null
  refresh: () => void
}

export const LogsContext = createContext<LogsContextValue | undefined>(
  undefined
)

export function LogsProvider({
  children,
  initLogs,
  pollInterval = 2000,
}: {
  children: ReactNode
  initLogs: TOTELLog[]
} & Omit<IPollingProps, 'shouldPoll'>) {
  const { addPanel, clearPanels } = useSurfaces()
  const { org } = useOrg()
  const { logStream } = useLogStream()
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const [offset, setOffset] = useState<string>()  
  const { data, error, headers, isLoading } = usePolling<TOTELLog[]>({
    path: `/api/orgs/${org.id}/log-streams/${logStream.id}/logs`,
    dependencies: [offset],
    headers: offset
      ? {
          'X-Nuon-API-Offset': offset,
        }
      : {},
    initData: initLogs,
    pollInterval,
    shouldPoll: logStream.open,
  })

  const [logs, setLogs] = useState<TOTELLog[]>(data)
  const { filteredLogs, ...filters } = useLogFilters(logs)

  useEffect(() => {
    setLogs((prev) => {
      const logMap = new Map(prev.map((log) => [log.id, log]))
      data.forEach((log) => logMap.set(log.id, log))
      return Array.from(logMap.values())
    })
    
    const logOffset = headers?.['x-nuon-api-next']
    if (logOffset) {
      setOffset(logOffset)
    }
  }, [data, headers])

  useArrowKeys({
    onDownArrow() {
      if (activeLog) {
        clearPanels()
        const activeLogIndex = filteredLogs.findIndex(
          (l) => l.id === activeLog.id
        )
        const nextLogIndex = activeLogIndex + 1
        const nextLog = filteredLogs?.at(
          nextLogIndex === filteredLogs?.length ? 0 : nextLogIndex
        )
        setTimeout(() => {
          setActiveLog(nextLog)
        }, 160)
      }
    },
    onUpArrow() {
      if (activeLog) {
        clearPanels()
        const activeLogIndex = filteredLogs.findIndex(
          (l) => l.id === activeLog.id
        )
        const prevLog = filteredLogs?.at(activeLogIndex - 1)
        setTimeout(() => {
          setActiveLog(prevLog)
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
            setActiveLog(undefined)
          }}
        />,
        activeLog.id
      )
    }
  }, [activeLog])

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
        refresh: () => {
          console.warn('logs refresh is not implemented')
        },
      }}
    >
      {children}
    </LogsContext.Provider>
  )
}
