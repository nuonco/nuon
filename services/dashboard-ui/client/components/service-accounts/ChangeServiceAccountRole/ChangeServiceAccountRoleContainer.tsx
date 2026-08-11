import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useRoleOptions } from '@/hooks/use-roles'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateServiceAccountRole } from '@/lib'
import type { TAccount } from '@/types'
import { ChangeServiceAccountRoleModal } from './ChangeServiceAccountRole'

const ChangeServiceAccountRoleModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const currentRole = account.roles?.[0]?.role_type || ''
  const identity = account.email || account.id || ''

  const { roleOptions } = useRoleOptions('service_account')

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ role }: { role: string }) =>
      updateServiceAccountRole({
        body: { role },
        accountId: account.id || '',
        orgId: org.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['service-accounts', org.id] })
      addToast(
        <Toast heading="Role updated" theme="success">
          <Text>Updated the role for {identity}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <ChangeServiceAccountRoleModal
      accountIdentity={identity}
      currentRole={currentRole}
      roleOptions={roleOptions}
      isPending={isPending}
      error={error}
      onSubmit={({ role }) => mutate({ role })}
      {...props}
    />
  )
}

export const ChangeServiceAccountRoleButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <ChangeServiceAccountRoleModalContainer account={account} />

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
