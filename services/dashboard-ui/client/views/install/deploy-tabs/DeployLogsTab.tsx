import { SSELogs } from '@/components/log-stream/SSELogs'
import { PageTitle } from '@/components/navigation/PageTitle'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'

export const DeployLogsTab = () => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  const logStream = deploy?.log_stream

  return (
    <>
      <PageTitle segments={['Deploy logs', install?.name]} />
      <LogStreamProvider logStreamId={logStream?.id}>
        <LogViewerProvider>
          <SSELogs filterClassName="top-0" />
        </LogViewerProvider>
      </LogStreamProvider>
    </>
  )
}
