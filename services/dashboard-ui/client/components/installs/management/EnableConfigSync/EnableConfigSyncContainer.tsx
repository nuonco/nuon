import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateInstall } from '@/lib'
import { EnableConfigSyncModal } from './EnableConfigSync'

interface IEnableConfigSync {}

export const EnableConfigSyncModalContainer = ({ ...props }: IEnableConfigSync & Omit<IModal, 'onSubmit'>) => {
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

  return (
    <EnableConfigSyncModal
      isManagedByConfig={isManagedByConfig}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const EnableConfigSyncButton = ({ ...props }: IEnableConfigSync & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()
  const modal = <EnableConfigSyncModalContainer />

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
