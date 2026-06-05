import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDriftScanAllComponentsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  isKickedOff: boolean
  error?: { error?: string } | null
  onSubmit: () => void
}

export const DriftScanAllComponentsModal = ({
  installName,
  isPending,
  isKickedOff,
  error,
  onSubmit,
  ...props
}: IDriftScanAllComponentsModal) => {
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
          <Icon variant="ScanIcon" size="24" />
          Drift scan all {installName} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting drift scan
          </span>
        ) : (
          'Drift scan all components'
        ),
        disabled: isKickedOff || isPending,
        onClick: onSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to deploy components'}
          </Banner>
        ) : null}
        <Text variant="base">
          This will perform a drift scan against the latest build of each
          component and the component deployments on your install.
        </Text>
      </div>
    </Modal>
  )
}
