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
import type { TAPIError, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { RetryStepModal } from './RetryStep'

interface IRetryStep {
  step: TWorkflowStep
}

export const RetryStepModalContainer = ({
  step,
  ...props
}: IRetryStep & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      retryWorkflowStep({
        orgId: org.id,
        workflowId: step.install_workflow_id,
        stepId: step.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Step retry initiated" theme="success">
          <Text>{toSentenceCase(step.name)} is being retried.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    onError: (err) => {
      addToast(
        <Toast heading="Failed to retry step" theme="error">
          <Text>There was an error while retrying this step.</Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <RetryStepModal
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}

export const RetryStepButton = ({
  step,
  ...props
}: IRetryStep & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RetryStepModalContainer step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Retry step
    </Button>
  )
}
