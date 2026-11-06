import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const RemoveImageRepository = ({
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
    heading="Remove image repository?"
    primaryActionTrigger={{
      variant: 'danger',
      children: isPending ? 'Removing repository' : 'Remove repository',
      disabled: isPending,
      onClick: onDelete,
    }}
  >
    <div className="flex flex-col gap-4">
      <Text>
        Removing <Text family="mono">{name}</Text> deletes its saved pull
        identity selection. It does not delete images, the connection, or cloud
        permissions.
      </Text>
      <Banner theme="warn">
        <Text weight="strong" as="span">
          Warning:
        </Text>{' '}
        New config saves and app syncs for this repository will fail until it is
        registered again. Existing configurations keep their pinned pull
        identity. To change the pull identity, add the repository again with
        the replacement and re-sync the app.
      </Banner>
      <FormErrorBanner error={error} fallback="Repository removal failed" />
    </div>
  </Modal>
)
