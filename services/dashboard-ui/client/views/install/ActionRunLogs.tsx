import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { PageTitle } from '@/components/navigation/PageTitle'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'

export const ActionRunLogsPage = () => {
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()
  const logStream = installActionRun?.log_stream

  return (
    <div className="flex flex-col gap-6">
      <PageTitle title={`Logs | ${install?.name}`} />
      {logStream ? (
        <LogStreamProvider logStreamId={logStream.id} shouldPoll={logStream.open}>
          <UnifiedLogsProvider>
            <LogViewerProvider>
              <InstallActionRunLogs actionConfig={installActionRun?.config} />
            </LogViewerProvider>
          </UnifiedLogsProvider>
        </LogStreamProvider>
      ) : (
        <div className="flex flex-col items-center gap-4 p-12">
          <Text variant="base" weight="strong">Waiting on log stream</Text>
          <Text variant="body" theme="neutral">Logs will appear here once the runner starts.</Text>
          <Button variant="ghost" onClick={() => window.location.reload()}>
            <Icon variant="ArrowClockwiseIcon" />
            Refresh Page
          </Button>
        </div>
      )}
    </div>
  )
}
