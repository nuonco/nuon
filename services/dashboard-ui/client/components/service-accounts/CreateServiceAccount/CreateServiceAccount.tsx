import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
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
  onSubmit: (params: { name: string; role: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState('')
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
        disabled: !name || !role || isPending,
        onClick: () => onSubmit({ name, role }),
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

        <div className="flex flex-col gap-2">
          <Label htmlFor="service-account-name">Name</Label>
          <Input
            id="service-account-name"
            placeholder="e.g. ci-deploy"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>

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
