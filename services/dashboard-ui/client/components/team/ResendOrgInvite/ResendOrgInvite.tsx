import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const ResendOrgInviteModal = ({
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
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="EnvelopeIcon" size="24" />
          Resend invite
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Resending...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="EnvelopeIcon" />
            Resend invite
          </span>
        ),
        disabled: isPending,
        onClick: () => onSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to resend invite'}
          </Banner>
        ) : null}
        <Text variant="base">
          Resend the invitation email to <strong>{email}</strong>?
        </Text>
      </div>
    </Modal>
  )
}
