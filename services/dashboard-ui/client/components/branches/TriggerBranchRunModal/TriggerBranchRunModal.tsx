import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

export interface ITriggerBranchRunModal extends Omit<IModal, 'onSubmit'> {
  branchName: string
  planOnly: boolean
  isPending: boolean
  onConfirm: () => void
}

export const TriggerBranchRunModal = ({
  branchName,
  planOnly,
  isPending,
  onConfirm,
  ...props
}: ITriggerBranchRunModal) => {
  return (
    <Modal
      heading={planOnly ? 'Trigger preview run?' : 'Trigger run?'}
      primaryActionTrigger={{
        children: isPending
          ? 'Triggering...'
          : planOnly
            ? 'Trigger preview'
            : 'Trigger run',
        disabled: isPending,
        onClick: onConfirm,
        variant: 'primary',
      }}
      {...props}
    >
      <Text>
        {planOnly
          ? `This starts a plan-only preview run for "${branchName}". Nothing will be deployed.`
          : `This starts a new run for "${branchName}" and deploys the current configuration to its install groups. This may take a few minutes.`}
      </Text>
    </Modal>
  )
}
