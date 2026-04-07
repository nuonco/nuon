import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getTerraformWorkspaceLock, unlockTerraformWorkspace } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { UnlockTerraformWorkspaceModal } from './UnlockTerraformWorkspace'

interface IUnlockTerraformWorkspace {
  workspaceId: string
  description?: string
  onSuccess?: () => void
}

export const UnlockTerraformWorkspaceButton = ({
  workspaceId,
  description = 'the workspace',
  onSuccess,
  ...props
}: IUnlockTerraformWorkspace & IButtonAsButton) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const { data: lock } = useQuery({
    queryKey: ['terraform-workspace-lock', org?.id, workspaceId],
    queryFn: () => getTerraformWorkspaceLock({ orgId: org.id, workspaceId }),
    enabled: !!workspaceId,
  })

  if (!lock) return null

  const modal = (
    <UnlockTerraformWorkspaceModalContainer
      workspaceId={workspaceId}
      description={description}
      onSuccess={onSuccess}
    />
  )

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {props?.isMenuButton ? null : <Icon variant="LockOpen" />}
      Unlock Terraform state
      {props?.isMenuButton ? <Icon variant="LockOpen" /> : null}
    </Button>
  )
}

export const UnlockTerraformWorkspaceModalContainer = ({
  workspaceId,
  description = 'the workspace',
  onSuccess,
  ...props
}: IUnlockTerraformWorkspace & Omit<IModal, 'onSubmit'>) => {
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      unlockTerraformWorkspace({ orgId: org.id, terraformWorkspaceId: workspaceId }),
    onSuccess: () => {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        status: 'ok',
        user,
        props: { orgId: org.id, installId: install.id, workspaceId },
      })
      addToast(
        <Toast heading="Terraform state unlocked" theme="success">
          <Text>The Terraform workspace for {description} has been unlocked.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({
        queryKey: ['terraform-workspace-lock', org?.id, workspaceId],
      })
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      trackEvent({
        event: 'terraform_workspace_state_unlock',
        status: 'error',
        user,
        props: { orgId: org.id, installId: install.id, workspaceId, err: err?.error },
      })
    },
  })

  return (
    <UnlockTerraformWorkspaceModal
      description={description}
      isPending={isLoading}
      error={error}
      onSubmit={() => execute()}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}
