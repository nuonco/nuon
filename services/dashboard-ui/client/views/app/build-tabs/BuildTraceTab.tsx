import { PageTitle } from '@/components/navigation/PageTitle'
import { TracePanel } from '@/components/spans/TracePanel'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'

export const BuildTraceTab = () => {
  const { build } = useBuild()
  const { app } = useApp()

  return (
    <>
      <PageTitle segments={['Build trace', app?.name]} />
      <TracePanel logStream={build?.log_stream} />
    </>
  )
}
