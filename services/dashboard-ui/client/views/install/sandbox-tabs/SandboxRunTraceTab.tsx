import { TracePanel } from '@/components/spans/TracePanel'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunTraceTab = () => {
  const { sandboxRun } = useSandboxRun()
  return <TracePanel logStream={sandboxRun?.log_stream} />
}
