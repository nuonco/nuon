'use client'

import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogsSkeleton as LogsViewerSkeleton } from '@/components/log-stream/Logs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider-temp'
import { LogViewerProvider } from '@/providers/log-viewer-provider-temp'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TOTELLog } from '@/types'
import type { IActionRunLogs } from './types'

export const ActionRunLogs = ({ actionRun, isAdhoc }: IActionRunLogs) => {
  const { org } = useOrg()

  const params = useQueryParams({
    order: actionRun?.log_stream?.open ? 'asc' : 'desc',
  })

  const { data: logs, isLoading: isLoadingLogs } = useQuery<TOTELLog[]>({
    dependencies: [actionRun?.log_stream?.id],
    path: actionRun?.log_stream?.id
      ? `/api/orgs/${org.id}/log-streams/${actionRun?.log_stream?.id}/logs${params}`
      : null,
  })

  if (!actionRun?.log_stream) {
    return null
  }

  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Action logs</Text>
      {isLoadingLogs && !logs?.length ? (
        <ActionRunLogsSkeleton />
      ) : (
        <LogStreamProvider
          shouldPoll={actionRun?.log_stream?.open}
          initLogStream={actionRun?.log_stream}
        >
          <UnifiedLogsProvider initLogs={logs}>
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
      )}
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
          <LogsViewerSkeleton />
        </div>
      </div>
    </div>
  )
}
