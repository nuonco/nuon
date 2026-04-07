import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { RadioInput } from '@/components/common/form/RadioInput'
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

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="UserPlus" size="24" />
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
            <Icon variant="UserPlus" />
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
        {hasSupportRole ? (
          <div className="flex flex-col gap-2">
            <Label>Role</Label>
            <RadioInput
              name="role_type"
              value="org_admin"
              checked={roleType === 'org_admin'}
              onChange={() => setRoleType('org_admin')}
              labelProps={{ labelText: 'Admin' }}
            />
            <RadioInput
              name="role_type"
              value="org_support"
              checked={roleType === 'org_support'}
              onChange={() => setRoleType('org_support')}
              labelProps={{ labelText: 'Support' }}
            />
          </div>
        ) : null}
      </div>
    </Modal>
  )
}
