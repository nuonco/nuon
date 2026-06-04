import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IEnableAutoApproveModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  isApproveAll: boolean
  isSuccess: boolean
  onSubmit: () => void
}

export const EnableAutoApproveModal = ({
  isPending,
  error,
  isApproveAll,
  isSuccess,
  onSubmit,
  ...props
}: IEnableAutoApproveModal) => {
  const buttonText = isApproveAll ? 'Disable auto approval' : 'Enable auto approval'
  const confirmText = isApproveAll
    ? 'Are you sure you want to disable auto approve for changes to this install?'
    : 'Are you sure you want to enable auto approve for changes to this install?'

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="info"
        >
          <Icon variant={isApproveAll ? "ToggleRightIcon" : "ToggleLeftIcon"} size="24" />
          {buttonText}?
        </Text>
      }
      primaryActionTrigger={{
        children: isSuccess ? (
          <span className="flex items-center gap-2">
            <Icon variant="CheckCircleIcon" /> Settings updated
          </span>
        ) : isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Updating settings...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant={isApproveAll ? "ToggleRightIcon" : "ToggleLeftIcon"} />
            {buttonText}
          </span>
        ),
        onClick: onSubmit,
        disabled: isPending || isSuccess,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || `Unable to ${isApproveAll ? 'disable' : 'enable'} auto approval`}
          </Banner>
        ) : null}

        <Text variant="body">
          {confirmText}
        </Text>

        {!isApproveAll && (
          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> When auto approve is enabled, all changes to this install will be automatically approved and applied without manual review.
            </Text>
          </Banner>
        )}
      </div>
    </Modal>
  )
}
