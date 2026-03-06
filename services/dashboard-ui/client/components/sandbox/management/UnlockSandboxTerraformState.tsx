import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
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
import { unlockTerraformWorkspace } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const UnlockSandboxTerraformStateButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <UnlockSandboxTerraformStateModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="LockOpen" />}
      Unlock Terraform state
      {props?.isMenuButton ? <Icon variant="LockOpen" /> : null}
    </Button>
  )
}

export const UnlockSandboxTerraformStateModal = ({
  ...props
}: IModal) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const workspaceId = install?.sandbox?.terraform_workspace?.id

  const {
    mutate: execute,
    isPending: isLoading,
    error,
  } = useMutation({
    mutationFn: () =>
      unlockTerraformWorkspace({
        orgId: org.id,
        terraformWorkspaceId: workspaceId!,
      }),
    onSuccess: () => {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          workspaceId,
        },
      })
      addToast(
        <Toast heading="Terraform state unlocked" theme="success">
          <Text>
            The sandbox Terraform workspace has been unlocked.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({
        queryKey: ['install', org?.id, install?.id],
      })
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          workspaceId,
          err: err?.error,
        },
      })
    },
  })

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
          >
            <Icon variant="LockOpen" size="24" />
            Unlock Terraform workspace
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            Force unlock the Terraform state for the sandbox
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Unlocking...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="LockOpen" />
            Force unlock
          </span>
        ),
        disabled: isLoading || !workspaceId,
        onClick: () => execute(),
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to unlock Terraform workspace'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="body" theme="neutral">
            Are you sure you want to force unlock this Terraform workspace? This
            should only be done if a previous operation failed to release the
            lock.
          </Text>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> Force unlocking a workspace that is
              actively in use by a running job may cause state corruption.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}
