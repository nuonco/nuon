import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteServiceAccountModal = ({
  accountIdentity,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountIdentity: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => {
  const [confirmValue, setConfirmValue] = useState('')

  const isConfirmValid = confirmValue === accountIdentity

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="WarningIcon" size="24" />
          Delete service account?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Deleting service account
          </span>
        ) : (
          'Delete service account'
        ),
        disabled: !isConfirmValid || isPending,
        onClick: () => onSubmit(),
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to delete service account'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="body" theme="neutral">
              {accountIdentity} will be deleted and any tokens issued to it will stop
              working immediately.
            </Text>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {accountIdentity}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-service-account-identity"
              placeholder="service account identity"
              type="text"
              value={confirmValue}
              onChange={(e) => setConfirmValue(e.target.value)}
              error={confirmValue.length > 0 && !isConfirmValid}
              errorMessage={
                confirmValue.length > 0 && !isConfirmValid ? "Doesn't match" : undefined
              }
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}
