import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IModal } from '@/components/surfaces/Modal'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { triggerBranchRun } from '@/lib'
import type { TAPIError, TAppBranch, TAppBranchConfig } from '@/types'
import { TriggerBranchRunModal } from './TriggerBranchRunModal'

interface ITriggerBranchRunModalContainer extends Omit<IModal, 'onSubmit'> {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  appId: string
  orgId: string
  planOnly: boolean
  onSuccess?: () => void
}

export const TriggerBranchRunModalContainer = ({
  branch,
  currentConfig,
  appId,
  orgId,
  planOnly,
  onSuccess,
  ...props
}: ITriggerBranchRunModalContainer) => {
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
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
    onSuccess: () => {
      addToast(
        <Toast theme="success" heading={planOnly ? 'Preview run triggered' : 'Run triggered'}>
          <Text>
            {planOnly
              ? 'A plan-only preview run has been queued.'
              : 'Your app branch run has been queued.'}
          </Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast theme="error" heading="Branch run failed">
          <Text>{error.error || 'Unable to trigger run.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <TriggerBranchRunModal
      branchName={branch.name || ''}
      planOnly={planOnly}
      isPending={isPending}
      onConfirm={() => mutate()}
      {...props}
    />
  )
}
