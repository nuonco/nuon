import { useRunner } from '@/hooks/use-runner'
import type { IButtonAsButton } from '@/components/common/Button'
import { ShutdownRunnerButton } from './ShutdownRunner'
import { ShutdownMngRunnerButton } from './ShutdownMngRunner'

interface IShutdownRunnerControl extends IButtonAsButton {
  runnerId: string
  showRunnerLabel?: boolean
}

export const ShutdownRunnerControl = ({ runnerId, showRunnerLabel, ...props }: IShutdownRunnerControl) => {
  const { isManaged } = useRunner()
  if (isManaged) {
    return <ShutdownMngRunnerButton runnerId={runnerId} showRunnerLabel={showRunnerLabel} {...props} />
  }
  return <ShutdownRunnerButton runnerId={runnerId} showRunnerLabel={showRunnerLabel} {...props} />
}
