import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'

interface IRunnerJobLogs {
  logStreamId?: string
}

export const RunnerJobLogs = ({ logStreamId }: IRunnerJobLogs) => {
  if (!logStreamId) return <LogsSkeleton />

  return (
    <LogStreamProvider logStreamId={logStreamId}>
      <LogViewerProvider>
        <SSELogs />
      </LogViewerProvider>
    </LogStreamProvider>
  )
}
