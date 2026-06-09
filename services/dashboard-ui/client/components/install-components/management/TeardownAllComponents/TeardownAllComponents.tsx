import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ITeardownAllComponentsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  isKickedOff: boolean
  error?: { error?: string } | null
  onSubmit: () => void
}

export const TeardownAllComponentsModal = ({
  installName,
  isPending,
  isKickedOff,
  error,
  onSubmit,
  ...props
}: ITeardownAllComponentsModal) => {
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
          <Icon variant="CloudArrowDownIcon" size="24" />
          Teardown all {installName} components?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting teardown
          </span>
        ) : (
          'Teardown all components'
        ),
        disabled: isKickedOff || isPending,
        onClick: onSubmit,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to teardown components'}
          </Banner>
        ) : null}
        <Text variant="base">
          This will remove all running component deployments from your install.
        </Text>
      </div>
    </Modal>
  )
}
