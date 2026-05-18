import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunLogsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const logStream = sandboxRun?.log_stream

  return (
    <LogStreamProvider logStreamId={logStream?.id}>
      <LogViewerProvider>
        <SSELogs filterClassName="top-0" />
      </LogViewerProvider>
    </LogStreamProvider>
  )
}
