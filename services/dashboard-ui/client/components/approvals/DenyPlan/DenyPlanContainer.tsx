import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Menu } from '@/components/common/Menu'
import { SplitButton, type ISplitButton } from '@/components/common/SplitButton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { addRespondedStep } from '@/hooks/use-responded-approvals'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveWorkflowStep } from '@/lib'
import type { TApproveWorkflowStepBody } from '@/lib/ctl-api/workflows/approve-workflow-step'
import type { TWorkflowStep } from '@/types'
import { DENY_MODAL_COPY } from '@/utils/approval-utils'
import { DenyPlanModal } from './DenyPlan'

type TDenyType = Exclude<
  TApproveWorkflowStepBody['response_type'],
  'approve' | 'retry'
>

interface IDenyPlan {
  step: TWorkflowStep
}

export const DenyPlanModalContainer = ({
  denyType,
  step,
  ...props
}: IDenyPlan & { denyType: TDenyType } & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const modalCopy = DENY_MODAL_COPY[step.approval.type]

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation({
    mutationFn: (params: {
      body: TApproveWorkflowStepBody
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
        <Toast heading="Plan denied" theme="success">
          <Text>The plan has been denied and will not be applied.</Text>
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
        <Toast heading="Failed to deny changes" theme="error">
          <Text>There was an error while trying deny these changes.</Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DenyPlanModal
      modalCopy={modalCopy}
      isPending={isPending}
      error={error}
      onSubmit={() =>
        execute({
          body: { note: 'Deny plan', response_type: denyType },
          workflowId: step.install_workflow_id,
          workflowStepId: step.id,
          approvalId: step?.approval?.id,
        })
      }
      {...props}
    />
  )
}

export const DenyPlanButton = ({
  step,
  ...props
}: IDenyPlan & Omit<ISplitButton, 'buttonProps' | 'dropdownProps'>) => {
  const { addModal } = useSurfaces()

  if (!step.approval) return null

  const openModal = (denyType: TDenyType) => {
    addModal(<DenyPlanModalContainer step={step} denyType={denyType} />)
  }

  return (
    <SplitButton
      buttonProps={{
        children: 'Deny plan',
        onClick: () => openModal('deny'),
      }}
      dropdownProps={{
        children: (
          <Menu>
            <Button
              className="!text-foreground"
              onClick={() => openModal('deny-skip-current')}
              size={props?.size}
            >
              Deny and continue
            </Button>
            <Button className="!text-foreground" size={props?.size} disabled>
              Deny and skip dependents
            </Button>
          </Menu>
        ),
        id: 'deny-plan-dropdown',
        alignment: 'right',
      }}
      {...props}
    />
  )
}
