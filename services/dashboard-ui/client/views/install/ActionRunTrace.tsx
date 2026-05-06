import { PageTitle } from '@/components/navigation/PageTitle'
import { TracePanel } from '@/components/spans/TracePanel'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'

export const ActionRunTracePage = () => {
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()

  return (
    <>
      <PageTitle title={`Trace | ${install?.name}`} />
      <TracePanel logStream={installActionRun?.log_stream} />
    </>
  )
}
