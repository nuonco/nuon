import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useBranch } from '@/hooks/use-branch'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useOrg } from '@/hooks/use-org'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import { deleteAppBranch } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { TriggerBranchRunModal } from '@/components/branches/TriggerBranchRunModal'
import { BranchDetailActions } from './BranchDetailActions'

interface IBranchDetailActionsContainer {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
}

const DeleteBranchModal = ({
  branch,
  appId,
  ...props
}: { branch: TAppBranch; appId: string } & IModal) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const navigate = useNavigate()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      deleteAppBranch({
        appId,
        branchId: branch.id!,
        orgId: org!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-branches'] })
      addToast(
        <Toast heading="Branch deleted" theme="success">
          <Text>Branch "{branch.name}" has been deleted.</Text>
        </Toast>,
      )
      removeModal(props.modalId)
      navigate(`/${org!.id}/apps/${appId}/branches`)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Branch deletion failed" theme="error">
          <Text>{err?.description || err?.error || 'Try again.'}</Text>
        </Toast>,
      )
    },
  })

  return (
    <Modal
      heading="Delete branch?"
      primaryActionTrigger={{
        children: isPending ? 'Deleting...' : 'Delete',
        disabled: isPending,
        onClick: () => mutate(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>
        This will permanently delete the branch "{branch.name}" and all its configs and runs.
      </Text>
    </Modal>
  )
}

export const BranchDetailActionsContainer = ({
  branch,
  currentConfig,
  appId,
  orgId,
}: IBranchDetailActionsContainer) => {
  const { refresh } = useBranch()
  const { addModal } = useSurfaces()

  const openTriggerModal = (planOnly: boolean) => {
    addModal(
      <TriggerBranchRunModal
        branch={branch}
        currentConfig={currentConfig}
        appId={appId}
        orgId={orgId}
        planOnly={planOnly}
        onSuccess={refresh}
      />
    )
  }

  return (
    <BranchDetailActions
      editButton={
        <EditBranchButton branch={branch} currentConfig={currentConfig} onSuccess={refresh} />
      }
      manageInstallsButton={
        <EditDeploymentPlanButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
        />
      }
      deleteButton={
        <Button
          isMenuButton
          variant="danger"
          onClick={() => {
            const modal = <DeleteBranchModal branch={branch} appId={appId} />
            addModal(modal)
          }}
        >
          Delete branch
          <Icon variant="TrashIcon" size={16} />
        </Button>
      }
      hasConfig={!!currentConfig}
      isTriggerPending={false}
      onTriggerRun={() => openTriggerModal(false)}
      onTriggerPreview={() => openTriggerModal(true)}
    />
  )
}
