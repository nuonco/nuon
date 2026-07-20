import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteApiTokenModal = ({
  tokenName,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  tokenName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="WarningIcon" size="24" />
        Delete API token?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Deleting token
        </span>
      ) : (
        'Delete token'
      ),
      disabled: isPending,
      onClick: () => onSubmit(),
      variant: 'danger',
    }}
    {...props}
  >
    <div className="flex flex-col gap-6">
      {error ? (
        <Banner theme="error">{error?.error || 'Unable to delete API token'}</Banner>
      ) : null}

      <Text variant="body" theme="neutral">
        {tokenName} will be deleted and can no longer be used to access the API. This
        cannot be undone.
      </Text>
    </div>
  </Modal>
)
