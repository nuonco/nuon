import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const ChangeRoleModal = ({
  accountEmail,
  currentRole,
  currentRoleLabel,
  isPending,
  error,
  roleOptions: fetchedRoleOptions,
  onSubmit,
  ...props
}: {
  accountEmail: string
  currentRole: string
  currentRoleLabel?: string
  isPending: boolean
  error: TAPIError | null
  roleOptions: { value: string; label: string; description?: string }[]
  onSubmit: (params: { roleType: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [roleType, setRoleType] = useState(currentRole || 'org_read_only')

  const roleOptions =
    !currentRole ||
    fetchedRoleOptions.some((option) => option.value === currentRole)
      ? fetchedRoleOptions
      : [
          ...fetchedRoleOptions,
          { value: currentRole, label: currentRoleLabel || currentRole },
        ]

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
        disabled: isPending || roleType === currentRole,
        onClick: () => onSubmit({ roleType }),
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
          Change the role for <strong>{accountEmail}</strong>.
        </Text>

        <Select
          value={roleType}
          onChange={(value) => setRoleType(value)}
          options={roleOptions}
          labelProps={{ labelText: 'Role' }}
        />
      </div>
    </Modal>
  )
}
