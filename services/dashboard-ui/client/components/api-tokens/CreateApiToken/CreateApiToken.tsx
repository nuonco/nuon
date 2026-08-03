import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DURATION_OPTIONS = [
  { value: '24h', label: '1 day' },
  { value: '168h', label: '1 week' },
  { value: '720h', label: '30 days' },
  { value: '2160h', label: '90 days' },
  { value: '8760h', label: '1 year' },
]

export const ROLE_OPTIONS = [
  { value: 'org_read_only', label: 'Read-only' },
  { value: 'org_builder', label: 'Builder' },
  { value: 'org_admin', label: 'Admin' },
]

export const CreateApiTokenModal = ({
  isPending,
  error,
  createdToken,
  onSubmit,
  onDone,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  createdToken: string | null
  onSubmit: (params: { name: string; duration: string; role: string }) => void
  onDone: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState('')
  const [duration, setDuration] = useState('720h')
  const [role, setRole] = useState('org_read_only')

  if (createdToken) {
    return (
      <Modal
        heading={
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="KeyIcon" size="24" />
            API token created
          </Text>
        }
        primaryActionTrigger={{
          children: 'Done',
          onClick: () => onDone(),
          variant: 'primary',
        }}
        {...props}
      >
        <div className="flex flex-col gap-4">
          <Banner theme="warn">
            Save this token somewhere safe. You won't be able to view it again.
          </Banner>
          <div className="flex items-center gap-2">
            <Text variant="body" family="mono" className="break-all">
              {createdToken}
            </Text>
            <ClickToCopyButton textToCopy={createdToken} />
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="KeyIcon" size="24" />
          Create API token
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating token
          </span>
        ) : (
          'Create token'
        ),
        disabled: !name || isPending,
        onClick: () => onSubmit({ name, duration, role }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create API token'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="token-name">Name</Label>
          <Input
            id="token-name"
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
          options={ROLE_OPTIONS}
          labelProps={{ labelText: 'Role' }}
        />

        <Select
          value={duration}
          onChange={(e) => setDuration(e.target.value)}
          options={DURATION_OPTIONS}
          labelProps={{ labelText: 'Expires after' }}
        />
      </div>
    </Modal>
  )
}
