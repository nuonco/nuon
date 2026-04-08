import type { IModal } from '@/components/surfaces/Modal'
import { RunnerProvider } from '@/providers/runner-provider'
import { useRunner } from '@/hooks/use-runner'
import { RestartRunnerModal } from './RestartRunnerModal'

const RestartRunnerModalInner = ({ runnerId, ...modalProps }: { runnerId: string } & IModal) => {
  const { isManaged } = useRunner()
  return <RestartRunnerModal runnerId={runnerId} isManaged={isManaged} {...modalProps} />
}

export const RestartRunnerModalContainer = ({ runnerId, ...modalProps }: { runnerId: string } & IModal) => (
  <RunnerProvider runnerId={runnerId}>
    <RestartRunnerModalInner runnerId={runnerId} {...modalProps} />
  </RunnerProvider>
)
