import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteServiceAccount } from '@/lib'
import type { TAccount } from '@/types'
import { DeleteServiceAccountModal } from './DeleteServiceAccount'

const DeleteServiceAccountModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const identity = account.email || account.id || ''

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => deleteServiceAccount({ accountId: account.id || '', orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['service-accounts', org.id] })
      addToast(
        <Toast heading="Service account deleted" theme="success">
          <Text>Deleted {identity} from {org.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Delete failed" theme="error">
          <Text>There was an error deleting {identity}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteServiceAccountModal
      accountIdentity={identity}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteServiceAccountButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteServiceAccountModalContainer account={account} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Delete service account
      {props?.isMenuButton ? <Icon variant="TrashIcon" /> : null}
    </Button>
  )
}
