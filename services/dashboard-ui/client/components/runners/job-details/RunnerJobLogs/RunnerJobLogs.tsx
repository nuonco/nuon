import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'

const ACTIVE_STATUSES = ['in-progress', 'queued', 'available']

interface IRunnerJobLogs {
  logStreamId?: string
  jobStatus?: string
}

export const RunnerJobLogs = ({ logStreamId, jobStatus }: IRunnerJobLogs) => {
  const isJobActive = ACTIVE_STATUSES.includes(jobStatus ?? '')

  if (!logStreamId) return <LogsSkeleton />

  return (
    <LogStreamProvider logStreamId={logStreamId} shouldPoll={isJobActive}>
      <UnifiedLogsProvider>
        <LogViewerProvider>
          <SSELogs />
        </LogViewerProvider>
      </UnifiedLogsProvider>
    </LogStreamProvider>
  )
}
