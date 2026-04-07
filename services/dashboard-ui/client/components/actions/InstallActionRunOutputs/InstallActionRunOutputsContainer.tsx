import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { InstallActionRunOutputs } from './InstallActionRunOutputs'

export const InstallActionRunOutputsContainer = () => {
  const { installActionRun } = useInstallActionRun()

  return <InstallActionRunOutputs installActionRun={installActionRun} />
}
