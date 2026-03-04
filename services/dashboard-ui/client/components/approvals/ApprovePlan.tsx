import { useMutation } from '@tanstack/react-query'
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
import { APPROVAL_MODAL_COPY } from '@/utils/approval-utils'

interface IApprovePlan {
  step: TWorkflowStep
}

export const ApprovePlanModal = ({ step, ...props }: IApprovePlan & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()

  const modalCopy = APPROVAL_MODAL_COPY[step.approval.type]

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof approveWorkflowStep>[0]['body']; workflowId: string; workflowStepId: string; approvalId: string }) =>
      approveWorkflowStep({
        body: params.body,
        orgId: org.id,
        workflowId: params.workflowId,
        workflowStepId: params.workflowStepId,
        approvalId: params.approvalId,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Plan approved" theme="success">
          <Text>The plan has been approved and the changes are being applied.</Text>
        </Toast>
      )
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Failed to approve changes" theme="error">
          <Text>There was an error while trying approve these changes.</Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="stronger"
        >
          {modalCopy.title}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Approving plan
          </span>
        ) : (
          'Approve plan'
        ),
        onClick: () => {
          execute({
            body: { note: 'Approved plan', response_type: 'approve' },
            workflowId: step.install_workflow_id,
            workflowStepId: step.id,
            approvalId: step?.approval?.id,
          })
        },
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

export const ApprovePlanButton = ({
  step,
  ...props
}: IApprovePlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ApprovePlanModal step={step} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Approve plan
    </Button>
  )
}
