import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunLogsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const logStream = sandboxRun?.log_stream

  if (!logStream) return <LogsSkeleton />

  return (
    <LogStreamProvider logStreamId={logStream.id} shouldPoll={logStream.open}>
      <UnifiedLogsProvider>
        <LogViewerProvider>
          <SSELogs filterClassName="top-0" />
        </LogViewerProvider>
      </UnifiedLogsProvider>
    </LogStreamProvider>
  )
}
