import { ActionRunLogsDeleted } from '@/components/actions/DeleteActionRunLogs'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TracePanel } from '@/components/spans/TracePanel'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'

export const ActionRunTracePage = () => {
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()
  const logStream = installActionRun?.log_stream

  return (
    <>
      <PageTitle segments={['Trace', install?.name]} />
      {logStream?.logs_deleted_at ? (
        <ActionRunLogsDeleted />
      ) : (
        <TracePanel logStream={logStream} />
      )}
    </>
  )
}
