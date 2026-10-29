import { useState } from 'react'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

export const RerunPreviewModal = ({
  prNumber,
  onConfirm,
  ...props
}: {
  prNumber: number
  onConfirm: () => Promise<unknown>
} & IModal) => {
  const { removeModal } = useSurfaces()
  const [pending, setPending] = useState(false)

  return (
    <Modal
      heading="Re-run preview?"
      size="sm"
      primaryActionTrigger={{
        children: pending ? 'Starting...' : 'Re-run preview',
        disabled: pending,
        onClick: () => {
          setPending(true)
          void onConfirm()
            .then(() => removeModal(props.modalId))
            .catch(() => setPending(false))
        },
      }}
      {...props}
    >
      <Text variant="body" theme="neutral">
        This starts another preview run for #{prNumber} with the current
        preview config.
      </Text>
    </Modal>
  )
}
