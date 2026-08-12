import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteRoleModal = ({
  roleTitle,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  roleTitle: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="WarningIcon" size="24" />
        Delete role?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Deleting role
        </span>
      ) : (
        'Delete role'
      ),
      disabled: isPending,
      onClick: () => onSubmit(),
      variant: 'danger',
    }}
    {...props}
  >
    <div className="flex flex-col gap-6">
      {error ? (
        <Banner theme="error">{error?.error || 'Unable to delete role'}</Banner>
      ) : null}

      <Text variant="body" theme="neutral">
        {roleTitle} will be deleted and revoked from everyone holding it,
        including API tokens and service accounts. This cannot be undone.
      </Text>
    </div>
  </Modal>
)
