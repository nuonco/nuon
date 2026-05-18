import { useSearchParams } from 'react-router'
import { useLogStreamData, useLogViewer } from '@/hooks/use-logs'
import { SSELogs } from './SSELogs'

export const SSELogsContainer = ({
  filterClassName = 'top-0',
}: {
  filterClassName?: string
}) => {
  const { isLoading, connectionState } = useLogStreamData()
  const { filteredLogs, filters, activeLog, handleActiveLog } = useLogViewer()
  const [searchParams] = useSearchParams()
  const deepLinkLogId = searchParams?.get('log')

  return (
    <SSELogs
      filterClassName={filterClassName}
      filteredLogs={filteredLogs}
      filters={filters}
      activeLog={activeLog}
      handleActiveLog={handleActiveLog}
      isLoading={isLoading}
      isConnected={connectionState === 'connected'}
      deepLinkLogId={deepLinkLogId}
    />
  )
}
