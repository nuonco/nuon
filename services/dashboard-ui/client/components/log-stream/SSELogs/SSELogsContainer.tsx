import { useSearchParams } from 'react-router'
import { useLogViewer, useUnifiedLogData } from '@/hooks/use-logs'
import { SSELogs } from './SSELogs'

export const SSELogsContainer = ({
  filterClassName = 'top-0',
}: {
  filterClassName?: string
}) => {
  const { loadMore, hasMore, isLoading, isStreamOpen } = useUnifiedLogData()
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
      loadMore={loadMore}
      hasMore={hasMore}
      isLoading={isLoading}
      isStreamOpen={isStreamOpen}
      deepLinkLogId={deepLinkLogId}
    />
  )
}
