'use client'

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useLogFilters, type TLogFiltersProps } from '@/hooks/use-log-filters'
import { useSurfaces } from '@/hooks/use-surfaces'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider-temp'
import type { TOTELLog } from '@/types'

/**
 * TEMP VERSION - Log viewer provider for active log management and filtering
 * 
 * - Handles active log selection and state management
 * - Manages keyboard navigation (arrow keys)
 * - Integrates with surfaces for log panels
 * - Uses useLogFilters for filtering
 * - Independent of data source (works with SSE, HTTP, or any log array)
 */

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
  // Get live logs from UnifiedLogsProvider
  const unifiedLogsContext = useContext(UnifiedLogsContext)
  if (!unifiedLogsContext) {
    throw new Error('LogViewerProvider must be used within a UnifiedLogsProvider')
  }
  
  const { logs } = unifiedLogsContext
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const filters = useLogFilters(logs || [])
  const { addPanel, removePanel } = useSurfaces()

  function handleActiveLog(id?: string) {
    setActiveLog(id ? filters.filteredLogs?.find((l) => l.id === id) : undefined)
  }

  // Keyboard navigation
  useArrowKeys({
    onDownArrow() {
      if (activeLog && filters.filteredLogs && filters.filteredLogs.length > 0) {
        removePanel(activeLog?.id)
        const activeLogIndex = filters.filteredLogs.findIndex((l) => l.id === activeLog.id)
        const nextLogIndex = activeLogIndex + 1
        const nextLog = filters.filteredLogs?.at(
          nextLogIndex === filters.filteredLogs?.length ? 0 : nextLogIndex
        )
        setTimeout(() => {
          handleActiveLog(nextLog?.id)
        }, 160)
      }
    },
    onUpArrow() {
      if (activeLog && filters.filteredLogs && filters.filteredLogs.length > 0) {
        removePanel(activeLog?.id)
        const activeLogIndex = filters.filteredLogs.findIndex((l) => l.id === activeLog.id)
        const prevLog = filters.filteredLogs?.at(activeLogIndex - 1)
        setTimeout(() => {
          handleActiveLog(prevLog?.id)
        }, 160)
      }
    },
  })

  // Manage log panels
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

// Hook for consuming the log viewer context
export const useLogViewer = () => {
  const context = useContext(LogViewerContext)
  if (context === undefined) {
    throw new Error('useLogViewer must be used within a LogViewerProvider')
  }
  return context
}