import { SSELogs } from '@/components/log-stream/SSELogs'
import { PageTitle } from '@/components/navigation/PageTitle'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { useInstall } from '@/hooks/use-install'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunLogsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { install } = useInstall()
  const logStream = sandboxRun?.log_stream

  return (
    <>
      <PageTitle segments={['Sandbox run logs', install?.name]} />
      <LogStreamProvider logStreamId={logStream?.id}>
        <LogViewerProvider>
          <SSELogs filterClassName="top-0" />
        </LogViewerProvider>
      </LogStreamProvider>
    </>
  )
}
