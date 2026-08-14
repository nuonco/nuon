import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

export const RevokeTriggerSecretModal = ({
  onConfirm,
  ...props
}: {
  onConfirm: () => void
} & IModal) => {
  const { removeModal } = useSurfaces()

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="WarningIcon" size="24" />
          Revoke secret?
        </Text>
      }
      primaryActionTrigger={{
        children: 'Revoke secret',
        variant: 'danger',
        onClick: () => {
          onConfirm()
          removeModal(props.modalId)
        },
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Banner theme="warn">
          Any integration signing requests with this secret will stop
          authenticating immediately.
        </Banner>
        <Text variant="body" theme="neutral">
          Revoking cannot be undone. Rotate to a new secret and update every
          source that calls this trigger.
        </Text>
      </div>
    </Modal>
  )
}

export const RevokeTriggerSecretButton = ({
  onConfirm,
  ...props
}: {
  onConfirm: () => void
} & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="danger"
      onClick={() => addModal(<RevokeTriggerSecretModal onConfirm={onConfirm} />)}
      {...props}
    >
      Revoke
    </Button>
  )
}
