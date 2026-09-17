import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IReprovisionStackModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
  onClose: () => void
}

export const ReprovisionStackModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IReprovisionStackModal) => {
  return (
    <Modal
      heading="Reprovision stack?"
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Reprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="StackPlusIcon" />
            Reprovision stack
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
            {(error as any)?.error || 'Unable to reprovision stack.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Reprovisioning will recreate the stack and runner for {installName}. Components will not be redeployed.
        </Text>

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> Actions and deployments won't be available
            while the runner is recreated.
          </Text>
        </Banner>
      </div>
    </Modal>
  )
}
