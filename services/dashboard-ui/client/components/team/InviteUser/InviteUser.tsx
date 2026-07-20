import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const InviteUserModal = ({
  hasSupportRole,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  hasSupportRole: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { email: string; roleType: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [email, setEmail] = useState('')
  const [roleType, setRoleType] = useState('org_admin')

  const roleOptions = [
    { value: 'org_admin', label: 'Admin' },
    ...(hasSupportRole ? [{ value: 'org_support', label: 'Support' }] : []),
    { value: 'org_read_only', label: 'Read-only' },
  ]

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="UserPlusIcon" size="24" />
          Invite team member
        </Text>
      }
      primaryActionTrigger={{
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
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to invite user to organization'}
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
        <Select
          value={roleType}
          onChange={(e) => setRoleType(e.target.value)}
          options={roleOptions}
          labelProps={{ labelText: 'Role' }}
        />
      </div>
    </Modal>
  )
}
