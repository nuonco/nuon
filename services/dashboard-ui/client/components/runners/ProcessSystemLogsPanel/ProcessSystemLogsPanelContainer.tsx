import { useRunnerProcess } from '@/hooks/use-runner-process'
import { ProcessSystemLogsPanel } from './ProcessSystemLogsPanel'

export const ProcessSystemLogsPanelContainer = ({
  runnerId,
  processId,
}: {
  runnerId?: string
  processId?: string
}) => {
  const { data: process, isLoading } = useRunnerProcess({ runnerId, processId })

  return (
    <ProcessSystemLogsPanel
      process={process}
      isLoading={isLoading}
      runnerId={runnerId}
    />
  )
}
