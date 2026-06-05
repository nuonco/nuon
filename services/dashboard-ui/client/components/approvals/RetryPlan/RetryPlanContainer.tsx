import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { retryWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { RETRY_MODAL_COPY } from '@/utils/approval-utils'
import { RetryPlanModal } from './RetryPlan'

interface IRetryPlan {
  step: TWorkflowStep
}

export const RetryPlanModalContainer = ({
  step,
  ...props
}: IRetryPlan & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const modalCopy = RETRY_MODAL_COPY[step.approval.type]

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation({
    mutationFn: () =>
      retryWorkflowStep({
        orgId: org.id,
        workflowId: step.install_workflow_id,
        stepId: step.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Plan retry initiated" theme="success">
          <Text>
            A new plan is being generated. Review the updated changes when
            ready.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Retry failed" theme="error">
          <Text>{err?.error || 'Unable to retry these changes.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <RetryPlanModal
      modalCopy={modalCopy}
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}

export const RetryPlanButton = ({
  step,
  ...props
}: IRetryPlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RetryPlanModalContainer step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Retry plan
    </Button>
  )
}
