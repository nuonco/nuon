import type { IModal } from '@/components/surfaces/Modal'
import { ShutdownMngRunnerModal } from '@/components/runners/management/ShutdownMngRunner'
import { ShutdownRunnerModal } from '@/components/runners/management/ShutdownRunner'

interface IRestartRunnerModal extends IModal {
  runnerId: string
  isManaged: boolean
}

export const RestartRunnerModal = ({ runnerId, isManaged, ...modalProps }: IRestartRunnerModal) => {
  if (isManaged) {
    return <ShutdownMngRunnerModal runnerId={runnerId} {...modalProps} />
  }
  return <ShutdownRunnerModal runnerId={runnerId} {...modalProps} />
}
