import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveWorkflowStep } from '@/lib'
import type { TAPIError } from '@/types'

type GroupAction = 'approve' | 'deny-skip-current'

interface IGroupActionTarget {
  orgId: string
  workflowId: string
  stepId: string
  approvalId: string
  groupName: string
}

const ACTION_COPY: Record<
  GroupAction,
  { title: string; heading: string; message: (groupName: string) => string; confirm: string; pending: string; variant: 'primary' | 'danger' }
> = {
  approve: {
    title: 'Approve install group plan',
    heading: 'Approve and deploy?',
    message: (groupName) =>
      `This approves the plan for ${groupName} and proceeds with deploying to the install group. This may take a few minutes.`,
    confirm: 'Approve plan',
    pending: 'Approving plan',
    variant: 'primary',
  },
  'deny-skip-current': {
    title: 'Skip install group',
    heading: 'Skip this install group?',
    message: (groupName) =>
      `This skips deploying to ${groupName} and continues the workflow without applying these changes.`,
    confirm: 'Skip install group',
    pending: 'Skipping',
    variant: 'danger',
  },
}

interface IConfirmGroupActionModal extends IModal {
  action: GroupAction
  target: IGroupActionTarget
}

const ConfirmGroupActionModal = ({ action, target, ...props }: IConfirmGroupActionModal) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const copy = ACTION_COPY[action]

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: () =>
      approveWorkflowStep({
        orgId: target.orgId,
        workflowId: target.workflowId,
        workflowStepId: target.stepId,
        approvalId: target.approvalId,
        body: { response_type: action, note: '' },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading={action === 'approve' ? 'Plan approved' : 'Install group skipped'} theme="success">
          <Text>
            {action === 'approve'
              ? `Deploying ${target.groupName}. This may take a few minutes.`
              : `Skipped ${target.groupName}.`}
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['branch-run'] })
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading={action === 'approve' ? 'Approval failed' : 'Skip failed'} theme="error">
          <Text>{err?.error || 'Unable to respond to this plan.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          {copy.title}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> {copy.pending}
          </span>
        ) : (
          copy.confirm
        ),
        onClick: () => execute(),
        disabled: isPending,
        variant: copy.variant,
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          {copy.heading}
        </Text>
        <Text variant="base">{copy.message(target.groupName)}</Text>
      </div>
    </Modal>
  )
}

const ACTION_LABEL: Record<GroupAction, { label: string; variant: 'primary' | 'danger' }> = {
  approve: { label: 'Approve', variant: 'primary' },
  'deny-skip-current': { label: 'Skip', variant: 'danger' },
}

export const GroupActionButton = ({
  action,
  target,
}: {
  action: GroupAction
  target: IGroupActionTarget
}) => {
  const { addModal } = useSurfaces()
  const { label, variant } = ACTION_LABEL[action]

  return (
    <Button variant={variant} onClick={() => addModal(<ConfirmGroupActionModal action={action} target={target} />)}>
      {label}
    </Button>
  )
}

interface IGroupApprovalActions {
  target: IGroupActionTarget
}

export const GroupApprovalActions = ({ target }: IGroupApprovalActions) => (
  <>
    <GroupActionButton action="deny-skip-current" target={target} />
    <GroupActionButton action="approve" target={target} />
  </>
)
