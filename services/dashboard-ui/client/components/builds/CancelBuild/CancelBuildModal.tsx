import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

export const CancelBuildModal = ({
  componentName,
  onConfirm,
  ...props
}: {
  componentName?: string
  onConfirm: () => void
} & IModal) => {
  const { removeModal } = useSurfaces()

  return (
    <Modal
      heading="Cancel build?"
      primaryActionTrigger={{
        children: 'Cancel build',
        variant: 'danger',
        onClick: () => {
          onConfirm()
          removeModal(props.modalId)
        },
      }}
      secondaryActionTrigger={{
        children: 'Keep building',
        onClick: () => removeModal(props.modalId),
      }}
      {...props}
    >
      <Text variant="body" theme="neutral">
        This stops the running {componentName ? `${componentName} ` : ''}build.
        You can start a new build afterward.
      </Text>
    </Modal>
  )
}
