import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeployAllComponentsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  isKickedOff: boolean
  error?: { error?: string } | null
  onSubmit: () => void
}

export const DeployAllComponentsModal = ({
  installName,
  isPending,
  isKickedOff,
  error,
  onSubmit,
  ...props
}: IDeployAllComponentsModal) => {
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
          <Icon variant="CloudArrowUpIcon" size="24" />
          Deploy all components to {installName}?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting deployments
          </span>
        ) : (
          'Deploy components'
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
        <Text variant="base" weight="stronger">
          Are you sure you want to deploy all components?
        </Text>
        <Text variant="base">
          This action will deploy the latest build of each component to your
          install.
        </Text>
      </div>
    </Modal>
  )
}
