import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunLogsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { install } = useInstall()

  return (
    <>
      <PageTitle segments={['Sandbox run logs', install?.name]} />
      <LogsPanel logStream={sandboxRun?.log_stream} />
    </>
  )
}
