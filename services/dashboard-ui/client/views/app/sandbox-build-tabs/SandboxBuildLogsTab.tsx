import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'

export const SandboxBuildLogsTab = () => {
  const { build } = useSandboxBuild()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Sandbox build logs', app?.name]} />
      <LogsPanel logStream={build?.log_stream} />
    </>
  )
}
