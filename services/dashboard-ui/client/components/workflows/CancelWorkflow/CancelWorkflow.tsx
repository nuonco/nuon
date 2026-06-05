import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface ICancelWorkflowModal extends Omit<IModal, 'onSubmit'> {
  workflowType: string
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const CancelWorkflowModal = ({
  workflowType,
  isPending,
  error,
  onSubmit,
  ...props
}: ICancelWorkflowModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="WarningIcon" size="24" />
          {`Cancel ${workflowType} workflow?`}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Canceling workflow
          </span>
        ) : (
          'Cancel workflow'
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error.error ||
              'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Once canceled, this {workflowType} workflow cannot be restarted. You
          will have to trigger a new workflow run.
        </Text>
      </div>
    </Modal>
  )
}
