import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateAccountRole } from '@/lib'
import type { TAccount } from '@/types'
import { ChangeRoleModal } from './ChangeRole'

const ChangeRoleModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const currentRole = account.roles?.[0]?.role_type || ''

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ roleType }: { roleType: string }) =>
      updateAccountRole({
        body: { role_type: roleType },
        accountId: account.id || '',
        orgId: org.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-accounts', org?.id] })
      addToast(
        <Toast heading="Role updated" theme="success">
          <Text>Updated the role for {account.email}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Update failed" theme="error">
          <Text>There was an error changing the role for {account.email}.</Text>
        </Toast>
      )
    },
  })

  return (
    <ChangeRoleModal
      accountEmail={account.email || ''}
      currentRole={currentRole}
      isPending={isPending}
      error={error}
      onSubmit={({ roleType }) => mutate({ roleType })}
      {...props}
    />
  )
}

export const ChangeRoleButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <ChangeRoleModalContainer account={account} />

  return (
    <Button
      variant="ghost"
      className="!p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Change role
      {props?.isMenuButton ? <Icon variant="UserCheckIcon" /> : null}
    </Button>
  )
}
