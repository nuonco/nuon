import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteConnection = ({
  name,
  error,
  isPending,
  onDelete,
  ...props
}: {
  name: string
  error?: TAPIError | null
  isPending: boolean
  onDelete: () => void
} & IModal) => (
  <Modal
    {...props}
    heading="Delete cloud connection?"
    primaryActionTrigger={{
      variant: 'danger',
      children: isPending ? 'Deleting connection' : 'Delete connection',
      disabled: isPending,
      onClick: onDelete,
    }}
  >
    <div className="flex flex-col gap-4">
      <Text>
        Deleting {name} removes this connection from Nuon. It does not delete
        the IAM role or any resources in AWS.
      </Text>
      <FormErrorBanner
        error={
          error
            ? {
                ...error,
                error: 'Cannot delete connection',
                description: error.description || error.error,
              }
            : error
        }
        fallback="Cannot delete connection"
      />
    </div>
  </Modal>
)
