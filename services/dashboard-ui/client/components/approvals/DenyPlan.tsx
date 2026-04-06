import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { SplitButton, type ISplitButton } from '@/components/common/SplitButton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { addRespondedStep } from '@/hooks/use-responded-approvals'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveWorkflowStep } from '@/lib'
import type { TApproveWorkflowStepBody } from '@/lib/ctl-api/workflows/approve-workflow-step'
import type { TWorkflowStep } from '@/types'
import { DENY_MODAL_COPY } from '@/utils/approval-utils'

type TDenyType = Exclude<
  TApproveWorkflowStepBody['response_type'],
  'approve' | 'retry'
>

interface IDenyPlan {
  step: TWorkflowStep
}

export const DenyPlanModal = ({
  denyType,
  step,
  ...props
}: IDenyPlan & {
  denyType: TDenyType
} & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()
  const queryClient = useQueryClient()

  const modalCopy = DENY_MODAL_COPY[step.approval.type]

  const { mutate: execute, isPending: isLoading, error } = useMutation({
    mutationFn: (params: { body: TApproveWorkflowStepBody; workflowId: string; workflowStepId: string; approvalId: string }) =>
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
            <Icon variant="Loading" /> Denying plan
          </span>
        ) : (
          'Deny plan'
        ),
        onClick: () => {
          execute({
            body: { note: 'Deny plan', response_type: denyType },
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

export const DenyPlanButton = ({
  step,
  ...props
}: IDenyPlan & Omit<ISplitButton, 'buttonProps' | 'dropdownProps'>) => {
  const { addModal } = useSurfaces()

  const openModal = (denyType: TDenyType) => {
    addModal(<DenyPlanModal step={step} denyType={denyType} />)
  }

  return (
    <SplitButton
      buttonProps={{
        children: 'Deny plan',
        onClick: () => {
          openModal('deny')
        },
      }}
      dropdownProps={{
        children: (
          <Menu>
            <Button
              className="!text-foreground"
              onClick={() => {
                openModal('deny-skip-current')
              }}
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
