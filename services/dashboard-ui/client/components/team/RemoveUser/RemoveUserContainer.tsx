import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { removeUser } from '@/lib'
import type { TAccount } from '@/types'
import { RemoveUserModal } from './RemoveUser'

const RemoveUserModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
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
    <RemoveUserModal
      accountEmail={account.email || ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const RemoveUserButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <RemoveUserModalContainer account={account} />

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
