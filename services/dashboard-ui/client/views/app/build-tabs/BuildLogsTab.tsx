import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'

export const BuildLogsTab = () => {
  const { build } = useBuild()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Build logs', app?.name]} />
      <LogsPanel logStream={build?.log_stream} />
    </>
  )
}
