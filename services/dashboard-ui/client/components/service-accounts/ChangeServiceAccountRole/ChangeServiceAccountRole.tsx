import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const ChangeServiceAccountRoleModal = ({
  accountIdentity,
  currentRole,
  roleOptions,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountIdentity: string
  currentRole: string
  roleOptions: SelectOption[]
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { role: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [role, setRole] = useState(currentRole || roleOptions[0]?.value || '')

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="UserCheckIcon" size="24" />
          Change role
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save'
        ),
        disabled: isPending || role === currentRole,
        onClick: () => onSubmit({ role }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to change role'}
          </Banner>
        ) : null}

        <Text>
          Change the role for <strong>{accountIdentity}</strong>.
        </Text>

        <Select
          value={role}
          onChange={(value) => setRole(value)}
          options={roleOptions}
          labelProps={{ labelText: 'Role' }}
        />
      </div>
    </Modal>
  )
}
