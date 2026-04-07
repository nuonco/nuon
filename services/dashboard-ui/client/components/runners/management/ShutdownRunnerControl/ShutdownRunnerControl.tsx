import type { IButtonAsButton } from '@/components/common/Button'
import { ShutdownRunnerButton } from '../ShutdownRunner'
import { ShutdownMngRunnerButton } from '../ShutdownMngRunner'

interface IShutdownRunnerControl extends IButtonAsButton {
  runnerId: string
  processId?: string
  showRunnerLabel?: boolean
  isManaged: boolean
}

export const ShutdownRunnerControl = ({
  runnerId,
  processId,
  showRunnerLabel,
  isManaged,
  ...props
}: IShutdownRunnerControl) => {
  if (isManaged) {
    return (
      <ShutdownMngRunnerButton
        runnerId={runnerId}
        processId={processId}
        showRunnerLabel={showRunnerLabel}
        {...props}
      />
    )
  }
  return (
    <ShutdownRunnerButton
      runnerId={runnerId}
      processId={processId}
      showRunnerLabel={showRunnerLabel}
      {...props}
    />
  )
}
