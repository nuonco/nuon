import { Text } from '../../atoms/Text'
import { Time } from '../../molecules/Time'
import { Modal } from '../surfaces/Modal'

export interface IResumeDraftModal {
  updatedAt?: string
  onResume: () => void
  onStartFresh: () => void
}

export const ResumeDraftModal = ({
  updatedAt,
  onResume,
  onStartFresh,
}: IResumeDraftModal) => (
  <Modal
    heading="Resume draft"
    dismissible={false}
    secondaryAction={{
      children: 'Start fresh',
      variant: 'secondary',
      onClick: onStartFresh,
    }}
    primaryAction={{
      children: 'Resume draft',
      variant: 'primary',
      onClick: onResume,
    }}
  >
    <Text color="secondary">
      You have unsaved changes from{' '}
      <Time value={updatedAt} format="relative" variant="body" color="secondary" />
      . Resume your draft or start fresh?
    </Text>
  </Modal>
)
