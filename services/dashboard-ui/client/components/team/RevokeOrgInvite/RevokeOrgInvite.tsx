import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const RevokeOrgInviteModal = ({
  email,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  email: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="ProhibitIcon" size="24" />
          Revoke invite?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Revoking...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ProhibitIcon" />
            Revoke invite
          </span>
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
            {error?.error || 'Unable to revoke invite'}
          </Banner>
        ) : null}
        <Text variant="base">
          Revoking the invitation for <strong>{email}</strong> will prevent them
          from accepting it. You can send a new invite later.
        </Text>
      </div>
    </Modal>
  )
}
