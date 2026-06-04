import { Toggle } from '@/components/common/form/Toggle'
import { Tooltip } from '@/components/common/Tooltip'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { EnableAutoApproveModalContainer } from './EnableAutoApproveContainer'

const MANAGED_BY_CONFIG_TIP = 'Managed by config. Disable config sync to edit.'

export const AutoApproveToggle = () => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll =
    hasInstallConfig &&
    install?.install_config?.approval_option === 'approve-all'
  const isManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const handleChange = () => {
    const modal = <EnableAutoApproveModalContainer />
    addModal(modal)
  }

  const toggle = (
    <Toggle
      checked={isApproveAll}
      onChange={handleChange}
      disabled={isManagedByConfig}
      label="Auto approval"
    />
  )

  if (isManagedByConfig) {
    return (
      <Tooltip tipContent={MANAGED_BY_CONFIG_TIP} position="left" tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs">
        {toggle}
      </Tooltip>
    )
  }

  return toggle
}
