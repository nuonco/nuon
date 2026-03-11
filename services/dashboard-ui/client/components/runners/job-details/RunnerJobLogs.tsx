import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { useRunnerJob } from '@/hooks/use-runner-job'

const ACTIVE_STATUSES = ['in-progress', 'queued', 'available']

export const RunnerJobLogs = () => {
  const { job } = useRunnerJob()
  const isJobActive = ACTIVE_STATUSES.includes(job.status ?? '')

  if (!job.log_stream_id) return <LogsSkeleton />

  return (
    <LogStreamProvider logStreamId={job.log_stream_id} shouldPoll={isJobActive}>
      <UnifiedLogsProvider>
        <LogViewerProvider>
          <SSELogs />
        </LogViewerProvider>
      </UnifiedLogsProvider>
    </LogStreamProvider>
  )
}
