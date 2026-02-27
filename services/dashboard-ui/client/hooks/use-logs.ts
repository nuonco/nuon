import { useContext } from 'react'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import { LogViewerContext } from '@/providers/log-viewer-provider'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import type { TOTELLog, TAPIError } from '@/types'

export const useUnifiedLogData = () => {
  const context = useContext(UnifiedLogsContext)
  if (context === undefined) {
    throw new Error('useUnifiedLogData must be used within a UnifiedLogsProvider')
  }
  return context
}

export const useLogViewer = () => {
  const context = useContext(LogViewerContext)
  if (context === undefined) {
    throw new Error('useLogViewer must be used within a LogViewerProvider')
  }
  return context
}

export const useUnifiedLogsComplete = () => {
  const logData = useUnifiedLogData()
  const viewer = useLogViewer()

  return {
    logs: logData.logs,
    isLoading: logData.isLoading,
    error: logData.error,
    connectionState: logData.connectionState,
    loadMore: logData.loadMore,
    hasMore: logData.hasMore,
    isStreamOpen: logData.isStreamOpen,
    activeLog: viewer.activeLog,
    filteredLogs: viewer.filteredLogs,
    filters: viewer.filters,
    handleActiveLog: viewer.handleActiveLog,
  }
}

export type UnifiedLogsContextValue = {
  logs: TOTELLog[]
  isLoading: boolean
  error: TAPIError | null
  connectionState: 'disconnected' | 'connecting' | 'connected' | 'reconnecting'
  loadMore: () => void
  hasMore: boolean
  isStreamOpen: boolean
}

export type LogViewerContextValue = {
  activeLog?: TOTELLog
  filteredLogs: TOTELLog[]
  filters: TLogFiltersProps
  handleActiveLog: (id?: string) => void
}
