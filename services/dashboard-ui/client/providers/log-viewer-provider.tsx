import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useLogFilters, type TLogFiltersProps } from '@/hooks/use-log-filters'
import { useSurfaces } from '@/hooks/use-surfaces'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import type { TOTELLog } from '@/types'

type LogViewerContextValue = {
  activeLog?: TOTELLog
  filteredLogs: TOTELLog[]
  filters: TLogFiltersProps
  handleActiveLog: (id?: string) => void
}

export const LogViewerContext = createContext<LogViewerContextValue | undefined>(undefined)

interface LogViewerProviderProps {
  children: ReactNode
}

export function LogViewerProvider({ children }: LogViewerProviderProps) {
  const unifiedLogsContext = useContext(UnifiedLogsContext)
  if (!unifiedLogsContext) {
    throw new Error('LogViewerProvider must be used within a UnifiedLogsProvider')
  }

  const { logs } = unifiedLogsContext
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const cycleDirectionRef = useRef<'up' | 'down' | undefined>()
  const filters = useLogFilters(logs || [])
  const { addPanel, updatePanel, removePanel } = useSurfaces()
  const navigate = useNavigate()

  function setLogParam(logId?: string) {
    const params = new URLSearchParams(window.location.search)
    if (logId) {
      params.set('log', logId)
    } else {
      params.delete('log')
    }
    navigate(`?${params.toString()}`, { replace: true })
  }

  function handleActiveLog(id?: string) {
    const log = id ? filters.filteredLogs?.find((l) => l.id === id) : undefined
    setActiveLog(log)
    setLogParam(log?.id)
  }

  useArrowKeys({
    onDownArrow() {
      if (activeLog && filters.filteredLogs && filters.filteredLogs.length > 0) {
        cycleDirectionRef.current = 'down'
        const idx = filters.filteredLogs.findIndex((l) => l.id === activeLog.id)
        const nextIdx = idx + 1 >= filters.filteredLogs.length ? 0 : idx + 1
        handleActiveLog(filters.filteredLogs[nextIdx]?.id)
      }
    },
    onUpArrow() {
      if (activeLog && filters.filteredLogs && filters.filteredLogs.length > 0) {
        cycleDirectionRef.current = 'up'
        const idx = filters.filteredLogs.findIndex((l) => l.id === activeLog.id)
        handleActiveLog(filters.filteredLogs.at(idx - 1)?.id)
      }
    },
  })

  const panelIdRef = useRef<string | undefined>()

  useEffect(() => {
    if (activeLog) {
      const panel = (
        <LogPanel
          log={activeLog}
          cycleDirection={cycleDirectionRef.current}
          onClose={() => handleActiveLog(undefined)}
        />
      )
      if (panelIdRef.current) {
        updatePanel(panelIdRef.current, panel)
      } else {
        cycleDirectionRef.current = undefined
        panelIdRef.current = 'log-panel'
        addPanel(panel, undefined, 'log-panel')
      }
    } else if (panelIdRef.current) {
      cycleDirectionRef.current = undefined
      removePanel(panelIdRef.current)
      panelIdRef.current = undefined
    }
  }, [activeLog])

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const logId = params.get('log')
    if (logId && !panelIdRef.current && filters.filteredLogs?.length) {
      handleActiveLog(logId)
    }
  }, [filters.filteredLogs?.length])

  return (
    <LogViewerContext.Provider
      value={{
        activeLog,
        filteredLogs: filters.filteredLogs || [],
        filters,
        handleActiveLog,
      }}
    >
      {children}
    </LogViewerContext.Provider>
  )
}

export const useLogViewer = () => {
  const context = useContext(LogViewerContext)
  if (context === undefined) {
    throw new Error('useLogViewer must be used within a LogViewerProvider')
  }
  return context
}
