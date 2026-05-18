import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { useDeploy } from '@/hooks/use-deploy'

export const DeployLogsTab = () => {
  const { deploy } = useDeploy()
  const logStream = deploy?.log_stream

  return (
    <LogStreamProvider logStreamId={logStream?.id}>
      <LogViewerProvider>
        <SSELogs filterClassName="top-0" />
      </LogViewerProvider>
    </LogStreamProvider>
  )
}
