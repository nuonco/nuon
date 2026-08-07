import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TOrgInvite } from '@/types'

export const InviteUserModal = ({
  isPending,
  isResendPending,
  error,
  invites,
  roleOptions,
  onSubmit,
  onResend,
  ...props
}: {
  isPending: boolean
  isResendPending?: boolean
  error: TAPIError | null
  invites?: TOrgInvite[]
  roleOptions: { value: string; label: string }[]
  onSubmit: (params: { email: string; roleType: string }) => void
  onResend?: (inviteId: string) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [email, setEmail] = useState('')
  const [roleType, setRoleType] = useState('org_admin')

  const normalizedEmail = email.trim().toLowerCase()
  const matchedInvite = normalizedEmail
    ? invites?.find(
        (i) =>
          i?.status !== 'accepted' &&
          i?.email?.trim().toLowerCase() === normalizedEmail
      )
    : undefined

  const primaryActionTrigger = matchedInvite
    ? {
        children: isResendPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Resending...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="EnvelopeIcon" />
            Resend invite
          </span>
        ),
        disabled: isResendPending,
        onClick: () =>
          matchedInvite.id && onResend?.(matchedInvite.id),
        variant: 'primary' as const,
      }
    : {
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Inviting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="UserPlusIcon" />
            Invite user
          </span>
        ),
        disabled: !email || isPending,
        onClick: () => onSubmit({ email, roleType }),
        variant: 'primary' as const,
      }

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="UserPlusIcon" size="24" />
          Invite team member
        </Text>
      }
      primaryActionTrigger={primaryActionTrigger}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to invite user to organization'}
          </Banner>
        ) : null}
        {matchedInvite ? (
          <Banner theme="info">
            {email} already has a pending invite. Resend it to send another
            email.
          </Banner>
        ) : null}
        <div className="flex flex-col gap-2">
          <Label htmlFor="invite-email">
            Email address of the user you want to invite
          </Label>
          <Input
            id="invite-email"
            placeholder="user@email.com"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        {matchedInvite ? null : (
          <Select
            value={roleType}
            onChange={(value) => setRoleType(value)}
            options={roleOptions}
            labelProps={{ labelText: 'Role' }}
          />
        )}
      </div>
    </Modal>
  )
}
