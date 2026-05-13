import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { syncSecrets } from '@/lib'
import { SyncSecretsModal } from './SyncSecrets'

interface ISyncSecrets {}

export const SyncSecretsModalContainer = ({ ...props }: ISyncSecrets & Omit<IModal, 'onSubmit'>) => {
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
      const workflowId = result.data.workflow_id
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
    <SyncSecretsModal
      installName={install.name}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const SyncSecretsButton = ({
  ...props
}: ISyncSecrets & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <SyncSecretsModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Sync secrets
      <Icon variant="KeyIcon" />
    </Button>
  )
}
