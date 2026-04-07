import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const RemoveUserModal = ({
  accountEmail,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountEmail: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => {
  const [confirmEmail, setConfirmEmail] = useState('')

  const isConfirmValid = confirmEmail === accountEmail

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="Warning" size="24" />
          Remove team member?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Removing user
          </span>
        ) : (
          'Remove user'
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
            {error?.error || 'Unable to remove user from organization'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to remove {accountEmail} from your organization?
            </Text>
            <Text variant="body" theme="neutral">
              This action will remove the user and revoke their access immediately.
            </Text>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {accountEmail}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-user-email"
              placeholder="user email"
              type="text"
              value={confirmEmail}
              onChange={(e) => setConfirmEmail(e.target.value)}
              error={confirmEmail.length > 0 && !isConfirmValid}
              errorMessage={confirmEmail.length > 0 && !isConfirmValid ? "Email doesn't match" : undefined}
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}
