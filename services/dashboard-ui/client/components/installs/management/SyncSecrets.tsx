import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { syncSecrets } from '@/lib'

interface ISyncSecrets {}

export const SyncSecretsModal = ({ ...props }: ISyncSecrets & IModal) => {
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      syncSecrets({
        orgId: org.id,
        installId: install.id,
        body: { plan_only: false },
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading="Secret sync started" theme="success">
          <Text>Secrets for {install.name} are being synchronized.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (error) => {
      addToast(
        <Toast heading="Secret sync failed" theme="error">
          <Text>Unable to sync secrets for {install.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant="Key" size="24" />
          Sync secrets?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Syncing secrets
          </span>
        ) : (
          'Sync secrets'
        ),
        onClick: () => mutate(),
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="strong">
          Are you sure you want to sync secrets for {install.name}?
        </Text>
        <Text variant="base">
          This will synchronize all secrets from your app configuration to the
          install environment.
        </Text>
      </div>
    </Modal>
  )
}

export const SyncSecretsButton = ({
  ...props
}: ISyncSecrets & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <SyncSecretsModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Sync secrets
      <Icon variant="Key" />
    </Button>
  )
}
