import { TracePanel } from '@/components/spans/TracePanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunTraceTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { install } = useInstall()
  return (
    <>
      <PageTitle segments={['Sandbox run trace', install?.name]} />
      <TracePanel logStream={sandboxRun?.log_stream} />
    </>
  )
}
