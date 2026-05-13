import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ICancelWorkflowsModal extends Omit<IModal, 'onSubmit'> {
  count: number
  isPending: boolean
  error?: string | null
  cancelResults?: {
    cancelled: string[]
    errors?: { workflow_id: string; error: string }[]
  } | null
  onSubmit: () => void
}

export const CancelWorkflowsModal = ({
  count,
  isPending,
  error,
  cancelResults,
  onSubmit,
  ...props
}: ICancelWorkflowsModal) => {
  const hasPartialErrors = cancelResults?.errors && cancelResults.errors.length > 0

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
          {`Cancel ${count} workflow${count === 1 ? '' : 's'}?`}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Canceling workflows
          </span>
        ) : (
          `Cancel ${count} workflow${count === 1 ? '' : 's'}`
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-2">
        {error && (
          <Banner theme="error">
            {error || 'An error occurred. Please refresh the page and try again.'}
          </Banner>
        )}
        {hasPartialErrors && (
          <Banner theme="error">
            {cancelResults.errors!.length} workflow{cancelResults.errors!.length === 1 ? '' : 's'} failed to cancel:
            <ul className="mt-1 list-disc pl-4">
              {cancelResults.errors!.map((e) => (
                <li key={e.workflow_id}>
                  {e.workflow_id}: {e.error}
                </li>
              ))}
            </ul>
          </Banner>
        )}
        <Text variant="base" weight="strong">
          Are you sure you want to cancel {count} workflow{count === 1 ? '' : 's'}?
        </Text>
        <Text variant="base">
          Once a workflow is canceled you cannot restart it. You will have to
          trigger new workflow runs.
        </Text>
      </div>
    </Modal>
  )
}
