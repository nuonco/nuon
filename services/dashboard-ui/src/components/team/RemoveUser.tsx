'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { removeUser } from '@/actions/orgs/remove-user'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TAccount } from '@/types'

export const RemoveUserModal = ({
  account,
  ...props
}: { account: TAccount } & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { data, error, isLoading, execute } = useServerAction({
    action: removeUser,
  })

  const [confirmEmail, setConfirmEmail] = useState('')

  const isConfirmValid = confirmEmail === account.email
  const canRemove = isConfirmValid && !isLoading

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>
          There was an error while trying to remove {account.email} from{' '}
          {org.name}.
        </Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
      </>
    ),
    errorHeading: `${account.email} was not removed.`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>
        User the {account?.email} was removed from {org?.name}.
      </Text>
    ),
    successHeading: `${account?.email} was removed.`,
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Warning" size="24" />
          {`Remove team member?`}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Removing user
          </span>
        ) : (
          'Remove user'
        ),
        disabled: !canRemove,
        onClick: () => {
          execute({ orgId: org.id, path, body: { user_id: account?.id } })
        },
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
              Are you sure you want to remove {account.email} from your organization?
            </Text>
            <Text variant="body" theme="neutral">
              This action will remove the user and revoke their access immediately.
            </Text>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {account.email}
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

export const RemoveUserButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <RemoveUserModal account={account} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Remove user
      {props?.isMenuButton ? <Icon variant="UserMinus" /> : null}
    </Button>
  )
}
