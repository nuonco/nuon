import { TracePanel } from '@/components/spans/TracePanel'
import { useDeploy } from '@/hooks/use-deploy'

export const DeployTraceTab = () => {
  const { deploy } = useDeploy()
  return <TracePanel logStream={deploy?.log_stream} />
}
