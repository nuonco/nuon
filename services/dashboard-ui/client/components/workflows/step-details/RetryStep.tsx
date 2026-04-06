import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { retryWorkflowStep } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface IRetryStep {
  step: TWorkflowStep
}

export const RetryStepModal = ({ step, ...props }: IRetryStep & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending: isLoading, error } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      retryWorkflowStep({
        body: { operation: 'retry-step', step_id: step.id },
        orgId: org.id,
        workflowId: step.install_workflow_id,
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
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          Retry step?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Retrying step
          </span>
        ) : (
          'Retry step'
        ),
        onClick: () => execute(),
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to retry this step?
        </Text>
        <Text variant="base">
          Retrying will rerun this workflow step. If successful, the workflow will continue from
          this point.
        </Text>
      </div>
    </Modal>
  )
}

export const RetryStepButton = ({ step, ...props }: IRetryStep & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RetryStepModal step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Retry step
    </Button>
  )
}
