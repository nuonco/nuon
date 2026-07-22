import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Select, type SelectOption } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const CreateServiceAccountModal = ({
  roleOptions,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  roleOptions: SelectOption[]
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { role: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [role, setRole] = useState(roleOptions[0]?.value ?? '')

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="RobotIcon" size="24" />
          Create service account
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating service account
          </span>
        ) : (
          'Create service account'
        ),
        disabled: !role || isPending,
        onClick: () => onSubmit({ role }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create service account'}
          </Banner>
        ) : null}

        <Text>
          Service accounts are non-human identities for automating access to the Nuon API.
        </Text>

        <Select
          value={role}
          onChange={(e) => setRole(e.target.value)}
          options={roleOptions}
          labelProps={{ labelText: 'Role' }}
        />
      </div>
    </Modal>
  )
}
