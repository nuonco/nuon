'use client'

import { usePathname } from 'next/navigation'
import { updateInstall } from '@/actions/installs/update-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

interface IEnableConfigSync {}

export const EnableConfigSyncModal = ({ ...props }: IEnableConfigSync & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const hasManagedBy = Boolean(install?.metadata?.managed_by)
  const isManagedByConfig =
    hasManagedBy && install?.metadata?.managed_by === 'nuon/cli/install-config'

  const { data, error, isLoading, execute } = useServerAction({
    action: updateInstall,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unable to update config sync for {install.name}.</Text>,
    errorHeading: 'Config sync update failed',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>
        Config sync has been {isManagedByConfig ? 'disabled' : 'enabled'} for {install.name}.
      </Text>
    ),
    successHeading: 'Config sync updated',
  })

  const buttonText = isManagedByConfig ? 'Disable Install Config Sync' : 'Enable Install Config Sync'
  const modalHeading = isManagedByConfig ? 'Disable Install Config Sync?' : 'Enable Install Config Sync?'

  const handleToggle = () => {
    execute({
      orgId: org.id,
      path,
      installId: install.id,
      body: {
        metadata: {
          managed_by: isManagedByConfig
            ? 'nuon/dashboard'
            : 'nuon/cli/install-config',
        },
      },
    })
  }

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="FileCloud" size="24" />
          {modalHeading}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {isManagedByConfig ? 'Disabling...' : 'Enabling...'}
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant={isManagedByConfig ? 'ToggleRight' : 'ToggleLeft'} />
            {buttonText}
          </span>
        ),
        onClick: handleToggle,
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        
        <Text variant="base">
          This Install can be managed via an Install Config file only after marking it as managed by Install Config.
        </Text>
        
        <Text variant="base">
          {isManagedByConfig
            ? 'Disabling this will stop any future syncs from the Install Config file.'
            : 'Enable this to allow syncing from an Install Config file.'}
        </Text>
      </div>
    </Modal>
  )
}

export const EnableConfigSyncButton = ({ ...props }: IEnableConfigSync & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()
  const modal = <EnableConfigSyncModal />

  const hasManagedBy = Boolean(install?.metadata?.managed_by)
  const isManagedByConfig =
    hasManagedBy && install?.metadata?.managed_by === 'nuon/cli/install-config'

  const buttonText = isManagedByConfig ? 'Disable install config sync' : 'Enable install config sync'

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {buttonText}
      <Icon variant="FileCloud" />
    </Button>
  )
}
