import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IShutdownMngRunnerModal extends IModal {
  label: string
  error: any
  isLoading: boolean
  onConfirm: () => void
}

export const ShutdownMngRunnerModal = ({
  label,
  error,
  isLoading,
  onConfirm,
  ...props
}: IShutdownMngRunnerModal) => {
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
            <Icon variant="Power" size="24" />
            {label}?
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Shutting down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Power" />
            {label}
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
            {error?.error || 'Unable to shutdown managed runner process.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this managed runner instance.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            This will destroy the managed runner instance. A new instance will be
            provisioned automatically to replace it.
          </Text>
          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>The VM instance will be terminated</li>
            <li>A new instance will be provisioned with the latest version</li>
            <li>All local state will be lost</li>
          </ul>
        </div>
      </div>
    </Modal>
  )
}
