import type { IModal } from '@/components/surfaces/Modal'
import { RunnerProvider } from '@/providers/runner-provider'
import { useRunner } from '@/hooks/use-runner'
import { ShutdownMngRunnerModal } from '@/components/runners/management/ShutdownMngRunner'
import { ShutdownRunnerModal } from '@/components/runners/management/ShutdownRunner'

const RestartRunnerInner = ({ runnerId, ...modalProps }: { runnerId: string } & IModal) => {
  const { isManaged } = useRunner()
  if (isManaged) {
    return <ShutdownMngRunnerModal runnerId={runnerId} {...modalProps} />
  }
  return <ShutdownRunnerModal runnerId={runnerId} {...modalProps} />
}

export const RestartRunnerModal = ({ runnerId, ...modalProps }: { runnerId: string } & IModal) => (
  <RunnerProvider runnerId={runnerId}>
    <RestartRunnerInner runnerId={runnerId} {...modalProps} />
  </RunnerProvider>
)
