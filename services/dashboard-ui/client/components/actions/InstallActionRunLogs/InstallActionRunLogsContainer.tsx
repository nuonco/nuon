import { useSearchParams } from 'react-router'
import { useUnifiedLogData, useLogViewer } from '@/hooks/use-logs'
import type { TActionConfig } from '@/types'
import { InstallActionRunLogs } from './InstallActionRunLogs'

export const InstallActionRunLogsContainer = ({
  actionConfig,
  layout = 'vertical',
}: {
  actionConfig: TActionConfig
  layout?: 'vertical' | 'horizontal'
}) => {
  const [searchParams] = useSearchParams()
  const { loadMore, hasMore, isLoading, isStreamOpen } = useUnifiedLogData()
  const { filteredLogs, activeLog, handleActiveLog, filters } = useLogViewer()

  return (
    <InstallActionRunLogs
      actionConfig={actionConfig}
      layout={layout}
      filteredLogs={filteredLogs}
      loadMore={loadMore}
      hasMore={hasMore}
      isLoading={isLoading}
      isStreamOpen={isStreamOpen}
      activeLog={activeLog}
      handleActiveLog={handleActiveLog}
      filters={filters}
      searchParamPanel={searchParams.get('panel')}
    />
  )
}
