import { TracePanel } from '@/components/spans/TracePanel'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'

export const DeployTraceTab = () => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  return (
    <>
      <PageTitle segments={['Deploy trace', install?.name]} />
      <TracePanel logStream={deploy?.log_stream} />
    </>
  )
}
