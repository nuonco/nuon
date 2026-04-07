import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateInstall } from '@/lib'

interface IEnableConfigSync {}

export const EnableConfigSyncModal = ({ ...props }: IEnableConfigSync & IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const hasManagedBy = Boolean(install?.metadata?.managed_by)
  const isManagedByConfig =
    hasManagedBy && install?.metadata?.managed_by === 'nuon/cli/install-config'

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      updateInstall({
        orgId: org.id,
        installId: install.id,
        body: {
          metadata: {
            managed_by: isManagedByConfig
              ? 'nuon/dashboard'
              : 'nuon/cli/install-config',
          },
        },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Config sync updated" theme="success">
          <Text>
            Config sync has been {isManagedByConfig ? 'disabled' : 'enabled'} for {install.name}.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (error) => {
      addToast(
        <Toast heading="Config sync update failed" theme="error">
          <Text>Unable to update config sync for {install.name}.</Text>
        </Toast>
      )
    },
  })

  const buttonText = isManagedByConfig ? 'Disable Install Config Sync' : 'Enable Install Config Sync'
  const modalHeading = isManagedByConfig ? 'Disable Install Config Sync?' : 'Enable Install Config Sync?'

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
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
        onClick: () => mutate(),
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
