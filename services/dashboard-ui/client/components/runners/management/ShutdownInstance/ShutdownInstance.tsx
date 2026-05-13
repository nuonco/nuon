import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IShutdownInstanceModal extends IModal {
  error: any
  isLoading: boolean
  onConfirm: () => void
}

export const ShutdownInstanceModal = ({
  error,
  isLoading,
  onConfirm,
  ...props
}: IShutdownInstanceModal) => {
  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="warn"
          >
            <Icon variant="CloudArrowDownIcon" size="24" />
            Restart runner instance?
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Restarting
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDownIcon" />
            Restart instance
          </span>
        ),
        disabled: isLoading,
        onClick: onConfirm,
        variant: 'primary' as const,
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to restart runner instance.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Restart this runner instance.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            The runner VM will be restarted.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
