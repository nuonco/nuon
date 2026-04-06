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
import { approveWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { RETRY_MODAL_COPY } from '@/utils/approval-utils'

interface IRetryPlan {
  step: TWorkflowStep
}

export const RetryPlanModal = ({ step, ...props }: IRetryPlan & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const modalCopy = RETRY_MODAL_COPY[step.approval.type]

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: () =>
      approveWorkflowStep({
        body: { note: 'Retry plan', response_type: 'retry' },
        orgId: org.id,
        workflowId: step.install_workflow_id,
        workflowStepId: step.id,
        approvalId: step?.approval?.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Plan retry initiated" theme="success">
          <Text>
            A new plan is being generated. Please review the updated changes when
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
        <Toast heading="Failed to retry changes" theme="error">
          <Text>There was an error while retrying these changes.</Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="stronger"
        >
          {modalCopy.title}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Retrying plan
          </span>
        ) : (
          'Retry plan'
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
            {(error as any)?.error ||
              'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          {modalCopy.heading}
        </Text>
        <Text variant="base">{modalCopy.message}</Text>
      </div>
    </Modal>
  )
}

export const RetryPlanButton = ({
  step,
  ...props
}: IRetryPlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RetryPlanModal step={step} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Retry plan
    </Button>
  )
}
