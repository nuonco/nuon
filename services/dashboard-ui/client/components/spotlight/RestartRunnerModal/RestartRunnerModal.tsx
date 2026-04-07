import type { IModal } from '@/components/surfaces/Modal'
import { ShutdownMngRunnerModal } from '@/components/runners/management/ShutdownMngRunner'
import { ShutdownRunnerModal } from '@/components/runners/management/ShutdownRunner'

interface IRestartRunnerModal extends IModal {
  runnerId: string
  processId?: string
  isManaged: boolean
}

export const RestartRunnerModal = ({ runnerId, processId = '', isManaged, ...modalProps }: IRestartRunnerModal) => {
  if (isManaged) {
    return <ShutdownMngRunnerModal runnerId={runnerId} processId={processId} {...modalProps} />
  }
  return <ShutdownRunnerModal runnerId={runnerId} processId={processId} {...modalProps} />
}
