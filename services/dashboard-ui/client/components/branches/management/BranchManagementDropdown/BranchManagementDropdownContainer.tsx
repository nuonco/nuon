import { useMemo } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useBranch } from '@/hooks/use-branch'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { BranchProvider } from '@/providers/branch-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TAppBranch } from '@/types'
import { deleteAppBranch } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { TriggerBranchRunModal } from '@/components/branches/TriggerBranchRunModal'
import { BranchManagementDropdown } from './BranchManagementDropdown'

const DeleteBranchModal = ({
  branch,
  appId,
  orgId,
  ...props
}: { branch: TAppBranch; appId: string; orgId: string } & IModal) => {
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const navigate = useNavigate()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      deleteAppBranch({ appId, branchId: branch.id!, orgId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-branches'] })
      addToast(
        <Toast heading="Branch deleted" theme="success">
          <Text>Branch "{branch.name}" has been deleted.</Text>
        </Toast>,
      )
      removeModal(props.modalId)
      navigate(`/${orgId}/apps/${appId}/branches`)
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

interface IBranchManagementMenu {
  appId: string
  orgId: string
}

const BranchManagementMenu = ({ appId, orgId }: IBranchManagementMenu) => {
  const { branch, refresh } = useBranch()
  const { addModal } = useSurfaces()

  const currentConfig = useMemo(() => {
    if (!branch.configs?.length) return undefined
    return [...branch.configs].sort(
      (a, b) => (b.config_number || 0) - (a.config_number || 0)
    )[0]
  }, [branch.configs])

  const handleTriggerRun = () => {
    addModal(
      <TriggerBranchRunModal
        branch={branch}
        currentConfig={currentConfig}
        appId={appId}
        orgId={orgId}
        planOnly={false}
        onSuccess={refresh}
      />
    )
  }

  const handleDelete = () => {
    const modal = <DeleteBranchModal branch={branch} appId={appId} orgId={orgId} />
    addModal(modal)
  }

  return (
    <BranchManagementDropdown
      dropdownId={branch.id!}
      detailHref={`/${orgId}/apps/${appId}/branches/${branch.id}`}
      editButton={
        <EditBranchButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
          isMenuButton
        />
      }
      deploymentPlanButton={
        <EditDeploymentPlanButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
          isMenuButton
        />
      }
      deleteButton={
        <Button isMenuButton variant="danger" onClick={handleDelete}>
          Delete branch
          <Icon variant="TrashIcon" size={16} />
        </Button>
      }
      hasConfig={!!currentConfig}
      isTriggerPending={false}
      onTriggerRun={handleTriggerRun}
    />
  )
}

export const BranchManagementDropdownContainer = ({
  branch,
  appId,
  orgId,
}: {
  branch: TAppBranch
  appId: string
  orgId: string
}) => {
  return (
    <BranchProvider
      branchId={branch.id!}
      shouldPoll={false}
      loadingElement={<Skeleton height="24px" width="24px" />}
      errorElement={null}
    >
      <SurfacesProvider>
        <BranchManagementMenu appId={appId} orgId={orgId} />
      </SurfacesProvider>
    </BranchProvider>
  )
}
