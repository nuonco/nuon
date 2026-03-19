import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { useBranch } from '@/hooks/use-branch'
import type { TAppBranch, TAppBranchConfig } from '@/types'
import { triggerBranchRun } from '@/lib'
import { EditBranchButton } from '@/components/branches/EditBranchNameModal'
import { EditInstallGroupsButton } from '@/components/branches/EditInstallGroupsModal'

interface IBranchDetailActions {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
}

export const BranchDetailActions = ({
  branch,
  currentConfig,
  appId,
  orgId,
}: IBranchDetailActions) => {
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
    <div className="flex items-center gap-3">
      <EditBranchButton branch={branch} currentConfig={currentConfig} onSuccess={refresh} />
      <EditInstallGroupsButton branch={branch} currentConfig={currentConfig} onSuccess={refresh} />

      <Button
        variant="primary"
        disabled={!currentConfig || triggerRunMutation.isPending}
        onClick={handleTriggerRun}
        title={
          !currentConfig
            ? 'Create a configuration first to trigger a run'
            : 'Trigger a new run with the current configuration'
        }
      >
        <Icon variant="Play" size={16} />
        {triggerRunMutation.isPending ? 'Triggering...' : 'Trigger run'}
      </Button>
    </div>
  )
}
