import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDriftScanSandboxModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  onSubmit: () => void
  onClose: () => void
}

export const DriftScanSandboxModal = ({
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IDriftScanSandboxModal) => {
  return (
    <Modal
      heading="Drift scan sandbox?"
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Scanning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Scan" />
            Drift scan sandbox
          </span>
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'primary' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to drift scan sandbox.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Are you sure you want to drift scan this sandbox?
        </Text>
      </div>
    </Modal>
  )
}
