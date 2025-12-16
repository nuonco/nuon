import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { LogsSkeleton as LogsViewerSkeleton } from '@/components/log-stream/Logs'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider-temp'
import { LogViewerProvider } from '@/providers/log-viewer-provider-temp'
import { getLogStreamLogs } from '@/lib'

export async function Logs({
  logStreamId,
  logStreamOpen,
  orgId,
}: {
  logStreamId: string
  logStreamOpen: boolean
  orgId: string
}) {
  const {
    data: logs,
    error,
    headers,
  } = await getLogStreamLogs({
    logStreamId,
    order: logStreamOpen ? 'asc' : 'desc',
    orgId,
  })

  return error ? (
    <LogsError />
  ) : (
    <UnifiedLogsProvider initLogs={logs}>
      <LogViewerProvider>
        <SSELogs filterClassName="-top-6" />
      </LogViewerProvider>
    </UnifiedLogsProvider>
  )
}

export const LogsSkeleton = () => {
  return (
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
        <LogsViewerSkeleton />
      </div>
    </div>
  )
}

export const LogsError = () => {
  return (
    <EmptyState
      emptyTitle="No logs found"
      emptyMessage="Unable to load logs for this sandbox run."
      variant="table"
    />
  )
}
