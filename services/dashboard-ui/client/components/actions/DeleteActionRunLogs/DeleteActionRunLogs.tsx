import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteActionRunLogsModal = ({
  actionName,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  actionName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & Omit<IModal, 'onSubmit'>) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="TrashIcon" size="24" />
        Delete run logs?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Deleting logs
        </span>
      ) : (
        'Delete logs'
      ),
      disabled: isPending,
      onClick: () => onSubmit(),
      variant: 'danger',
    }}
    {...props}
  >
    <div className="flex flex-col gap-6">
      {error ? (
        <Banner theme="error">
          {error?.description || error?.error || 'Unable to delete run logs'}
        </Banner>
      ) : null}

      <Text variant="body" theme="neutral">
        All logs and trace spans from this {actionName} run will be permanently
        deleted. The run and its outputs are kept.
      </Text>

      <Banner theme="warn">
        <Text variant="body">
          <strong>Warning:</strong> This cannot be undone. If these logs contain
          a secret, rotate it as well.
        </Text>
      </Banner>
    </div>
  </Modal>
)

export const ActionRunLogsDeleted = () => (
  <EmptyState
    variant="history"
    emptyTitle="Logs deleted"
    emptyMessage="An org admin deleted the logs and trace spans for this run."
  />
)

export const DeleteActionRunLogsButton = ({
  onOpenModal,
  ...props
}: { onOpenModal: () => void } & Omit<IButtonAsButton, 'children'>) => (
  <Button variant="danger" onClick={onOpenModal} {...props}>
    Delete logs
    <Icon variant="TrashIcon" />
  </Button>
)
