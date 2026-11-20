import { Toggle } from '@/components/common/form/Toggle'
import { Tooltip } from '@/components/common/Tooltip'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRunnerSwitchLabel } from '@/utils/install-utils'
import { UseRunnerToggleModalContainer } from './UseRunnerToggleContainer'

const SWITCHING_TIP =
  'A reprovision is in progress. Try again once it finishes.'

export const RunnerModeToggle = () => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const usesRunner = install?.install_config?.use_runner ?? true
  const isSwitching = Boolean(getInstallRunnerSwitchLabel(install))

  const handleChange = () => {
    const modal = <UseRunnerToggleModalContainer />
    addModal(modal)
  }

  const toggle = (
    <Toggle
      checked={usesRunner}
      onChange={handleChange}
      disabled={isSwitching}
      label="Use a runner"
    />
  )

  if (isSwitching) {
    return (
      <Tooltip
        tipContent={SWITCHING_TIP}
        position="left"
        tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs"
      >
        {toggle}
      </Tooltip>
    )
  }

  return toggle
}
