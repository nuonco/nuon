import { Toggle } from '@/components/common/form/Toggle'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import {
  ConfirmOverrideModalContainer,
  EnableAutoApproveModalContainer,
} from './EnableAutoApproveContainer'

export const AutoApproveToggle = () => {
  const { addModal } = useSurfaces()
  const { install } = useInstall()

  const hasInstallConfig = Boolean(install?.install_config)
  const isApproveAll =
    hasInstallConfig &&
    install?.install_config?.approval_option === 'approve-all'
  const isInstallManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  const handleChange = () => {
    if (isInstallManagedByConfig) {
      const overrideModal = (
        <ConfirmOverrideModalContainer
          onConfirm={() => {
            const mainModal = <EnableAutoApproveModalContainer />
            addModal(mainModal)
          }}
        />
      )
      addModal(overrideModal)
    } else {
      const modal = <EnableAutoApproveModalContainer />
      addModal(modal)
    }
  }

  return (
    <Toggle
      checked={isApproveAll}
      onChange={handleChange}
      label="Auto approval"
    />
  )
}
