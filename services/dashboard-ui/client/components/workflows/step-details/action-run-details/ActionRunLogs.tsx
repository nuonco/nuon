import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { IActionRunLogs } from './types'

export const ActionRunLogs = ({ actionRun, isAdhoc }: IActionRunLogs) => {
  if (!actionRun?.log_stream) {
    return null
  }

  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Action logs</Text>
      <LogStreamProvider
        shouldPoll={actionRun.log_stream.open}
        logStreamId={actionRun.log_stream.id}
      >
        <UnifiedLogsProvider>
          <LogViewerProvider>
            {isAdhoc ? (
              <SSELogs />
            ) : (
              <InstallActionRunLogs
                actionConfig={actionRun?.config}
                layout="horizontal"
              />
            )}
          </LogViewerProvider>
        </UnifiedLogsProvider>
      </LogStreamProvider>
    </div>
  )
}

export const ActionRunLogsSkeleton = () => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        <Skeleton height="32px" width="120px" />
        <Skeleton height="32px" width="100px" />
        <Skeleton height="32px" width="140px" />
        <Skeleton height="32px" width="110px" />
      </div>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Skeleton height="36px" width="320px" />
            <Skeleton height="17px" width="85px" />
          </div>
          <div className="flex items-center gap-4">
            <Skeleton height="32px" width="86px" />
            <Skeleton height="32px" width="135px" />
            <Skeleton height="32px" width="140px" />
          </div>
        </div>
        <div>
          <LogsSkeleton />
        </div>
      </div>
    </div>
  )
}
