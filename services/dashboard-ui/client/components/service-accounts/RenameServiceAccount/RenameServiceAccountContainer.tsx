import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateServiceAccount } from '@/lib'
import type { TAccount } from '@/types'
import { RenameServiceAccountModal } from './RenameServiceAccount'

const RenameServiceAccountModalContainer = ({
  account,
  ...props
}: { account: TAccount } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const currentName = account.name || ''
  const identity = account.name || account.email || account.id || ''

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ name }: { name: string }) =>
      updateServiceAccount({
        body: { name },
        accountId: account.id || '',
        orgId: org.id,
      }),
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['service-accounts', org.id] })
      addToast(
        <Toast heading="Service account renamed" theme="success">
          <Text>
            Renamed {identity} to {name}.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <RenameServiceAccountModal
      accountIdentity={identity}
      currentName={currentName}
      isPending={isPending}
      error={error}
      onSubmit={({ name }) => mutate({ name })}
      {...props}
    />
  )
}

export const RenameServiceAccountButton = ({
  account,
  ...props
}: { account: TAccount } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <RenameServiceAccountModalContainer account={account} />

  return (
    <Button
      variant="ghost"
      className="!p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Rename
      {props?.isMenuButton ? <Icon variant="PencilSimpleIcon" /> : null}
    </Button>
  )
}
