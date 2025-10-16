import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { LogsSkeleton as LogsViewerSkeleton } from '@/components/log-stream/Logs'
import { LogsProvider } from '@/providers/logs-provider'
import { getLogsByLogStreamId } from '@/lib'

export async function Logs({
  logStreamId,
  orgId,
}: {
  logStreamId: string
  orgId: string
}) {
  const { data: logs, error } = await getLogsByLogStreamId({
    logStreamId,
    orgId,
  })

  return error ? (
    <LogsError />
  ) : (
    <LogsProvider initLogs={logs}>
      <InstallActionRunLogs />
    </LogsProvider>
  )
}

export const LogsSkeleton = () => {
  return (
    <div className="flex items-start flex-auto divide-x">
      <div className="flex flex-col gap-2 w-fit md:min-w-64 pr-2 h-full">
        <Skeleton height="32px" width="100%" />
        <Skeleton height="32px" width="100%" />
        <Skeleton height="32px" width="100%" />
        <Skeleton height="32px" width="100%" />
      </div>
      <div className="flex flex-col gap-4 pl-2 w-full">
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
    </div>
  )
}

export const LogsError = () => {
  return (
    <EmptyState
      className="!my-8"
      emptyTitle="No logs found"
      emptyMessage="Unable to load logs for this action run."
      variant="table"
    />
  )
}
