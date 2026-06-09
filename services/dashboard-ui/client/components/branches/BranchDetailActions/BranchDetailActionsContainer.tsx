import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useBranch } from '@/hooks/use-branch'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import { triggerBranchRun } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditDeploymentPlanButton } from '@/components/branches/DeploymentPlanEditor'
import { BranchDetailActions } from './BranchDetailActions'

interface IBranchDetailActionsContainer {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
}

export const BranchDetailActionsContainer = ({
  branch,
  currentConfig,
  appId,
  orgId,
}: IBranchDetailActionsContainer) => {
  const { addToast } = useToast()
  const { refresh } = useBranch()

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
      hasConfig={!!currentConfig}
      isTriggerPending={triggerRunMutation.isPending}
      onTriggerRun={handleTriggerRun}
      onTriggerPreview={handleTriggerPreview}
    />
  )
}
