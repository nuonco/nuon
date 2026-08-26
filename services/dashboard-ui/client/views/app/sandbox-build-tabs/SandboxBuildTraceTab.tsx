import { PageTitle } from '@/components/navigation/PageTitle'
import { TracePanel } from '@/components/spans/TracePanel'
import { useApp } from '@/hooks/use-app'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'

export const SandboxBuildTraceTab = () => {
  const { build } = useSandboxBuild()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Sandbox build trace', app?.name]} />
      <TracePanel logStream={build?.log_stream} />
    </>
  )
}
