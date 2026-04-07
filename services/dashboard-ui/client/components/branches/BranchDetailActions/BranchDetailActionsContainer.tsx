import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useBranch } from '@/hooks/use-branch'
import type { TAppBranch, TAppBranchConfig } from '@/types'
import { triggerBranchRun } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditInstallGroupsButton } from '@/components/branches/EditInstallGroupsModal'
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
    mutationFn: () =>
      triggerBranchRun({
        appId,
        branchId: branch.id!,
        orgId,
        request: {
          config_id: currentConfig?.id,
          force: false,
        },
      }),
    onSuccess: () => {
      addToast(
        <Toast theme="success" heading="Run triggered successfully">
          <Text>Your app branch run has been queued.</Text>
        </Toast>
      )
      refresh()
    },
    onError: (error: any) => {
      const errorMessage =
        typeof error === 'string'
          ? error
          : error.user_error ||
            error.error ||
            error.description ||
            'Failed to trigger run'
      addToast(
        <Toast theme="error" heading="Failed to trigger run">
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

    triggerRunMutation.mutate()
  }

  return (
    <BranchDetailActions
      editButton={
        <EditBranchButton branch={branch} currentConfig={currentConfig} onSuccess={refresh} />
      }
      manageInstallsButton={
        <EditInstallGroupsButton branch={branch} currentConfig={currentConfig} onSuccess={refresh} />
      }
      hasConfig={!!currentConfig}
      isTriggerPending={triggerRunMutation.isPending}
      onTriggerRun={handleTriggerRun}
    />
  )
}
