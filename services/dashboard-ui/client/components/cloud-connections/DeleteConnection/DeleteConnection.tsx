import { Banner } from '@/components/common/Banner'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteConnection = ({
  name,
  imageConfigs = 0,
  error,
  isPending,
  onDelete,
  ...props
}: {
  name: string
  imageConfigs?: number
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
        the cloud identity or any cloud resources.
      </Text>
      {imageConfigs > 0 && (
        <Banner theme="warn">
          This connection is used by {imageConfigs} saved image{' '}
          {imageConfigs === 1 ? 'configuration' : 'configurations'}. Deleting it
          will cause future builds and metadata fetches using{' '}
          {imageConfigs === 1 ? 'that configuration' : 'those configurations'}{' '}
          to fail. To restore access, verify a replacement connection and
          re-sync the affected apps.
        </Banner>
      )}
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
