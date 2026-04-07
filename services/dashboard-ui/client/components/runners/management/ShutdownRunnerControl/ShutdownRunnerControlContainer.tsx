import { useRunner } from '@/hooks/use-runner'
import { ShutdownRunnerControl } from './ShutdownRunnerControl'
import type { IButtonAsButton } from '@/components/common/Button'

interface IShutdownRunnerControlContainer extends IButtonAsButton {
  runnerId: string
  processId?: string
  showRunnerLabel?: boolean
}

export const ShutdownRunnerControlContainer = ({
  runnerId,
  processId,
  showRunnerLabel,
  ...props
}: IShutdownRunnerControlContainer) => {
  const { isManaged } = useRunner()

  return (
    <ShutdownRunnerControl
      runnerId={runnerId}
      processId={processId}
      showRunnerLabel={showRunnerLabel}
      isManaged={isManaged}
      {...props}
    />
  )
}
