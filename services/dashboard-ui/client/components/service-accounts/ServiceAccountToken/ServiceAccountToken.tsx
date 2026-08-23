import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
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

export const CreateServiceAccountTokenModal = ({
  accountIdentity,
  defaultDuration = '8760h',
  isPending,
  error,
  createdToken,
  onSubmit,
  onDone,
  ...props
}: {
  accountIdentity: string
  defaultDuration?: string
  isPending: boolean
  error: TAPIError | null
  createdToken: string | null
  onSubmit: (params: { duration: string; invalidate: boolean }) => void
  onDone: () => void
} & Omit<IModal, 'onSubmit'>) => {
  const [duration, setDuration] = useState(defaultDuration)
  const [invalidate, setInvalidate] = useState(false)

  if (createdToken) {
    return (
      <Modal
        heading={
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="KeyIcon" size="24" />
            Token created
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
          Create token
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
        disabled: isPending,
        onClick: () => onSubmit({ duration, invalidate }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create token'}
          </Banner>
        ) : null}

        <Text>
          Create a token for <strong>{accountIdentity}</strong>.
        </Text>

        <Select
          value={duration}
          onChange={(value) => setDuration(value)}
          options={DURATION_OPTIONS}
          labelProps={{ labelText: 'Expires after' }}
        />

        <CheckboxInput
          checked={invalidate}
          onChange={(e) => setInvalidate(e.target.checked)}
          labelProps={{
            labelText: 'Invalidate existing tokens for this service account',
          }}
        />
      </div>
    </Modal>
  )
}
