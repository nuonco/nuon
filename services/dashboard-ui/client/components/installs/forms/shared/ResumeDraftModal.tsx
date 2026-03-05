import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'

interface IResumeDraftModal extends IModal {
  draftTimestamp: string
  onResume: () => void
  onStartFresh: () => void
}

function formatRelativeTime(timestamp: string): string {
  const now = Date.now()
  const then = new Date(timestamp).getTime()
  const diffMs = now - then
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
}

export const ResumeDraftModal = ({
  draftTimestamp,
  onResume,
  onStartFresh,
  ...props
}: IResumeDraftModal) => {
  return (
    <Modal
      heading="Resume draft?"
      primaryActionTrigger={{
        children: 'Resume draft',
        onClick: onResume,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Start fresh',
        onClick: onStartFresh,
      }}
      {...props}
    >
      <Text>
        You have unsaved changes from {formatRelativeTime(draftTimestamp)}.
        Would you like to resume your draft or start fresh?
      </Text>
    </Modal>
  )
}
