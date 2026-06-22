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
import { triggerBranchRun, deleteAppBranch } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
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
  const { addToast } = useToast()
  const { refresh } = useBranch()
  const { addModal } = useSurfaces()

  const triggerRunMutation = useMutation({
    mutationFn: (planOnly: boolean) =>
      triggerBranchRun({
        appId,
        branchId: branch.id!,
        orgId,
        request: {
          config_id: currentConfig?.id,
          force: false,
          plan_only: planOnly,
        },
      }),
    onSuccess: (_, planOnly) => {
      addToast(
        <Toast theme="success" heading={planOnly ? 'Preview run triggered' : 'Run triggered'}>
          <Text>{planOnly ? 'A plan-only preview run has been queued.' : 'Your app branch run has been queued.'}</Text>
        </Toast>
      )
      refresh()
    },
    onError: (error: TAPIError) => {
      const errorMessage = error.error || 'Unable to trigger run.'
      addToast(
        <Toast theme="error" heading="Branch run failed">
          <Text>{errorMessage}</Text>
        </Toast>
      )
    },
  })

  const handleTriggerRun = () => {
    if (!currentConfig) {
      addToast(
        <Toast theme="error" heading="No configuration available">
          <Text>Create a config first before triggering a run.</Text>
        </Toast>
      )
      return
    }

    triggerRunMutation.mutate(false)
  }

  const handleTriggerPreview = () => {
    if (!currentConfig) {
      addToast(
        <Toast theme="error" heading="No configuration available">
          <Text>Create a config first before triggering a preview.</Text>
        </Toast>
      )
      return
    }

    triggerRunMutation.mutate(true)
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
          className="!text-red-800 dark:!text-red-500"
          onClick={() => {
            const modal = <DeleteBranchModal branch={branch} appId={appId} />
            addModal(modal)
          }}
        >
          <Icon variant="TrashIcon" size={16} />
          Delete branch
        </Button>
      }
      hasConfig={!!currentConfig}
      isTriggerPending={triggerRunMutation.isPending}
      onTriggerRun={handleTriggerRun}
      onTriggerPreview={handleTriggerPreview}
    />
  )
}
