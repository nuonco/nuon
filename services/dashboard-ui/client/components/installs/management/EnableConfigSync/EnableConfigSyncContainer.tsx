import { useMutation, useQueryClient } from '@tanstack/react-query'
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
import { DisableConfigSyncModal, EnableConfigSyncModal } from './EnableConfigSync'

const DisableConfigSyncModalContainer = ({ ...props }: Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      updateInstall({
        orgId: org.id,
        installId: install.id,
        body: { metadata: { managed_by: 'nuon/dashboard' } },
      }),
    onSuccess: () => {
      removeModal(props.modalId)
      queryClient.invalidateQueries({ queryKey: ['install', org.id, install.id] })
      addToast(
        <Toast heading="Config sync disabled" theme="success">
          <Text>Config sync has been disabled for {install.name}.</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Config sync update failed" theme="error">
          <Text>Unable to update config sync for {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DisableConfigSyncModal
      installName={install?.name ?? ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

const EnableConfigSyncModalContainer = ({ ...props }: Omit<IModal, 'onSubmit'>) => {
  const queryClient = useQueryClient()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      updateInstall({
        orgId: org.id,
        installId: install.id,
        body: { metadata: { managed_by: 'nuon/cli/install-config' } },
      }),
    onSuccess: () => {
      removeModal(props.modalId)
      queryClient.invalidateQueries({ queryKey: ['install', org.id, install.id] })
      addToast(
        <Toast heading="Config sync enabled" theme="success">
          <Text>Config sync has been enabled for {install.name}.</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Config sync update failed" theme="error">
          <Text>Unable to update config sync for {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <EnableConfigSyncModal
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const EnableConfigSyncButton = ({ ...props }: IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const isManagedByConfig = install?.metadata?.managed_by === 'nuon/cli/install-config'

  const handleClick = () => {
    if (isManagedByConfig) {
      addModal(<DisableConfigSyncModalContainer />)
    } else {
      addModal(<EnableConfigSyncModalContainer />)
    }
  }

  return (
    <Button onClick={handleClick} {...props}>
      {isManagedByConfig ? 'Disable config sync' : 'Enable config sync'}
      <Icon variant="FileCloudIcon" />
    </Button>
  )
}
