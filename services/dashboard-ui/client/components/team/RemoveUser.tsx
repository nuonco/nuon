import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { removeUser } from '@/lib'
import type { TAccount } from '@/types'

export const RemoveUserModal = ({
  account,
  ...props
}: { account: TAccount } & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [confirmEmail, setConfirmEmail] = useState('')

  const isConfirmValid = confirmEmail === account.email

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () => removeUser({ body: { user_id: account.id }, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading={`${account.email} was removed.`} theme="success">
          <Text>User {account.email} was removed from {org.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading={`${account.email} was not removed.`} theme="error">
          <Text>There was an error removing {account.email} from {org.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="Warning" size="24" />
          Remove team member?
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
        disabled: !isConfirmValid || isLoading,
        onClick: () => mutate(),
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
      onClick={() => addModal(modal)}
      {...props}
    >
      Remove user
      {props?.isMenuButton ? <Icon variant="UserMinus" /> : null}
    </Button>
  )
}
