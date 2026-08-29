import { useMemo } from 'react'
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
import type { TAPIError, TAppBranch, TAppBranchConfig, TAppBranchRunPreviewMode } from '@/types'
import { deleteAppBranch, triggerBranchRun } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { TriggerBranchRunModal } from '@/components/branches/TriggerBranchRunModal'
import {
  PreviewBranchRunModalContainer,
  quickPreviewFromDefaults,
} from '@/components/branches/PreviewBranchRunModal'
import { previewDefaultsFromConfig } from '@/components/branches/shared/PreviewDefaultsEditor'
import { resolveInstallName } from '@/components/branches/shared/preview-run-utils'
import { BranchDetailActions, type PreviewQuickAction } from './BranchDetailActions'

interface IBranchDetailActionsContainer {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
  showManage?: boolean
  showTriggerNudge?: boolean
}

export const DeleteBranchModal = ({
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
  showManage,
  showTriggerNudge,
}: IBranchDetailActionsContainer) => {
  const { refresh } = useBranch()
  const { addModal } = useSurfaces()
  const { addToast } = useToast()

  const previewDefaults = useMemo(
    () => previewDefaultsFromConfig(currentConfig?.preview_config),
    [currentConfig?.preview_config]
  )

  const { mutate: triggerQuickPreview, isPending: isQuickPreviewPending } = useMutation({
    mutationFn: (mode: TAppBranchRunPreviewMode) => {
      const quick = quickPreviewFromDefaults(currentConfig?.preview_config, mode)
      return triggerBranchRun({
        appId,
        branchId: branch.id!,
        orgId,
        request: {
          config_id: currentConfig?.id,
          preview_run: {
            source: 'branch',
            git_ref: branch.name,
            mode: quick.mode !== previewDefaults.mode ? quick.mode : undefined,
            install_id: quick.installId || undefined,
          },
        },
      })
    },
    onSuccess: () => {
      addToast(
        <Toast theme="success" heading="Preview run triggered">
          <Text>Your preview run has been queued.</Text>
        </Toast>
      )
      refresh()
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast theme="error" heading="Preview run failed">
          <Text>{error.error || 'Unable to trigger preview run.'}</Text>
        </Toast>
      )
    },
  })

  const previewQuickActions = useMemo((): PreviewQuickAction[] => {
    if (!previewDefaults.installId) return []
    const installLabel = resolveInstallName(
      previewDefaults.installId,
      currentConfig?.preview_config
    )

    return [
      {
        label: `Plan only · ${installLabel}`,
        mode: 'plan-only',
        onClick: () => triggerQuickPreview('plan-only'),
      },
      {
        label: `Apply · ${installLabel}`,
        mode: 'apply',
        onClick: () => triggerQuickPreview('apply'),
      },
    ]
  }, [currentConfig?.preview_config, previewDefaults.installId, triggerQuickPreview])

  const openTriggerModal = () => {
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

  const openPreviewModal = () => {
    addModal(
      <PreviewBranchRunModalContainer
        branch={branch}
        currentConfig={currentConfig}
        appId={appId}
        orgId={orgId}
        onSuccess={refresh}
      />
    )
  }

  return (
    <BranchDetailActions
      editButton={
        <EditBranchButton
          isMenuButton
          branch={branch}
          currentConfig={currentConfig}
          onSuccess={refresh}
        />
      }
      deploymentPlanButton={
        <EditDeploymentPlanButton
          isMenuButton
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
      isTriggerPending={isQuickPreviewPending}
      showManage={showManage}
      showTriggerNudge={showTriggerNudge}
      previewQuickActions={previewQuickActions}
      onTriggerRun={openTriggerModal}
      onTriggerPreviewModal={openPreviewModal}
    />
  )
}
