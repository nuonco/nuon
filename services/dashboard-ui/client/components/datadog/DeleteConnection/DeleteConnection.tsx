import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TDatadogConnection } from '@/types'

export const DeleteConnectionModal = ({
  connection,
  isPending,
  onConfirm,
  ...props
}: {
  connection: TDatadogConnection
  isPending: boolean
  onConfirm: () => void
} & Omit<IModal, 'onSubmit'>) => (
  <Modal
    heading="Delete Datadog connection"
    primaryActionTrigger={{
      children: isPending ? 'Deleting…' : 'Delete connection',
      disabled: isPending,
      onClick: onConfirm,
      variant: 'danger',
    }}
    {...props}
  >
    <div className="flex flex-col gap-2">
      <Text>
        This deletes <strong>{connection.name}</strong> and stops fanning
        Nuon events out to its Datadog tenant. Existing managed monitors
        belonging to this connection will also be removed in Datadog.
      </Text>
      <Text variant="subtext" theme="neutral">
        The DD API key isn't revoked on the Datadog side — rotate or delete
        it there if you no longer want Nuon to be able to write events.
      </Text>
    </div>
  </Modal>
)
