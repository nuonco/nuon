import { useContext } from 'react'
import { LogStreamContext, type LogStreamContextValue } from '@/providers/log-stream-provider'
import { LogViewerContext } from '@/providers/log-viewer-provider'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import type { TOTELLog, TAPIError } from '@/types'

export const useLogStreamData = (): LogStreamContextValue => {
  const context = useContext(LogStreamContext)
  if (context === undefined) {
    throw new Error('useLogStreamData must be used within a LogStreamProvider')
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

export type { LogStreamContextValue }

export type LogViewerContextValue = {
  activeLog?: TOTELLog
  filteredLogs: TOTELLog[]
  filters: TLogFiltersProps
  handleActiveLog: (id?: string) => void
}
