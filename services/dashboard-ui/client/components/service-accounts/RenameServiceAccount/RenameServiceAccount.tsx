import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const RenameServiceAccountModal = ({
  accountIdentity,
  currentName,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountIdentity: string
  currentName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { name: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState(currentName)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="PencilSimpleIcon" size="24" />
          Rename service account
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
        disabled: !name || isPending || name === currentName,
        onClick: () => onSubmit({ name }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to rename service account'}
          </Banner>
        ) : null}

        <Text>
          Rename <strong>{accountIdentity}</strong>.
        </Text>

        <div className="flex flex-col gap-2">
          <Label htmlFor="service-account-rename">Name</Label>
          <Input
            id="service-account-rename"
            placeholder="e.g. ci-deploy"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
      </div>
    </Modal>
  )
}
