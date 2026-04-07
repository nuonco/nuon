import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { addRespondedStep } from '@/hooks/use-responded-approvals'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { APPROVAL_MODAL_COPY } from '@/utils/approval-utils'
import { ApprovePlanModal } from './ApprovePlan'

interface IApprovePlan {
  step: TWorkflowStep
}

export const ApprovePlanModalContainer = ({
  step,
  ...props
}: IApprovePlan & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const modalCopy = APPROVAL_MODAL_COPY[step.approval.type]

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation({
    mutationFn: (params: {
      body: Parameters<typeof approveWorkflowStep>[0]['body']
      workflowId: string
      workflowStepId: string
      approvalId: string
    }) =>
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
          <Text>
            The plan has been approved and the changes are being applied.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      addRespondedStep(step.id)
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
    <ApprovePlanModal
      modalCopy={modalCopy}
      isPending={isPending}
      error={error}
      onSubmit={() =>
        execute({
          body: { note: 'Approved plan', response_type: 'approve' },
          workflowId: step.install_workflow_id,
          workflowStepId: step.id,
          approvalId: step?.approval?.id,
        })
      }
      {...props}
    />
  )
}

export const ApprovePlanButton = ({
  step,
  ...props
}: IApprovePlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ApprovePlanModalContainer step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Approve plan
    </Button>
  )
}
