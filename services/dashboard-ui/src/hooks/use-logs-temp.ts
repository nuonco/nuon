'use client'

import { useContext } from 'react'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider-temp'
import { LogViewerContext } from '@/providers/log-viewer-provider-temp'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import type { TOTELLog, TAPIError } from '@/types'

/**
 * TEMP VERSION - Hook exports for easy testing of the new unified log system
 * 
 * Usage example:
 * 
 * import { UnifiedLogsProvider } from '@/providers/unified-logs-provider-temp'
 * import { LogViewerProvider } from '@/providers/log-viewer-provider-temp'
 * import { useUnifiedLogData, useLogViewer } from '@/hooks/use-logs-temp'
 * 
 * function MyComponent() {
 *   return (
 *     <UnifiedLogsProvider initLogs={[]}>
 *       <LogViewerProvider>
 *         <LogList />
 *       </LogViewerProvider>
 *     </UnifiedLogsProvider>
 *   )
 * }
 * 
 * function LogList() {
 *   const { logs, isLoading, loadMore } = useUnifiedLogData()
 *   const { activeLog, handleActiveLog, filteredLogs, filters } = useLogViewer()
 *   
 *   // Your component logic here...
 * }
 */

/**
 * Hook to access unified log data (SSE or HTTP based on stream state)
 */
export const useUnifiedLogData = () => {
  const context = useContext(UnifiedLogsContext)
  if (context === undefined) {
    throw new Error('useUnifiedLogData must be used within a UnifiedLogsProvider')
  }
  return context
}

/**
 * Hook to access log viewer state (active log, filtering, etc.)
 */
export const useLogViewer = () => {
  const context = useContext(LogViewerContext)
  if (context === undefined) {
    throw new Error('useLogViewer must be used within a LogViewerProvider')
  }
  return context
}

/**
 * Combined hook that provides both data and viewer state
 * Use this when you need both in the same component
 */
export const useUnifiedLogsComplete = () => {
  const logData = useUnifiedLogData()
  const viewer = useLogViewer()
  
  return {
    // Data from UnifiedLogsProvider
    logs: logData.logs,
    isLoading: logData.isLoading,
    error: logData.error,
    connectionState: logData.connectionState,
    loadMore: logData.loadMore,
    hasMore: logData.hasMore,
    isStreamOpen: logData.isStreamOpen,
    
    // Viewer state from LogViewer
    activeLog: viewer.activeLog,
    filteredLogs: viewer.filteredLogs,
    filters: viewer.filters,
    handleActiveLog: viewer.handleActiveLog,
  }
}

/**
 * Type definitions for the unified system
 */
// Re-export context value types (these are defined in the provider files)
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