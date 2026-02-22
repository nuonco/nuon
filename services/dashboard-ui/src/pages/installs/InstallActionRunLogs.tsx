import { useParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { InstallActionRunLogs as InstallActionRunLogsComponent } from '@/components/actions/InstallActionRunLogs'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { LogsSkeleton as LogsViewerSkeleton } from '@/components/log-stream/Logs'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider-temp'
import { LogViewerProvider } from '@/providers/log-viewer-provider-temp'
import type { TInstallActionRun, TInstallAction, TLogStreamLog } from '@/types'

const LogsSkeleton = () => {
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

const LogsError = () => {
  return (
    <EmptyState
      className="!my-8"
      emptyTitle="No logs found"
      emptyMessage="Unable to load logs for this action run."
      variant="table"
    />
  )
}

export default function InstallActionRunLogs() {
  const { orgId, installId, actionId, runId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installActionRun } = usePolling<TInstallActionRun>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}/runs/${runId}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const { data: installAction } = useQuery<TInstallAction>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}`,
  })

  const { data: logs, error, isLoading } = usePolling<TLogStreamLog[]>({
    path: installActionRun?.log_stream?.id
      ? `/api/ctl-api/v1/log-streams/${installActionRun.log_stream.id}/logs?order=${installActionRun.log_stream.open ? 'asc' : 'desc'}`
      : '',
    shouldPoll: installActionRun?.log_stream?.open,
    pollInterval: 5000,
    dependencies: [installActionRun?.log_stream?.id],
  })

  if (isLoading) {
    return <LogsSkeleton />
  }

  if (error) {
    return <LogsError />
  }

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name || '',
          },
          {
            path: `/${orgId}/installs/${installId}/actions`,
            text: 'Actions',
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}`,
            text: installAction?.action_workflow?.name || 'Action',
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}`,
            text: `${installActionRun?.trigger_type || ''} run`,
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}/logs`,
            text: `Logs`,
          },
        ]}
      />
      <LogStreamProvider
        initLogStream={installActionRun?.log_stream || null}
        shouldPoll={installActionRun?.log_stream?.open}
      >
        <UnifiedLogsProvider initLogs={logs || []}>
          <LogViewerProvider>
            <InstallActionRunLogsComponent actionConfig={installActionRun?.config} />
          </LogViewerProvider>
        </UnifiedLogsProvider>
      </LogStreamProvider>
    </>
  )
}
